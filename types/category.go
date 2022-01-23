package types

import "strings"

// A Category is a "folder" for grouping translation-keys together

// swagger:model Category
type Category struct {
	Entity
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// If the category-key is "___root___", it should be considered as if there are no categories,
	// but just a flat list of items
	Key            string   `json:"key"`
	ProjectID      string   `json:"project_id,omitempty"`
	TranslationIDs []string `json:"translation_ids,omitempty"`
	// A category may have one or more Child-categories.
	SubCategories []Category
}

// Splits the key into multiple keys. Root-values are removed
func (c Category) Keys() []string {
	list := []string{}
	if c.Key == RootCategory {
		return list
	}
	return strings.Split(c.Key, ".")
}

const (
	// RootCategories are accessible without a key, but we do need a key.
	// A bit dirty.
	RootCategory = "___root___"
)
