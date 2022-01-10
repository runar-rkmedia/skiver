//go:generate echo "Generating spec..."
//go:generate swagger generate spec -i base-swagger.yml -x models -o swagger.yml --scan-models
//go:generate echo "Generating model..."
//go:generate swagger generate model -f swagger.yml
//go:generate echo "Validating spec..."
//go:generate swagger validate swagger.yml
//go:generate echo "Generating frontend-types"
//go:generate sh -c "cd frontend && yarn gen"
//go:generate echo "done"
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "embed"
	"net/http/pprof"

	"github.com/NYTimes/gziphandler"
	swaggerMiddleware "github.com/go-openapi/runtime/middleware"
	"github.com/runar-rkmedia/gabyoall/api/utils"
	"github.com/runar-rkmedia/gabyoall/logger"
	"github.com/runar-rkmedia/skiver/bboltStorage"
	cfg "github.com/runar-rkmedia/skiver/config"
	"github.com/runar-rkmedia/skiver/frontend"
	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/localuser"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/translator"
	"github.com/runar-rkmedia/skiver/types"
)

// TODO: update to use debug.BuildInfo or see https://github.com/carlmjohnson/versioninfo/

var (
	//go:embed swagger.yml
	swaggerYml string
	// These are added at build...
	Version      string
	BuildDateStr string
	BuildDate    time.Time
	GitHash      string
	isDev        = true
	IsDevStr     = "1"

	serverStartTime = time.Now()
)

func init() {
	if BuildDateStr != "" {
		t, err := time.Parse("2006-01-02T15:04:05", BuildDateStr)
		if err != nil {
			panic(fmt.Errorf("Failed to parse build-date: %w", err))
		}
		BuildDate = t
	}
	if IsDevStr != "1" {
		isDev = false
	}
}

var (
	pwsalt            = "devsalt-123-123-123-123"
	maxBodySize int64 = 1_000_000 // 1MB
)

func getDefaultDBLocation() string {
	if s, err := os.Stat("./skiver.bbolt"); err == nil && !s.IsDir() {
		return "./skiver.bbolt"
		// When running in a
	} else if s, err := os.Stat("./storage"); err == nil && s.IsDir() {
		return "./storage/skiver.bbolt"
	}
	return "./skiver.bbolt"
}

// After considerations, this naming is really bad. it is not actually publishing, but simply subscribing to changes.
type PubSubPublisher interface {
	Publish(kind, variant string, contents interface{})
}

type MultiPublisher struct {
	publishers map[string]PubSubPublisher
}

func NewMultiPublisher() MultiPublisher {
	return MultiPublisher{map[string]PubSubPublisher{}}
}
func (m *MultiPublisher) Publish(kind, variant string, contents interface{}) {
	for _, v := range m.publishers {
		go v.Publish(kind, variant, contents)
	}
}
func (m *MultiPublisher) AddSubscriber(name string, publisher PubSubPublisher) error {
	m.publishers[name] = publisher
	return nil
}

type logPublisher struct {
	l logger.AppLogger
}

func (m *logPublisher) Publish(kind, variant string, contents interface{}) {
	m.l.Debug().Str("kind", kind).Str("variant", variant).Interface("contents", contents).Msg("Event received")
}

type Translator interface {
	Translate(text, from, to string) (string, error)
}
type translationHook struct {
	translator Translator
	l          logger.AppLogger
	db         types.Storage
}

