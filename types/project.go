package types

// A Project is a semi-contained entity. Other projects may use translations from other projects,
// if the translations are either referred to directly, or the tags are included within the project.

type Project struct {
	Entity
	Title       string
	Description string
	// If present, any translations with tags matching will also be included in the exported translations
	// If the project contains conflicting translations, the project has presedence.
	IncludedTags []string
}
