package handlers

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/ghodss/yaml"
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
    thisIsFine: Great
    thisIsFine_superb: Fantastic
    _strangeNonContext: foo
    _strange_context: baz
'no': # Fun with norwegian in yaml!
  general:
    thisIsFine_superb: Fantastisk
`,
			&Import{
				Categories: map[string]types.ExtendedCategory{"general": {
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
									Value:    "",
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
			var j map[string]interface{}
			err := yamlUnmarshalAllowTabs(tt.fields, &j)
			if err != nil {
				t.Errorf("TEST-INPUT_ERROR: Failed to unmarshal: %s %s", err, tt.fields)
				return
			}
			base := types.Project{}
			base.ID = "proj-123"
			base.CreatedBy = "jim"
			base.OrganizationID = "org-123"
			got, warnings, err := ImportI18NTranslation(_test_locales, tt.localeHint, base, "test-import", j)
			if !tt.wantErr {
				testza.AssertNoError(t, err)
			} else if got == nil {
				t.Error("expected error, but none was returned")
			}
			if err := internal.Compare("result", got, tt.expects, internal.CompareOptions{
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

			// testza.AssertEqual(t, tt.expects, got)
		})
	}
}

// Tabs are annoying in yaml, so lets just convert it.
func yamlUnmarshalAllowTabs(s string, j interface{}) error {
	s = strings.ReplaceAll(s, "\t", "  ")
	return yaml.Unmarshal([]byte(s), j)
}

var (
	_test_locales = []types.Locale{
		{
			Entity:   types.Entity{ID: "loc-en"},
			Iso639_1: "en",
			Iso639_2: "en",
			Iso639_3: "eng",
			IETF:     "en-US",
		},
		{
			Entity:   types.Entity{ID: "loc-no"},
			Iso639_1: "no",
			Iso639_2: "no",
			Iso639_3: "nor",
			IETF:     "nb-NO",
		},
	}
)
