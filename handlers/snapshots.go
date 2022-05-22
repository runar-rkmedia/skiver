package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/uploader"
	"github.com/runar-rkmedia/skiver/utils"
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

func uploadSnapshot(l logger.AppLogger, uploaders []uploader.FileUploader, tag string, snap types.ProjectSnapshot) ([]types.UploadMeta, error) {
	if len(uploaders) == 0 {
		return nil, nil
	}

	var tags []string
	if _, err := semver.StrictNewVersion(tag); err == nil {
		ts, err := utils.ResolveSemver(tag)
		if err != nil {
			return nil, err
		}
		for _, v := range ts {
			tags = append(tags, v.String())
		}
	} else {
		tags = []string{tag}
	}
	if len(tags) == 0 {
		l.Fatal().Msg("Tags cannot be empty, the tag-input was: " + tag)
	}
	locales := utils.SortedMapKeys(snap.Project.Locales)

	localeKeys := []importexport.LocaleKeyEnum{
		importexport.LocaleKeyEnumIETF,
		importexport.LocaleKeyEnumISO1,
		importexport.LocaleKeyEnumISO2,
		importexport.LocaleKeyEnumISO3,
	}

	var m []types.UploadMeta
	uploadedLocales := map[string]struct{}{}
	for _, localeKey := range localeKeys {
		i18n, err := importexport.ExportExtendedProjectToI18Next(l, snap.Project,
			locales,
			localeKey,
		)
		if err != nil {
			return nil, err
		}

		for locale, content := range i18n {
			// a locale (the struct) have localekeys which may or may not be unique.
			// THerefore, we track if we already have uploaded that locale for the localeKey.
			// for instance, the norwegian language has the same locale-code for iso_639_2 and 3:
			// "iso_639_1": "nb",
			// "iso_639_2": "nob",
			// "iso_639_3": "nob",
			// "ietf": "nb-NO",
			// TODO: it may save some cpu-resources if we did not have to regenerate the same i18n-struct
			// but it is probably ok
			if _, ok := uploadedLocales[locale]; ok {
				continue
			}
			var localeID string
			for _, l := range snap.Project.Locales {
				switch localeKey {
				case importexport.LocaleKeyEnumIETF:
					if l.IETF == locale {
						localeID = l.ID
					}
				case importexport.LocaleKeyEnumISO1:
					if l.Iso639_1 == locale {
						localeID = l.ID
					}
				case importexport.LocaleKeyEnumISO2:
					if l.Iso639_2 == locale {
						localeID = l.ID
					}
				case importexport.LocaleKeyEnumISO3:
					if l.Iso639_3 == locale {
						localeID = l.ID
					}
				}

			}
			uploadedLocales[locale] = struct{}{}
			aliases := make([]string, len(tags))
			for i := 0; i < len(tags); i++ {
				aliases[i] = fmt.Sprintf("%s_%s_%s_%s.json", snap.OrganizationID, snap.Project.ID, locale, tags[i])
			}

			b, err := json.Marshal(content)
			if err != nil {
				return m, err
			}

			for _, u := range uploaders {
				r := bytes.NewReader(b)
				um, err := u.AddPublicFileWithAliases(aliases, r, r.Size(), "application/json", "")
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(um); i++ {
					um[i].Locale = localeID
					um[i].LocaleKey = localeKey.String()
					um[i].Tag = tags[i]
				}

				m = append(m, um...)
			}

		}
	}

	return m, nil
}

func uploadSnapshotForProjectAndUpdateIt(db types.Storage, l logger.AppLogger, uploaders []uploader.FileUploader, tag string, snap types.ProjectSnapshot) {
	var uploadMetas []types.UploadMeta
	// Upload the snapshots, if there are any uploaders configured
	if len(uploaders) > 0 {
		ums, err := uploadSnapshot(l, uploaders, tag, snap)
		if err != nil {
			l.Error().Err(err).Msg("Failed to upload snapshot")
			return
		}
		uploadMetas = ums
	}
	p, err := db.GetProject(snap.Project.ID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to retrieve project from database with new uploads")
		return
	}
	if p == nil {
		l.Error().Err(err).Msg("project was nil when attempting to retrieve it after uploading snapshots")
		return

	}
	pSnap := p.Snapshots[tag]
	pSnap.UploadMeta = uploadMetas
	// In previous snapshots, we should remove all UploadMeta for all tags that
	// have the same url This is mostly used when we are overwriting files for
	// instance when using semantic versioning and the we then overwrite for
	// instance the major-version file, since this is a newer tag.
	urls := map[string]struct{}{}
	for _, um := range uploadMetas {
		urls[um.URL] = struct{}{}
	}
	for k, v := range p.Snapshots {
		if len(v.UploadMeta) > 0 {
			var um []types.UploadMeta
			for i := 0; i < len(v.UploadMeta); i++ {
				if _, ok := urls[v.UploadMeta[i].URL]; ok {
					continue
				}
				um = append(um, v.UploadMeta[i])
			}
			v.UploadMeta = um
			p.Snapshots[k] = v
		}
	}
	p.Snapshots[tag] = pSnap
	_, err = db.UpdateProject(p.ID, *p)
	if err != nil {
		l.Error().Err(err).Msg("Failed to update database with new uploads for project")
		return
	}
	l.Info().Interface("UploadMeta", uploadMetas).Msg("Snapshot was updated with referances to externally uploaded snapshots")
}

// PostSnapshot creates a snapshot if there does not exist one already with the same hash.
func PostSnapshot(uploaders []uploader.FileUploader) AppHandler {
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
			return nil, ErrApiDatabase("Project", err)
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
		if len(uploaders) > 0 {
			go uploadSnapshotForProjectAndUpdateIt(rc.Context.DB, rc.L, uploaders, *j.Tag, s)
		}

		return updatedProject, err
	}
}

var (
	CodeInternalServerError requestContext.ErrorCodes = "Internal server error"
)
