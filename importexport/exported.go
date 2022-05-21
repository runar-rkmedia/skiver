package importexport

import (
	"embed"
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
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

type LocaleKeyEnum struct {
	name string
}

func (f LocaleKeyEnum) String() string                   { return f.name }
func (f LocaleKeyEnum) Is(comparitor LocaleKeyEnum) bool { return f.name == comparitor.name }

// Deprecated, only used for earlier enums
func (f LocaleKeyEnum) From(s string) LocaleKeyEnum { return LocaleKeyEnum{s} }

var (
	LocaleKeyEnumIETF = LocaleKeyEnum{string(LocaleKeyIETF)}
	LocaleKeyEnumISO1 = LocaleKeyEnum{string(LocaleKeyISO1)}
	LocaleKeyEnumISO2 = LocaleKeyEnum{string(LocaleKeyISO2)}
	LocaleKeyEnumISO3 = LocaleKeyEnum{string(LocaleKeyISO3)}
)

func addi18nnode(c types.ExtendedCategory, localeID string) (I18N, error) {
	node := I18N{
		Nodes: make(map[string]I18N),
	}
	if c.HasTranslationForLocale(localeID) {
	outer:
		for _, t := range c.Translations {
			if t.Deleted != nil {
				continue outer
			}
		inner:
			for _, tv := range t.Values {
				if tv.Deleted != nil {
					continue inner
				}
				if tv.LocaleID != localeID {
					continue inner
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

	locKeys := utils.SortedMapKeys(locs)
	for _, k := range locKeys {
		l := locs[k]
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
		catKeys := utils.SortedMapKeys(l.Categories)
		for _, k := range catKeys {
			cat := l.Categories[k]
			// if cat.Deleted != nil {
			// 	continue
			// }
			// if !cat.HasTranslationForLocaleDeep(l.ID) {
			// 	continue
			// }
			n, err := addi18nnode(cat, l.ID)

			if err != nil {
				return node, err
			}
			if cat.IsRoot() {
				node.Nodes[key] = n
			} else {
				// keys := cat.Path()
				// cKey := keys[len(keys)-1]
				// node.Nodes[key].Nodes[cat.Key] = n
				kkk := node.Nodes[key]
				kkk.AddNode(n, cat.Path())
				node.Nodes[key] = kkk
				// node.AddNode(n, cat.Path())
			}
		}
	}
	return
}

func (j *I18N) AddNode(node I18N, path []string) {
	if len(path) == 1 {
		j.Nodes[path[len(path)-1]] = node
		return
	}
	if n, ok := j.Nodes[path[0]]; ok {
		n.AddNode(node, path[1:])
		j.Nodes[path[0]] = n

		return
	}
	n := I18N{
		Value: "",
		Nodes: map[string]I18N{},
	}
	n.AddNode(node, path[1:])
	j.Nodes[path[0]] = n
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
	return nil, fmt.Errorf("Expected rootnode, but received: %#v %#v", m, in)
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
