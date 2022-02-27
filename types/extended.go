package types

import (
	"errors"
	"fmt"
)

type ExtendedProject struct {
	Project      `json:"project"`
	Exists       *bool                       `json:"exists,omitempty"`
	Categories   map[string]ExtendedCategory `json:"categories"`
	CategoryTree CategoryTreeNode            `json:"category_tree"`
	Locales      map[string]Locale           `json:"locales"`
}

type ExtendedCategory struct {
	Category
	// TODO: change to map
	Exists       *bool                          `json:"exists,omitempty"`
	Translations map[string]ExtendedTranslation `json:"translations,omitempty"`
}
type ExtendedTranslation struct {
	Translation `json:"translation"`
	Exists      *bool                       `json:"exists,omitempty"`
	Values      map[string]TranslationValue `json:"values"`
}
type ExtendedLocale struct {
	Locale     `json:"locale"`
	Categories map[string]ExtendedCategory `json:"categories"`
}

func (t Translation) Extend(db Storage, options ...ExtendOptions) (et ExtendedTranslation, err error) {
	opts, err := getExtendOptions(options)
	if err != nil {
		return
	}
	et.Translation = t
	if len(et.ValueIDs) == 0 {
		return
	}
	et.Values = map[string]TranslationValue{}
	for _, tid := range et.ValueIDs {
		t, err := db.GetTranslationValue(tid)
		if err != nil {
			return et, err
		}
		if t == nil {
			continue
		}
		key := tid
		if opts.ByKeyLike {
			key = t.LocaleID
		}
		et.Values[key] = *t
	}
	return
}
func (c Category) Extend(db Storage, options ...ExtendOptions) (ec ExtendedCategory, err error) {
	opts, err := getExtendOptions(options)
	if err != nil {
		return
	}
	ec.Category = c
	if len(ec.TranslationIDs) == 0 {
		return
	}
	// for _, sc := range c.SubCategories {
	// 	esc, err := sc.Extend(db, options...)
	// 	if err != nil {
	// 		return ec, err
	// 	}
	// 	ec.SubCategories = append(ec.SubCategories, esc)
	// }
	ec.Translations = map[string]ExtendedTranslation{}
	for _, tid := range ec.TranslationIDs {
		t, err := db.GetTranslation(tid)
		if err != nil {
			return ec, err
		}
		if t == nil {
			continue
		}
		et, err := t.Extend(db, opts)
		if err != nil {
			return ec, err
		}
		key := tid
		if opts.ByKeyLike {
			key = et.Key
		}
		ec.Translations[key] = et
	}
	return
}

type ExtendOptions struct {
	ByID, ByKeyLike  bool
	LocaleFilter     []string
	LocaleFilterFunc func(locale Locale) bool
	ErrOnNoLocales   bool
}

func (o ExtendOptions) Validate() error {
	if o.ByID && o.ByKeyLike {
		return fmt.Errorf("ExtendOptions cannot have both ByID and ByKeyLike")
	}
	if !o.ByID && !o.ByKeyLike {
		return fmt.Errorf("ExtendOptions must have one of ByID or ByKeyLike")
	}
	return nil
}
func getExtendOptions(options []ExtendOptions) (ExtendOptions, error) {
	if len(options) == 0 {
		return ExtendOptions{ByID: true}, nil
	}
	return options[0], options[0].Validate()
}

var (
	ErrNoLocales = errors.New("List of locales was empty")
)

func (p Project) Extend(db Storage, options ...ExtendOptions) (ep ExtendedProject, err error) {
	ep.Project = p
	opts, err := getExtendOptions(options)
	if err != nil {
		return ep, err
	}
	locales, err := db.GetLocales()

	if err != nil {
		return
	}
	if len(opts.LocaleFilter) > 0 {
	outer:
		for k, v := range locales {
			for _, key := range opts.LocaleFilter {
				if key == k || v.IETF == key || v.Iso639_3 == key || v.Iso639_2 == key || v.Iso639_1 == key {
					continue outer
				}
				delete(locales, k)
			}
		}
	}
	if opts.LocaleFilterFunc != nil {
		for k, v := range locales {
			keep := opts.LocaleFilterFunc(v)
			if !keep {
				delete(locales, k)

			}
		}
	}
	if opts.ErrOnNoLocales && len(locales) == 0 {
		err = ErrNoLocales
		return
	}
	ep.Locales = locales
	if len(ep.CategoryIDs) == 0 {
		return
	}
	ep.Categories = map[string]ExtendedCategory{}
	// Sorted by category-path-length
	for _, cid := range ep.CategoryIDs {
		cat, err := db.GetCategory(cid)
		if err != nil {
			return ep, err
		}
		if cat == nil {
			continue
		}
		ec, err := cat.Extend(db, opts)
		if err != nil {
			return ep, err
		}
		key := cid
		if opts.ByKeyLike {
			key = ec.Key
		}
		ep.Categories[key] = ec
	}
	ep.CategoryTree = CreateCategoryTreeNode(ep.Categories)
	return
}

