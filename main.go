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
	"context"
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "embed"
	"net/http/pprof"

	"github.com/NYTimes/gziphandler"
	swaggerMiddleware "github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/bboltStorage"
	cfg "github.com/runar-rkmedia/skiver/config"
	"github.com/runar-rkmedia/skiver/frontend"
	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/localuser"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/translator"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
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
	publishers   map[string]PubSubPublisher
	publishFuncs map[string]func(kind, variant string, contents interface{})
}

func NewMultiPublisher() MultiPublisher {
	return MultiPublisher{map[string]PubSubPublisher{}, map[string]func(kind string, variant string, contents interface{}){}}
}
func (m *MultiPublisher) Publish(kind, variant string, contents interface{}) {
	for _, v := range m.publishers {
		go v.Publish(kind, variant, contents)
	}
	for _, v := range m.publishFuncs {
		go v(kind, variant, contents)
	}
}
func (m *MultiPublisher) AddSubscriber(name string, publisher PubSubPublisher) error {
	m.publishers[name] = publisher
	return nil
}
func (m *MultiPublisher) AddSubscriberFunc(name string, publisher func(kind, variant string, contents interface{})) error {
	m.publishFuncs[name] = publisher
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
	// TODO: this function really should cache

	debug := m.l.HasDebug()
	if kind == string(bboltStorage.PubTypeTranslationValue) && variant == string(bboltStorage.PubVerbCreate) {
		tv, ok := contents.(types.TranslationValue)
		if !ok {
			m.l.Error().Interface("content", contents).Msg("Failed to convert contents to TranslationValue")
			return
		}
		orgId := tv.OrganizationID
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

		// We need the project-settings, so we resolve the project
		var project *types.Project
		{
			t, err := m.db.GetTranslation(tv.TranslationID)
			if err != nil {
				m.l.Error().Err(err).Msg("failed to lookup translation")
				return
			}
			if t == nil {
				m.l.Error().Err(err).Msg("Missing translation")
				return
			}
			cat, err := m.db.GetCategory(t.CategoryID)
			if err != nil {
				m.l.Error().Err(err).Msg("failed to lookup category")
				return
			}
			if t == nil {
				m.l.Error().Err(err).Msg("Missing category")
				return
			}
			p, err := m.db.GetProject(cat.ProjectID)
			if err != nil {
				m.l.Error().Err(err).Msg("failed to lookup project")
				return
			}
			if t == nil {
				m.l.Error().Err(err).Msg("Missing project")
				return
			}
			project = p

		}
		if len(project.LocaleIDs) == 0 {
			if m.l.HasDebug() {
				m.l.Debug().Interface("project", project).Msg("project does not have any locale-ids")
			}
			return
		}

		_locales, err := m.db.GetLocales()
		if err != nil {
			m.l.Error().Err(err).Msg("failed to lookup locales")
			return
		}
		locales := map[string]types.Locale{}
		for k, v := range project.LocaleIDs {
			if !v.AutoTranslation {
				continue
			}
			if l, ok := _locales[k]; ok {
				locales[k] = l
			}

		}

		runtime.Breakpoint()

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
			tv.OrganizationID = orgId
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

	if config.Authentication.SessionLifeTime == 0 {
		config.Authentication.SessionLifeTime = time.Hour
	}
	if config.Authentication.SessionLifeTime < time.Minute {
		l.Fatal().Str("Authentication.SessionLifeTime", config.Authentication.SessionLifeTime.String()).Msg("SessionLifeTime cannot be shorter than a minute. That would just be really annoying.")
	}
	events := NewMultiPublisher()
	l.Info().
		Str("version", Version).
		Time("buildDate", BuildDate).
		Time("buildDateLocal", BuildDate.Local()).
		Str("gitHash", GitHash).
		Str("db", cfg.DBLocation).
		Int("pid", os.Getpid()).
		Msg("Starting")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	// IMPORTANT: database publishes changes, but for performance-reasons, it should not be used until the listener (ws) is started.
	db, err := bboltStorage.NewBbolt(l, cfg.DBLocation, &events)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	defer db.DB.Close()
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
		// defer func() { quitSelfservice <- struct{}{} }()
		go func() {

			tick := time.Tick(time.Second * 15)

			for {
				select {
				case <-tick:
					utils.SelfCheck(utils.SelfCheckLimit{
						MemoryMB:   1000,
						GoRoutines: 10000,
						Streaks:    5,
					}, logger.GetLogger("self-check"), 0)
				}
			}
		}()
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
	if cfg.Debug {
		handler.Handle("/debug/vars/", expvar.Handler())
		handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		handler.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		handler.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		handler.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		handler.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	}
	// TODO: consider using a buffered channel.
	handler.Handle("/ws/", handlers.NewWsHandler(logger.GetLoggerWithLevel("ws", "debug"), pubsub.Ch, handlers.WsOptions{}))
	exportCache := cache.New(time.Hour, time.Hour)
	events.AddSubscriberFunc("exportCache", func(kind, variant string, contents interface{}) {
		//  TODO: delete only thoose belonging to a project, etc.
		exportCache.Flush()
	})

	go func() {

		org, err := types.SeedUsers(&db, nil, pw.Hash)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to seed users")
		}
		if org != nil {
			err = types.SeedLocales(&db, org.ID, nil)
			if err != nil {
				l.Fatal().Err(err).Msg("Failed to seed Locale")
			}
		}
	}()
	handler.Handle("/api/ping", handlers.PingHandler(handler))

	info := types.ServerInfo{
		ServerStartedAt: serverStartTime,
		GitHash:         GitHash,
		Version:         Version,
		BuildDate:       BuildDate,
	}
	p, ok := ctx.DB.(localuser.Persistor)
	if !ok {
		ctx.L.Warn().Str("type", fmt.Sprintf("%T", ctx.DB)).Msg("DB does not implement the localUser.Persistor-interface")
	}
	userSessions, err := localuser.NewUserSessionInMemory(localuser.UserSessionOptions{TTL: config.Authentication.SessionLifeTime}, uuid.NewString, p)
	if err != nil {
		ctx.L.Fatal().Err(err).Msg("Failed to set up userSessions")
	}

	apiHandler := http.StripPrefix("/api/",
		handlers.EndpointsHandler(
			ctx, userSessions, pw, info, []byte(swaggerYml), exportCache,
		),
	)
	if config.Gzip {
		apiHandler = gziphandler.GzipHandler(apiHandler)
	}
	handler.Handle("/api/",
		// requires go.1.18
		// ioutil.MaxBytesHandler(
		apiHandler,

		// maxBodySize,
	// 	)
	)
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
		w, err := utils.NewDirWatcher("./frontend/dist")
		if err != nil {
			panic(err)
		}
		go func() {
			for {
				select {
				case _, ok := <-w.Events:
					if !ok {
						fmt.Println("not ok")
						continue
					}
					events.Publish("dist", "change", nil)
				}

			}
		}()
		if err != nil {
			panic(err)
		}
	} else {
		handler.Handle("/", frontend.DistServer)
	}
	l.Info().Str("address", cfg.Address).Int("port", cfg.Port).Bool("redirectHttpToHttps", useCert && cfg.RedirectPort != 0).Bool("tls", useCert).Msg("Creating listener")

	serverErrors := make(chan error, 1)
	srv := http.Server{
		Addr: address, Handler: handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     Logger(logger.GetLogger("http-server")),
	}
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
				serverErrors <- http.ListenAndServe(redirectAddress, http.HandlerFunc(redirectTLS))
			}()

		}
		go func() {
			serverErrors <- srv.ListenAndServeTLS(config.Api.CertFile, config.Api.CertKey)
		}()
	} else {
		go func() {
			serverErrors <- srv.ListenAndServe()
		}()
	}
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to create listener")
	}
	select {
	case err := <-serverErrors:
		l.Error().Err(err).Msg("A server-error occured")
		return
	case sig := <-shutdown:
		events.Publish("system", "shutdown", sig)

		// Any outstanding requests gets some time to complete
		l.Error().Interface("signal", sig).Msg("Received signal, starting shutdown")

		defer l.Info().Interface("signal", sig).Msg("Shutdown complete")

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			l.Error().Err(err).Msg("Failed to stop server gracefully.")
			return
		}
		return
	}

}

// Tiny wrapper for use with standard lolgger
// TODO: move to upstream logger-lib
type NewLog struct {
	logger *logger.AppLogger
}

func (l *NewLog) Write(p []byte) (n int, err error) {
	l.logger.Error().Msg(string(p))
	return len(p), nil
}

func Logger(l logger.AppLogger) *log.Logger {
	lg := NewLog{&l}
	return log.New(&lg, "", 0)
}
