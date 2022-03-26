package handlers

import (
	"net/http"

	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func GetCategory(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		categories, err := db.GetCategories()
		if err != nil {
			return nil, ErrApiDatabase("Category", err)
		}
		return categories, nil
	}
}
func PostCategory(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		var j models.CategoryInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}

		c := types.Category{
			// TranslationInput: j,
			ProjectID:   *j.ProjectID,
			Key:         *j.Key,
			Description: j.Description,
			Title:       *j.Title,
		}
		c.CreatedBy = session.User.ID
		c.OrganizationID = session.Organization.ID
		category, err := db.CreateCategory(c)
		if err != nil {
			return nil, ErrApiDatabase("Category", err)
		}
		return category, nil
	}
}
