package importexport

import (
	"bytes"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/types"
)

func LocaleListToDict(list []types.Locale) map[string]types.Locale {
	l := map[string]types.Locale{}
	for _, v := range list {
		if v.ID == "" {
			v.ID = v.IETF
		}
		l[v.IETF] = v
	}
	return l
}

func Test_Export(t *testing.T) {
	monkey.Patch(time.Now, func() time.Time { return time.Date(2022, 01, 23, 18, 24, 37, 0, time.UTC) })

	tests := []struct {
		name    string
		input   string
		options ExportI18NOptions
		project types.ExtendedProject
	}{
		{
			name: "Should match output",
			options: ExportI18NOptions{
				LocaleKey:    "Iso639_3",
				LocaleFilter: []string{},
			},
			project: types.ExtendedProject{
				Locales: LocaleListToDict(types.DefaultLocales),
				Project: types.Project{
					Title:       "Project Foo",
					Description: "Foo is Bar for Baz",
				},
				Categories: map[string]types.ExtendedCategory{
					"cat-a": {
						Category: types.Category{
							Title: "General Category",
							Key:   "General",
						},
						Translations: map[string]types.ExtendedTranslation{
							"t-a": {
								Translation: types.Translation{
									Key:   "Welcome",
									Title: "Welcoming the user to Foo",
									Variables: map[string]interface{}{
										"userName": "Rock",
									},
								},
								Values: map[string]types.TranslationValue{"en-GB": {
									LocaleID: "en-GB",
									Value:    "Welcome, {{user}}",
								}},
							},
						},
						SubCategories: []types.ExtendedCategory{
							{
								Category: types.Category{
									Title: "For use in forms",
									Key:   "Forms",
								},
								SubCategories: []types.ExtendedCategory{
									{
										Category: types.Category{
											Title:       "Buttons - You click them",
											Description: "Surely you know what they are!",
											Key:         "Buttons",
										},
										Translations: map[string]types.ExtendedTranslation{
											"t-b": {
												Translation: types.Translation{
													Key:       "GoToCheckout",
													Title:     "The submit-button",
													Variables: map[string]interface{}{"count": 42},
												},
												Values: map[string]types.TranslationValue{
													"locale-en": {LocaleID: "en-GB", Value: "Go to checkout ({{count}})"},
													"locale-no": {LocaleID: "nb-NO", Value: "Gå til utsjekk ({{count}})"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			i18n, err := ExportI18N(tt.project, tt.options)

			testza.AssertNoError(t, err)
			// b, err := json.MarshalIndent(i18n.ToMap(), "", "  ")
			// testza.AssertNoError(t, err)
			internal.MatchSnapshot(t, "json", i18n.ToMap())
			// testza.AssertEqual(t, tt.wantI18n, i18n)

			node, err := ExportI18N(tt.project, ExportI18NOptions{})
			testza.AssertNoError(t, err)
			internal.MatchSnapshot(t, "node.yaml", node)
			internal.MatchSnapshot(t, "node-map.yaml", node.ToMap())

			var w bytes.Buffer
			err = ExportByGoTemplate("typescript.tmpl", tt.project, i18n, &w)
			testza.AssertNoError(t, err)
			ts := w.String()
			testza.AssertNotNil(t, ts)
			testza.AssertNotEqual(t, "", ts)
			pretty, err := Prettier(ts)
			testza.AssertNoError(t, err)
			if err != nil {
				t.Errorf("prettier failed: %s", err)
				t.Logf("The input was: %s", ts)

				return
			}
			testza.AssertNotNil(t, pretty)
			testza.AssertNotEqual(t, "", pretty)
			internal.MatchSnapshot(t, "ts", []byte(pretty))
		})
	}
}
