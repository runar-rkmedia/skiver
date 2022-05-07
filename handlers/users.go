package handlers

import (
	"net/http"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type UserStorage interface {
	FindUsers(max int, filter ...types.User) (map[string]types.User, error)
	UpdateUser(id string, payload types.UpdateUserPayload) (types.User, error)
	GetUser(userId string) (*types.User, error)
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

type PasswordKeeper interface {
	Verify(hash []byte, pw string) (bool, error)
	Hash(pw string) ([]byte, error)
}

func ChangePassword(db UserStorage, pwKeeper PasswordKeeper, sessionReplace SessionReplacer) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		var j models.ChangePasswordInput
		err := rc.ValidateBody(&j, false)
		if err != nil {
			return nil, err
		}
		if *j.Password == j.NewPassword {
			return nil, NewApiError("Both the current and new password was the same", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
		}
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		u, err := db.GetUser(session.User.ID)
		if err != nil {
			return nil, err
		}
		ok, err := pwKeeper.Verify(u.PW, *j.Password)
		if err != nil {
			return nil, NewApiErr(err, http.StatusBadGateway, "PwVerifyErr")
		}
		if !ok {
			return nil, NewApiError("Incorrect password", http.StatusBadRequest, "PwMismatch")
		}
		hashed, err := pwKeeper.Hash(j.NewPassword)
		if err != nil {
			return nil, NewApiErr(err, http.StatusBadGateway, "UpdatePW:PwHash")
		}
		payload := types.UpdateUserPayload{PW: &hashed}
		payload.UpdatedBy = session.User.ID
		f := false
		payload.TemporaryPassword = &f
		user, err := db.UpdateUser(session.User.ID, payload)
		if err != nil {
			return nil, err
		}
		sessionReplace.UpdateAllSessionsForUser(user.ID, user)
		rc.L.Info().Str("userID", user.ID).Msg("User changed password")

		return Ok, nil

	}
}

type SessionCreator interface {
	NewSession(user types.User, organization types.Organization, userAgent string, opts ...types.UserSessionOptions) (s types.Session)
}

type SessionReplacer interface {
	UpdateAllSessionsForUser(userId string, user types.User) error
}

func CreateToken(sessionCreator SessionCreator) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		var j models.CreateTokenInput
		err := rc.ValidateBody(&j, false)
		if err != nil {
			return nil, err
		}
		d := time.Duration(*j.TTLHours * int64(time.Hour))
		if d > time.Hour*24*365*5 {
			return nil, NewApiError("Duration is above current limit. Please use a shorter duration", http.StatusBadRequest, "TokenMaxDuration")
		}
		if d < time.Minute {
			return nil, NewApiError("Duration is below current limit. Please use a longer duration", http.StatusBadRequest, "TokenMaxDuration")
		}
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return "", err
		}
		newSession := sessionCreator.NewSession(session.User, session.Organization, *j.Description, types.UserSessionOptions{TTL: d})
		if newSession.Token == "" {
			rc.L.Error().
				Str("userID", session.User.ID).
				Str("userOrgID", session.User.OrganizationID).
				Msg("The session created for user was unexpectedly empty")
			return "", NewApiError("Sesssion was unexpectedly empty", http.StatusBadGateway, "EmptySession")
		}
		if err != nil {
			return nil, err
		}
		response := models.TokenResponse{
			Description: newSession.UserAgent,
			Expires:     strfmt.DateTime(newSession.Expires),
			Issued:      strfmt.DateTime(newSession.Issued),
			Token:       newSession.Token,
		}
		return response, nil

	}
}

var Ok = models.OkResponse{Ok: boolPointer(true)}
