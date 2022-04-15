package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/dustin/go-humanize"
	"github.com/ghodss/yaml"
	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/handlers"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

var (
	// These are added at build...
	version   string
	date      string
	buildDate time.Time
	builtBy   string
	commit    string
)

func init() {
	if date != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", date)
		if err != nil {
			panic(fmt.Errorf("Failed to parse build-date: %w", err))
		}
		buildDate = t
	}
}

type URLY struct {
	url.URL
}
type secret string

func (u URLY) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}
func (u URLY) MarshalTOML() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}
func (u secret) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}
func (u secret) MarshalTOML() ([]byte, error) {
	return []byte(`""`), nil
}
func getFile(f string) (*os.File, bool) {
	if f == "" {
		return nil, false
	}
	file, err := os.Open(CLI.Unused.Source)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false
		}
		l.Fatal().Err(err).Msg("Failed to open file")
	}
	return file, true
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	l.Debug().Str("cmd", cmd).Bool("found", err == nil).Msg("Checking for existance of command")
	return err == nil
}

func runPrettier(filepath string, contents io.Reader) ([]byte, error) {
	if commandExists(CLI.PrettierDSlimPath) {
		l.Debug().Msg("prettier_d_slim is available")
		if contents == nil {
			f, err := os.Open(filepath)
			if err != nil {
				return nil, fmt.Errorf("failed to read the file-contents prior to runnning command")
			}
			defer f.Close()
			contents = f
		}
		return runCmd(CLI.PrettierPath+" -w --ignore-path NOEXIST --stdin --stdin-filepath", filepath, contents)
	}
	if commandExists(CLI.PrettierPath) {
		l.Debug().Msg("prettier is available. Consider using prettier_d_slim if you want improved speed")
		return runCmd(CLI.PrettierPath+" -w --ignore-path NOEXIST", filepath, contents)
	}
	return nil, nil
}
func runCmd(command string, fPath string, stdin io.Reader) ([]byte, error) {
	a := strings.Split(command, " ")
	cmd := a[0]
	args := []string{}
	if len(a) > 1 {
		args = a[1:]
	}
	args = append(args, fPath)
	c := exec.Command(cmd, args...)
	// bufin := strings.NewReader(s)
	// c.Stdin = bufin
	if stdin != nil {
		c.Stdin = stdin
	}
	if l.HasDebug() {
		l.Debug().
			Str("path", fPath).
			Str("cmd", cmd).
			Interface("args", args).
			Msg("Running command on replacement")
	}
	out, err := c.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("Failed to run onReplaceCmd %s %s %v: %w", c.Path, string(out), c.Args, err)
	}
	return out, nil
}

type cli struct {
	Endpoint          URLY     `help:"Endpoint for skiver" env:"SKIVER_ENDPOINT" json:"endpoint"`
	Project           string   `help:"Project-id/ShortName" env:"SKIVER_PROJECT" json:"project"`
	Token             secret   `help:"Token used for authentication" env:"SKIVER_TOKEN" json:"token"`
	Locale            string   `help:"Locale to use" env:"SKIVER_LOCALE" json:"locale"`
	Version           struct{} `help:"Print version information" cmd:"" json:"-" toml:"-"`
	WithPrettier      bool     `help:"Where available, will attempt to run prettier, or prettier_d if available"`
	PrettierPath      string   `help:"Path-override for prettier" default:"prettier"`
	PrettierDSlimPath string   `help:"Path-override for prettier_d_slim, which should be faster than regular prettier" default:"prettier_d_slim"`

	Import struct {
		Source string `help:"Source-file for import" arg:"" env:"SKIVER_IMPORT_SOURCE" json:"source"`
	} `help:"Import from file" cmd:"" json:"import"`
	Generate struct {
		Path   string `help:"Ouput file to write to" type:"path" env:"SKIVER_GENERATE_PATH" json:"path"`
		Format string `help:"Generate files from export. Common formats are: i18n,typescript." json:"format" required:"true"`
	} `help:"Generate files from project etc." cmd:"" json:"generate"`
	Unused struct {
		Source string `help:"Source-file to check-against. If ommitted, the upstream project is used as source" json:"source"`
		Dir    string `help:"Directory for source-code" type:"existingdir" arg:"" required:"" json:"dir"`
	} `help:"Find unused translation-keys" cmd:"" json:"unused"`

	Inject struct {
		IgnoreFilter []string `help:"Ignore-filter for files" json:"ignore_filter"`
		DryRun       bool     `help:"Enable dry-run" json:"dry_run"`
		OnReplace    string   `help:"Command to run on file after replacement, like prettier" json:"on_replace"`
		Dir          string   `help:"Directory for source-code" type:"existingdir" arg:"" json:"dir"`
	} `help:"Inject helper-comments into source-files" cmd:"" json:"inject"`
	Config struct {
		Format string   `enum:"json,yaml,toml" default:"toml" json:"format"`
		Paths  struct{} `help:"Print paths used" cmd:"" json:"-"`
		Show   struct {
		} `help:"Print effective config" cmd:"" json:"-"`
	} `help:"Configuration" cmd:"" json:"-"`
	LogFormat string `help:"Format to log as" default:"human" enum:"json,human" json:"log_format"`
	LogLevel  string `help:"Level for logging." default:"info" enum:"trace,debug,info,warn,error,panic" json:"log_level"`
	Verbose   int64  `help:"More verbose logging. Overrides log-level." type:"counter" short:"v" json:"-"`
	Quiet     int64  `help:"Quiet. Overrides log-level" type:"counter" short:"s" json:"-"`
}

