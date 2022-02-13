package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/ghodss/yaml"
	"github.com/go-test/deep"
	"github.com/gookit/color"
	"github.com/pelletier/go-toml"
)

var (
	DefaultCompareOptions = CompareOptions{true, true, true, false, false, false}
	// red                   = color.FgRed.Render
	cName = color.Green.Render
	cGot  = color.Yellow.Render
	cDiff = color.Red.Render
	cWant = color.Cyan.Render
)

func Compare(name string, got, want interface{}, options ...CompareOptions) error {

	opts := DefaultCompareOptions
	if len(options) > 0 {
		opts = options[0]
	}

	// At least on of these must be applied
	if !opts.Reflect && !opts.Diff {
		panic("Invlid options")
	}

	if opts.TOML || opts.JSON {
		opts.Yaml = false
	}
	if opts.Diff {
		if diff := deep.Equal(got, want); diff != nil {
			var g interface{} = got
			var w interface{} = want
			var d interface{}
			if opts.IsString {
				gotS, ok := got.(string)
				if !ok {
					return fmt.Errorf("failed to diff 'got', as input was not of type string, as defined via options")
				}
				wantS, ok := want.(string)
				if !ok {
					return fmt.Errorf("failed to diff 'want', as input was not of type string, as defined via options")
				}
				if strings.HasPrefix(gotS, "{") {
					err := json.Unmarshal([]byte(gotS), &g)
					if err != nil {
						return err
					}
					err = json.Unmarshal([]byte(wantS), &w)
					if err != nil {
						return err
					}
					if opts.Yaml {
						g = MustYaml(g)
						w = MustYaml(w)
						// d = MustYaml(diff)
					}
					if opts.JSON {
						g = MustJSON(g)
						w = MustJSON(w)
						// d = MustJSON(diff)
					}
					// Toml looks weird with some inputs
					if opts.TOML {
						g = MustToml(g)
						w = MustToml(w)
					}
					diff = deep.Equal(g, w)

				}

			} else {

				if opts.Yaml {
					g = MustYaml(got)
					w = MustYaml(want)
					// d = MustYaml(diff)
				}
				if opts.JSON {
					g = MustJSON(got)
					w = MustJSON(want)
					// d = MustJSON(diff)
				}
				// Toml looks weird with some inputs
				if opts.TOML {
					g = MustToml(got)
					w = MustToml(want)
				}
			}

			if false {

				return fmt.Errorf("%s", lineDiff(g, w))
			}
			d = MustYaml(diff)
			f := "%[1]s: \n%[2]v\n%[3]s\n\ndiff:\n%[4]s\nwant:\n%[5]v"
			return fmt.Errorf(f, cName(name), cGot(withLineNumbers(g)), lineDiff(g, w), cDiff(d), cWant(withLineNumbers(w)))
		}
	}
	if opts.Reflect {
		if !reflect.DeepEqual(got, want) {
			var g interface{} = got
			var w interface{} = want
			if opts.Yaml {
				g = MustYaml(got)
				w = MustYaml(want)
			}
			if opts.JSON {
				g = MustJSON(got)
				w = MustYaml(want)
			}
			if opts.Yaml {
				g = MustToml(got)
				w = MustToml(want)
			}
			return fmt.Errorf("%s: \n%v\nwant:\n%v", cName(name), cGot(g), cWant(w))
		}
	}
	return nil
}

type CompareOptions struct {
	// Produces a diff of the result, but may in some edgecases not detect all errors (like differences in timezones)
	Diff,
	// Uses a traditional reflect.DeepEqual to perform tests.
	Reflect,
	// Produces output in yaml for readability
	Yaml bool
	JSON     bool
	TOML     bool
	IsString bool
}

func MustYaml(j interface{}) string {
	b, err := yaml.Marshal(j)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal: %w\n\n%v", err, j))
	}
	return string(b)
}
func MustJSON(j interface{}) string {
	b, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		panic(fmt.Errorf("Failed to marshal: %w\n\n%v", err, j))
	}
	return string(b)
}
func MustToml(j interface{}) string {
	s := MustJSON(j)
	// var k interface{}
	var k map[string]interface{}
	err := json.Unmarshal([]byte(s), &k)
	if err != nil {
		return s
	}
	b, err := toml.Marshal(k)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal: %w\n\n%v", err, s))
	}
	return string(b)
}

func withLineNumbers(si interface{}) string {
	s, ok := si.(string)
	if !ok {
		return "Failed to add line-numbers"
	}

	split := strings.Split(s, "\n")
	str := ""
	for i := 0; i < len(split); i++ {
		if split[i] == "" && i == len(split)-1 {
			continue
		}
		str += fmt.Sprintf("%02d: %s\n", i+1, split[i])
	}
	return str
}

