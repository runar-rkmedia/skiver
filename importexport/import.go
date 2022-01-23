package importexport

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/runar-rkmedia/skiver/interpolator"
	"github.com/runar-rkmedia/skiver/interpolator/lexer"
	"github.com/runar-rkmedia/skiver/interpolator/parser"
	"github.com/runar-rkmedia/skiver/types"
)

// The Import reprecents values that should be created in the database.
// The keys in each map is of no importance except for temporarily tracking these values
type Import struct {
	Categories map[string]types.ExtendedCategory
}

type WarningLevel string
type WarningKind string

const (
	WarningLevelMinor               WarningLevel = "minor"
	WarningLevelMajor               WarningLevel = "major"
	WarningKindTranslationVariables WarningKind  = "translation-variable"
	WarningKindTranslationReference WarningKind  = "translation-reference"
)

type Warning struct {
	Message string       `json:"message"`
	Error   error        `json:"error,omitempty"`
	Details interface{}  `json:"details,omitempty"`
	Level   WarningLevel `json:"level"`
	Kind    WarningKind  `json:"kind"`
}

func newWarning(msg string, kind WarningKind, level WarningLevel) Warning {
	return Warning{
		Message: msg,
		Level:   level,
		Kind:    kind,
	}
}

// i18n-translations are either:
// a root-elemenent of type locale as key, or the category as key, or the translation as key.
// Every leaf-node must be of type string
func ImportI18NTranslation(
	locales []types.Locale,
	localeHint *types.Locale,
	base types.Project,
	source types.CreatorSource,
	input map[string]interface{},
) (*Import, []Warning, error) {
	parso := parser.NewParser(nil)
	var w []Warning

	localeLength := len(locales)
	if input == nil || len(input) == 0 {
		return nil, w, fmt.Errorf("Empty input")
	}
	if base.ID == "" {
		return nil, w, fmt.Errorf("base.projectId (ProjectID) is required")
	}
	if base.CreatedBy == "" {
		return nil, w, fmt.Errorf("base.createdBy is required")
	}
	if base.OrganizationID == "" {
		return nil, w, fmt.Errorf("base.organizationID is required")
	}
	if source == "" {
		return nil, w, fmt.Errorf("source is required")
	}
	if localeLength == 0 && localeHint == nil {
		return nil, w, fmt.Errorf("No locales")
	}
	mv, err := GetMapPaths(input)
	if err != nil {
		return nil, w, err
	}
	// We dont need to sort, but it is nice to have idempotency where we can
	sort.Slice(mv, sortMapPath(mv))
	var imp = Import{
		Categories: make(map[string]types.ExtendedCategory),
	}
	for i := 0; i < len(mv); i++ {
		node := getNode(mv[i])
		if node.Value == "" && node.Root == "" && node.MidPath == "" {
			return nil, w, fmt.Errorf("one of the values failed to parse: %s", strings.Join(mv[i], "."))
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
				return nil, w, fmt.Errorf("Failed to resolve as locale, attempted to parse '%s' as locale from value. You can add a locale as input-hint, or specify the locale within the body. The first value that failed to parse was: '%s'", split[0], strings.Join(mv[i], "."))
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
			cat.ProjectID = base.ID
			cat.CreatedBy = base.CreatedBy
			cat.OrganizationID = base.OrganizationID
			if cat.Translations == nil {
				cat.Translations = make(map[string]types.ExtendedTranslation)
			}
			if _, ok := cat.Translations[translation]; !ok {
				cat.Translations[translation] = types.ExtendedTranslation{}
			}
			t := cat.Translations[translation]
			t.Key = translation
			t.CreatedBy = base.CreatedBy
			t.OrganizationID = base.OrganizationID
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

			// Attempt to infer variables and nested references from the translation-values used.
			parsed, parseErr := parso.Parse(translationValue)
			if parseErr != nil {
				warn := newWarning("There was a problem parsing the translation-value.", WarningKindTranslationVariables, WarningLevelMajor)
				warn.Error = parseErr
			}
			for i, n := range parsed.Nodes {
				switch n.Token.Kind {
				case lexer.TokenNestingPrefix:
					if n.Left == nil {
						warn := newWarning(
							fmt.Sprintf(
								"Attempted to interpret a translation-value and infer any references, but failed. %s (%d) did not have a left-node. This occured in category %s translation %s value %s at %d-%d",
								n.Token.Kind, i, category, translation, translationValue, n.Token.Start, n.Token.End),
							WarningKindTranslationReference,
							WarningLevelMinor,
						)
						warn.Details = parsed.Nodes
						w = append(w, warn)
						continue
					}
					key := strings.TrimSpace(n.Left.Token.Literal)
					if key == "" {
						warn := newWarning(
							fmt.Sprintf(
								"Attempted to interpred a translation-value and infer any references, but the value was empty. %s (%d.Left). This occured in category %s translation %s value %s at %d-%d",
								n.Token.Kind, i, category, translation, translationValue, n.Left.Token.Start, n.Left.Token.End),
							WarningKindTranslationVariables,
							WarningLevelMinor,
						)
						warn.Details = parsed.Nodes
						w = append(w, warn)
						continue

					}
					if t.Variables == nil {
						t.Variables = make(map[string]interface{})
					}
					if _, ok := t.Variables["_refs:"+key]; !ok {
						if n.Right != nil {
							t.Variables["_refs:"+key] = n.Right.Token.Literal
						} else {
							t.Variables["_refs:"+key] = nil

						}

					}

				case lexer.TokenPrefix:
					if n.Left == nil {
						warn := newWarning(
							fmt.Sprintf(
								"Attempted to interpret a translation-value and infer any variables used, but failed. %s (%d) did not have a left-node. This occured in category %s translation %s value %s at %d-%d",
								n.Token.Kind, i, category, translation, translationValue, n.Token.Start, n.Token.End),
							WarningKindTranslationVariables,
							WarningLevelMinor,
						)
						warn.Details = parsed.Nodes
						w = append(w, warn)
						continue
					}
					key := strings.TrimSpace(n.Left.Token.Literal)
					if key == "" {
						warn := newWarning(
							fmt.Sprintf(
								"Attempted to interpred a translation-value and infer any variables used, but the value was empty. %s (%d.Left). This occured in category %s translation %s value %s at %d-%d",
								n.Token.Kind, i, category, translation, translationValue, n.Left.Token.Start, n.Left.Token.End),
							WarningKindTranslationVariables,
							WarningLevelMinor,
						)
						warn.Details = parsed.Nodes
						w = append(w, warn)
						continue

					}
					if t.Variables == nil {
						t.Variables = make(map[string]interface{})
					}
					t.Variables[key] = getValueForVariableKey(key)
				}
			}
			tv.LocaleID = locale.ID
			tv.CreatedBy = base.CreatedBy
			tv.OrganizationID = base.OrganizationID
			tv.Source = source

			t.Values[tvId] = tv
			cat.Translations[translation] = t
			imp.Categories[category] = cat
		}

	}

	return &imp, w, nil
}
func getValueForVariableKey(key string) interface{} {
	key = strings.ToLower(key)
	if val, ok := interpolator.DefaultInterpolationExamples[key]; ok {
		return val
	}
	return "???"
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
func GetMapPaths(input interface{}, paths ...string) (n [][]string, err error) {
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
			nn, err := GetMapPaths(v, p...)
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
