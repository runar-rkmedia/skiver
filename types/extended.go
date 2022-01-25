package types

import (
	"fmt"
)

type ExtendedProject struct {
	Project
	Exists     *bool `json:"exists,omitempty"`
	Categories map[string]ExtendedCategory
	Locales    map[string]Locale
}

type ExtendedCategory struct {
	Category
	SubCategories []ExtendedCategory `json:"sub_categories,omitempty"`
	Exists        *bool              `json:"exists,omitempty"`
	Translations  map[string]ExtendedTranslation
}
type ExtendedTranslation struct {
	Translation
	Exists *bool `json:"exists,omitempty"`
	Values map[string]TranslationValue
}
type ExtendedLocale struct {
	Locale
	Categories map[string]ExtendedCategory
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
	ByID, ByKeyLike bool
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
	ep.Locales = locales
	if len(ep.CategoryIDs) == 0 {
		return
	}
	ep.Categories = map[string]ExtendedCategory{}
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
	return
}

func (ep ExtendedProject) ByLocales() (map[string]ExtendedLocale, error) {
	el := map[string]ExtendedLocale{}

	// TODO: needs improvement... It should only add leafnodes that match the language
	for _, c := range ep.Categories {
		for _, l := range ep.Locales {
			if c.HasTranslationForLocaleDeep(l.ID) {
				loc, ok := el[l.ID]
				if !ok {
					loc = ExtendedLocale{Locale: l, Categories: map[string]ExtendedCategory{}}
				}
				loc.Categories[c.ID] = c
				el[loc.ID] = loc
			}

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
func (el ExtendedCategory) HasTranslationForLocaleDeep(localeID string) bool {
	if el.HasTranslationForLocale(localeID) {
		return true
	}
	for _, c := range el.SubCategories {
		if c.HasTranslationForLocaleDeep(localeID) {
			return true
		}
	}
	return false
}

// TOOD: this just replaces earlier categories with the leaf-category.
func (ep ExtendedProject) traverseCategory(c ExtendedCategory, el map[string]ExtendedLocale) error {

	for _, t := range c.Translations {
		for _, tv := range t.Values {
			loc, ok := el[tv.LocaleID]
			if !ok {
				if l, ok := ep.Locales[tv.LocaleID]; !ok {

					return fmt.Errorf("Could not find locale for id %s", tv.LocaleID)
				} else {
					loc = ExtendedLocale{Locale: l, Categories: map[string]ExtendedCategory{}}
				}
			}
			loc.Categories[c.ID] = c
			el[tv.LocaleID] = loc
		}
	}
	for _, cc := range c.SubCategories {
		err := ep.traverseCategory(cc, el)
		if err != nil {
			return err
		}

	}
	return nil
}
