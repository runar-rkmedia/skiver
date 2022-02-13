package importexport

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/types"
)

func TestImport(t *testing.T) {
	tests := []struct {
		name            string
		localeHint      *types.Locale
		fields          string
		expects         *Import
		expectsWarnings []Warning
		wantErr         bool
	}{

		{
			"With languages",
			nil,
			`
en:
  general:
    thisIsFine: Great
    thisIsFine_superb: Fantastic
    _strangeNonContext: foo
    _strange_context: baz
		form: 
		  buttons: 
			  submit: "Submit {{count}} items"
		  inputLabel: 
			  email: "Email"
			  age: "Between {{minAge}} and {{maxAge}}"
'no': # Fun with norwegian in yaml!
  general:
    thisIsFine_superb: Fantastisk
`,
			&Import{
				Categories: map[string]types.ExtendedCategory{
					"general.form.buttons": {
						Category: types.Category{
							Entity: types.Entity{
								CreatedBy:      "jim",
								OrganizationID: "org-123",
							},
							Title:     "Buttons",
							Key:       "general.form.buttons",
							ProjectID: "proj-123",
						},
						Translations: map[string]types.ExtendedTranslation{
							"submit": {
								Translation: types.Translation{
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Key:   "submit",
									Title: "Submit",
									Variables: map[string]interface{}{
										"count": 42,
									},
								},
								Values: map[string]types.TranslationValue{
									"loc-en": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										Value:    "Submit {{count}} items",
										LocaleID: "loc-en",
										Source:   "test-import",
									},
								},
							},
						},
					},
					"general.form.inputLabel": {
						Category: types.Category{
							Entity: types.Entity{
								CreatedBy:      "jim",
								OrganizationID: "org-123",
							},
							Title:       "Input label",
							Description: "",
							Key:         "general.form.inputLabel",
							ProjectID:   "proj-123",
						},
						Translations: map[string]types.ExtendedTranslation{
							"age": {
								Translation: types.Translation{
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Key:   "age",
									Title: "Age",
									Variables: map[string]interface{}{
										"minAge": "???",
										"maxAge": "???",
									},
								},
								Values: map[string]types.TranslationValue{
									"loc-en": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										Value:    "Between {{minAge}} and {{maxAge}}",
										LocaleID: "loc-en",
										Source:   "test-import",
									},
								},
							},
							"email": {
								Translation: types.Translation{
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Key:   "email",
									Title: "Email",
								},
								Values: map[string]types.TranslationValue{
									"loc-en": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										Value:    "Email",
										LocaleID: "loc-en",
										Source:   "test-import",
									},
								},
							},
						},
					},
					"general.form": {
						Category: types.Category{
							Entity:    types.Entity{CreatedBy: "jim", OrganizationID: "org-123"},
							Title:     "Form",
							Key:       "general.form",
							ProjectID: "proj-123",
						},
						Translations: map[string]types.ExtendedTranslation{},
					},
					"general": {
						Category: types.Category{
							Entity: types.Entity{
								CreatedBy:      "jim",
								OrganizationID: "org-123",
							},
							Title:       "General",
							Description: "",
							Key:         "general",
							ProjectID:   "proj-123",
						},
						Translations: map[string]types.ExtendedTranslation{

							"_strangeNonContext": {
								Translation: types.Translation{
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Key:   "_strangeNonContext",
									Title: "Strange non context",
								},
								Values: map[string]types.TranslationValue{
									"loc-en": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										Value:    "foo",
										LocaleID: "loc-en",
										Source:   "test-import",
									},
								},
							},

							"_strange": {
								Translation: types.Translation{
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Key:   "_strange",
									Title: "Strange",
								},
								Values: map[string]types.TranslationValue{
									"loc-en": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										LocaleID: "loc-en",
										Source:   "test-import",
										Context:  map[string]string{"context": "baz"},
									},
								},
							},

							"thisIsFine": {
								Translation: types.Translation{
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Key:   "thisIsFine",
									Title: "This is fine",
								},
								Values: map[string]types.TranslationValue{
									"loc-en": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										Value:    "Great",
										LocaleID: "loc-en",
										Source:   "test-import",
										Context: map[string]string{
											"superb": "Fantastic",
										},
									},
									"loc-no": {
										Entity: types.Entity{
											CreatedBy:      "jim",
											OrganizationID: "org-123",
										},
										LocaleID: "loc-no",
										Source:   "test-import",
										Context: map[string]string{
											"superb": "Fantastisk",
										},
									},
								},
							},
						},
					},
				},
			},
			nil,
			false,
		},
		{
			"interpolation",
			nil,
			`
en:
	meaningOfLife: The meaning of $t(life) is {{count}}
	life: life
"no":
	meaningOfLife: Meningen med $t(life) er {{count}}
	life: livet
`,
			&Import{
				Categories: map[string]types.ExtendedCategory{types.RootCategory: {
					Category: types.Category{
						Entity: types.Entity{
							CreatedBy:      "jim",
							OrganizationID: "org-123",
						},
						Title:       "Root",
						Description: "",
						Key:         types.RootCategory,
						ProjectID:   "proj-123",
					},
					Translations: map[string]types.ExtendedTranslation{
						"meaningOfLife": {
							Translation: types.Translation{
								Entity: types.Entity{
									CreatedBy:      "jim",
									OrganizationID: "org-123",
								},
								Key:   "meaningOfLife",
								Title: "Meaning of life",
								Variables: map[string]interface{}{
									"count":      42,
									"_refs:life": nil,
								},
							},
							Values: map[string]types.TranslationValue{
								"loc-en": {
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Value:    "The meaning of $t(life) is {{count}}",
									LocaleID: "loc-en",
									Source:   "test-import",
								},
								"loc-no": {
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Value:    "Meningen med $t(life) er {{count}}",
									LocaleID: "loc-no",
									Source:   "test-import",
								},
							},
						},
						"life": {
							Translation: types.Translation{
								Entity: types.Entity{
									CreatedBy:      "jim",
									OrganizationID: "org-123",
								},
								Key:   "life",
								Title: "Life",
							},
							Values: map[string]types.TranslationValue{
								"loc-en": {
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Value:    "life",
									LocaleID: "loc-en",
									Source:   "test-import",
								},
								"loc-no": {
									Entity: types.Entity{
										CreatedBy:      "jim",
										OrganizationID: "org-123",
									},
									Value:    "livet",
									LocaleID: "loc-no",
									Source:   "test-import",
								},
							},
						},
					},
				},
				},
			},
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if strings.Contains(tt.name, "inter") {
				return
			}
			var j map[string]interface{}
			err := internal.YamlUnmarshalAllowTabs(tt.fields, &j)
			if err != nil {
				t.Errorf("TEST-INPUT_ERROR: Failed to unmarshal: %s %s", err, tt.fields)
				return
			}
			node, err := importAsI18Nodes(j)
			testza.AssertNoError(t, err)
			// testza.AssertEqualValues(t, nil, node)

			nodeMap := node.ToMap()
			testza.AssertEqual(t, nodeMap, j, "input -> i18nNodes -> nodeMap should equal the input-value")

			if err := internal.Compare("nodemap", nodeMap, j); err != nil {
				t.Error(err)
			}
			merged, err := node.MergeAsIfRootIsLocale(types.Test_locales)
			testza.AssertNoError(t, err)
			internal.MatchSnapshot(t, "merged.yaml", merged)

			base := types.Project{}
			base.ID = "proj-123"
			base.CreatedBy = "jim"
			base.OrganizationID = "org-123"
			// Should this function create all categories of previous levels?
			got, warnings, err := ImportI18NTranslation(types.Test_locales, tt.localeHint, base, "test-import", j)
			if !tt.wantErr {
				testza.AssertNoError(t, err)
			} else if got == nil {
				t.Error("expected error, but none was returned")
			}
			// testza.AssertEqual(t, tt.expects, got)
			// return
			if err := internal.Compare("import", got, tt.expects, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				Yaml:    true,
				JSON:    false,
			}); err != nil {
				t.Log("input", tt.fields)
				t.Error(err)
			}
			if err := internal.Compare("warnings", warnings, tt.expectsWarnings, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				Yaml:    true,
			}); err != nil {
				t.Log("input", tt.fields)
				t.Error(err)
			}
			p := types.ExtendedProject{Categories: got.Categories, Locales: map[string]types.Locale{}}
			p.CategoryTree = types.CreateCategoryTreeNode(p.Categories)
			for _, v := range types.Test_locales {

				p.Locales[v.ID] = v
			}

			export, err := ExportI18N(p, ExportI18NOptions{
				LocaleKey:    "",
				LocaleFilter: []string{},
			})
			// internal.PrintMultiLineYaml("ppppp", export)
			// t.FailNow()
			if err := internal.Compare("export of resulting import should match input (ignoring order)", export.ToMap(), j, internal.CompareOptions{
				Diff:    true,
				Reflect: true,
				Yaml:    true,
			}); err != nil {
				t.Error(err)
			}
		})
	}
}