func (m *translationHook) Publish(kind, variant string, contents interface{}) {

	debug := m.l.HasDebug()
	if kind == string(bboltStorage.PubTypeTranslationValue) && variant == string(bboltStorage.PubVerbCreate) {
		tv, ok := contents.(types.TranslationValue)
		if !ok {
			m.l.Error().Interface("content", contents).Msg("Failed to convert contents to TranslationValue")
			return
		}
		if tv.Source == types.CreatorSourceTranslator {
			if debug {
				m.l.Debug().Interface("content", contents).Msg("ignoring TranslationValue since it was sourced from me")
			}
			return
		}
		if strings.TrimSpace(tv.Value) == "" {
			m.l.Error().Interface("content", contents).Msg("Received TranslationValue, but the value appeared to be empty")
			return
		}
		locales, err := m.db.GetLocales()
		if err != nil {
			m.l.Error().Err(err).Msg("failed to lookup locales")
			return
		}
		tvs, err := m.db.GetTranslationValuesFilter(0, types.TranslationValue{TranslationID: tv.TranslationID})
		if err != nil {
			m.l.Error().Err(err).Msg("failed to lookup translationvalues")
			return
		}
		existingTranslations := map[string]string{}
		for k, v := range tvs {
			existingTranslations[v.LocaleID] = k
		}
		sourceLocale := locales[tv.LocaleID]
		for _, l := range locales {
			if l.ID == sourceLocale.ID {
				if debug {
					m.l.Debug().Interface("locale", l).Msg("Skipped locale, since it is the source-locale")
				}
				continue
			}
			if _, ok := existingTranslations[l.ID]; ok {
				if debug {
					m.l.Debug().Interface("locale", l).Msg("Skipped locale, since it is already translated")
				}
				continue
			}
			if sourceLocale.Iso639_1 == l.Iso639_1 {
				if debug {
					m.l.Debug().Interface("locale", l).Msg("skipping translation, since Iso639_1 is the same as the source")
				}
				continue

			}
			// TODO: check if source and target are supported.
			source := sourceLocale.Iso639_1
			target := l.Iso639_1
			result, err := m.translator.Translate(tv.Value, source, target)
			if err != nil {
				m.l.Error().Err(err).Str("source", source).Str("target", target).Msg("failed during translation")
				continue
			}
			if result == "" {
				m.l.Warn().Str("result", result).Msg("The translation returned a empty result")
				continue
			}
			tv := types.TranslationValue{
				Value:         result,
				LocaleID:      l.ID,
				TranslationID: tv.TranslationID,
				Source:        types.CreatorSourceTranslator,
			}
			tv.CreatedBy = string(tv.Source)
			_, err = m.db.CreateTranslationValue(tv)

			if err != nil {
				m.l.Error().Err(err).Msg("Failed to create translation-value")
				continue
			}

		}

	}
}

