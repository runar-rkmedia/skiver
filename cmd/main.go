package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/types"
)

func required(v, s string) {
	if v != "" {
		return
	}
	fmt.Println(s, "is required")
	flag.Usage()
	os.Exit(1)
}

var CLI struct {
	Endpoint *url.URL `help:"Endpoint for skiver" `
	Project  string   `help:"Project-id/ShortName" required:""`
	Token    string   `help:"Token used for authenticaion"`
	Locale   string   `help:"Locale to use"`

	Import struct {
		Source *os.File `help:"Source-file for import" arg:""`
	} `help:"Import from file" cmd:""`

	Generate struct {
		Path   string   `help:"Ouput file to write to" type:"path"`
		TsKeys struct{} `help:"Generate a typescript key file for typesafe referance of translation-keys with TsDoc filled information from project" cmd:""`
	} `help:"Generate files from project etc." cmd:""`

	Inject struct {
		Dir string `help:"Directory for source-code" type:"existingdir" arg:""`
	} `help:"Inject helper-comments into source-files" cmd:""`
	Config struct {
		Show struct{} `help:"Print effective config" cmd:""`
	} `help:"Configuration" cmd:"" json:"-"`
	LogFormat string `help:"Human or json" default:"human"`
	Verbose   int    `help:"More verbose logging" type:"counter"`
	Quiet     int    `help:"Quiet" type:"counter"`
}

func main() {

	var paths []string
	if p, err := os.Getwd(); err == nil {
		paths = append(paths, path.Join(p, "skiver-cli.json"))
	}
	if p, err := os.UserConfigDir(); err == nil {
		paths = append(paths, path.Join(p, "skiver-cli", "config.json"))
	}
	if p, err := os.UserHomeDir(); err == nil {
		paths = append(paths, path.Join(p, "skiver-cli.json"))
	}
	ctx := kong.Parse(&CLI,
		kong.Name("Skiver CLI"),
		kong.Description("Interactions with skiver, a developer-focused translation-service"),
		kong.Configuration(kong.JSON, paths...),
	)
	level := "info"
	if CLI.Verbose > 0 {
		level = "debug"
	}
	if CLI.Quiet > 0 {
		level = "warn"
	}
	l := logger.InitLogger(logger.LogConfig{
		Format: CLI.LogFormat,
		Level:  level,
	})
	api := NewAPI(l, CLI.Endpoint.String())
	api.SetToken(CLI.Token)
	switch ctx.Command() {
	case "config show":
		bebo, _ := json.MarshalIndent(CLI, "", "  ")
		fmt.Println(string(bebo))
		os.Exit(0)

	case "import <source>":
		l.Debug().Msg("importing")
		err := api.Import(CLI.Project, "i18n", CLI.Import.Source)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to import")
		}
		l.Info().Msg("Successful import")
	case "generate ts-keys":
		var w io.Writer
		format := "typescript"
		locale := CLI.Locale
		if locale == "" {
			l.Fatal().Msg("Locale is required")
		}
		ll := l.Debug().Str("project", CLI.Project).
			Str("format", format)
		if CLI.Generate.Path != "" {
			outfile, err := os.OpenFile(CLI.Generate.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				log.Fatal(err)

				return
			}
			w = outfile
			ll = ll.Str("path", CLI.Generate.Path)
		}
		if w == nil {
			w = os.Stdout
			ll = ll.Bool("stdout", true)
		}

		if l.HasDebug() {
			ll.Msg("Generating file")
		}
		err := api.Export(CLI.Project, format, locale, w)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed export")
		}
		l.Info().Msg("Successful export")
	default:
		l.Fatal().Str("command", ctx.Command()).Msg("Not implemented yet")
	}
}

type Api struct {
	l        logger.AppLogger
	endpoint string
	cookies  []*http.Cookie
	login    *types.LoginResponse
	client   *http.Client
}

func NewAPI(l logger.AppLogger, endpoint string) Api {
	c := http.Client{Timeout: time.Minute}
	return Api{
		l:        l,
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &c,
	}
}

