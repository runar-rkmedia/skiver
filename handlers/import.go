package handlers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/runar-rkmedia/skiver/types"
)

// The Import reprecents values that should be created in the database.
// The keys in each map is of no importance except for temporarily tracking these values
type Import struct {
	Categories map[string]types.ExtendedCategory
}

// i18n-translations are either:
// a root-elemenent of type locale as key, or the category as key, or the translation as key.
// Every leaf-node must be of type string
func ImportI18NTranslation(
	locales []types.Locale,
	localeHint *types.Locale,
	projectID string,
	createdBy string,
	source types.CreatorSource,
	input map[string]interface{},
) (*Import, error) {
	localeLength := len(locales)
	if input == nil || len(input) == 0 {
		return nil, fmt.Errorf("Empty input")
	}
	if projectID == "" {
		return nil, fmt.Errorf("ProjectID is required")
	}
	if createdBy == "" {
		return nil, fmt.Errorf("source is required")
	}
	if source == "" {
		return nil, fmt.Errorf("source is required")
	}
	if localeLength == 0 && localeHint == nil {
		return nil, fmt.Errorf("No locales")
	}
	mv, err := getMapPaths(input)
	if err != nil {
		return nil, err
	}
	// We dont need to sort, but it is nice to have idempotency where we can
	sort.Slice(mv, sortMapPath(mv))
	var imp = Import{
		Categories: make(map[string]types.ExtendedCategory),
	}
	for i := 0; i < len(mv); i++ {
		node := getNode(mv[i])
		if node.Value == "" && node.Root == "" && node.MidPath == "" {
			return nil, fmt.Errorf("one of the values failed to parse: %s", strings.Join(mv[i], "."))
		}

		var locale types.Locale
		if localeHint != nil {
			locale = *localeHint
		}
		if locale.ID == "" {
			split := strings.Split(node.Root, ".")
			for i := 0; i < localeLength; i++ {

				switch split[0] {
				case locales[i].ID, locales[i].IETF, locales[i].Iso639_3, locales[i].Iso639_2, locales[i].Iso639_1:
					locale = locales[i]
					break
				}
			}
			if locale.ID == "" {
				return nil, fmt.Errorf("Failed to resolve as locale, attempted to parse '%s' as locale from value. You can add a locale as input-hint, or specify the locale within the body. The first value that failed to parse was: '%s'", split[0], strings.Join(mv[i], "."))
			}
			category := ""

			category = strings.Join(split[1:], ".")
			if category == "" {
				category = types.RootCategory
			}
			// if len(split) > 2 {
			// }
			// translation := node.MidPath
			translation, context := cutLast(node.MidPath, "_")
			translationValue := node.Value
			if _, ok := imp.Categories[category]; !ok {
				imp.Categories[category] = types.ExtendedCategory{}
			}
			cat := imp.Categories[category]
			cat.Key = category
			cat.Title = InferTitle(category)
			cat.ProjectID = projectID
			cat.CreatedBy = createdBy
			if cat.Translations == nil {
				cat.Translations = make(map[string]types.ExtendedTranslation)
			}
			if _, ok := cat.Translations[translation]; !ok {
				cat.Translations[translation] = types.ExtendedTranslation{}
			}
			t := cat.Translations[translation]
			t.Key = translation
			t.CreatedBy = createdBy
			t.Title = InferTitle(translation)
			if t.Values == nil {
				t.Values = make(map[string]types.TranslationValue)
			}
			tvId := locale.ID
			if _, ok := t.Values[tvId]; !ok {
				t.Values[tvId] = types.TranslationValue{}
			}
			tv := t.Values[tvId]
			if context != "" {
				if tv.Context == nil {
					tv.Context = map[string]string{}
				}
				tv.Context[context] = translationValue
			} else {
				tv.Value = translationValue
			}
			tv.LocaleID = locale.ID
			tv.CreatedBy = createdBy
			tv.Source = source

			t.Values[tvId] = tv
			cat.Translations[translation] = t
			imp.Categories[category] = cat
		}

	}

	return &imp, nil
}

func cutLast(s, sep string) (string, string) {
	if s == "" {
		return "", ""
	}
	split := strings.Split(s, sep)
	i := len(split) - 1
	if i == 0 {
		return s, ""
	}
	first := strings.Join(split[:i], sep)
	last := strings.Join(split[i:], sep)
	if first == "" {
		if last == "" {
			return "", ""
		}
		unos, dos := cutLast(last, sep)
		return sep + unos, dos
	}

	return first, last
}

// TODO: implement
func CleanKey(s string) string {
	return s
}

// TODO: This should be a bit smarter.
func InferTitle(s string) string {
	if s == "" {
		return ""
	}
	if len(s) == 1 {
		return strings.ToUpper(s)
	}
	str := strings.ReplaceAll(stringy.New(s).SnakeCase().ToLower(), "_", " ")
	return strings.ToTitle(str[:1]) + str[1:]
}

type Node struct{ Root, MidPath, Value string }

func getNode(mapPath []string) (n Node) {
	l := len(mapPath)
	if l == 0 {
		return
	}
	n.Value = mapPath[l-1]
	if l == 1 {
		return
	}
	if l == 2 {
		n.Root = mapPath[0]
		return
	}
	n.MidPath = mapPath[l-2]
	n.Root = strings.Join(mapPath[0:l-2], ".")
	return
}

// returns the map as a slice of paths where the last item in the path is the value.
// Fot translations, this is ok, since they are strings in any case.
func getMapPaths(input interface{}, paths ...string) (n [][]string, err error) {
	switch input.(type) {
	case nil:
		return
	case float64, int:

		f := fmt.Sprintf("%v", input.(float64))
		paths = append(paths, f)
		n = append(n, paths)
		return n, err

	case string:
		paths = append(paths, input.(string))
		n = append(n, paths)
		return n, err
	case map[string]interface{}:
		for k, v := range input.(map[string]interface{}) {
			p := make([]string, len(paths)+1)
			copy(p, paths)
			p[len(p)-1] = k
			nn, err := getMapPaths(v, p...)
			if err != nil {
				return n, err
			}
			n = append(n, nn...)

		}
		return
	}
	return nil, fmt.Errorf("unhandled type: %t %#v", input, input)
}

// Use with sort.Slice()
func sortMapPath(got [][]string) func(i, j int) bool {
	return func(i, j int) bool {

		li := len(got[i])
		lj := len(got[j])
		if li == lj {

			I := strings.Join(got[i], "|")
			J := strings.Join(got[j], "|")
			return I < J
		}
		return li < lj
	}
}
