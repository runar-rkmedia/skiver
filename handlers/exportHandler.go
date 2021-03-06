package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func getExport(l logger.AppLogger, exportCache Cache, db types.Storage, opt importexport.ExportOptions) (toWriter interface{}, contentType string, err error) {
	projectKey := opt.Project
	locales := opt.Locales
	localeKey := opt.LocaleKey
	format := opt.Format
	tag := opt.Tag

	flatten := !opt.NoFlatten
	switch format {
	case "typescript":
		break
	case "raw":
		break
	case "i18n":
		break
	default:

		validFormats := []string{"i18n", "raw", "typescript"}
		valid := false
		for _, v := range validFormats {
			if format == v {
				valid = true
			}
		}
		if !valid {
			err = NewApiErr(fmt.Errorf("invalid format: %s. Valid formats are: %s", format, validFormats), http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
			return
		}
	}
	if projectKey == "" {
		err = NewApiError("A project must be selected", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
		return
	}
	cacheKeys := []string{opt.InOrg, format, localeKey, tag}
	cacheKeys = append(cacheKeys, locales...)
	cacheKeys = append(cacheKeys, projectKey)
	sort.Strings(cacheKeys)
	cacheKey := strings.Join(cacheKeys, "%")
	if flatten {
		cacheKey += "F"
	}
	if exportCache != nil {
		if v, ok := exportCache.Get(cacheKey); ok {
			return v, "", nil
		}
	}

	defer func() {
		if exportCache == nil {
			return
		}
		if err != nil {
			return
		}
		if tag != "" {
			exportCache.SetDefault(cacheKey, toWriter)
		} else {
			// A very short cache-time for exports that are pulled directly from live-data.
			exportCache.Set(cacheKey, toWriter, time.Second*3)
		}
	}()

	ps, err := db.GetProjectByIDOrShortName(projectKey)
	if err != nil {
		err = ErrApiDatabase("Project", err)
		return
	}
	if ps == nil {
		err = ErrApiNotFound("Project", projectKey)
		return
	}
	// In the future, inOrg is required. for now it is optional for clients expecting skiver before v0.5.4
	if opt.InOrg != "" {
		if ps.OrganizationID != opt.InOrg {
			// retry by using looking up the name in the org:
			org, err := db.FindOrganizationByIdOrTitle(opt.InOrg)
			if err != nil {
				err = ErrApiDatabase("Organization", err)
				return nil, "", err
			}
			if org == nil {
				err = ErrApiNotFound("Project", projectKey)
				return nil, "", err
			}
		}
	}
	var ep types.ExtendedProject

	if tag != "" {
		var snapshotMeta types.ProjectSnapshotMeta
		if s, ok := ps.Snapshots[tag]; ok {
			snapshotMeta = s
		}
		if snapshotMeta.SnapshotID == "" {
			// TODO: check if tag is semver and resolve
			l.Warn().Msg("Not implemented semver-check on tags")

		}
		if snapshotMeta.SnapshotID == "" {
			err = NewApiError("Tag not found", http.StatusNotFound, "TagNotFound", ps.Snapshots)
			return
		}
		s, err := db.GetSnapshot(snapshotMeta.SnapshotID)
		if err != nil {
			err = NewApiErr(err, http.StatusInternalServerError, string(requestContext.CodeErrSnapshot))
			return toWriter, "", err
		}
		if s == nil {
			err = NewApiError("The snapshot was not found", http.StatusNotFound, string(CodeInternalServerError))
			return toWriter, "", err
		}
		ep = s.Project

	} else {
		// These options shouild be the same as when creating snapshots.
		ep, err = ps.Extend(db, types.ExtendOptions{
			LocaleFilter:   locales,
			ByKeyLike:      false,
			ByID:           true,
			ErrOnNoLocales: true,
			LocaleFilterFunc: func(locale types.Locale) bool {
				for k, v := range ps.LocaleIDs {
					if locale.ID != k {
						continue
					}
					if v.Publish {
						return true
					}
				}
				return false
			},
		})

	}
	if err != nil {
		err = fmt.Errorf("Error extending project '%s' (%s): %w", ep.Title, ep.ID, err)
		return
	}
	writer, contentType, err := importexport.ExportExtendedProject(l, ep, opt.Locales, importexport.LocaleKeyEnum{}.From(opt.LocaleKey),
		importexport.Format{}.From(opt.Format),
		opt.Locales)
	if err != nil {
		err = NewApiErr(err, http.StatusBadGateway, "ExportExtended")
	}
	return writer, contentType, err
}

func GetExport(
	exportCache Cache,
) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (toWriter interface{}, apiErr error) {
		AddAccessControl(r, rw)

		// In versions prior to v0.5.4, exports did not not include this organization-id.
		// It an upcoming release, we plan to allow the users to set CORS-settings per organization-level, and also per project-level.
		// In addition, we may want to require an api-key to get the the export, so that organizations cannot easily list other orgs translation-files
		// CORS-requests by default do not include query-parameters.
		// This requires that the path includes both organization and project.
		//
		// To not break existing clients, we keep the old behaviour. Since we are not in v1 yet, backwards-compatibility is not a
		// requirement, but we try to be nice and keep this behaviour until all clients have upgraded and are ok with this change.
		params := GetParams(r)
		orgKey := params.ByName("org")
		projectKey := params.ByName("project")
		// If orgKey is exclicidly set, use the current user's organization.
		// We could of course allow omiting the the org in this case, but this is a source for production-errors for clients
		// if they use that route for customers, in the belief that that url would be viewable anonoumously.
		if orgKey == "me" {
			session, err := GetRequestSession(r)
			if err != nil {
				return nil, err
			}
			orgKey = session.Organization.ID

			// This check will be removed lated on
		} else if strings.Contains(orgKey, "p=") {
			orgKey = ""
		}

		q, err := ExtractParams(r)
		format := "i18n" // Default format
		localeKey := ""
		tag := ""
		var locales []string
		flatten := true
		for k, v := range q {
			switch strings.ToLower(k) {
			case "locale", "l":
				locales = v
			case "tag", "t":
				if len(v) > 1 {
					rc.WriteError("tag specified more than once", requestContext.CodeErrInputValidation)
					return
				}
				tag = v[0]
			case "format", "f":
				if len(v) > 1 {
					rc.WriteError("format specified more than once", requestContext.CodeErrInputValidation)
					return
				}
				format = v[0]
			case "project", "p":
				rc.L.Warn().
					Str("org", orgKey).
					Str("project", projectKey).
					Str("raw", v[0]).
					Msg("A client requested an export with project using the deprecated query-option")
				if projectKey != "" {
					rc.WriteError("project is already specified", requestContext.CodeErrInputValidation)
					return

				}
				if len(v) > 1 {
					rc.WriteError("project specified more than once", requestContext.CodeErrInputValidation)
					return
				}
				projectKey = v[0]
			case "locale_key":
				if len(v) > 1 {
					rc.WriteError("locale_key specified more than once", requestContext.CodeErrInputValidation)
					return
				}
				localeKey = v[0]
			case "no_flatten":
				flatten = false
			}
		}

		toWriter, contentType, err := getExport(rc.L, exportCache, rc.Context.DB, importexport.ExportOptions{
			InOrg:     orgKey,
			Project:   projectKey,
			Locales:   locales,
			LocaleKey: localeKey,
			Format:    format,
			Tag:       tag,
			NoFlatten: !flatten,
		})
		if err != nil {
			return toWriter, err
		}
		if toWriter == nil {
			return nil, NewApiError("No content", http.StatusNoContent, "NoContent:export")
		}
		if contentType != "" {
			rw.Header().Set("Content-Type", contentType)
			rw.Write(toWriter.([]byte))
			return nil, nil
		}
		return toWriter, err
	}
}
