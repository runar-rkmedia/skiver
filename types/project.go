package types

// A Project is a semi-contained entity. Other projects may use translations from other projects,
// if the translations are either referred to directly, or the tags are included within the project.

// swagger:model Project
type Project struct {
	Entity
	ShortName    string   `json:"short_name"`
	Title        string   `json:"title"`
	Description  string   `json:"description,omitempty"`
	IncludedTags []string `json:"included_tags,omitempty"`
	CategoryIDs  []string `json:"category_ids,omitempty"`
}