// TODO: add test with "missing" root-node, and missing SubCategories but with existing subsubcategories
func CreateCategoryTreeNode(extendedCategories map[string]ExtendedCategory) CategoryTreeNode {
	node := CategoryTreeNode{}
	catForPathLength := [][]string{}
	for key, ec := range extendedCategories {

		length := len(ec.Path())
		for len(catForPathLength) <= length {

			catForPathLength = append(catForPathLength, []string{})
		}
		catForPathLength[length] = append(catForPathLength[length], key)
	}

	for _, mapKeys := range catForPathLength {
		for _, mapKey := range mapKeys {
			cat := extendedCategories[mapKey]
			path := cat.Path()
			length := len(path)
			// FIXME: make recursive, don't overwrite previous values.
			if length == 0 {
				node = CategoryTreeNode{ExtendedCategory: cat, Categories: make(map[string]CategoryTreeNode)}
				continue
			}
			if length == 1 {
				if node.ID == "" {
					node = CategoryTreeNode{ExtendedCategory: ExtendedCategory{}, Categories: make(map[string]CategoryTreeNode)}
				}
				node.Categories[path[0]] = CategoryTreeNode{ExtendedCategory: cat, Categories: make(map[string]CategoryTreeNode)}
				continue
			}
			if length == 2 {
				node.Categories[path[0]].Categories[path[1]] = CategoryTreeNode{ExtendedCategory: cat, Categories: make(map[string]CategoryTreeNode)}
				continue
			}
			if length == 3 {
				node.Categories[path[0]].Categories[path[1]].Categories[path[2]] = CategoryTreeNode{ExtendedCategory: cat, Categories: make(map[string]CategoryTreeNode)}
				continue
			}
			if length == 4 {
				node.Categories[path[0]].Categories[path[1]].Categories[path[2]].Categories[path[3]] = CategoryTreeNode{ExtendedCategory: cat, Categories: make(map[string]CategoryTreeNode)}
				continue
			}
			if length == 5 {
				node.Categories[path[0]].Categories[path[1]].Categories[path[2]].Categories[path[3]].Categories[path[4]] = CategoryTreeNode{ExtendedCategory: cat, Categories: make(map[string]CategoryTreeNode)}
				continue
			}
			panic(fmt.Sprintf("Out of bounds: %v", path))

		}

	}
	return node
}

type CategoryTreeNode struct {
	ExtendedCategory
	Categories map[string]CategoryTreeNode `json:"categories,omitempty"`
}

func (node CategoryTreeNode) HasTranslationForLocaleDeep(localeID string) bool {
	if node.HasTranslationForLocale(localeID) {
		return true
	}
	for _, v := range node.Categories {
		if v.HasTranslationForLocaleDeep(localeID) {
			return true
		}
	}
	return false
}

func (ep ExtendedProject) ByLocales() (map[string]ExtendedLocale, error) {
	el := map[string]ExtendedLocale{}

	for _, c := range ep.Categories {
		for _, l := range ep.Locales {
			if !c.HasTranslationForLocale(l.ID) {
				continue
			}
			loc, ok := el[l.ID]
			if !ok {
				loc = ExtendedLocale{Locale: l, Categories: map[string]ExtendedCategory{}}
			}

			loc.Categories[c.Key] = c
			el[loc.ID] = loc

		}
	}

	return el, nil
}

func (el ExtendedCategory) HasTranslationForLocale(localeID string) bool {

	for _, t := range el.Translations {
		for _, tv := range t.Values {
			if tv.LocaleID == localeID {
				return true
			}
		}
	}
	return false
}
