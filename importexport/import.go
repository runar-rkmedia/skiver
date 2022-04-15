package importexport

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/runar-rkmedia/skiver/interpolator"
	"github.com/runar-rkmedia/skiver/interpolator/lexer"
	"github.com/runar-rkmedia/skiver/interpolator/parser"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
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

func importAsI18Nodes(input interface{}) (I18N, error) {
	node := I18N{}
	switch t := input.(type) {
	case string:
		node.Value = t
		return node, nil
	case map[string]interface{}:
		node.Nodes = map[string]I18N{}
		for k, v := range t {
			n, err := importAsI18Nodes(v)
			if err != nil {
				return node, err
			}
			// Skip empty nodes
			if n.Value == "" && len(n.Nodes) == 0 {
				continue
			}
			node.Nodes[k] = n

		}
		return node, nil

	}
	return node, fmt.Errorf("Unhandled type for %#v", input)
}

// non-Recursively traverses the node-tree to find all categories and fill any value in the import
func importFromCategoryNode(base types.Project, source types.CreatorSource, key string, node I18NWithLocales) ([]types.ExtendedCategory, error, []Warning) {
	var warnings []Warning

	cats := []types.ExtendedCategory{}
	cat := types.ExtendedCategory{
		Translations: map[string]types.ExtendedTranslation{},
	}
	cat.Key = key
	cat.Title = InferTitle(key)
	cat.ProjectID = base.ID
	cat.CreatedBy = base.CreatedBy
	cat.OrganizationID = base.OrganizationID

	if len(node.Value) > 0 {
		var t types.ExtendedTranslation
		if tk, ok := cat.Translations[t.Key]; ok {
			t = tk
		} else {
			t = types.ExtendedTranslation{}
		}
		warnings = append(warnings, translationFromNode(&t, key, base, source, node, cat.Key)...)
		if _, ok := cat.Translations[t.Key]; ok {
			panic("I existx")
		}
		cat.Translations[t.Key] = t
	}

	if len(node.Nodes) > 0 {
		// Each node here may be either a subcategory, or translation.
		// cat.SubCategories = []types.ExtendedCategory{}
		keys := make([]string, len(node.Nodes))
		i := 0
		for k := range node.Nodes {
			keys[i] = k
			i++
		}
		sort.Strings(keys)

		for _, scKey := range keys {
			// If the child-node does not itself have nodes, they are considered translations.
			// E.g. it depends on the next level of nodes
			childNode := node.Nodes[scKey]

			if len(childNode.Value) > 0 {

				translationKey, _ := SplitTranslationAndContext(scKey, "_")
				var t types.ExtendedTranslation
				if tk, ok := cat.Translations[translationKey]; ok {
					t = tk
				} else {
					t = types.ExtendedTranslation{}
				}
				warnings = append(warnings, translationFromNode(&t, scKey, base, source, childNode, cat.Key)...)
				cat.Translations[translationKey] = t
			} else {

				sc, err, w := importFromCategoryNode(base, source, scKey, childNode)
				if len(w) > 0 {
					warnings = append(warnings, w...)
				}
				if err != nil {
					return cats, fmt.Errorf("Error occured parsing subCategories for node %#v: %w", node, err), warnings
				}
				for _, v := range sc {
					v.Key = cat.Key + "." + v.Key
					cats = append(cats, v)

				}

			}

		}
	}
	cats = append(cats, cat)

	return cats, nil, warnings
}

