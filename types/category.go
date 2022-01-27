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
	SubCategories []Category `json:"sub_categories,omitempty"`
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

func (c Category) IsRoot() bool {
	return c.Key == RootCategory
}
func (e Category) Namespace() string {
	return e.Kind()
}
func (e Category) Kind() string {
	return string(PubTypeCategory)
}

// Used to filter and search along with Category.Filter(CategoryFilter)
type CategoryFilter struct {
	OrganizationID string
	Key            string
	ID             string
	SubCategory    []CategoryFilter
}

// Used to filter and search
func (cat Category) Filter(f CategoryFilter) bool {
	if f.OrganizationID != "" && f.OrganizationID != cat.OrganizationID {
		return false
	}
	if f.Key != "" && f.Key != cat.Key {
		return false
	}
	if f.ID != "" && f.ID != cat.ID {
		return false
	}
	if len(f.SubCategory) != 0 {
		for _, sf := range f.SubCategory {
			match := cat.Filter(sf)
			if match {
				return true
			}
		}
	}
	return true
}

func (cat Category) AsUniqueFilter() CategoryFilter {
	return CategoryFilter{
		OrganizationID: cat.OrganizationID,
		Key:            cat.Key,
		ID:             cat.ID,
		SubCategory:    []CategoryFilter{},
	}
}