func (a *Api) SetToken(token string) {
	c := http.Cookie{
		Name:  "token",
		Value: token,
	}
	a.cookies = append(a.cookies, &c)
}
func (a *Api) Login(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("Missing username/password")
	}
	payload := struct{ Username, Password string }{Username: username, Password: password}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("Failed to marshal login-payload: %w", err)
	}
	r, err := http.NewRequest(http.MethodPost, a.endpoint+"/api/login/", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("failed to create login-request: %w", err)
	}
	r.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("login-request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed reading body of login-request: %w", err)
	}
	if res.StatusCode >= 300 {
		var j models.APIError
		err = json.Unmarshal(body, &j)
		if err == nil {
			return fmt.Errorf("Login-request return a %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("Login-request return a %d-response: %s", res.StatusCode, string(body))
	}
	var j types.LoginResponse
	err = json.Unmarshal(body, &j)
	if err != nil {
		return fmt.Errorf("failed reading body of login-request: %w", err)
	}
	a.cookies = res.Cookies()
	if a.l.HasDebug() {
		a.l.Debug().
			Int("statusCode", res.StatusCode).
			Str("path", res.Request.URL.String()).
			Str("method", res.Request.Method).
			Interface("login-response", j).
			Msg("Result of request")
	}
	return nil

}
func (a Api) Import(projectName string, kind string, reader io.Reader) error {
	if len(a.cookies) == 0 {
		return fmt.Errorf("Not logged in")
	}
	r, err := http.NewRequest(http.MethodPost, a.endpoint+"/api/import/"+kind+"/"+projectName, reader)
	if err != nil {
		return fmt.Errorf("failed to create import-request: %w", err)
	}
	r.Header.Add("Content-Type", "application/json")
	for _, c := range a.cookies {
		r.AddCookie(c)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("import-request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed reading body of import-request: %w", err)
	}
	if res.StatusCode >= 300 {
		var j models.APIError
		err = json.Unmarshal(body, &j)
		if err == nil {
			return fmt.Errorf("import-request return a %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("import-request return a %d-response: %s", res.StatusCode, string(body))
	}
	var j handlers.ImportResult
	err = json.Unmarshal(body, &j)
	if err != nil {
		return fmt.Errorf("failed reading body of import-request: %w", err)
	}
	if a.l.HasDebug() {
		a.l.Debug().
			Int("statusCode", res.StatusCode).
			Str("path", res.Request.URL.String()).
			Str("method", res.Request.Method).
			Interface("import-warnings", j.Warnings).
			Int("translation-creations", len(j.Changes.TranslationCreations)).
			Int("category-creations", len(j.Changes.CategoryCreations)).
			Int("translation-value-creations", len(j.Changes.TranslationValueUpdates)).
			Int("translation-value-creations", len(j.Changes.TranslationsValueCreations)).
			Msg("Result of request")
	}
	return nil

}

func (a Api) Export(projectName string, format string, locale string, writer io.Writer) error {
	if len(a.cookies) == 0 {
		return fmt.Errorf("Not logged in")
	}
	r, err := http.NewRequest(http.MethodGet, a.endpoint+"/api/export/", nil)
	if err != nil {
		return fmt.Errorf("failed to create export-request: %w", err)
	}
	q := r.URL.Query()
	q.Set("format", format)
	q.Set("locale", locale)
	q.Set("project", projectName)
	r.URL.RawQuery = q.Encode()
	r.Header.Add("Content-Type", "application/json")
	for _, c := range a.cookies {
		r.AddCookie(c)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return fmt.Errorf("export-request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed reading body of export-request: %w", err)
		}
		var j models.APIError
		err = json.Unmarshal(body, &j)
		if err == nil {
			return fmt.Errorf("export-request return a %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("export-request return a %d-response: %s", res.StatusCode, string(body))
	}
	written, err := io.Copy(writer, res.Body)
	if a.l.HasDebug() {
		a.l.Debug().
			Int("statusCode", res.StatusCode).
			Str("path", res.Request.URL.String()).
			Str("method", res.Request.Method).
			Int64("written-bytes", written).
			Str("written-text", humanize.Bytes(uint64(written))).
			Msg("Result of request")
	}
	return nil

}
