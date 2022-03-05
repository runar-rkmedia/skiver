package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func NewApiError(msg string, statusCode int, code string, details ...interface{}) requestContext.APIError {
	return NewApiErr(errors.New(msg), statusCode, code, details...)
}
func NewApiErr(err error, statusCode int, code string, details ...interface{}) requestContext.APIError {
	e := requestContext.APIError{
		Err: requestContext.Error{
			Code:    requestContext.ErrorCodes(code),
			Message: err.Error(),
		},
		InternalError: err,
		StatusCode:    statusCode,
	}
	switch len(details) {
	case 0:
		break
	case 1:
		e.Details = details[0]
	default:
		lg := logger.GetLogger("Developer-mistake")
		lg.Error().Msg("Recieved more than one detail-object. This is a developer-mistake and should be fixed.")
		m := map[string]interface{}{}
		for i := 0; i < len(details); i++ {
			m[fmt.Sprintf("___%d___", i)] = details[i]
		}
		e.Details = m
	}

	return e
}

// PostSnapshot creates a snapshot if there does not exist one already with the same hash.
func PostSnapshot() AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (output interface{}, err error) {

		session, err := GetRequestSession(r)
		if err != nil {
			return
		}

		var j models.CreateSnapshotInput

		if err = rc.ValidateBody(&j, false); err != nil {
			err = NewApiErr(err, http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
			return
		}
		project, err := rc.Context.DB.GetProject(*j.ProjectID)
		if err != nil {
			err = NewApiErr(err, http.StatusInternalServerError, string(requestContext.CodeErrProject))
			return
		}
		if project == nil {
			err = NewApiError("Project not found", http.StatusNotFound, string(requestContext.CodeErrNotFoundProject))
			return
		}
		if s, ok := project.Snapshots[*j.Tag]; ok {
			err = NewApiError("The tag already exists", http.StatusNotFound, "TaxExists", s)
			return
		}

		projectLocaleMap := map[string]bool{}
		for k, v := range project.LocaleIDs {
			if !v.Publish {
				continue
			}
			projectLocaleMap[k] = true

		}
		if len(projectLocaleMap) == 0 {
			err = NewApiError("There are no published locales for this project", http.StatusBadRequest, "No published project-locales")
			return
		}

		ep, err := project.Extend(rc.Context.DB, types.ExtendOptions{
			ByID:      true,
			ByKeyLike: false,
			LocaleFilterFunc: func(l types.Locale) bool {
				return projectLocaleMap[l.ID]
			},
			ErrOnNoLocales: true,
		})
		if err != nil {
			err = NewApiErr(err, http.StatusInternalServerError, string(requestContext.CodeErrProject))
			return
		}
		s, err := ep.CreateSnapshot(session.User.ID)
		if err != nil {
			rc.WriteErr(err, requestContext.CodeErrProject)
			return
		}

		var projectSnapshot types.ProjectSnapshot
		existing, err := rc.Context.DB.FindOneSnapshot(s)
		if err != nil {
			return
		}
		if existing != nil {
			projectSnapshot = *existing
		} else {
			ss, err := rc.Context.DB.CreateSnapshot(s)
			if err != nil {
				err = NewApiErr(err, http.StatusInternalServerError, string(requestContext.CodeErrProject))
				return nil, err
			}
			projectSnapshot = ss
		}

		if project.Snapshots == nil {
			project.Snapshots = map[string]types.ProjectSnapshotMeta{}
		}
		project.Snapshots[*j.Tag] = types.ProjectSnapshotMeta{
			Description: j.Description,
			CreatedBy:   session.User.ID,
			SnapshotID:  projectSnapshot.ID,
			CreatedAt:   time.Now(),
			Hash:        s.ProjectHash,
		}

		updatedProject, err := rc.Context.DB.UpdateProject(project.ID, *project)
		if err != nil {
			return
		}

		return updatedProject, err
	}
}

var (
	CodeInternalServerError requestContext.ErrorCodes = "Internal server error"
)
