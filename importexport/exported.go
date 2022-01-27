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
				in := I18N{Value: tv.Value}
				node.Nodes[t.Key] = in
				if len(tv.Context) > 0 {
					for contextKey, val := range tv.Context {
						node.Nodes[t.Key+"_"+contextKey] = I18N{Value: val}

					}
				}
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
			if v.IsRoot() {
				node.Nodes[key] = n
			} else {
				node.Nodes[key].Nodes[v.Key] = n
			}
		}
	}
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

type I18N struct {
	Value string          `json:"value,omitempty"`
	Nodes map[string]I18N `json:"nodes,omitempty"`
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
		val := v.ToMap()
		if val == "" {
			fmt.Println("kkk", k, v)
			continue
		}
		m[k] = val
	}

	return m

}
func (in I18NWithLocales) Merge(v I18NWithLocales) I18NWithLocales {
	for k, v := range v.Value {
		in.Value[k] = v
	}
	for k, v := range v.Nodes {
		if ex, ok := in.Nodes[k]; ok {
			in.Nodes[k] = ex.Merge(v)
		} else {
			in.Nodes[k] = v
		}
	}
	return in
}
func (in I18N) ToLocaleAwere(locale string) I18NWithLocales {
	node := I18NWithLocales{Value: map[string]string{}, Nodes: map[string]I18NWithLocales{}}

	if in.Value != "" {
		node.Value[locale] = in.Value
	}

	for k, v := range in.Nodes {
		node.Nodes[k] = v.ToLocaleAwere(locale)
	}
	return node
}

// Aweseome name
type LocaleLookerUpper interface {
	GetLocaleID(string) string
}

func (in I18N) MergeAsIfRootIsLocale(localeGetter LocaleLookerUpper) (I18NWithLocales, error) {
	node := I18NWithLocales{Nodes: map[string]I18NWithLocales{}}
	if in.Value != "" {
		return node, fmt.Errorf("did not expect values in root-node: %#v", node)
	}
	if len(in.Nodes) == 0 {
		return node, nil
	}

	for localeKey, localeNode := range in.Nodes {
		localeID := localeKey
		if localeGetter != nil {
			localeID = localeGetter.GetLocaleID(localeKey)
			if localeID == "" {
				return node, fmt.Errorf("Failed to look up locale-id for locale-key: %s", localeKey)
			}
		}

		for k, v := range localeNode.Nodes {

			newNode := v.ToLocaleAwere(localeID)
			if ex, ok := node.Nodes[k]; ok {
				// TODO: Merge
				node.Nodes[k] = ex.Merge(newNode)
				// return node, fmt.Errorf("Must merge \n%#v \n%#v", ex, v)
			} else {
				node.Nodes[k] = newNode
			}
		}
	}
	return node, nil
}

type I18NWithLocales struct {
	// The keys should be the ID of the locale
	Value map[string]string          `json:"value,omitempty"`
	Nodes map[string]I18NWithLocales `json:"nodes,omitempty"`
}