var (
	CLI                  = cli{}
	l   logger.AppLogger = logger.GetLogger("")
)

func marshal(o any, format string) []byte {
	switch format {
	case "json":
		j, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			l.Fatal().Err(err).Msg("failed to marshal (json)")
		}
		return j
	case "toml":
		b := bytes.Buffer{}
		enc := toml.NewEncoder(&b)
		enc.CompactComments(true)
		enc.ArraysWithOneElementPerLine(true)
		enc.SetTagComment("help")
		enc.SetTagName("json")
		err := enc.Encode(o)
		if err != nil {
			l.Fatal().Err(err).Msg("failed to marshal (toml)")
		}
		return b.Bytes()
	}
	j, err := yaml.Marshal(o)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to marshal (yaml)")
	}
	return j
}

func main() {
	godotenv.Load()
	var jsonPaths []string
	var yamlPaths []string
	var tomlPaths []string
	if p, err := os.Getwd(); err == nil {
		jsonPaths = append(jsonPaths, path.Join(p, "skiver-cli.json"))
		yamlPaths = append(yamlPaths, path.Join(p, "skiver-cli.yaml"))
		tomlPaths = append(tomlPaths, path.Join(p, "skiver-cli.toml"))
	}
	if p, err := os.UserConfigDir(); err == nil {
		jsonPaths = append(jsonPaths, path.Join(p, "skiver-cli", "config.json"))
		yamlPaths = append(yamlPaths, path.Join(p, "skiver-cli", "config.yaml"))
		tomlPaths = append(tomlPaths, path.Join(p, "skiver-cli", "config.toml"))
	}
	if p, err := os.UserHomeDir(); err == nil {
		jsonPaths = append(jsonPaths, path.Join(p, "skiver-cli.json"))
		yamlPaths = append(yamlPaths, path.Join(p, "skiver-cli.yaml"))
		tomlPaths = append(tomlPaths, path.Join(p, "skiver-cli.toml"))
	}
	ctx := kong.Parse(&CLI,
		kong.Name("Skiver CLI"),
		kong.Description("Interactions with skiver, a developer-focused translation-service"),
		kong.Configuration(kong.JSON, jsonPaths...),
		kong.Configuration(kongyaml.Loader, yamlPaths...),
		kong.Configuration(TomlLoader, tomlPaths...),
	)

	level := "info"
	if CLI.LogLevel != "" {
		level = CLI.LogLevel
	}
	fmt.Println("level", level)
	switch CLI.Quiet {
	case 1:
		level = "warn"
	case 2:
		level = "error"
	case 3:
		level = "fatal"
	case 4:
		level = "panic"
	}
	switch CLI.Verbose {
	case 1:
		level = "debug"
	case 2:
		level = "trace"
	case 3:
		level = "trace"
	}
	l = logger.InitLogger(logger.LogConfig{
		Format: CLI.LogFormat,
		Level:  level,
	})
	var api Api
	api = NewAPI(l, CLI.Endpoint.String())
	api.SetToken(string(CLI.Token))
	switch ctx.Command() {
	case "version":
		b := struct {
			Version   string
			Revision  string
			BuildDate *time.Time
		}{
			Version:   version,
			Revision:  commit,
			BuildDate: &buildDate,
		}

		bb, _ := toml.Marshal(b)
		fmt.Println(string(bb))
		os.Exit(0)
	case "config paths":

		exists := func(f string) bool {
			_, err := os.Stat(f)
			if err == nil {
				return true
			}
			if errors.Is(err, os.ErrNotExist) {
				return false
			}
			return true
		}
		var existing []string
		var nonExisting []string
		var allPaths []string
		allPaths = append(allPaths, jsonPaths...)
		allPaths = append(allPaths, yamlPaths...)
		allPaths = append(allPaths, tomlPaths...)
		for _, v := range allPaths {
			if exists(v) {
				existing = append(existing, v)
			} else {
				nonExisting = append(nonExisting, v)
			}
		}
		j := struct {
			AllPaths         []string `help:"List of paths the cli will check for configuration"`
			ExistingPaths    []string `help:"List of paths the cli found configurations in"`
			NonExistingPaths []string `help:"List of paths the cli did not find any configurations in"`
		}{
			AllPaths:         allPaths,
			ExistingPaths:    existing,
			NonExistingPaths: nonExisting,
		}

		fmt.Println(string(marshal(j, CLI.Config.Format)))
		os.Exit(0)
	case "config show":
		fmt.Println(string(marshal(CLI, CLI.Config.Format)))
		os.Exit(0)

	case "import <source>":
		l.Debug().Msg("importing")
		source, exists := getFile(CLI.Import.Source)
		if !exists {
			l.Fatal().Msg("File not found")
		}
		err := api.Import(CLI.Project, "i18n", CLI.Locale, source)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to import")
		}
		l.Info().Msg("Successful import")
	case "generate":

		var w io.Writer
		format := CLI.Generate.Format
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
			defer outfile.Close()
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
		l.Debug().Msg("Export completed")
		if CLI.WithPrettier && CLI.Generate.Path != "" {
			out, err := runPrettier(CLI.Generate.Path, nil)
			if err != nil {
				l.Error().Err(err).Str("out", string(out)).Msg("Failed to run prettier on output")
			}
		}
		l.Info().Msg("Successful export")
	case "unused <dir>":
		// TODO: also check if translations are used within other translations.
		// For instance, a translation may only be used by refereance.
		source, _ := getFile(CLI.Unused.Source)
		translationKeys, regex := buildTranslationMapWithRegex(l, source, api, CLI.Project, CLI.Locale)
		found := map[string]bool{}
		foundCh := make(chan string)
		quitCh := make(chan struct{})
		replacementFunc := func(groups []string) (replacement string, changed bool) {
			foundCh <- groups[1]
			return "", false
		}

		go func() {
			for {
				select {
				case <-quitCh:
					return
				case f := <-foundCh:
					found[f] = true
				}
			}
		}()
		filter := []string{"ts", "tsx"}

		in := NewInjector(l, CLI.Unused.Dir, true, "", CLI.Inject.IgnoreFilter, filter, regex, replacementFunc)
		err := in.Inject()
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to inject")
		}
		quitCh <- struct{}{}
		var unused []string
		for k := range translationKeys {
			if found[k] {
				continue
			}
			unused = append(unused, k)
		}
		sort.Strings(unused)

		count := len(unused)
		fmt.Println(strings.Join(unused, "\n"))
		if count > 0 {
			l.Info().Int("count-unused", len(unused)).Msg("Found some possibly unused translation-keys")
		} else {
			l.Info().Msg("Found no unused translation-keys")
		}

	case "inject <dir>":
		m := BuildTranslationKeyFromApi(api, l, CLI.Project, CLI.Locale)
		sorted := utils.SortedMapKeys(m)
		regex := buildTranslationKeyRegexFromMap(sorted)
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
		err := in.Inject()
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to inject")
		}
		l.Info().Msg("Done")
	default:
		l.Fatal().Str("command", ctx.Command()).Msg("Not implemented yet")
	}
}
func BuildTranslationKeyFromApi(api Api, l logger.AppLogger, projectKeyLike, localeLike string) map[string]map[string]string {
	buf := bytes.Buffer{}
	err := api.Export(projectKeyLike, "raw", localeLike, &buf)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to get exported project")
	}
	var ep types.ExtendedProject
	err = json.Unmarshal(buf.Bytes(), &ep)
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to unmarshal exported project")
	}
	m, err := FlattenExtendedProject(ep, []string{localeLike})
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to flatten exported project")
	}
	if len(m) == 0 {
		l.Fatal().Msg("Found no matches")
	}
	return m
}