func translationFromNode(t *types.ExtendedTranslation, key string, base types.Project, source types.CreatorSource, node I18NWithLocales, categoryKey string) []Warning {
	var warnings []Warning

	tranlationKey, context := SplitTranslationAndContext(key, "_")
	if t.Key == "" {

		t.Key = tranlationKey
		t.Title = InferTitle(tranlationKey)
		t.CreatedBy = base.CreatedBy
		t.OrganizationID = base.OrganizationID
		t.Values = map[string]types.TranslationValue{}
	}
	nodeValueKeys := utils.SortedMapKeys(node.Value)
	for _, localeId := range nodeValueKeys {
		value := node.Value[localeId]
		var tv types.TranslationValue
		if tk, ok := t.Values[localeId]; ok {
			tv = tk
		} else {

			tv = types.TranslationValue{LocaleID: localeId}
		}
		tv.Source = source

		tv.CreatedBy = base.CreatedBy
		tv.OrganizationID = base.OrganizationID
		w, variables := InferVariables(value, categoryKey, t.Key)
		if len(w) != 0 {
			warnings = append(warnings, w...)
		}
		if context != "" {
			if tv.Context == nil {
				tv.Context = map[string]string{}
			}
			tv.Context[context] = value
		} else if value != "" {
			tv.Value = value
		}
		t.Values[localeId] = tv

		if len(variables) > 0 {
			if t.Variables == nil {
				t.Variables = map[string]interface{}{}
			}
			for k, v := range variables {
				if ex, ok := t.Variables[k]; ok {
					if ex == v {
						continue
					}
					warnings = append(warnings, Warning{
						Message: "duplicate inferred values with different values detected",
						Details: struct {
							A, B interface{}
						}{A: ex, B: ex},
						Level: WarningLevelMinor,
						Kind:  WarningKindTranslationVariables,
					})
				}
				t.Variables[k] = v

			}
		}

	}
	return warnings
}

var (
	inferVariablesRegex = regexp.MustCompile(`{{\s*([^\s,}]*)[^}]*}}`)
)

