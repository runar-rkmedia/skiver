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

// Locale represents a language, dialect etc.
// swagger:response
type localeResponse struct {
	// in:body
	Body Locale
}

type Translations struct {
	Entity
	LocaleID         string
	Namespace        string
	Prefix           string
	Key              string
	Title            string
	Description      string
	Aliases          []string
	TagIDS           []string
	Context          string
	Variables        map[string]interface{}
	ExampleVariables map[string]interface{}
}
