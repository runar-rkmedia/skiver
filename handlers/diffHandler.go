package handlers

import (
	"net/http"

	"github.com/r3labs/diff/v2"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/utils"
)

func GetDiff(exportCache Cache) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		var input models.DiffSnapshotInput
		err := rc.ValidateBody(&input, false)
		if err != nil {
			return nil, err
		}

		if areEqaul(*input.A, *input.B) {
			return nil, NewApiError("Cannot diff with equal objects", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
		}
		var a interface{}
		var b interface{}

		if input.A.Raw == nil {
			a, _, err = getExport(rc.L, exportCache, rc.Context.DB, importexport.ExportOptions{
				Project: *input.A.ProjectID,
				Tag:     input.A.Tag,
				Format:  input.Format,
			})
			if err != nil {
				return nil, err
			}
		} else {
			a = input.A.Raw
		}
		if input.B.Raw == nil {
			b, _, err = getExport(rc.L, exportCache, rc.Context.DB, importexport.ExportOptions{
				Project: *input.B.ProjectID,
				Tag:     input.B.Tag,
				Format:  input.Format,
			})
			if err != nil {
				return nil, err
			}
		} else {
			b = input.B.Raw
		}

		return utils.NewProjectDiff(a, b, input)
	}
}

// DiffOfObjects returns a changelog with options set for use with for instance i18n-json.
func DiffOfObjects(a, b interface{}) (diff.Changelog, error) {
	return diff.Diff(a, b, diff.DisableStructValues(), diff.AllowTypeMismatch(true))
}

// swagger:response DiffResponse
type diffResponse struct {
	// In: body
	Data utils.ProjectDiffResponse
}

func getKey(s models.SnapshotSelector) string {
	return *s.ProjectID + s.Tag
}
func areEqaul(a, b models.SnapshotSelector) bool {
	return getKey(a) == getKey(b)
}
