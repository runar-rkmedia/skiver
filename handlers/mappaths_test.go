package handlers

import (
	"sort"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/ghodss/yaml"
	"github.com/runar-rkmedia/skiver/internal"
)

// NOTE: the tested function below does not sort the keys, and that does not matter for its functionality.
// However, for the test it is important, to be able to get consistant results

func TestGetMapPath(t *testing.T) {
	tests := []struct {
		name    string
		fields  string
		expects [][]string
		wantErr bool
	}{

		{
			"Simple string",
			"foo",
			[][]string{{"foo"}},
			false,
		},
		{
			"Simple number",
			"123",
			[][]string{{"123"}},
			false,
		},

		{
			"single key value",
			`
bar: foo
`,
			[][]string{{"bar", "foo"}},
			false,
		},

		{
			"threee key value",
			`
bar: foo
bar2: foo2
bar3: foo3
`,
			[][]string{
				{"bar2", "foo2"},
				{"bar3", "foo3"},
				{"bar", "foo"},
			},
			false,
		},

		{
			"single nested",
			`
bar:
  foo: jib
`,
			[][]string{{"bar", "foo", "jib"}},
			false,
		},
		{
			"multiple nested",
			`
bar:
  foo: jib
dob:
  lab: must
gar:
  far: 
    mar: 
      rar: zar 		
      jar: car
      kar: 
        lar: 837384
boing: bang
barg: blipp
bopp:
  bipp: kipp
`,
			[][]string{
				{"barg", "blipp"},
				{"boing", "bang"},
				{"bar", "foo", "jib"},
				{"bopp", "bipp", "kipp"},
				{"dob", "lab", "must"},
				{"gar", "far", "mar", "jar", "car"},
				{"gar", "far", "mar", "rar", "zar"},
				{"gar", "far", "mar", "kar", "lar", "837384"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j interface{}
			err := yaml.Unmarshal([]byte(tt.fields), &j)
			if err != nil {
				t.Errorf("TEST-INPUT_ERROR: Failed to unmarshal: %s %s", err, tt.fields)
				return
			}
			got, err := getMapPaths(j)

			if !tt.wantErr {
				testza.AssertNoError(t, err)
			} else if got == nil {
				t.Error("expected error, but none was returned")
			}

			sort.Slice(got, sortMapPath(got))

			if err := internal.Compare("result", got, tt.expects, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				Yaml:    false,
				JSON:    true,
			}); err != nil {
				t.Log("value: ", j, tt.fields)
				t.Error(err)
			}
		})
	}
}
