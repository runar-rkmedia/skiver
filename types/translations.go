package types

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

// swagger:model Translation
type Translation struct {
	Entity
	Aliases             []string               `json:"aliases"`
	ParentTranslationID string                 `json:"parent_translation,omitempty"`
	Description         string                 `json:"description,omitempty"`
	Key                 string                 `json:"key,omitempty"`
	CategoryID          string                 `json:"category,omitempty"`
	Tag                 []string               `json:"tags,omitempty"`
	Title               string                 `json:"title,omitempty"`
	Variables           map[string]interface{} `json:"variables,omitempty"`
	ValueIDs            []string               `json:"value_ids,omitempty"`
}

// swagger:model TranslationValue
type TranslationValue struct {
	Entity
	// The pre-interpolated value to use  with translations
	// Example: The {{productName}} fires up to {{count}} bullets of {{subject}}.
	Value string `json:"value,omitempty"`
	// locale ID
	LocaleID string `json:"locale_id,omitempty"`
	// Translation ID
	TranslationID string `json:"translation_id,omitempty"`
	// Indicating from where the value was created from, usually user, but could be a tranlator-service, like Bing.
	Source  CreatorSource     `json:"source,omitempty"`
	Context map[string]string `json:"context,omitempty"`
}

func (e TranslationValue) Namespace() string {
	return e.Kind()
}
func (e TranslationValue) Kind() string {
	return string(PubTypeTranslationValue)
}

type CreatorSource string

var (
	CreatorSourceUser       CreatorSource = "user"
	CreatorSourceTranslator CreatorSource = "system-translator"
	CreatorSourceImport     CreatorSource = "user-import"
)

// Locale represents a language, dialect etc.
// swagger:response
type localeResponse struct {
	// in:body
	Body Locale
}

// TranslationValue is the direct translated value of a TranslationKey for a Locale
// swagger:response
type translationValueResponse struct {
	// in:body
	Body TranslationValue
}
