package handlers

import (
	"testing"

	"github.com/runar-rkmedia/skiver/internal"
)

// NOTE: the tested function below does not sort the keys, and that does not matter for its functionality.
// However, for the test it is important, to be able to get consistant results

func TestNode(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		expects Node
	}{

		{
			"Simple string",
			[]string{"foo"},
			Node{Value: "foo"},
		},
		{
			"Should return value and root",
			[]string{"bar", "foo"},
			Node{Value: "foo", Root: "bar"},
		},
		{
			"Should return value,root, with midpath",
			[]string{"baz", "bar", "foo"},
			Node{Value: "foo", Root: "baz", MidPath: "bar"},
		},
		{
			"Should return value,root, with midpath, where root gets dot-path",
			[]string{"jar", "jib", "baz", "bar", "foo"},
			Node{Value: "foo", Root: "jar.jib.baz", MidPath: "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNode(tt.fields)

			if err := internal.Compare("result", got, tt.expects, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				Yaml:    false,
				JSON:    true,
			}); err != nil {
				t.Log("input", tt.fields)
				t.Error(err)
			}
		})
	}
}
