package handlers

import (
	"net/http"

	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func GetProjects(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, ErrApiInternalErrorMissingSession
		}
		projectFilter := types.Project{}
		projectFilter.OrganizationID = session.Organization.ID
		projects, err := db.FindProjects(0, projectFilter)
		if err != nil {
			return nil, ErrApiDatabase("Project", err)
		}
		return projects, err
	}
}
func UpdateProject(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, ErrApiInternalErrorMissingSession
		}

		if !session.User.CanUpdateProjects {
			return nil, ErrApiNotAuthorized("Project", "update")
		}
		var j models.UpdateProjectInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}

		p, err := db.GetProject(*j.ID)
		if err != nil {
			return nil, ErrApiDatabase("Project", err)
		}
		if p == nil || session.User.OrganizationID != p.OrganizationID {
			return nil, ErrApiNotFound("Project", *j.ID)
		}
		payload := types.Project{
			Title:       j.Title,
			Description: j.Description,
			ShortName:   j.ShortName,
		}
		payload.UpdatedBy = session.User.ID
		if len(j.Locales) > 0 {
			payload.LocaleIDs = map[string]types.LocaleSetting{}
			for lID, ls := range j.Locales {
				payload.LocaleIDs[lID] = types.LocaleSetting{
					Enabled:         ls.Enabled,
					Publish:         ls.Publish,
					AutoTranslation: ls.AutoTranslation,
				}

			}
		}
		project, err := db.UpdateProject(*j.ID, payload)
		if err != nil {
			return nil, ErrApiDatabase("Project", err)
		}

		return project, err
	}
}
func CreateProject(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, ErrApiInternalErrorMissingSession
		}

		if !session.User.CanCreateProjects {
			return nil, ErrApiNotAuthorized("Project", "create")
		}
		var j models.ProjectInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}

		p := types.Project{
			Title:       *j.Title,
			Description: j.Description,
			ShortName:   *j.ShortName,
			LocaleIDs:   map[string]types.LocaleSetting{},
		}
		if len(j.Locales) > 0 {
			for lID, ls := range j.Locales {
				p.LocaleIDs[lID] = types.LocaleSetting{
					Enabled:         ls.Enabled,
					Publish:         ls.Publish,
					AutoTranslation: ls.AutoTranslation,
				}
			}
		}

		p.CreatedBy = session.User.ID
		p.OrganizationID = session.Organization.ID
		project, err := db.CreateProject(p)
		if err != nil {
			return nil, ErrApiDatabase("Project", err)
		}
		return project, err
	}
}