func MatchSnapshot(t *testing.T, extension string, b interface{}) {
	t.Helper()
	var v []byte
	switch val := b.(type) {
	case []byte:
		v = val
	case string:
		v = []byte(v)
	}
	switch {
	case strings.HasSuffix(extension, "json"):
		bb, err := json.MarshalIndent(b, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		v = bb
	case strings.HasSuffix(extension, "yaml"):
		bb, err := yaml.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		v = bb
	case strings.HasSuffix(extension, "toml"):
		bb, err := toml.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		v = bb
	}
	matchSnapshot(t, extension, v, false)
}
func OverwriteSnapshot(t *testing.T, extension string, b []byte) {
	matchSnapshot(t, extension, b, true)
}
func matchSnapshot(t *testing.T, extension string, b []byte, overWrite bool) {
	t.Helper()
	if overWrite {
		ci, err := os.LookupEnv("CI")
		if err {
			t.Fatalf("This looks like a CI-environment, and the developer forgot to disable overwriting of snapshots, which would make the test pointless. the CI-env-var is: %s", ci)
			return
		}
	}
	filePath := path.Join("testdata", "snapshots", strings.ReplaceAll(t.Name(), "/", "-")+"."+extension)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		t.Fatalf("failed during filepath.Abs %s: %s", filePath, err)
		return
	}
	file, err := os.Open(filePath)
	if overWrite || errors.Is(err, os.ErrNotExist) {
		err := ioutil.WriteFile(filePath, b, 0677)
		if err != nil {
			t.Fatalf("Failed writing snapshot to %s", absPath)
		}
		t.Fatalf("Wrote snapshot to %s", absPath)
		return
	} else if err != nil {
		t.Fatalf("failed to open file %s: %s", filePath, err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read the file %s: %s", filePath, err)
	}
	want := strings.TrimSpace(string(data))
	got := strings.TrimSpace(string(b))
	if want == got {
		return
	}

	smallest := len(got)
	if len(want) < smallest {
		smallest = len(want)
	}
	for i := 0; i < smallest; i++ {
		g := got[i]
		w := want[i]
		if g != w {
			linenumber := strings.Count(got[:i], "\n")
			t.Errorf("First mismatch at line %d index %d %s", linenumber, i, got[:i])
			break

		}

	}

	if strings.HasPrefix(got, "{") {

		if err := Compare(t.Name()+"("+extension+")", got, want, CompareOptions{Yaml: true, IsString: false, Reflect: false, Diff: true}); err != nil {
			t.Error(err)
			return
		}
	}
	t.Error(StringCompare(t, want, got))
}

func lineDiff(wanti, goti interface{}) string {
	got := goti.(string)
	want := wanti.(string)
	if got == want {
		return ""
	}
	return LineDiff(got, want)
}

func LineDiff(want, got string) string {
	if got == want {
		return ""
	}
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	lenGot := len(gotLines)
	lenWant := len(wantLines)
	smallest := lenGot
	longest := lenGot
	if lenWant < smallest {
		smallest = lenWant
	} else {
		longest = lenWant
	}

	var errs []string
	for lineNumber := 0; lineNumber < longest; lineNumber++ {
		g := ""
		w := ""
		if lenGot > lineNumber {
			g = gotLines[lineNumber]
		}
		if lenWant > lineNumber {
			w = wantLines[lineNumber]
		}
		ln := fmt.Sprintf("%2d: ", lineNumber+1)
		if g == w {
			errs = append(errs, color.Gray.Render(ln+g))
			continue
		}
		errs = append(errs, fmt.Sprintf("%s%s", cDiff(ln), cGot(g)))
		errs = append(errs, fmt.Sprintf("%s%s", cDiff(ln), cWant(w)))
	}
	if lenWant != lenGot {
		errs = append(errs, fmt.Sprintf("(...Linenumbers differ %d vs %d)", lenGot-1, lenWant-1))
	}
	if len(errs) > 0 {
		return strings.Join(errs, "\n")
	}
	return fmt.Sprintf("%s | %s", got, want)
}

func StringCompare(t *testing.T, want, got string) error {

	if want != got {
		testza.AssertEqual(t, want, got)
		// dmp := diffmatchpatch.New()
		// fileAdmp, fileBdmp, dmpStrings := dmp.DiffLinesToChars(want, got)
		// diffs := dmp.DiffMain(fileAdmp, fileBdmp, false)
		// diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
		// diffs = dmp.DiffCleanupSemantic(diffs)

		// t.Errorf("Result was not as expected: %s", diffs)
		f := "Result not as expected, want \n%[1]s:\n%[2]v\ngot:\n%[3]s"
		if true {
			f = "Result not as expected, want \n%[1]s:got: \n%[3]v\ndiff:\n%[2]s"
			// f = "Result not as expected, diff \n%[2]s:got: \n%[3]v\nwant:\n%[1]s"
		}
		return fmt.Errorf(f, cWant(withLineNumbers(want)), LineDiff(want, got), cGot(withLineNumbers(got)))
	}
	return nil
}

// Tabs are annoying in yaml, so lets just convert it.
func YamlUnmarshalAllowTabs(s string, j interface{}) error {
	s = strings.ReplaceAll(s, "\t", "  ")
	return yaml.Unmarshal([]byte(s), j)
}

func PrintMultiLineYaml(title, content interface{}) {
	_, file, no, ok := runtime.Caller(1)
	if ok {
		if _, filec, ok := strings.Cut(file, "/skiver/"); ok {
			file = filec
		}
		fmt.Printf("\n%s:%s ", color.Gray.Render(file), color.Blue.Render(no))
	}
	fmt.Printf(AsMultiLineYaml(title, content))
}
func TPrintMultiLineYaml(t *testing.T, title, content interface{}) {
	t.Helper()
	t.Log(AsMultiLineYaml(title, content))
}
func AsMultiLineYaml(title, content interface{}) string {
	return fmt.Sprintf("%s (%T)\n%s", color.Yellow.Render(title), content, color.Cyan.Render(withLineNumbers(MustYaml(content))))
}
