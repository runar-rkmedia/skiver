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
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "embed"
	"net/http/pprof"

	"github.com/NYTimes/gziphandler"
	swaggerMiddleware "github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
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
	"github.com/zserge/metric"
)

// TODO: update to use debug.BuildInfo or see https://github.com/carlmjohnson/versioninfo/

var (
	//go:embed swagger.yml
	swaggerYml string
	// These are added at build...
	version   string
	date      string
	buildDate time.Time
	builtBy   string
	commit    string
	isDev     = true
	IsDevStr  = "1"

	serverStartTime = time.Now()
)

func init() {
	if date != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", date)
		if err != nil {
			panic(fmt.Errorf("Failed to parse build-date: %w", err))
		}
		buildDate = t
	}
	if IsDevStr != "1" || commit != "" {
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
	if m.l.HasTrace() {

		m.l.Trace().Str("kind", kind).Str("variant", variant).Interface("contents", contents).Msg("Event received")
		return
	}
	m.l.Debug().Str("kind", kind).Str("variant", variant).Msg("Event received")
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
	if kind != string(types.PubTypeTranslationValue) {
		return
	}
	if variant != string(types.PubVerbCreate) && variant != string(types.PubVerbUpdate) {
		return
	}
	tv, ok := contents.(types.TranslationValue)
	if !ok {
		m.l.Error().Interface("content", contents).Msg("Failed to convert contents to TranslationValue")
		return
	}
	orgId := tv.OrganizationID
	if tv.Source == types.CreatorSourceImport {
		if debug {
			m.l.Debug().Interface("content", contents).Msg("ignoring TranslationValue since it was sourced from an import")
		}
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

	// We need the project-settings, so we resolve the project
	var project *types.Project
	{
		t, err := m.db.GetTranslation(tv.TranslationID)
		if err != nil {
			m.l.Error().Str("translationID", tv.TranslationID).Err(err).Msg("failed to lookup translation")
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
		// TODO: implement for contexts too
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

func getInstanceHash() string {
	rand.Seed(time.Now().Unix())
	n := rand.Int63n(1_000_000)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, n)
	if err != nil {
		panic(err)
	}
	w := utils.HashName(buf.Bytes())

	s := strings.Split(w, "-")
	j := strings.Join(s[:2], "-")
	return j
}
func gethostNameHash() string {
	h, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	w := utils.HashName([]byte(h))
	s := strings.Split(w, "-")
	j := strings.Join(s[:2], "-")
	return j
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
		WithCaller: config.LogLevel == "debug" || commit == "",
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
		Str("version", version).
		Time("buildDate", buildDate).
		Time("buildDateLocal", buildDate.Local()).
		Str("gitHash", commit).
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
		handler.Handle("/debug/metrics", metric.Handler(metric.Exposed))
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

	info := struct {
		types.ServerInfo
		sync.RWMutex
	}{
		types.ServerInfo{
			ServerStartedAt: serverStartTime,
			GitHash:         commit,
			Version:         version,
			BuildDate:       buildDate,
			Instance:        getInstanceHash(),
			HostHash:        gethostNameHash(),
		},
		sync.RWMutex{},
	}

	go func() {
		cacheFile := "./latest.cache.json"
		if isDev {
			stat, err := os.Stat(cacheFile)
			if err == nil {
				diff := time.Now().Sub(stat.ModTime())
				if diff < time.Hour {

					b, err := os.ReadFile(cacheFile)
					if err == nil && b != nil {
						err := json.Unmarshal(b, &info.LatestRelease)
						if err != nil {
							l.Error().Err(err).Msg("Failed to unmarshal latest-release from cache")
						} else {
							return

						}
					} else {
						l.Error().Err(err).Msg("Failed to read latest-release from cache")
					}
				}
			}

		}
		ticker := time.NewTicker(time.Hour * 1)
		for ; true; <-ticker.C {
			r, err := types.GetLatestVersion(http.DefaultClient)
			if err != nil {
				l.Error().Err(err).Msg("Failed to check latest release-version")
				continue
			}
			info.Lock()
			info.LatestRelease = r
			info.Unlock()
			if isDev {
				b, _ := json.Marshal(r)
				if err := os.WriteFile(cacheFile, b, 0677); err != nil {
					l.Error().Err(err).Msg("Failed to write release-cache")
				}
			}
		}
	}()
	p, ok := ctx.DB.(localuser.Persistor)
	if !ok {
		ctx.L.Warn().Str("type", fmt.Sprintf("%T", ctx.DB)).Msg("DB does not implement the localUser.Persistor-interface")
	}
	userSessions, err := localuser.NewUserSessionInMemory(types.UserSessionOptions{TTL: config.Authentication.SessionLifeTime}, uuid.NewString, p)
	if err != nil {
		ctx.L.Fatal().Err(err).Msg("Failed to set up userSessions")
	}

	router := httprouter.New()
	router.HandleMethodNotAllowed = true
	router.HandleOPTIONS = true
	router.RedirectTrailingSlash = true
	// router.PanicHandler = func(rw http.ResponseWriter, r *http.Request, i interface{}) {
	// 	// TODO: in this handler, we should probably get rc from r.context
	// 	rc := ctx.NewReqContext(rw, r)
	// 	l.Error().
	// 		Str("path", r.URL.Path).
	// 		Str("method", r.Method).
	// 		Interface("panic-data", i).Msg("Panic")

	// 	rc.WriteError("Internal error. I am terribly sorry, but I must have overlooked something.", "Internal panic")
	// }
	auth := handlers.NewAuthHandler(userSessions)

	type routeOptions struct {
		sessionRole func(s types.Session, r *http.Request) error
	}
	// This is still being fleshed out...
	// Should use middleware-pattern
	c := func(name string, h handlers.AppHandler, opts ...routeOptions) httprouter.Handle {
		expvar.Publish("endpoint-"+name, metric.NewHistogram())
		expvar.Publish("endpoint-count-"+name, metric.NewCounter())
		var options routeOptions
		if len(opts) > 0 {
			options = opts[0]
		}
		return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
			startTime := time.Now()
			defer func() {
				expvar.Get("endpoint-" + name).(metric.Metric).Add(float64(time.Since(startTime).Milliseconds()))
				expvar.Get("endpoint-count-" + name).(metric.Metric).Add(1)
			}()

			if len(p) > 0 {
				ctx := r.Context()
				ctx = context.WithValue(ctx, httprouter.ParamsKey, p)
				r = r.WithContext(ctx)
			}

			handlers.AddAccessControl(r, rw)
			// This turned out a bit hacky, but it leaves a nicer pattern for route-handlers
			// since they don't need to care about auth, logging, or writing
			// TODO: change the details here
			if l.HasDebug() {
				l.Debug().Str("Path", r.URL.Path).Str("method", r.Method).Str("handler", name).Interface("params", p).Msg("Incoming request")
				defer func() {
					h := rw.Header().Clone()
					for k := range h {
						l := strings.ToLower(k)
						if strings.Contains(l, "cookie") || strings.Contains(l, "auth") {
							h.Del(k)
						}
					}
					l.Debug().Str("Path", r.URL.Path).Str("method", r.Method).Str("handler", name).Interface("outgoing-headers", h).Msg("Outgoing response")
				}()
			}
			rc := ctx.NewReqContext(rw, r)
			_r, err := auth(rw, r)
			if err != nil {
				rc.WriteErr(err, "Internal server error")
				return
			}
			r = _r

			if options.sessionRole != nil {
				s, err := handlers.GetRequestSession(r)
				if err != nil {
					rc.WriteError("Authentication required", requestContext.CodeErrAuthenticationRequired)
					return
				}
				err = options.sessionRole(s, r)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrAuthoriziation)
					return
				}
			}

			if err != nil {
				rc.WriteErr(err, "Authentication Error")
				return
			}
			output, err := h(rc, rw, r)
			if err != nil {
				rc.WriteErr(err, "")
				return
			}
			if output != nil {
				rc.WriteOutput(output, http.StatusOK)
			} else {
				if rc.L.HasDebug() {
					rc.L.Warn().Msg("No output produced. This may be a false warning as we are migrating to a new httpMux-pattern")
				}
			}
		}
	}
	// We are migrating to using httprouter, but not all routes have been migrated
	router.GET("/api/export/:params", c("GetExport", handlers.GetExport(exportCache)))
	router.GET("/api/export/", c("GetExportx", handlers.GetExport(exportCache)))
	router.GET("/api/user/", c("GetSimpleUsers", handlers.ListUsers(&db, true)))
	router.GET("/api/missing/", c("GetMissing", handlers.GetMissing(&db)))
	router.POST("/api/missing/:locale/:project", c("ReportMissing", handlers.PostMissing(&db)))
	router.GET("/api/category/", c("GetCategory", handlers.GetCategory(&db)))
	router.POST("/api/category/", c("PostCategory", handlers.PostCategory(&db), routeOptions{
		sessionRole: func(s types.Session, r *http.Request) error {
			if !s.User.CanCreateTranslations {
				return fmt.Errorf("You are not authorized to manage translations")
			}
			return nil
		},
	}))
	router.PUT("/api/category/", c("UpdateCategory", handlers.UpdateCategory(&db), routeOptions{
		sessionRole: func(s types.Session, r *http.Request) error {
			if !s.User.CanCreateTranslations {
				return fmt.Errorf("You are not authorized to manage translations")
			}
			return nil
		},
	}))
	router.GET("/api/users/", c("GetUsers", handlers.ListUsers(&db, false), routeOptions{
		sessionRole: func(s types.Session, r *http.Request) error {
			if !s.User.CanUpdateUsers {
				return fmt.Errorf("You are not authorized to manage users")
			}
			return nil
		},
	}))
	router.POST("/api/user/password", c("ChangePassword", handlers.ChangePassword(&db, &pw)))
	router.POST("/api/user/token", c("CreateToken", handlers.CreateToken(userSessions)))
	router.POST("/api/project/snapshotdiff/", c("DiffSnapshot", handlers.GetDiff(exportCache)))
	router.DELETE("/api/translation/:id/", c("DeleteTranslation", handlers.DeleteTranslation(),
		routeOptions{sessionRole: func(s types.Session, r *http.Request) error {
			if !s.User.CanUpdateTranslations {
				return fmt.Errorf("You are not authorized to delete translations")
			}
			return nil
		}}))
	router.PUT("/api/translation/", c("UpdateTranslation", handlers.UpdateTranslation(),
		routeOptions{sessionRole: func(s types.Session, r *http.Request) error {
			if !s.User.CanUpdateTranslations {
				return fmt.Errorf("You are not authorized to update translations")
			}
			return nil
		}}))
	router.POST("/api/translation/", c("CreateTranslation", handlers.CreateTranslation(),
		routeOptions{sessionRole: func(s types.Session, r *http.Request) error {
			if !s.User.CanCreateTranslations {
				return fmt.Errorf("You are not authorized to create translations")
			}
			return nil
		}}))
	router.GET("/api/translation/", c("GetTranslation", handlers.GetTranslations()))
	router.GET("/api/serverInfo/", c("GetServerInfo", handlers.GetServerInfo(&db, func() *types.ServerInfo {
		info.RLock()
		defer info.RUnlock()
		return &info.ServerInfo

	})))
	router.POST("/api/project/snapshot/", c("PostSnapshot", handlers.PostSnapshot(), routeOptions{sessionRole: func(s types.Session, _ *http.Request) error {
		if !s.User.CanUpdateProjects {
			return fmt.Errorf("You are not authorized to create snapshots")
		}
		return nil
	}}))

	apiHandler := http.StripPrefix("/api/",
		handlers.EndpointsHandler(ctx, userSessions, pw, []byte(swaggerYml)),
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
	handler.Handle("/api/project/snapshot/", router)
	handler.Handle("/api/project/snapshotdiff/", router)
	handler.Handle("/api/export/", router)
	handler.Handle("/api/translation/", router)
	handler.Handle("/api/users/", router)
	handler.Handle("/api/user/", router)
	handler.Handle("/api/wordcloud/", router)
	handler.Handle("/api/missing/", router)
	handler.Handle("/api/serverInfo/", router)
	handler.Handle("/api/category/", router)
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
