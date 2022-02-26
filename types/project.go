package types

import "time"

// A Project is a semi-contained entity. Other projects may use translations from other projects,
// if the translations are either referred to directly, or the tags are included within the project.

// swagger:model Project
type Project struct {
	Entity
	ShortName    string                         `json:"short_name"`
	Title        string                         `json:"title"`
	Description  string                         `json:"description,omitempty"`
	IncludedTags []string                       `json:"included_tags,omitempty"`
	CategoryIDs  []string                       `json:"category_ids,omitempty"`
	LocaleIDs    map[string]LocaleSetting       `json:"locales,omitempty"`
	Snapshots    map[string]ProjectSnapshotMeta `json:"snapshots,omitempty"`
}

type ProjectSnapshotMeta struct {
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
	SnapshotID  string    `json:"id"`
	Hash        uint64    `json:"hash"`
}
type LocaleSetting struct {
	// If set, the locale will be visible for editing.
	Enabled bool `json:"enabled"`
	// If set, the associated translations will be published in releases.
	// This is useful for when adding new locales, and one don't want to publish it to users until it is complete
	Publish bool `json:"publish"`
	// If set, will allow registered translation-services to translate from other languages to this locale.
	// This might help speed up translations for new locales.
	// See the Config or Organization-settings for instructions on how to set up translation-services.
	//
	// * Organization-settings are not yet available.
	//
	// TODO: implement organization-settings
	AutoTranslation bool `json:"auto_translation"`
}
