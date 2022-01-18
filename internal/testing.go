package internal

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ghodss/yaml"
	"github.com/go-test/deep"
	"github.com/gookit/color"
	"github.com/pelletier/go-toml"
)

var (
	DefaultCompareOptions = CompareOptions{true, true, true, false, false}
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
			// Toml looks great on diffs!
			d := MustToml(diff)
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
	JSON bool
	TOML bool
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
