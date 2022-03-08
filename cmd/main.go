package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
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
		IgnoreFilter []string `help:"Ignore-filter for files"`
		DryRun       bool     `help:"Enable dry-run"`
		OnReplace    string   `help:"Command to run on file after replacement, like prettier"`
		Dir          string   `help:"Directory for source-code" type:"existingdir" arg:""`
	} `help:"Inject helper-comments into source-files" cmd:""`
	Config struct {
		Show struct{} `help:"Print effective config" cmd:""`
	} `help:"Configuration" cmd:"" json:"-"`
	LogFormat string `help:"Human or json" default:"human"`
	Verbose   int    `help:"More verbose logging" type:"counter" short:"v"`
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
	case "inject <dir>":
		buf := bytes.Buffer{}
		err := api.Export(CLI.Project, "raw", CLI.Locale, &buf)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to get exported project")
		}
		var ep types.ExtendedProject
		err = json.Unmarshal(buf.Bytes(), &ep)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to unmarshal exported project")
		}
		m, err := FlattenExtendedProject(ep, []string{CLI.Locale})
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to flatten exported project")
		}
		if len(m) == 0 {
			l.Fatal().Msg("Found no matches")
		}
		// os.WriteFile("export-ignore-me.json", buf.Bytes(), 0677)
		reg := `(.*")(`
		var regexKeys = make([]string, len(m))
		i := 0
		for k := range m {
			r := regexp.QuoteMeta(k)
			regexKeys[i] = r
			i++
		}
		reg += strings.Join(regexKeys, "|") + `)("(?: as any)?)(.*)`

		regex := regexp.MustCompile(reg)
		replacementFunc := func(groups []string) (replacement string, changed bool) {
			if len(groups) < 3 {
				return "", false

			}
			prefix := groups[0]
			// If the line is a comment, we dont care about replacing it
			if strings.HasPrefix(strings.TrimSpace(prefix), "//") {
				return "", false
			}
			key := groups[1]
			suffix := groups[2]
			rest := strings.Join(groups[3:], "")
			skiverComment := "// skiver: "
			var prevSkiverComment string
			if i := strings.Index(rest, skiverComment); i >= 0 {
				prevSkiverComment = rest[i:]
				rest = rest[0:i]
			}

			found, ok := m[key]
			if !ok {
				panic("Not found")
			}
			var ts string
			if len(found) == 0 {
				return "", false
			}
			f := utils.SortedMapKeys(found)
			for _, k := range f {
				if found[k] == "" {
					continue
				}
				ts += fmt.Sprintf("(%s) %s; ", k, found[k])

			}
			if ts == "" {
				return "", false
			}

			ts = skiverComment + ts
			ts = newLineReplacer.Replace(ts)
			ts = strings.TrimSuffix(ts, " ")
			if prevSkiverComment != "" {
				if prevSkiverComment == ts {
					return "", false
				}
			}

			if strings.TrimSpace(rest) == "," {
				return prefix + key + suffix + ", " + ts, true
			}

			return prefix + key + suffix + ts + "\n" + rest, true

		}
		filter := []string{"ts", "tsx"}

		in := NewInjector(l, CLI.Inject.Dir, CLI.Inject.DryRun, CLI.Inject.OnReplace, CLI.Inject.IgnoreFilter, filter, regex, replacementFunc)
		err = in.Inject()
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to inject")
		}
		l.Info().Msg("Done")
	default:
		l.Fatal().Str("command", ctx.Command()).Msg("Not implemented yet")
	}
}

var newLineReplacer = strings.NewReplacer("\n", "", "\r", "")

func GetFileAndContent(dryRun bool, fn string, fi os.FileInfo) (f *os.File, content []byte, err error) {

	if dryRun {
		f, err = os.Open(fn)
	} else {
		f, err = os.OpenFile(fn, os.O_RDWR, 0666)
	}

	if err != nil {
		return
	}

	content = make([]byte, fi.Size())
	n, err := f.Read(content)
	if err != nil {
		return
	}
	if int64(n) != fi.Size() {
		err = fmt.Errorf("Thw whole file was not read, only %s of %s", humanize.Bytes(uint64(n)), humanize.Bytes(uint64(fi.Size())))
	}

	return
}

func matchesLocale(l types.Locale, locales []string) string {
	for _, loc := range locales {
		if l.ID == loc {
			return loc
		}
		if l.IETF == loc {
			return loc
		}
		if l.Iso639_3 == loc {
			return loc
		}
		if l.Iso639_2 == loc {
			return loc
		}
		if l.Iso639_1 == loc {
			return loc
		}

	}

	return ""
}
func FlattenExtendedProject(ep types.ExtendedProject, locales []string) (map[string]map[string]string, error) {
	m := map[string]map[string]string{}
	for _, c := range ep.CategoryTree.Categories {
		for _, t := range c.Translations {
			key := c.Key + "." + t.Key
			for _, tv := range t.Values {
				loc := matchesLocale(ep.Locales[tv.LocaleID], locales)
				if loc == "" {
					continue
				}
				mm := map[string]string{}
				if tv.Value != "" {
					mm[loc] = tv.Value
				}
				for k, c := range tv.Context {
					mm[loc+"_"+k] = c
				}

				if len(mm) == 0 {
					continue
				}
				m[key] = mm

			}

		}

	}

	return m, nil
}

