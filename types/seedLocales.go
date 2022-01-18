package types

import (
	"fmt"
)

type localeSeeder interface {
	CreateLocale(locale Locale) (Locale, error)
	GetLocaleFilter(filter ...Locale) (*Locale, error)
}

// We should use this list: https://www.iana.org/assignments/language-subtag-registry/language-subtag-registry
// See also https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl

func SeedLocales(db localeSeeder, organizationID string, locales []Locale) error {
	if len(locales) == 0 {
		// TODO: build this data from som external api/resours
		// For now, it is just a tiny sample of what I persionally am going to need now.
		locales = []Locale{
			{
				IETF:      "en-GB",
				Iso639_1:  "en",
				Iso639_2:  "eng",
				Iso639_3:  "eng",
				Title:     "British",
				Fallbacks: []string{"eng", "en"},
			},
			{
				IETF:      "en-US",
				Iso639_1:  "en",
				Iso639_2:  "eng",
				Iso639_3:  "eng",
				Title:     "US English",
				Fallbacks: []string{"eng", "en"},
			},
			{
				IETF:      "nb-NO",
				Iso639_1:  "nb",
				Iso639_2:  "nob",
				Iso639_3:  "nob",
				Title:     "Norwegian bokmål",
				Fallbacks: []string{"nn-NO", "no", "dan", "swe", "eng"},
			},
			{
				IETF:      "nn-NO",
				Iso639_1:  "nn",
				Iso639_2:  "nno",
				Iso639_3:  "nno",
				Title:     "Norwegian Nynorsk",
				Fallbacks: []string{"nb-NO", "no", "dan", "swe", "eng"},
			},
		}
	}
	if v, err := db.GetLocaleFilter(locales[0]); err != nil {
		return fmt.Errorf("error occured while checking for locale: %w", err)
	} else if v != nil {
		return nil
	}
	for i := 0; i < len(locales); i++ {
		if locales[i].CreatedBy == "" {
			locales[i].CreatedBy = "seeder"
		}
		if locales[i].OrganizationID == "" {
			locales[i].OrganizationID = organizationID
		}
		if _, err := db.CreateLocale(locales[i]); err != nil {
			return fmt.Errorf("failed to create locale %s: %w", locales[i].IETF, err)
		}
	}
	return nil
}