func buildTranslationKeyRegexFromMap(sorted []string) *regexp.Regexp {
	reg := `(.*")(`
	var regexKeys = make([]string, len(sorted))
	i := 0
	for _, k := range sorted {
		r := regexp.QuoteMeta(k)
		regexKeys[i] = r
		i++
	}
	reg += strings.Join(regexKeys, "|") + `)("(?: as any)?)(.*)`
	return regexp.MustCompile(reg)

}

// Creates a flattened map of translationKeys
// The source can either be a file (i18next), or it will fallback to getting from the api
func buildTranslationMapWithRegex(l logger.AppLogger, fromSourceFile *os.File, api Api, project, locale string) (map[string]struct{}, *regexp.Regexp) {
	translationKeys := map[string]struct{}{}
	// var r1 *regexp.Regexp
	if fromSourceFile != nil {
		b, err := ioutil.ReadAll(fromSourceFile)
		if err != nil {
			l.Fatal().Err(err).Msg("Failed to read from source")
		}
		var j map[string]interface{}
		if err := json.Unmarshal(b, &j); err != nil {
			l.Fatal().Err(err).Msg("Failed to unmarshal source")
		}
		flat := Flatten(j)
		// strip context
		for k := range flat {
			ts := strings.Split(k, ".")
			lastTs := ts[len(ts)-1]
			key, _ := importexport.SplitTranslationAndContext(lastTs, "_")
			joined := strings.Join(append(ts[:len(ts)-1], key), ".")
			translationKeys[joined] = struct{}{}
		}
	} else {
		mm := BuildTranslationKeyFromApi(api, l, project, locale)
		for k := range mm {
			translationKeys[k] = struct{}{}
		}
	}

	sorted := utils.SortedMapKeys(translationKeys)
	regex := buildTranslationKeyRegexFromMap(sorted)
	return translationKeys, regex
}

