package types

import (
	"fmt"
	"strings"
)

type ExtendedProject struct {
	Project
	Exists     *bool `json:"exists,omitempty"`
	Categories map[string]ExtendedCategory
	Locales    map[string]Locale
}
type ExtendedCategory struct {
	Category
	Exists       *bool `json:"exists,omitempty"`
	Translations map[string]ExtendedTranslation
}
type ExtendedTranslation struct {
	Translation
	Exists *bool `json:"exists,omitempty"`
	Values map[string]TranslationValue
}

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
		key := cid
		if opts.ByKeyLike {
			key = ec.Key
		}
		ep.Categories[key] = ec
	}
	return
}

func ExportI18N(ex ExtendedProject, options ExportI18NOptions) (i18n I18N, err error) {
	i18n = make(I18N)
	if options.LocaleKey == "" {
		options.LocaleKey = LocaleKeyISO1
	}
	for _, l := range ex.Locales {
		key := getLocaleKey(options.LocaleKey, l)
		if key == "" {
			err = fmt.Errorf("the locale-key was empty")
			return
		}
		if len(options.LocaleFilter) > 0 {
			found := false
		inner:
			for _, v := range options.LocaleFilter {
				if key == v {
					found = true
					break inner
				}
			}
			if !found {
				continue
			}
		}
		i18n[key] = map[string]map[string]string{}
	}
	for _, c := range ex.Categories {
		cKey := strings.TrimSpace(c.Key)
		for _, l := range ex.Locales {
			key := getLocaleKey(options.LocaleKey, l)
			if i18n[key] == nil {
				continue
			}
			i18n[key][cKey] = map[string]string{}
		}
		for _, t := range c.Translations {
			tKey := t.Key
			if t.Context != "" {
				tKey = tKey + "_" + t.Context
			}
			for _, v := range t.Values {
				if v.Value == "" {
					continue
				}
				locale, ok := ex.Locales[v.LocaleID]
				if !ok {
					err = fmt.Errorf("the locale was not found in the export")
					return
				}
				localeKey := getLocaleKey(options.LocaleKey, locale)
				if i18n[localeKey] == nil {
					continue
				}

				i18n[localeKey][cKey][tKey] = v.Value
			}
		}
	}
	return
}

func getLocaleKey(key LocaleKey, locale Locale) string {
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

type I18N map[string]map[string]map[string]string