func (in Injecter) VisitFile(fPath string, info fs.FileInfo) (bool, error) {
	l := logger.With(in.l.With().Str("dir", in.Dir).Logger())

	f, b, err := GetFileAndContent(in.DryRun, fPath, info)
	if err != nil {
		return false, fmt.Errorf("Failed to read file %s: %w", fPath, err)
	}
	s := string(b)
	changed := false
	replacement := ReplaceAllStringSubmatchFunc(in.Regex, s, func(groups []string, start, end int) string {
		if len(groups) < 4 {
			l.Fatal().Interface("groups", groups).Msg("Expected to have 4 groups (whole match, prefix, content and suffix)")
		}
		repl, hasChange := in.ReplacementFunc(groups[1:])
		if !hasChange {
			return groups[0]
		}
		changed = true
		return repl
	})

	if !changed {
		return false, nil
	}

	if in.DryRun {
		fmt.Println(fPath)
		fmt.Println(replacement)
		return false, nil
	}
	f.Seek(0, 0)
	n, err := f.Write([]byte(replacement))
	if err != nil {
		return true, fmt.Errorf("Error writing replacement to file '%s': %s",
			fPath, err)
	}
	if int64(n) < info.Size() {
		err := f.Truncate(int64(n))
		if err != nil {
			return true, fmt.Errorf("Error truncating file '%s' to size %d",
				fPath, n)
		}
	}
	if in.OnReplaceCmd != "" {
		a := strings.Split(in.OnReplaceCmd, " ")
		cmd := a[0]
		args := []string{}
		if len(a) > 1 {
			args = a[1:]
		}
		args = append(args, fPath)
		c := exec.Command(cmd, args...)
		bufin := strings.NewReader(s)
		c.Stdin = bufin
		if l.HasDebug() {
			l.Debug().
				Str("path", fPath).
				Str("cmd", cmd).
				Interface("args", args).
				Msg("Running command on replacement")
		}
		if out, err := c.CombinedOutput(); err != nil {
			return true, fmt.Errorf("Failed to run onReplaceCmd %s %s %v: %w", c.Path, string(out), c.Args, err)
		}
	}
	return true, nil
}

type Injecter struct {
	l               logger.AppLogger
	Dir             string
	DryRun          bool
	OnReplaceCmd    string
	ExtensionFilter map[string]bool
	IgnoreFilter    []string
	Regex           *regexp.Regexp
	ReplacementFunc ReplacementFunc
}
type ReplacementFunc = func(groups []string) (s string, changed bool)

type st = struct {
	FilePath string
	Start    time.Time
	Duration time.Duration
}
type _written []st

func (st _written) Len() int {
	return len(st)
}
func (st _written) Less(i, j int) bool {
	return st[i].Duration < st[j].Duration
}
func (st _written) Swap(i, j int) {
	st[i], st[j] = st[j], st[i]
}

func NewInjector(l logger.AppLogger, dir string, dryRun bool, onReplace string, ignoreFilter []string, extFilter []string, regex *regexp.Regexp, replacementFunc ReplacementFunc) Injecter {

	in := Injecter{
		l:               l,
		DryRun:          dryRun,
		Dir:             dir,
		OnReplaceCmd:    onReplace,
		IgnoreFilter:    ignoreFilter,
		ExtensionFilter: map[string]bool{},
		Regex:           regex,
		ReplacementFunc: replacementFunc,
	}

	for _, ext := range extFilter {
		in.ExtensionFilter[ext] = true
	}

	return in
}

func (in Injecter) Inject() error {
	in.l.Debug().
		Str("dir", in.Dir).
		Msg("Started injection in path")
	paths := map[string]fs.FileInfo{}
	wg := sync.WaitGroup{}

	type s = struct {
		FilePath string
		Info     fs.FileInfo
	}
	concurrency := runtime.NumCPU()
	ch := make(chan s, concurrency)
	writtenCh := make(chan st)

	var written _written

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() error {
			for {
				select {
				case ss := <-ch:
					sst := st{FilePath: ss.FilePath, Start: time.Now()}
					changed, err := in.VisitFile(ss.FilePath, ss.Info)
					sst.Duration = time.Now().Sub(sst.Start)
					if err != nil {
						in.l.Fatal().Err(err).Str("path", ss.FilePath).Msg("Failed replacement in file")
					}
					if in.l.HasDebug() {
						in.l.Debug().
							Str("path", sst.FilePath).
							Str("duration", sst.Duration.String()).
							Bool("changed", changed).
							Msg("Completed replacement in file")
					}
					if changed {
						writtenCh <- sst
					}
					wg.Done()
				}
			}
		}()
	}

	go func() {
		for {
			select {
			case sst := <-writtenCh:
				written = append(written, sst)
			}
		}
	}()

	var walker filepath.WalkFunc = func(fPath string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		ext := strings.TrimPrefix(path.Ext(name), ".")
		if _, ok := in.ExtensionFilter[ext]; !ok {
			return nil
		}
		for _, ignore := range in.IgnoreFilter {
			// TODO: use gitignore etc.
			if strings.Contains(name, ignore) {
				return nil
			}

		}
		paths[fPath] = info
		ch <- s{fPath, info}
		wg.Add(1)
		return nil
	}
	if err := filepath.Walk(in.Dir, walker); err != nil {
		return err
	}
	in.l.Debug().
		Str("dir", in.Dir).
		Int("count", len(paths)).
		Int("concurrency", concurrency).
		Msg("Started injection for files")

	wg.Wait()

	in.l.Debug().
		Str("dir", in.Dir).
		Int("count", len(paths)).
		Int("concurrency", concurrency).
		Int("writtenCount", len(written)).
		Str("duration", time.Now().Sub(start).String()).
		Msg("Completed injection for files")

	if in.l.HasDebug() {
		sort.Sort(written)
		for _, v := range written {

			fmt.Println(v.Duration, v.FilePath)

		}
	}

	return nil

}

func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string, int, int) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			groups = append(groups, str[v[i]:v[i+1]])
		}

		result += str[lastIndex:v[0]] + repl(groups, v[0], v[1])
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
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