var newLineReplacer = strings.NewReplacer("\n", "", "\r", "")

// Flatten takes a map and returns a new one where nested maps are replaced
// by dot-delimited keys.
func Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := Flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}

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
	if ep.CategoryTree.Translations != nil {
		c := ep.CategoryTree
		for _, t := range ep.CategoryTree.Translations {
			key := c.Key + "." + t.Key
			if c.Key == "" {
				key = t.Key
			}
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
	for _, c := range ep.CategoryTree.Categories {
		for _, t := range c.Translations {
			key := c.Key + "." + t.Key
			if c.Key == "" {
				key = t.Key
			}
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
		mk, err := FlattenExtendedProject(types.ExtendedProject{CategoryTree: c, Locales: ep.Locales}, locales)
		if err != nil {
			return m, err
		}
		for k, v := range mk {
			m[k] = v
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
		if _, err := runCmd(in.OnReplaceCmd, fPath, strings.NewReader(s)); err != nil {
			return true, err
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
		if info == nil {
			return fmt.Errorf("fileInfo was nil for %s", fPath)
		}
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
func (a Api) Import(projectName string, kind string, locale string, reader io.Reader) error {
	if len(a.cookies) == 0 {
		return fmt.Errorf("Not logged in")
	}
	r, err := http.NewRequest(http.MethodPost, a.endpoint+"/api/import/"+kind+"/"+projectName+"/"+locale, reader)
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
			return fmt.Errorf("import-request returned %d-response: %s (%s) %#v", res.StatusCode, j.Error.Message, j.Error.Code, j.Details)
		}

		return fmt.Errorf("import-request returned %d-response: %s", res.StatusCode, string(body))
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

// This is simply copied from https://github.com/alecthomas/kong-yaml/blob/master/yaml.go
// with the simple change to use toml instead

func TomlLoader(r io.Reader) (kong.Resolver, error) {
	decoder := toml.NewDecoder(r)
	config := map[interface{}]interface{}{}
	err := decoder.Decode(&config)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("TOML config decode error: %w", err)
	}
	return kong.ResolverFunc(func(context *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
		// Build a string path up to this flag.
		path := []string{}
		for n := parent.Node(); n != nil && n.Type != kong.ApplicationNode; n = n.Parent {
			path = append([]string{n.Name}, path...)
		}
		path = append(path, flag.Name)
		path = strings.Split(strings.Join(path, "-"), "-")
		return find(config, path), nil
	}), nil
}

func find(config map[interface{}]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return convertToStringMap(config)
	}
	for i := 0; i < len(path); i++ {
		prefix := strings.Join(path[:i+1], "-")
		if child, ok := config[prefix].(map[interface{}]interface{}); ok {
			return find(child, path[i+1:])
		}
	}
	return config[strings.Join(path, "-")]
}

func convertToStringMap(in map[interface{}]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k.(string)] = v
	}

	return out
}
