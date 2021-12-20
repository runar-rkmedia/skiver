package types

// TODO: expand to support more levels
func ExportI18N(project Project, translations []Translations, locales []Locale) (i18n I18N) {
	tl := len(translations)
	for li := 0; li < len(locales); li++ {
		var t = map[string]map[string]string{}
		for ti := 0; ti < tl; ti++ {
			if translations[ti].LocaleID == locales[li].ID {
				if t[translations[ti].Prefix] == nil {
					t[translations[ti].Prefix] = map[string]string{}
				}
				key := translations[ti].Key
				if translations[ti].Context != "" {
					key += "_" + translations[ti].Context
				}
				t[translations[ti].Prefix][key] = translations[ti].Value

			}

		}
	}

	return nil
}

type I18N map[string]interface{}
