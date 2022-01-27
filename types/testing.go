package types

type Locales []Locale

func (t Locales) GetLocaleID(key string) string {
	for _, l := range t {
		if l.ID == key {
			return l.ID
		}
		if l.IETF == key {
			return l.ID
		}
		if l.Iso639_3 == key {
			return l.ID
		}
		if l.Iso639_2 == key {
			return l.ID
		}
		if l.Iso639_1 == key {
			return l.ID
		}
	}

	return ""
}

var (
	Test_locales = Locales{
		{
			Entity:   Entity{ID: "loc-en"},
			Iso639_1: "en",
			Iso639_2: "en",
			Iso639_3: "eng",
			IETF:     "en-US",
		},
		{
			Entity:   Entity{ID: "loc-no"},
			Iso639_1: "no",
			Iso639_2: "no",
			Iso639_3: "nor",
			IETF:     "nb-NO",
		},
	}
)
