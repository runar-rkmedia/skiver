package types

// A Project is a semi-contained entity. Other projects may use translations from other projects,
// if the translations are either referred to directly, or the tags are included within the project.

// swagger:model Project
type Project struct {
	Entity
	ProjectInput
}

type ProjectInput struct {
	// Required: true
	// Max Length: 400
	// Min Length: 2
	// example: My Great Project
	Title string `json:"title"`
	// Max Length: 8000
	// example: Project-description
	Description string `json:"description"`
	// If present, any translations with tags matching will also be included in the exported translations
	// If the project contains conflicting translations, the project has presedence.
	// example: ["actions", "general"]
	IncludedTags []string `json:"included_tags"`
}

// swagger:parameters createProject
type projectInput struct {

	// required:true
	// in:body
	Body ProjectInput
}
