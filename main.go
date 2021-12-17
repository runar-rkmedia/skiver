package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	_ "embed"
	"net/http/pprof"

	"github.com/NYTimes/gziphandler"
	"github.com/dustin/go-humanize"
	"github.com/qri-io/jsonschema"
	"github.com/runar-rkmedia/gabyoall/api/utils"
	"github.com/runar-rkmedia/gabyoall/logger"
	"github.com/runar-rkmedia/skiver/bboltStorage"
	cfg "github.com/runar-rkmedia/skiver/config"
	_ "github.com/runar-rkmedia/skiver/docs"
	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

var (
	//go:embed docs/generated-swagger.yml
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

type ApiConfig struct {
	Address      string
	RedirectPort int
	Port         int
	CertFile     string
	CertKey      string
	DBLocation   string
	logger.LogConfig
}

type PubSub struct {
	ch chan handlers.Msg
}

func (ps *PubSub) Publish(kind, variant string, content interface{}) {
	ps.ch <- handlers.Msg{
		Kind:     kind,
		Variant:  variant,
		Contents: content,
	}
}

func getDefaultDBLocation() string {
	if s, err := os.Stat("./skiver.bbolt"); err == nil && !s.IsDir() {
		return "./skiver.bbolt"
		// When running in a
	} else if s, err := os.Stat("./storage"); err == nil && s.IsDir() {
		return "./storage/skiver.bbolt"
	}
	return "./skiver.bbolt"
}

//g//o:generate sh -c "cd ../frontend && yarn gen"
func main() {
	err := cfg.InitConfig()
	if err != nil {
		panic(err)
	}
	// TODO: owen the config!
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
	pubsub := PubSub{make(chan handlers.Msg)}
	db, err := bboltStorage.NewBbolt(l, cfg.DBLocation, &pubsub)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	ctx := requestContext.Context{
		L:               l,
		DB:              &db,
		StructValidater: &jsonschema.Schema{},
	}
	go utils.SelfCheck(utils.SelfCheckLimit{
		MemoryMB:   1000,
		GoRoutines: 10000,
		Streaks:    5,
		Interval:   time.Second * 15,
	}, logger.GetLogger("self-check"), 0)

	address := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	handler := http.NewServeMux()
	handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	handler.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	handler.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	handler.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	handler.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	// TODO: consider using a buffered channel.
	handler.Handle("/ws/", handlers.NewWsHandler(logger.GetLoggerWithLevel("ws", "debug"), pubsub.ch, handlers.WsOptions{}))

	handler.Handle("/api/",
		gziphandler.GzipHandler(
			http.StripPrefix("/api/", EndpointsHandler(ctx))))

	useCert := false
	if cfg.CertFile != "" {
		_, err := os.Stat(cfg.CertFile)
		if err == nil {
			useCert = true
		}

	}

	if isDev {
		// In development, we serve the file directly.
		// handler.Handle("/", http.FileServer(http.Dir("./frontend/dist/")))
	} else {
		// handler.Handle("/", frontend.DistServer)
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
		srv.ListenAndServe()
	}
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to create listener")
	}

}

type AccessControl struct {
	AllowOrigin string
	MaxAge      time.Duration
}

var (
	accessControl = AccessControl{
		AllowOrigin: "_any_",
		MaxAge:      24 * time.Hour,
	}
	pingByte = []byte{}
)

func EndpointsHandler(ctx requestContext.Context) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "ping" {
			rw.Write(pingByte)
			return
		}
		h := rw.Header()
		switch accessControl.AllowOrigin {
		case "_any_":
			h.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		default:
			h.Set("Access-Control-Allow-Origin", accessControl.AllowOrigin)
		}
		h.Set("Access-Control-Allow-Headers", "x-request-id, content-type, jmes-path")
		h.Set("Access-Control-Max-Age", fmt.Sprintf("%0.f", accessControl.MaxAge.Seconds()))
		if r.Method == "OPTIONS" {
			h.Set("Cache-Control", fmt.Sprintf("public, max-age=%0.f", accessControl.MaxAge.Seconds()))
			h.Set("Vary", "origin")

			return
		}
		rc := requestContext.NewReqContext(&ctx, r, rw)
		var body []byte
		var err error
		isGet := r.Method == http.MethodGet
		isPost := r.Method == http.MethodPost
		// isDelete := r.Method == http.MethodDelete
		isPut := r.Method == http.MethodPut
		path := r.URL.Path
		paths := strings.Split(strings.TrimSuffix(path, "/"), "/")

		if rc.ContentKind > 0 && (isPost || isPut) {
			body, err = ioutil.ReadAll(r.Body)
			if err != nil {
				rc.WriteErr(err, requestContext.CodeErrReadBody)
			}
		}

		switch paths[0] {
		case "swagger", "swagger.yaml", "swagger.yml":
			rw.Header().Set("Content-Type", "text/vnd.yaml")
			rw.Header().Set("Content-Disposition", `attachment; filename="swagger-skiver.yaml"`)
			rw.Write([]byte(swaggerYml))
			return
		case "serverInfo":
			if isGet && len(paths) == 1 {
				info := types.ServerInfo{
					ServerStartedAt: serverStartTime,
					GitHash:         GitHash,
					Version:         Version,
					BuildDate:       BuildDate,
				}
				size, sizeErr := ctx.DB.Size()
				if sizeErr != nil {
					ctx.L.Warn().Err(sizeErr).Msg("Failed to retrieve size of database")
				} else {
					info.DatabaseSize = size
					info.DatabaseSizeStr = humanize.Bytes(uint64(size))
				}

				rc.WriteAuto(info, err, "serverInfo")
				return
			}
		case "endpoint":
			rw.Write(body)
			return
			// // Create endpoint
			// if isPost && len(paths) == 1 {
			// 	var input types.EndpointPayload
			// 	if err := rc.ValidateBytes(body, &input); err != nil {
			// 		return
			// 	}
			// 	e, err := ctx.DB.CreateEndpoint(input)
			// 	rc.WriteAuto(e, err, requestContext.CodeErrDBCreateEndpoint)
			// 	return
			// }
			// // List endpoints
			// if isGet && len(paths) == 1 {
			// 	es, err := ctx.DB.Endpoints()
			// 	rc.WriteAuto(es, err, requestContext.CodeErrEndpoint)
			// 	return
			// }
			// // Get endpoint
			// if isGet && len(paths) == 2 {
			// 	es, err := ctx.DB.Endpoint(paths[1])
			// 	rc.WriteAuto(es, err, requestContext.CodeErrEndpoint)
			// 	return
			// }
			// // Update endpoint
			// if isPut && len(paths) == 2 {
			// 	var input types.EndpointPayload
			// 	if err := rc.ValidateBytes(body, &input); err != nil {
			// 		return
			// 	}
			// 	e, err := ctx.DB.UpdateEndpoint(paths[1], input)
			// 	rc.WriteAuto(e, err, requestContext.CodeErrDBUpdateEndpoint)
			// 	return
			// }
			// // Delete endpoint
			// if isDelete && len(paths) == 2 {
			// 	e, err := ctx.DB.SoftDeleteEndpoint(paths[1])
			// 	rc.WriteAuto(e, err, requestContext.CodeErrDBDeleteEndpoint)
			// 	return
			// }
			// http.FileServer(frontend.StaticFiles).ServeHTTP(rc.Rw, rc.rw)

			rc.WriteError(fmt.Sprintf("No route registerd for: %s %s", r.Method, r.URL.Path), requestContext.CodeErrNoRoute)
		}
	}
}
