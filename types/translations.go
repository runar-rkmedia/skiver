package types

type Locale struct {
	iso2  string
	iso3  string
	title string
	// List of other Locales in preferred order for fallbacks
	fallbacks []string
}

type Translations struct {
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
