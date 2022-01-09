package types

// A Category is a "folder" for grouping translation-keys together

// swagger:model Category
type Category struct {
	Entity
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Key            string   `json:"key"`
	ProjectID      string   `json:"project_id,omitempty"`
	TranslationIDs []string `json:"translation_ids,omitempty"`
}

const (
	// RootCategories are accessible without a key, but we do need a key.
	// A bit dirty.
	RootCategory = "___root___"
)
