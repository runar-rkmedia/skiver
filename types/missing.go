package types

// swagger:model MissingTranslation
type MissingTranslation struct {
	Entity
	ProjectID     string `json:"project_id"`
	CategoryID    string `json:"category_id"`
	TranslationID string `json:"translation_id"`
	LocaleID      string `json:"locale_id"`
	// The reported project (may not exist), as reported by the client.
	Project string `json:"project"`
	// The reported category (may not exist), as reported by the client.
	Category string `json:"category"`
	// The reported translation (may not exist), as reported by the client.
	Translation string `json:"translation"`
	// The reported locale (may not exist), as reported by the client.
	Locale string `json:"locale"`
	// Number of times it has been reported.
	Count int `json:"count"`

	// It is probably not important to record every UserAgent, but the first and last is probably useful

	FirstUserAgent  string `json:"first_user_agent"`
	LatestUserAgent string `json:"latest_user_agent"`
}
