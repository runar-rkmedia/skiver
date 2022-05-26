package importexport

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

// Expects a map of arbitriry nestings of
func FlattenStringMap(prefix string, src map[string]interface{}, dest map[string]string, allowOverwrites bool) error {
	if len(prefix) > 0 {
		prefix += "."
	}
	for k, v := range src {
		switch child := v.(type) {
		case map[string]string:
			for key, v := range child {
				newKey := prefix + k + "." + key
				if !allowOverwrites {
					if val, exists := dest[newKey]; exists {
						return fmt.Errorf("Key already exists in destination for '%s'. The previous value was '%s' and the new value is '%s'", newKey, val, v)
					}
				}
				dest[newKey] = v
			}
		case map[string]interface{}:
			FlattenStringMap(prefix+k, child, dest, allowOverwrites)
		case string:
			newKey := prefix + k
			if !allowOverwrites {
				if val, exists := dest[newKey]; exists {
					return fmt.Errorf("Key already exists in destination for '%s'. The previous value was '%s' and the new value is '%s'", newKey, val, v)
				}
			}
			dest[newKey] = child
		default:
			return fmt.Errorf("Unhandled type in %s: '%#T' for: %#v", prefix+k, v, v)
		}
	}
	return nil
}

// FlattenExtendedCategories returns a map of categories and translations by their keys
func FlattenExtendedCategories(ec map[string]types.ExtendedCategory) (map[string]types.ExtendedCategory, map[string]types.ExtendedTranslation) {
	catsByKeys := map[string]types.ExtendedCategory{}
	translationsByKeys := map[string]types.ExtendedTranslation{}

	for _, c := range ec {
		catsByKeys[c.Key] = c
		for _, t := range c.Translations {
			translationsByKeys[c.Key+"."+t.Key] = t
		}
	}

	return catsByKeys, translationsByKeys
}

type ChangeRequest struct {
	Kind    string
	Payload interface{}
}

func DescribeProjectContent(p types.ExtendedProject, src map[string]interface{}) ([]ChangeRequest, error) {
	var changeSet []ChangeRequest

	flat := map[string]string{}
	err := FlattenStringMap("", src, flat, false)
	if err != nil {
		return changeSet, err
	}
	if len(flat) == 0 {
		return changeSet, nil
	}
	cats, translations := FlattenExtendedCategories(p.Categories)

	for k, v := range flat {
		if c, ok := cats[k]; ok {
			if c.Title == v {
				// No changes needed
				continue
			}
			change := types.Category{Title: v, Key: c.Key}
			change.ID = c.ID
			changeSet = append(changeSet, ChangeRequest{
				Kind:    "CategoryTitle",
				Payload: change,
			})
		}
		if t, ok := translations[k]; ok {
			if t.Title == v {
				// No changes needed
				continue
			}
			change := types.Translation{Title: v, Key: t.Key}
			change.ID = t.ID
			changeSet = append(changeSet, ChangeRequest{
				Kind:    "TranslationTitle",
				Payload: change,
			})
			continue
		}
		sorted := utils.SortedMapKeys(translations)
		suggestions := utils.SuggestionsFor(k, sorted, 0, 0)
		return changeSet, fmt.Errorf("failed to locate a category or translation for key '%s' in project '%s' (%s). Did you perhaps mean one of these?: %v", k, p.Title, p.ID, suggestions.ToStringSlice())
	}

	return changeSet, nil
}
