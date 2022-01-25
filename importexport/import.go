package importexport

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gobeam/stringy"
	"github.com/runar-rkmedia/skiver/internal"
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
			node.Nodes[k] = n

		}
		return node, nil

	}
	return node, fmt.Errorf("Unhandled type for %#v", input)
}

type Locales []types.Locale

func (t Locales) GetLocaleID(key string) string {
	for _, l := range t {
		if l.ID == key {
			return l.ID
		}
		if l.IETF == key {
			return l.ID
		}
		if l.Iso639_3 == key {
			return l.ID
		}
		if l.Iso639_2 == key {
			return l.ID
		}
		if l.Iso639_1 == key {
			return l.ID
		}
	}

	return ""
}

// non-Recursively traverses the node-tree to find all categories and fill any value in the import
func importFromCategoryNode(base types.Project, source types.CreatorSource, key string, node I18NWithLocales) (types.ExtendedCategory, error) {

	cat := types.ExtendedCategory{
		Translations: map[string]types.ExtendedTranslation{},
	}
	cat.Key = key
	cat.Title = InferTitle(key)
	cat.ProjectID = base.ID
	cat.CreatedBy = base.CreatedBy
	cat.OrganizationID = base.OrganizationID

	if len(node.Value) > 0 {
		t := types.ExtendedTranslation{}
		t.Key = key
		tranlationKey, context := cutLast(key, "_")
		t.Title = InferTitle(tranlationKey)
		fmt.Println(context)
		t.CreatedBy = base.CreatedBy
		t.OrganizationID = base.OrganizationID
		t.Values = map[string]types.TranslationValue{}
		nodeValueKeys := sortedMapKeys(node.Value)
		for _, localeId := range nodeValueKeys {
			// TODO: infer variables, etc.
			value := node.Value[localeId]
			tv := types.TranslationValue{LocaleID: localeId, Value: value}
			tv.Source = source

			tv.CreatedBy = base.CreatedBy
			tv.OrganizationID = base.OrganizationID
			t.Values[localeId] = tv

			w, variables := InferVariables(tv.Value, cat.Key, t.Key)
			if len(variables) > 0 {
				if t.Variables == nil {
					t.Variables = map[string]interface{}{}
				}
				for k, v := range variables {
					if ex, ok := t.Variables[k]; ok {
						if ex == v {
							continue
						}
						w = append(w, Warning{
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

		cat.Translations[t.Key] = t
	}

	if len(node.Nodes) > 0 {
		cat.SubCategories = []types.ExtendedCategory{}
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

				t := types.ExtendedTranslation{}
				tranlationKey, context := cutLast(scKey, "_")
				fmt.Println(context)
				t.Key = tranlationKey
				t.Title = InferTitle(tranlationKey)
				t.CreatedBy = base.CreatedBy
				t.OrganizationID = base.OrganizationID
				t.Values = map[string]types.TranslationValue{}
				nodeValueKeys := sortedMapKeys(childNode.Value)
				for _, localeId := range nodeValueKeys {
					// TODO: infer variables, etc.
					value := childNode.Value[localeId]
					tv := types.TranslationValue{LocaleID: localeId}
					tv.Source = source

					tv.CreatedBy = base.CreatedBy
					tv.OrganizationID = base.OrganizationID
					t.Values[localeId] = tv
					w, variables := InferVariables(tv.Value, cat.Key, t.Key)
					if context != "" {
						if tv.Context == nil {
							tv.Context = map[string]string{}
						}
						tv.Context[context] = value
					} else {
						tv.Value = value
					}

					if len(variables) > 0 {
						if t.Variables == nil {
							t.Variables = map[string]interface{}{}
						}
						for k, v := range variables {
							if ex, ok := t.Variables[k]; ok {
								if ex == v {
									continue
								}
								w = append(w, Warning{
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
				cat.Translations[t.Key] = t
			}
			sc, err := importFromCategoryNode(base, source, scKey, childNode)
			if err != nil {
				return cat, fmt.Errorf("Error occured parsing subCategories for node %s: %w", internal.MustJSON(node), err)
			}
			cat.SubCategories = append(cat.SubCategories, sc)

		}
	}

	return cat, nil
}

func thing(key string, base types.Project, source types.CreatorSource, childNode I18NWithLocales, categoryKey string) types.ExtendedTranslation {

	t := types.ExtendedTranslation{}
	tranlationKey, context := cutLast(key, "_")
	fmt.Println(context)
	t.Key = tranlationKey
	t.Title = InferTitle(tranlationKey)
	t.CreatedBy = base.CreatedBy
	t.OrganizationID = base.OrganizationID
	t.Values = map[string]types.TranslationValue{}
	nodeValueKeys := sortedMapKeys(childNode.Value)
	for _, localeId := range nodeValueKeys {
		// TODO: infer variables, etc.
		value := childNode.Value[localeId]
		tv := types.TranslationValue{LocaleID: localeId}
		tv.Source = source

		tv.CreatedBy = base.CreatedBy
		tv.OrganizationID = base.OrganizationID
		t.Values[localeId] = tv
		w, variables := InferVariables(tv.Value, categoryKey, t.Key)
		if context != "" {
			if tv.Context == nil {
				tv.Context = map[string]string{}
			}
			tv.Context[context] = value
		} else {
			tv.Value = value
		}

		if len(variables) > 0 {
			if t.Variables == nil {
				t.Variables = map[string]interface{}{}
			}
			for k, v := range variables {
				if ex, ok := t.Variables[k]; ok {
					if ex == v {
						continue
					}
					w = append(w, Warning{
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
	return t
}

func sortedMapKeys[T any](input map[string]T) []string {
	keys := make([]string, len(input))
	i := 0
	for k := range input {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func InferVariables(translationValue, category, translation string) ([]Warning, map[string]interface{}) {
	w := []Warning{}
	variables := make(map[string]interface{})

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
	locales Locales,
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
	nodeKeys := sortedMapKeys(nodes.Nodes)
	// The root-nodes are probably the key for the locale, so we unwind one level of the nodes-tree
	rootNode := types.ExtendedCategory{Translations: map[string]types.ExtendedTranslation{}}
	for _, key := range nodeKeys {
		node := nodes.Nodes[key]

		cat, err := importFromCategoryNode(base, source, key, node)
		if err != nil {
			return &imp, w, err
		}
		// special case for rootnodes
		if len(node.Nodes) == 0 {
			if len(cat.SubCategories) > 0 {
				return &imp, w, fmt.Errorf("Did not expect root-category to have subcategories")
			}
			for k, v := range cat.Translations {
				rootNode.Translations[k] = v
			}
			continue
		}
		imp.Categories[key] = cat
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
