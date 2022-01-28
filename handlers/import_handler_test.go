package handlers

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/bboltStorage"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/types"
)

// Tests by using a mocked DB, setting up with defaults, and running the import
// over the input, then checking to see if a export of the whole project
// matches the input-data.
func TestImportHandler(t *testing.T) {
	tests := []struct {
		name       string
		localeHint *types.Locale
		fields     string
	}{
		{
			"simple",
			nil,
			`
en:
  foo: 
	  bar
`,
		},
		{
			"Nested categories",
			nil,
			`
en:
  General:
    Forms:
      Buttons:
        GoToCheckout: Go to checkout ({{count}})
    Welcome: Welcome, {{user}}
nb:
  Abc: abc
	Abc_reverse: cba
  General:
    Forms:
      Buttons:
        GoToCheckout: Gå til utsjekk ({{count}})
`,
		},
	}

	for _, tt := range tests {
		internal.NewMockTimeNow()
		t.Run(tt.name, func(t *testing.T) {

			// 1. Setup
			bb := bboltStorage.NewMockDB(t)
			err := bb.StandardSeed()
			testza.AssertNoError(t, err)
			base := types.Project{}
			base.CreatedBy = "jim"
			base.OrganizationID = "org-123"
			base.ID = "proj-123"
			base.ShortName = "proj"
			base.Title = "proj"
			project, err := bb.CreateProject(base)
			testza.AssertNoError(t, err)

			// 2. Import from input
			var j map[string]interface{}
			err = internal.YamlUnmarshalAllowTabs(tt.fields, &j)
			testza.AssertNoError(t, err)
			impo, err := ImportIntoProject(bb, "i18n", base.CreatedBy, project, "", false, j)
			testza.AssertNil(t, err)
			testza.AssertNotNil(t, impo)

			// 3. Export the whole project.
			if p, err := bb.GetProject(project.ID); err == nil {
				project = *p
			} else {
				t.Error(err)
			}
			ep, err := project.Extend(bb)
			testza.AssertNoError(t, err)

			export, err := importexport.ExportI18N(ep, importexport.ExportI18NOptions{})
			testza.AssertNoError(t, err)
			yml := internal.MustYaml(export)
			if strings.Contains(yml, types.RootCategory) {
				t.Errorf("Root-category (%s) should not exist in the export, but was found:\n%s", types.RootCategory, yml)
			}

			toMap := export.ToMap()
			if err := internal.Compare("node-map", toMap, j); err != nil {
				t.Error(err)
			}
		})
	}
}
