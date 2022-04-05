package types

import (
	"testing"

	"github.com/runar-rkmedia/skiver/internal"
)

func TestCreateCategoryTreeNode(t *testing.T) {
	tests := []struct {
		name               string
		extendedCategories map[string]ExtendedCategory
		want               CategoryTreeNode
	}{
		{
			"Should work with nested where the previous key does not exist yet",
			map[string]ExtendedCategory{
				"baz.bar":     cat("baz.bar"),
				"foo":         cat("foo"),
				"foo.bar.baz": cat("foo.bar.baz"),
			},
			CategoryTreeNode{
				Categories: map[string]CategoryTreeNode{
					"foo": {ExtendedCategory: cat("foo"), Categories: map[string]CategoryTreeNode{
						"bar": {Categories: map[string]CategoryTreeNode{
							"baz": {ExtendedCategory: cat("foo.bar.baz")},
						}},
					}},
					"baz": {Categories: map[string]CategoryTreeNode{
						"bar": {ExtendedCategory: cat("baz.bar")},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateCategoryTreeNode(tt.extendedCategories)
			err := internal.Compare("got ", got, tt.want)
			if err != nil {
				t.Error(err)
			}
			t.Log(got)
		})
	}
}

func cat(key string) ExtendedCategory {
	c := ExtendedCategory{}
	c.Key = key
	return c
}
