package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
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

func UpdateCategory(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}

		params := httprouter.ParamsFromContext(r.Context())
		cid := params.ByName("id")
		var j models.UpdateCategoryInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}
		if cid == "" {
			cid = j.ID
		}
		if cid == "" {
			return nil, ErrApiMissingArgument("ID")
		}

		c := types.Category{
			Key:         j.Key,
			Description: j.Description,
			Title:       j.Title,
		}
		c.CreatedBy = session.User.ID
		c.OrganizationID = session.Organization.ID
		category, err := db.UpdateCategory(cid, c)
		if err != nil {
			return nil, ErrApiDatabase("Category", err)
		}
		return category, nil
	}
}
