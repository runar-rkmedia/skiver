package handlers

import (
	"net/http"
	"time"

	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func getOrgForJoinID(db types.OrgStorage, joinID string) (*types.Organization, error) {

	if joinID == "" {
		return nil, ErrApiMissingArgument("JoinID")
	}

	orgs, err := db.GetOrganizations()
	if err != nil {
		return nil, ErrApiDatabase("Organization", err)
	}
	var org *types.Organization
	for _, o := range orgs {
		if o.JoinID == joinID {
			org = &o
			break
		}
	}
	if org == nil {
		return nil, ErrApiNotFound("Organization", joinID)

	}
	if org.JoinIDExpires.Before(time.Now()) {
		return nil, ErrApiNotFound("Organization", joinID)
	}
	return org, nil
}

func GetOrgForJoinID(db types.Storage) AppHandler {
	return func(rc requestContext.ReqContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
		joinID := GetParams(r).ByName("join-id")
		org, err := getOrgForJoinID(db, joinID)
		return org, err
	}
}

func JoinOrgFromJoinID(db types.Storage, pw PasswordKeeper) AppHandler {

	return func(rc requestContext.ReqContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
		joinID := GetParams(r).ByName("join-id")
		org, err := getOrgForJoinID(db, joinID)
		if err != nil {
			return nil, err
		}

		var joinInput models.JoinInput
		err = rc.ValidateBody(&joinInput, false)
		if err != nil {
			return nil, err
		}

		pass, err := pw.Hash(*joinInput.Password)
		if err != nil {
			rc.L.Error().Err(err).Msg("there was an error with hashing the password")
			return nil, ErrApiInternalError("Failure in password-creator", "Password", err)
		}
		u := types.User{
			Entity: types.Entity{
				CreatedAt:      time.Time{},
				CreatedBy:      "join",
				OrganizationID: org.ID,
			},
			UserName:              *joinInput.Username,
			Active:                true,
			Store:                 types.UserStoreLocal,
			TemporaryPassword:     false,
			PW:                    pass,
			CanCreateOrganization: false,
			CanCreateUsers:        false,
			CanCreateProjects:     true,
			CanCreateTranslations: true,
			CanCreateLocales:      false,
			CanUpdateOrganization: false,
			CanUpdateUsers:        false,
			CanUpdateProjects:     true,
			CanUpdateTranslations: true,
			CanUpdateLocales:      false,
			CanManageSnapshots:    true,
		}
		existingUsers := false
		{
			orgUsers, err := db.FindUsers(1, types.User{Entity: types.Entity{OrganizationID: org.ID}})
			if err != nil {
				return nil, ErrApiDatabase("User", err)
			}
			existingUsers = len(orgUsers) > 0
		}
		if existingUsers {
			u.CanUpdateOrganization = true
			// user is the first to join, should have organization-administrative permissions
		}

		user, err := db.CreateUser(u)
		if err != nil {
			return nil, ErrApiDatabase("User", err)
		}
		// TODO: loginUser
		out := types.LoginResponse{
			User:         user,
			Organization: *org,
			Ok:           true,
		}
		return out, nil

	}
}
