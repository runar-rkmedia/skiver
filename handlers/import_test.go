package handlers

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/ghodss/yaml"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/types"
)

func TestImport(t *testing.T) {
	tests := []struct {
		name       string
		localeHint *types.Locale
		fields     string
		expects    *Import
		wantErr    bool
	}{

		{
			"With languages",
			nil,
			`
en:
  general:
    thisIsFine: Great
'no': # Fun with norwegian in yaml!
  general:
    thisIsFine: Greit
`,
			&Import{
				Categories: map[string]types.ExtendedCategory{"general": {
					Category: types.Category{
						Entity: types.Entity{
							CreatedBy: "jim",
						},
						Title:       "General",
						Description: "",
						Key:         "general",
						ProjectID:   "proj-123",
					},
					Translations: map[string]types.ExtendedTranslation{
						"thisIsFine": {
							Translation: types.Translation{
								Entity: types.Entity{
									CreatedBy: "jim",
								},
								Key:   "thisIsFine",
								Title: "This is fine",
							},
							Values: map[string]types.TranslationValue{
								"Greatloc-en": {
									Entity: types.Entity{
										CreatedBy: "jim",
									},
									Value:    "Great",
									LocaleID: "loc-en",
									Source:   "test-import",
								},
								"Greitloc-no": {
									Entity: types.Entity{
										CreatedBy: "jim",
									},
									Value:    "Greit",
									LocaleID: "loc-no",
									Source:   "test-import",
								},
							},
						},
					},
				},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j map[string]interface{}
			err := yaml.Unmarshal([]byte(tt.fields), &j)
			if err != nil {
				t.Errorf("TEST-INPUT_ERROR: Failed to unmarshal: %s %s", err, tt.fields)
				return
			}
			got, err := ImportI18NTranslation(_test_locales, tt.localeHint, "proj-123", "jim", "test-import", j)
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

			testza.AssertEqual(t, tt.expects, got)
		})
	}
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
