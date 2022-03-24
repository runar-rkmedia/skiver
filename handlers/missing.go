package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type MissingStorage interface {
	ReportMissing(key types.MissingTranslation) (*types.MissingTranslation, error)
	GetMissingKeysFilter(max int, filter ...types.MissingTranslation) (map[string]types.MissingTranslation, error)
}

func GetMissing(db MissingStorage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		m, err := db.GetMissingKeysFilter(0)
		if err != nil {
			return nil, NewApiErr(err, http.StatusBadGateway, string(requestContext.CodeErrReportMissing))
		}
		return m, err
	}
}
func PostMissing(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		params := GetParams(r)

		projectinput := params.ByName("project")
		localeinput := params.ByName("locale")
		if projectinput == "" {
			return nil, ErrApiMissingArgument("project")
		}
		if localeinput == "" {
			return nil, ErrApiMissingArgument("locale")
		}

		project, err := db.GetProjectByIDOrShortName(projectinput)
		if err != nil {
			return nil, ErrApiDatabase("Project", err)
		}
		if project == nil {
			return nil, ErrApiNotFound("Project", projectinput)
		}

		// The default-settings of i18next's AddMissing request does not add the correct Content-Type.
		// Just to be nice, we attempt to read the body anyway...
		rc.ContentKind = requestContext.OutputJson
		body, err := io.ReadAll(r.Body)
		if err != nil {

			return nil, NewApiErr(err, http.StatusBadGateway, string(requestContext.CodeErrReadBody))
		}
		var j models.ReportMissingInput
		err = rc.ValidateBytes(body, &j)
		if err != nil {
			return nil, nil
		}
		var errs []string
		var mts []types.MissingTranslation
		for k := range j {
			splitted := strings.Split(k, ".")
			category := splitted[0]
			var translation string
			if len(splitted) > 0 {
				translation = splitted[1]
			}
			// requires go 1.18
			// category, translation, _ := strings.Cut(k, ".")
			mt := types.MissingTranslation{
				Locale:          localeinput,
				Project:         projectinput,
				Translation:     translation,
				Category:        category,
				LatestUserAgent: r.UserAgent(),
			}
			mt.OrganizationID = project.OrganizationID

			session, err := GetRequestSession(r)
			if session.User.ID != "" && session.Organization.ID == project.OrganizationID {
				mt.CreatedBy = session.User.ID
				mt.OrganizationID = session.User.ID
			}
			if mt.CreatedBy == "" {
				mt.CreatedBy = "anonymous"
			}

			mtt, err := db.ReportMissing(mt)
			if err != nil {
				rc.L.Err(err).Interface("mt", mt).Msg("failed to report message")
				errs = append(errs, err.Error())
				continue
			}
			mts = append(mts, *mtt)
		}

		if errs != nil && len(errs) > 0 {

			return nil, NewApiErr(fmt.Errorf("%d/%d missing translations failed to report: %#v", len(errs), len(j), errs), http.StatusBadGateway, string(requestContext.CodeErrProject))
		}
		return mts, nil
	}
}
