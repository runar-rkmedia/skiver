package importexport

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

func TestImportBug(t *testing.T) {
	tests := []struct {
		name       string
		localeHint *types.Locale
		input      string
		result     func(*testing.T, *Import, []Warning, error)
	}{
		{

			// This was a bug related to the input below resulting in a
			// ghost-category Approval.Action, which should not exist. This only
			// happened when what should have been considered a translation had an
			// empty value. The code then considered this empty translation a
			// category instead.
			"No context with empty value",
			nil,
			`
en:
  Approval:
		Action: ''
		Action_MyContext: 'Value'
`,
			func(t *testing.T, i *Import, w []Warning, e error) {
				// There should be only a single category for this input
				testza.AssertEqual(t, utils.SortedMapKeys(i.Categories), []string{"Approval"})
				testza.AssertEqual(t, len(i.Categories), 1)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := types.Project{}
			base.ID = "proj-123"
			base.CreatedBy = "jim"
			base.OrganizationID = "org-123"
			var j map[string]interface{}
			err := internal.YamlUnmarshalAllowTabs(tt.input, &j)
			if err != nil {
				t.Errorf("TEST-INPUT_ERROR: Failed to unmarshal: %s %s", err, tt.input)
				return
			}
			got, warnings, err := ImportI18NTranslation(types.Test_locales, tt.localeHint, base, "test-import", j)
			tt.result(t, got, warnings, err)
		})
	}
}
