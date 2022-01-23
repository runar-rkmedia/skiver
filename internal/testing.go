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
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/andreyvit/diff"
	"github.com/ghodss/yaml"
	"github.com/go-test/deep"
	"github.com/gookit/color"
	"github.com/pelletier/go-toml"
	"github.com/sergi/go-diff/diffmatchpatch"
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
			d = MustToml(diff)
			// Toml looks great on diffs!
			return fmt.Errorf("YAML: %s: \n%v\ndiff:\n%s\nwant:\n%v", cName(name), cGot(g), cDiff(d), cWant(w))
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
			return fmt.Errorf("YAML: %s: \n%v\nwant:\n%v", cName(name), cGot(g), cWant(w))
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

func MatchSnapshot(t *testing.T, extension string, b []byte) {
	matchSnapshot(t, extension, b, false)
}
func OverwriteSnapshot(t *testing.T, extension string, b []byte) {
	matchSnapshot(t, extension, b, true)
}
func matchSnapshot(t *testing.T, extension string, b []byte, overWrite bool) {
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

func StringCompare(t *testing.T, want, got string) error {

	if want != got {
		testza.AssertEqual(t, want, got)
		dmp := diffmatchpatch.New()
		fileAdmp, fileBdmp, dmpStrings := dmp.DiffLinesToChars(want, got)
		diffs := dmp.DiffMain(fileAdmp, fileBdmp, false)
		diffs = dmp.DiffCharsToLines(diffs, dmpStrings)
		diffs = dmp.DiffCleanupSemantic(diffs)

		// t.Errorf("Result was not as expected: %s", diffs)
		return fmt.Errorf("Result not as expected:\n%v", diff.LineDiff(want, got))
	}
	return nil
}
