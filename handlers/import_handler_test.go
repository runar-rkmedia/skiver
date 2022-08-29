package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/go-common/logger"
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
			"simple, single level",
			nil,
			`
en:
  foo: bar
	foo_baz: foobar
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
		l := logger.GetLoggerWithLevel("test", "fatal")
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
			s := strings.ReplaceAll(tt.fields, "\t", "  ")
			err = internal.YamlUnmarshalAllowTabs(tt.fields, &j)
			testza.AssertNoError(t, err)
			// err = yaml.Unmarshal([]byte(s), &j)
			r, _ := http.NewRequest(http.MethodPost, "", strings.NewReader(s))
			r.Header.Set("Content-Type", "text/vnd.yaml")
			impo, err := ImportIntoProject(l, bb, "i18n", base.CreatedBy, project, "", []byte(s), r, ImportIntoProjectOptions{NoDryRun: true})
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

			testza.AssertNoError(t, err)
			export, err := importexport.ExportI18N(ep, importexport.ExportI18NOptions{})
			testza.AssertNoError(t, err)

			toMap := export.ToMap()
			fmt.Println(j)
			if err := internal.Compare("Export of result of import should match import-input", toMap, j); err != nil {
				t.Error(err)
			}
		})
	}
}
