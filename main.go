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
	l.Info().
		Str("version", Version).
		Time("buildDate", BuildDate).
		Time("buildDateLocal", BuildDate.Local()).
		Str("gitHash", GitHash).
		Str("db", cfg.DBLocation).
		Msg("Starting")
	if len(config.TranslatorServices) > 1 {
		l.Fatal().Msg("currently, only a single translator-service can be used.")
	}
	if len(config.TranslatorServices) > 0 {
		o := config.TranslatorServices[0]
		_, err := translator.NewTranslator(translator.TranslatorOptions{
			Kind:     o.Kind,
			ApiToken: o.ApiToken,
			Endpoint: o.Endpoint,
		})
		if err != nil {
			l.Fatal().Err(err).Msg("failed to set up translator-services")
		}
	}
	pubsub := handlers.NewPubSubChannel()
	// IMPORTANT: database publishes changes, but for performance-reasons, it should not be used until the listener (ws) is started.
	db, err := bboltStorage.NewBbolt(l, cfg.DBLocation, &pubsub)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to initialize storage")
	}
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

	types.SeedUsers(&db, nil, pw.Hash)
	types.SeedLocales(&db, nil)
	handler.Handle("/api/ping", handlers.PingHandler(handler))

	info := types.ServerInfo{
		ServerStartedAt: serverStartTime,
		GitHash:         GitHash,
		Version:         Version,
		BuildDate:       BuildDate,
	}

	handler.Handle("/api/",
		http.MaxBytesHandler(
			gziphandler.GzipHandler(
				http.StripPrefix("/api/",
					handlers.EndpointsHandler(
						ctx, pw, info, []byte(swaggerYml),
					),
				),
			),
			maxBodySize))

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
