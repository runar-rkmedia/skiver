package handlers

import (
	"net/http"

	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type UserStorage interface {
	FindUsers(max int, filter ...types.User) (map[string]types.User, error)
	UpdateUser(id string, payload types.User) (types.User, error)
}

func ListUsers(db UserStorage, simpleusers bool) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		f := types.User{}
		f.OrganizationID = session.User.OrganizationID

		// Sanity-check
		if f.OrganizationID == "" {
			return nil, ErrApiInternalErrorMissingSession
		}

		users, err := db.FindUsers(0, f)
		if err != nil {
			return nil, err
		}
		if !simpleusers {
			return users, nil
		}

		u := map[string]models.SimpleUser{}
		for _, v := range users {
			u[v.ID] = models.SimpleUser(v.UserName)
		}
		return u, nil
	}
}

func ChangePassword(db UserStorage, pwHasher func(s string) ([]byte, error)) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		var j models.ChangePasswordInput
		err = rc.ValidateBody(&j, false)
		if err != nil {
			return nil, err
		}
		hashed, err := pwHasher(*j.Password)
		if err != nil {
			return nil, NewApiErr(err, http.StatusBadGateway, "UpdatePW:PwHash")
		}
		user, err := db.UpdateUser(session.User.ID, types.User{PW: hashed})
		if err != nil {
			return nil, err
		}
		rc.L.Info().Str("userID", user.ID).Msg("User changed password")

		return Ok, nil

	}
}

var Ok = models.OkResponse{Ok: boolPointer(true)}