func main() {
	err := cfg.InitConfig()
	if err != nil {
		panic(err)
	}
	config := cfg.GetConfig()
	cfg := config.Api
	if cfg.Address == "" {
		cfg.Address = "0.0.0.0"
	}
	if cfg.Port == 0 {
		cfg.Port = 80
	}
	if config.LogFormat == "" {
		config.LogFormat = "json"
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if cfg.DBLocation == "" {
		cfg.DBLocation = getDefaultDBLocation()
	}
	logger.InitLogger(logger.LogConfig{
		Level:  config.LogLevel,
		Format: config.LogFormat,
		// We add this option during local development, but also if loglevel is debug
		WithCaller: config.LogLevel == "debug" || GitHash == "",
	})
	l := logger.GetLogger("main")
	events := NewMultiPublisher()
	l.Info().
		Str("version", Version).
		Time("buildDate", BuildDate).
		Time("buildDateLocal", BuildDate.Local()).
		Str("gitHash", GitHash).
		Str("db", cfg.DBLocation).
		Int("pid", os.Getpid()).
		Msg("Starting")
	// IMPORTANT: database publishes changes, but for performance-reasons, it should not be used until the listener (ws) is started.
	db, err := bboltStorage.NewBbolt(l, cfg.DBLocation, &events)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	if len(config.TranslatorServices) > 1 {
		l.Fatal().Msg("currently, only a single translator-service can be used.")
	}
	if l.HasDebug() {
		events.AddSubscriber("log", &logPublisher{logger.GetLogger("events")})
	}
	if len(config.TranslatorServices) > 0 {
		o := config.TranslatorServices[0]
		t, err := translator.NewTranslator(translator.TranslatorOptions{
			Kind: o.Kind,

			ApiToken: o.ApiToken,
			Endpoint: o.Endpoint,
		})
		if err != nil {
			l.Fatal().Err(err).Msg("failed to set up translator-services")
		}
		hook := translationHook{
			translator: t,
			l:          logger.GetLogger("translation-hook"),
			db:         &db,
		}
		events.AddSubscriber("translation-hook", &hook)
	}
	pubsub := handlers.NewPubSubChannel()
	events.AddSubscriber("msg", &pubsub)

	pw := localuser.NewPwHasher([]byte(pwsalt))

	ctx := requestContext.Context{
		L:  l,
		DB: &db,
		StructValidater: func(m interface{}) error {
			v, ok := m.(models.Validator)
			if !ok {
				l.Fatal().Interface("type", m).Msg("does not implement Validator")
			}
			return models.Validate(v)
		},
	}
	if config.SelfCheck {
		go utils.SelfCheck(utils.SelfCheckLimit{
			MemoryMB:   1000,
			GoRoutines: 10000,
			Streaks:    5,
			Interval:   time.Second * 15,
		}, logger.GetLogger("self-check"), 0)
	}

	address := net.JoinHostPort(cfg.Address, strconv.Itoa(cfg.Port))
	handler := http.NewServeMux()
	handler.Handle("/docs", swaggerMiddleware.SwaggerUI(swaggerMiddleware.SwaggerUIOpts{
		BasePath:         "/",
		Path:             "",
		SpecURL:          "/api/swagger.yml",
		SwaggerURL:       "",
		SwaggerPresetURL: "",
		SwaggerStylesURL: "",
		Favicon32:        "",
		Favicon16:        "",
		Title:            "Skiver",
	}, handler))
	handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	handler.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	handler.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	handler.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	handler.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	// TODO: consider using a buffered channel.
	handler.Handle("/ws/", handlers.NewWsHandler(logger.GetLoggerWithLevel("ws", "debug"), pubsub.Ch, handlers.WsOptions{}))

	err = types.SeedUsers(&db, nil, pw.Hash)
	if err != nil {
		panic(err)
	}
	err = types.SeedLocales(&db, nil)
	if err != nil {
		panic(err)
	}
	handler.Handle("/api/ping", handlers.PingHandler(handler))

	info := types.ServerInfo{
		ServerStartedAt: serverStartTime,
		GitHash:         GitHash,
		Version:         Version,
		BuildDate:       BuildDate,
	}

	handler.Handle("/api/",
		// requires go.1.18
		// http.MaxBytesHandler(
		gziphandler.GzipHandler(
			http.StripPrefix("/api/",
				handlers.EndpointsHandler(
					ctx, pw, info, []byte(swaggerYml),
				),
			),
			// ),
			// maxBodySize,
		))

	useCert := false
	if cfg.CertFile != "" {
		_, err := os.Stat(cfg.CertFile)
		if err == nil {
			useCert = true
		}

	}

	if isDev {
		// In development, we serve the file directly.
		handler.Handle("/", http.FileServer(http.Dir("./frontend/dist/")))
	} else {
		handler.Handle("/", frontend.DistServer)
	}
	l.Info().Str("address", cfg.Address).Int("port", cfg.Port).Bool("redirectHttpToHttps", useCert && cfg.RedirectPort != 0).Bool("tls", useCert).Msg("Creating listener")
	srv := http.Server{Addr: address, Handler: handler}
	if useCert {
		// TODO: re-read the certificate before it expires.
		if cfg.RedirectPort != 0 {
			redirectTLS := func(w http.ResponseWriter, r *http.Request) {
				newAddress := "https://" + r.Host
				if cfg.Port != 443 {
					newAddress += fmt.Sprintf(":%d", cfg.Port)
				}
				http.Redirect(w, r, newAddress+r.RequestURI, http.StatusMovedPermanently)
			}
			go func() {
				redirectAddress := fmt.Sprintf("%s:%d", cfg.Address, cfg.RedirectPort)
				if err := http.ListenAndServe(redirectAddress, http.HandlerFunc(redirectTLS)); err != nil {
					l.Fatal().Err(err).Str("redirectAddress", redirectAddress).Msg("Failed to create redirect-listener")

				}
			}()

		}
		err = srv.ListenAndServeTLS("server.crt", "server.key")
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to create listener")
	}

}
