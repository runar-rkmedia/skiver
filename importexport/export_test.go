package importexport

import (
	"bytes"
	"testing"
	"time"

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
	internal.NewMockTimeNow()

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
										"appName":  "20XX",
									},
								},
								Values: map[string]types.TranslationValue{"en-GB": {
									LocaleID: "en-GB",
									Value:    "Welcome, {{user}} to {{year}}",
								}},
							},
						},
					},
					"cat-b": {
						Category: types.Category{
							Title: "For use in forms",
							Key:   "General.Forms",
						},
					},
					"cat-c": {
						Category: types.Category{
							Title:       "Buttons - You click them",
							Description: "Surely you know what they are!",
							Key:         "General.Forms.Buttons",
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
		{
			name: "Should ignore deleted values",
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
									Entity: types.Entity{Deleted: &time.Time{}},
									Key:    "Welcome",
									Title:  "Welcoming the user to Foo",
									Variables: map[string]interface{}{
										"userName": "Rock",
										"appName":  "20XX",
									},
								},
								Values: map[string]types.TranslationValue{"en-GB": {
									LocaleID: "en-GB",
									Value:    "Welcome, {{user}} to {{year}}",
								}},
							},
						},
					},
					"cat-b": {
						Category: types.Category{
							Title: "For use in forms",
							Key:   "General.Forms",
						},
					},
					"cat-c": {
						Category: types.Category{
							Title:       "Buttons - You click them",
							Description: "Surely you know what they are!",
							Key:         "General.Forms.Buttons",
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
		{
			// There was a bug where root-categories generated keys like `.Foo` (leading dot)
			name: "Should handle root-category correctly",
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
							Title: "Root",
							Key:   "",
						},
						Translations: map[string]types.ExtendedTranslation{
							"t-a": {
								Translation: types.Translation{
									Key:   "404Page",
									Title: "The missing page",
								},
								Values: map[string]types.TranslationValue{"en-GB": {
									LocaleID: "en-GB",
									Value:    "This page is missing",
								}},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			tt.project.CategoryTree = types.CreateCategoryTreeNode(tt.project.Categories)
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

			// {

			// 	pretty, err := Prettier(ts)
			// 	testza.AssertNoError(t, err)
			// 	if pretty == "" {
			// 		t.Fatal("empty output from prettier")
			// 	}
			// 	internal.MatchSnapshot(t, "ts", []byte(pretty))
			// }
			testza.AssertNotNil(t, ts)
			testza.AssertNotEqual(t, "", ts)
			internal.MatchSnapshot(t, "ts", w.Bytes())
		})
	}
}
