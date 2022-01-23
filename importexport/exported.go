package importexport

import (
	"embed"
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
)

var (
	//go:embed templates
	Content embed.FS
)

type ExportI18NOptions struct {
	// Must be a key of locale
	LocaleKey    LocaleKey
	LocaleFilter []string
}

type LocaleKey string

const (
	LocaleKeyIETF LocaleKey = "ietf"
	LocaleKeyISO1 LocaleKey = "iso1"
	LocaleKeyISO2 LocaleKey = "iso2"
	LocaleKeyISO3 LocaleKey = "iso3"
)

func addi18nnode(c types.ExtendedCategory, localeID string) (I18N, error) {
	node := I18N{}
	if len(c.SubCategories) > 0 {
		for _, sc := range c.SubCategories {
			if !sc.HasTranslationForLocaleDeep(localeID) {
				continue
			}
			nn, err := addi18nnode(sc, localeID)
			if err != nil {
				return node, err
			}
			if node.Nodes == nil {
				node.Nodes = make(map[string]I18N)
			}
			node.Nodes[sc.Key] = nn
		}
	}
	if c.HasTranslationForLocale(localeID) {
		for _, t := range c.Translations {
			for _, tv := range t.Values {
				if tv.LocaleID != localeID {
					continue
				}
				if node.Nodes == nil {
					node.Nodes = make(map[string]I18N)
				}
				node.Nodes[t.Key] = I18N{Value: tv.Value}
			}
		}
	}

	return node, nil
}

func ExportI18N(ex types.ExtendedProject, options ExportI18NOptions) (node I18N, err error) {
	locs, err := ex.ByLocales()
	if err != nil {
		return node, err
	}

	for _, l := range locs {
		key := getLocaleKey(options.LocaleKey, l.Locale)
		if key == "" {
			err = fmt.Errorf("the locale-key was empty")
			return
		}
		if len(options.LocaleFilter) > 0 {
			found := false
		innerb:
			for _, v := range options.LocaleFilter {
				if key == v {
					found = true
					break innerb
				}
			}
			if !found {
				continue
			}
		}
		if node.Nodes == nil {
			node.Nodes = make(map[string]I18N)
		}
		node.Nodes[key] = I18N{
			Nodes: make(map[string]I18N),
		}
		for _, v := range l.Categories {
			if !v.HasTranslationForLocaleDeep(l.ID) {
				continue
			}
			n, err := addi18nnode(v, l.ID)
			if err != nil {
				return node, err
			}
			node.Nodes[key].Nodes[v.Key] = n

		}
	}

	// if 1 == 1 {
	// 	return i18n, err
	// }
	// if options.LocaleKey == "" {
	// 	options.LocaleKey = LocaleKeyISO1
	// }
	// for _, l := range ex.Locales {
	// 	key := getLocaleKey(options.LocaleKey, l)
	// 	if key == "" {
	// 		err = fmt.Errorf("the locale-key was empty")
	// 		return
	// 	}
	// 	if len(options.LocaleFilter) > 0 {
	// 		found := false
	// 	inner:
	// 		for _, v := range options.LocaleFilter {
	// 			if key == v {
	// 				found = true
	// 				break inner
	// 			}
	// 		}
	// 		if !found {
	// 			continue
	// 		}
	// 	}
	// 	i18n[key] = map[string]map[string]string{}
	// }
	// for _, c := range ex.Categories {
	// 	cKey := strings.TrimSpace(c.Key)
	// 	for _, l := range ex.Locales {
	// 		key := getLocaleKey(options.LocaleKey, l)
	// 		if i18n[key] == nil {
	// 			continue
	// 		}
	// 		i18n[key][cKey] = map[string]string{}
	// 	}
	// 	for _, t := range c.Translations {
	// 		tKey := t.Key
	// 		for _, v := range t.Values {
	// 			if v.Value == "" {
	// 				continue
	// 			}
	// 			locale, ok := ex.Locales[v.LocaleID]
	// 			if !ok {
	// 				err = fmt.Errorf("the locale '%s' was not found in the export, %#v", v.LocaleID, ex.Locales)
	// 				return
	// 			}
	// 			localeKey := getLocaleKey(options.LocaleKey, locale)
	// 			if i18n[localeKey] == nil {
	// 				continue
	// 			}

	// 			i18n[localeKey][cKey][tKey] = v.Value
	// 			if v.Context != nil {
	// 				for context, v := range v.Context {
	// 					tKey = tKey + "_" + context
	// 					i18n[localeKey][cKey][tKey] = v
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	return
}

func getLocaleKey(key LocaleKey, locale types.Locale) string {
	switch key {
	case LocaleKeyIETF:
		return locale.IETF
	case LocaleKeyISO1:
		return locale.Iso639_1
	case LocaleKeyISO2:
		return locale.Iso639_2
	case LocaleKeyISO3:
		return locale.Iso639_3
	}
	return locale.Iso639_1
}

// type I18N map[string]map[string]map[string]string

type I18N struct {
	Value string
	Nodes map[string]I18N
}

func I18NNodeToI18Next(in I18N) (map[string]interface{}, error) {

	m := in.ToMap()
	if mm, ok := m.(map[string]interface{}); ok {
		return mm, nil
	}
	return nil, fmt.Errorf("Expected rootrode, but received: %#v", m)
}

func (in I18N) ToMap() interface{} {

	if in.Value != "" {
		return in.Value
	}
	if len(in.Nodes) == 0 {
		return ""
	}
	m := map[string]interface{}{}

	for k, v := range in.Nodes {
		m[k] = v.ToMap()
	}

	return m

}
