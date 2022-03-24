package handlers

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func DeleteTranslation() AppHandler {

	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}

		params := httprouter.ParamsFromContext(r.Context())
		tid := params.ByName("id")
		var j models.DeleteInput
		err = rc.ValidateBody(&j, false)
		if err != nil {
			return nil, err
		}
		if tid == "" {
			return nil, NewApiError("Missing id", http.StatusBadRequest, string(requestContext.CodeErrIDEmpty))
		}
		var deleteTime *time.Time
		if j.Undelete {
			deleteTime = nil
		} else {
			if j.ExpiryDate == nil {
				t := time.Now().Add(time.Hour * 24 * 365 * 290)
				deleteTime = &t
			} else {
				deleteTime = (*time.Time)(j.ExpiryDate)
				if deleteTime.Sub(time.Now()) < time.Hour*23+time.Minute*55 {
					return nil, NewApiError("ExpiryDate must be at least 24 hours into the future", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
				}
			}
		}
		return rc.Context.DB.SoftDeleteTranslation(tid, session.User.ID, deleteTime)
	}
}

func UpdateTranslation() AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		if !session.User.CanUpdateTranslations {
			return nil, ErrApiNotAuthorized("Translation", "update")
		}
		params := httprouter.ParamsFromContext(r.Context())
		tid := params.ByName("id")
		var j models.UpdateTranslationInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}
		if tid == "" {
			tid = *j.ID
		}

		if tid == "" {
			rc.WriteError("Missing id", requestContext.CodeErrIDEmpty)
			return nil, ErrApiMissingArgument("ID")
		}
		existing, err := rc.Context.DB.GetTranslation(tid)
		if err != nil {
			rc.WriteErr(err, requestContext.CodeErrTranslation)
		}
		if existing == nil || existing.OrganizationID != session.User.OrganizationID {
			return nil, ErrApiNotFound("Translation", tid)
		}

		t := types.Translation{
			Key:         existing.Key,
			Title:       *j.Title,
			Description: *j.Description,
		}
		if j.Variables != nil {
			if v, ok := j.Variables.(map[string]interface{}); ok {

				t.Variables = v
			} else {
				return existing, ErrApiInputValidation("key variables are invalid", "Translation")
			}
		}

		updated, err := rc.Context.DB.UpdateTranslation(tid, t)
		if err != nil {

		}
		return updated, ErrApiDatabase("Translation", err)

	}
}
func CreateTranslation() AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}

		if !session.User.CanCreateTranslations {
			return nil, ErrApiNotAuthorized("Translation", "create")
		}
		var j models.TranslationInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}

		t := types.Translation{
			// TranslationInput: j,
			CategoryID:  *j.CategoryID,
			Key:         *j.Key,
			Description: j.Description,
			Title:       j.Title,
			Variables:   j.Variables,
		}
		t.CreatedBy = session.User.ID
		t.OrganizationID = session.Organization.ID
		translation, err := rc.Context.DB.CreateTranslation(t)
		return translation, ErrApiDatabase("Translation", err)
	}
}
func GetTranslations() AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {

		translations, err := rc.Context.DB.GetTranslations()
		if err != nil {
			return translations, NewApiErr(err, http.StatusBadRequest, string(requestContext.CodeErrTranslation))
		}
		return translations, err
	}
}
