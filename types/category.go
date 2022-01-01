package types

// A Category is a "folder" for grouping translation-keys together

// swagger:model Category
type Category struct {
	Entity
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Key            string   `json:"key"`
	ProjectID      string   `json:"project_id"`
	TranslationIDs []string `json:"translation_ids,omitempty"`
}