func InferVariables(translationValue, category, translation string) ([]Warning, map[string]interface{}) {
	w := []Warning{}
	variables := make(map[string]interface{})

	// The parser below is not yet quite up the the task, so we attempt to infer the easy, common ones with regex first:
	matches := inferVariablesRegex.FindAllStringSubmatch(translationValue, -1)
	if len(matches) > 0 {
		for _, v := range matches {
			if len(v) < 2 {
				continue
			}
			variables[v[1]] = getValueForVariableKey(v[1])

		}
	}

	parso := parser.NewParser(nil)
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
			if n.Right != nil {
				// The Literal value here *might* be a json-value. In this case, we
				// should treat each key-value-pair as variables.
				jsonLike := strings.TrimSpace(n.Right.Token.Literal)
				jsonLike = strings.ReplaceAll(jsonLike, `\"`, `"`)
				var varJson map[string]interface{}
				err := json.Unmarshal([]byte(jsonLike), &varJson)
				if err != nil {
					warn := newWarning("Attempted to interpret the token-arguments as json, but encountered an error", WarningKindTranslationVariables, WarningLevelMinor)
					warn.Details = struct {
						Error    error
						ErrorStr string
						Argument string
					}{
						Error:    err,
						ErrorStr: err.Error(),
						Argument: jsonLike,
					}
					w = append(w, warn)

				} else {
					for key := range varJson {
						if _, ok := variables[key]; ok {
							continue
						}
						variables[key] = getValueForVariableKey(key)
					}

				}
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
			if _, ok := variables["_refs:"+key]; !ok {
				if n.Right != nil {
					variables["_refs:"+key] = n.Right.Token.Literal
				} else {
					variables["_refs:"+key] = nil

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
			variables[key] = getValueForVariableKey(key)
		}
	}
	return w, variables
}

// i18n-translations are either:
// a root-elemenent of type locale as key, or the category as key, or the translation as key.
// Every leaf-node must be of type string
func ImportI18NTranslation(
	// TODO: check if we actually need the locales/localeHint here, or perhaps we should use a wrapper-func
	locales types.Locales,
	localeHint *types.Locale,
	base types.Project,
	source types.CreatorSource,
	input map[string]interface{},
) (*Import, []Warning, error) {
	var w []Warning

	localeLength := len(locales)
	if len(input) == 0 {
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
	var imp = Import{
		Categories: make(map[string]types.ExtendedCategory),
	}
	var locale types.Locale
	if localeHint != nil {
		locale = *localeHint
	}

	nodes_, err := importAsI18Nodes(input)
	if err != nil {
		return nil, w, err
	}
	// At root-level, we expect nodes to be set, and have values.
	if len(nodes_.Nodes) == 0 {
		return nil, w, fmt.Errorf("No nodes found in import")
	}
	if nodes_.Value != "" {
		return nil, w, fmt.Errorf("Did not expect root.node.Value to be set: %#v", nodes_)
	}
	var nodes I18NWithLocales
	if locale.ID == "" {
		// We have an input with one or more locales, where the root-node is a locale-key
		nodes, err = nodes_.MergeAsIfRootIsLocale(locales)
	} else {
		// The input does not include the locale, so we assign each node the locale
		nodes = nodes_.ToLocaleAwere(locale.ID)
	}

	// We dont need to sort, but it is nice to have idempotency where we can For
	// instance, the same failing input will always fail at the same place. It is
	// really annoying to deal with multiple errors if each time one submits one
	// never knows if the changed input actually had any effect since it randomly
	// shows an error from a different node
	nodeKeys := utils.SortedMapKeys(nodes.Nodes)
	// The root-nodes are probably the key for the locale, so we unwind one level of the nodes-tree
	rootNode := types.ExtendedCategory{Translations: map[string]types.ExtendedTranslation{}}
	for _, key_ := range nodeKeys {
		node := nodes.Nodes[key_]
		isRoot := len(node.Nodes) == 0
		// var cats []types.ExtendedCategory

		cats, err, impwarnings := importFromCategoryNode(base, source, key_, node)
		if len(impwarnings) > 0 {
			w = append(w, impwarnings...)
		}
		if err != nil {
			return &imp, w, err
		}
		if len(cats) == 0 {
			continue
		}
		for _, cat := range cats {

			key := cat.Key
			// special case for rootnodes
			if isRoot {
				// if len(cat.SubCategories) > 0 {
				// 	return &imp, w, fmt.Errorf("Did not expect root-category to have subcategories")
				// }
				// if len(cat.Category.SubCategories) > 0 {
				// 	return &imp, w, fmt.Errorf("Did not expect root-category to have subcategories")
				// }
				// oh dear...
				catTKeys := utils.SortedMapKeys(cat.Translations)
				for _, k := range catTKeys {
					v := cat.Translations[k]
					if ex, ok := rootNode.Translations[k]; ok {
						// merge translation-values. We only care about the .Value and .Context.
						for vk, v := range v.Values {
							if exv, ok := ex.Values[vk]; ok {
								if exv.Value == "" {
									exv.Value = v.Value
								}
								if len(v.Context) > 0 {
									if len(exv.Context) > 0 {
										for ck, c := range v.Context {
											exv.Context[ck] = c
										}
									} else {
										exv.Context = v.Context
									}

								}
								ex.Values[vk] = exv
							} else {

								ex.Values[vk] = v
							}

						}

						rootNode.Translations[ex.Key] = ex

					} else {

						rootNode.Translations[v.Key] = v
					}
				}
				continue
			}
			if _, ok := imp.Categories[key]; ok {
				// TODO: handle merge
				panic("already exists a key for this category " + key)

			} else {

				imp.Categories[key] = cat
			}
		}
	}
	if len(rootNode.Translations) > 0 {
		rootNode.Key = types.RootCategory
		rootNode.Title = InferTitle(rootNode.Key)
		rootNode.ProjectID = base.ID
		rootNode.CreatedBy = base.CreatedBy
		rootNode.OrganizationID = base.OrganizationID
		imp.Categories[types.RootCategory] = rootNode
	}
	return &imp, w, nil
}
func getValueForVariableKey(key string) interface{} {
	key = strings.ToLower(key)
	if val, ok := interpolator.DefaultInterpolationExamples[key]; ok {
		return val
	}
	for k, v := range interpolator.DefaultInterpolationExamples {
		if strings.HasSuffix(key, k) {
			return v
		}
	}
	return "???"
}

func SplitTranslationAndContext(s, sep string) (string, string) {
	if s == "" {
		return "", ""
	}
	// keys can have leading uderscores, but that is not a context.
	trimmed := strings.TrimPrefix(s, "_")
	before, after, _ := strings.Cut(trimmed, sep)
	if trimmed != s {
		before = strings.Repeat("_", len(s)-len(trimmed)) + before
	}
	return before, after
}

// TODO: implement
func CleanKey(s string) string {
	return s
}

// TODO: This should be a bit smarter.
func InferTitle(s string) string {
	if s == "" {
		return "Root"
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
