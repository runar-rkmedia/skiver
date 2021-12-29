package types

import "github.com/runar-rkmedia/skiver/models"

// # See https://en.wikipedia.org/wiki/Language_code for more information
// TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
type Locale struct {
	Entity
	// Represents the ISO-639-1 string, e.g. en
	Iso639_1 string `json:"iso_639_1"`
	// Represents the ISO-639-2 string, e.g. eng
	Iso639_2 string `json:"iso_639_2"`
	// Represents the ISO-639-3 string, e.g. eng
	Iso639_3 string `json:"iso_639_3"`
	// Represents the IETF language tag, e.g. en / en-US
	IETF  string `json:"ietf"`
	Title string `json:"title"`
	// List of other Locales in preferred order for fallbacks
	Fallbacks []string `json:"fallbacks,omitempty"`
}

// Locale represents a language, dialect etc.
// swagger:response
type localeResponse struct {
	// in:body
	Body Locale
}

// swagger:model Translation
type Translation struct {
	Entity
	// aliases
	Aliases []string `json:"aliases"`

	// Used as a variation for the key
	Context string `json:"context,omitempty"`

	// Description for the key, its use and where the key is used.
	Description string `json:"description,omitempty"`

	// Final part of the identifiying key.
	// With the example-input, the complete generated key would be store.product.description
	// Example: description
	Key string `json:"key,omitempty"`

	// locale ID
	LocaleID string `json:"locale_id,omitempty"`

	// Can be a dot-separated path-like string
	// Example: store.products
	Prefix string `json:"prefix,omitempty"`

	// project ID
	ProjectID string `json:"project,omitempty"`

	// tag
	Tag []string `json:"tags"`

	// Title with short description of the key
	Title string `json:"title,omitempty"`

	// The pre-interpolated value to use  with translations
	// Example: The {{productName}} fires up to {{count}} bullets of {{subject}}.
	Value string `json:"value,omitempty"`

	// Variables used within the translation.
	// This helps with giving translators more context,
	// The value for the translation will be used in examples.
	// Example: {"count":3,"productName":"X-Buster","subject":"compressed solar energy"}
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// swagger:parameters createTranslation
type translationInput struct {

	// required:true
	// in:body
	Body models.TranslationInput
}
