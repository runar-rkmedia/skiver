package types

import (
	"errors"
	"fmt"
)

type ExtendedProject struct {
	Project      `json:"project"`
	Exists       *bool                       `json:"exists,omitempty"`
	Categories   map[string]ExtendedCategory `json:"categories" diff:"-"`
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
		if !opts.IncludeDeleted && t.Deleted != nil {
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
		if !opts.IncludeDeleted && t.Deleted != nil {
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
	IncludeDeleted   bool
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
		if !opts.IncludeDeleted && cat.Deleted != nil {
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

func CreateCategoryTreeNode(extendedCategories map[string]ExtendedCategory) CategoryTreeNode {
	// The structure should have a node for each element in the path, so even with
	// just a single path like `foo.bar.baz`, we still create the root-node, the
	// foo-node, the bar-node, and the baz-node
	node := CategoryTreeNode{}
	// A slice which is indexed by the length of the path
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
			// path := append([]string{""}, cat.Path()...)
			path := cat.Path()
			node.add(path, cat)
		}
	}
	return node
}

// Adds the ExtendedCategory to the node at the given path
// Will create nodes as needed.
// The path is expected to be the ExtendedCategory.Path()
func (node *CategoryTreeNode) add(path []string, ec ExtendedCategory) error {
	if node.Categories == nil {
		node.Categories = map[string]CategoryTreeNode{}
	}
	switch len(path) {
	case 0:
		err := node.isEmpty()
		if err != nil {
			return err
		}
		node.ExtendedCategory = ec
		return nil
	case 1:
		err := node.Categories[path[0]].isEmpty()
		if err != nil {
			return err
		}
		node.Categories[path[0]] = CategoryTreeNode{ExtendedCategory: ec}
		return nil
	}
	rest := path[1:]
	next := node.Categories[path[0]]
	next.add(rest, ec)
	node.Categories[path[0]] = next
	return nil
}

func (node CategoryTreeNode) isEmpty() error {
	if node.ExtendedCategory.ID != "" {
		return fmt.Errorf("Node is not empty; it has ID-field")
	}
	if node.ExtendedCategory.Key != "" {
		return fmt.Errorf("Node is not empty; it has Key-field")
	}
	if len(node.Categories) > 0 {
		return fmt.Errorf("Node is not empty; it has categories")
	}
	if len(node.ExtendedCategory.Translations) > 0 {
		return fmt.Errorf("Node is not empty; it has Translations")
	}
	return nil
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
