package types

// # See https://en.wikipedia.org/wiki/Language_code for more information
type Locale struct {
	Entity
	// Represents the ISO-639-1 string, e.g. en
	Iso639_1 string
	// Represents the ISO-639-2 string, e.g. eng
	Iso639_2 string
	// Represents the ISO-639-3 string, e.g. eng
	Iso639_3 string
	// Represents the IETF language tag, e.g. en / en-US
	IETF  string
	Title string
	// List of other Locales in preferred order for fallbacks
	Fallbacks []string
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
