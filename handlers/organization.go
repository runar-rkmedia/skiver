package handlers

import (
	"net/http"
	"time"

	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

func GetOrganization(db types.OrgStorage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {

		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		if session.User.CanCreateOrganization {
			orgs, err := db.GetOrganizations()
			if err != nil {
				return orgs, ErrApiDatabase("Project", err)
			}
			return orgs, nil

		}
		org, err := db.GetOrganization(session.User.OrganizationID)
		out := map[string]*types.Organization{org.ID: org}
		if err != nil {
			return out, ErrApiDatabase("Project", err)
		}
		return out, nil
	}
}
func CreateOrganization(db types.OrgStorage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {

		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		if !session.User.CanCreateOrganization {
			return nil, ErrApiNotAuthorized("Organization", "create")
		}
		var j models.OrganizationInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}

		l := types.Organization{
			Title: *j.Title,
			// Initially set to expire within 30 days.
			JoinIDExpires: time.Now().Add(30 * 24 * time.Hour),
		}
		l.JoinID, err = utils.GetRandomName()
		if err != nil {
			return nil, NewApiErr(err, http.StatusBadGateway, "GetRandomName")
		}
		l.CreatedBy = session.User.ID
		org, err := db.CreateOrganization(l)
		if err != nil {
			return org, ErrApiDatabase("Organization", err)
		}
		return org, nil
	}
}
func UpdateOrganization(db types.OrgStorage) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		session, err := GetRequestSession(r)
		if err != nil {
			return nil, err
		}
		session.User.CanUpdateOrganization = true
		if !session.User.CanUpdateOrganization {
			return nil, ErrApiNotAuthorized("Organization", "update")
		}
		var j models.UpdateOrganizationInput
		if err := rc.ValidateBody(&j, false); err != nil {
			return nil, err
		}
		if *j.ID != session.User.OrganizationID {
			return nil, ErrApiNotAuthorized("Organization other than own organization", "update")
		}
		if j.JoinIDExpires != nil {
			org, err := db.GetOrganization(session.User.OrganizationID)
			if err != nil {
				return nil, ErrApiDatabase("Organization", err)
			}
			payload := types.UpdateOrganizationPayload{
				JoinID:        &j.JoinID,
				JoinIDExpires: (*time.Time)(j.JoinIDExpires),
				UpdatedBy:     session.User.ID,
			}

			if payload.JoinID == nil || *payload.JoinID == "" {
				id, err := utils.GetRandomName()
				if err != nil {
					return nil, NewApiErr(err, http.StatusBadGateway, "GetRandomName")
				}
				payload.JoinID = &id
			} else if org.JoinID != j.JoinID {
				return nil, ErrApiNotFound("JoinId", j.JoinID)
			}
			updated, err := db.UpdateOrganization(org.ID, payload)
			if err != nil {
				return nil, ErrApiDatabase("Organization", err)
			}
			return updated, err
		}
		return nil, NewApiError("No changes required", http.StatusAlreadyReported, "NoChanges")
	}
}
