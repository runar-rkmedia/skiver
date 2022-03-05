package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

func GetExport(
	exportCache Cache,
) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (toWriter interface{}, apiErr error) {
		ctx := rc.Context

		q, err := ExtractParams(r)
		fmt.Println("t??", q)
		format := "i18n" // Default format
		localeKey := ""
		tag := ""
		// FIXME: why do we allow multiple projects here?
		var projects []string
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
				projects = v
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
		switch format {
		case "typescript":
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
				rc.WriteErr(fmt.Errorf("invalid format: %s. Valid formats are: %s", format, validFormats), requestContext.CodeErrInputValidation)
				return
			}
		}
		if len(projects) != 1 {
			apiErr = NewApiError("A project must be selected", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
			return
		}
		cacheKeys := []string{format, localeKey, tag}
		cacheKeys = append(cacheKeys, locales...)
		cacheKeys = append(cacheKeys, projects...)
		sort.Strings(cacheKeys)
		cacheKey := strings.Join(cacheKeys, "%")
		if flatten {
			cacheKey += "F"
		}
		// if v, ok := exportCache.Get(cacheKey); ok {
		// 	rw.Write(v.([]byte))
		// 	return
		// }

		ps, err := ctx.DB.GetProjectByIDOrShortName(projects[0])
		if err != nil {
			rc.WriteErr(err, requestContext.CodeErrProject)
			return
		}
		if ps == nil {
			apiErr = NewApiError("Project not found", http.StatusNotFound, string(requestContext.CodeErrNotFoundProject))
			return
		}
		var ep types.ExtendedProject

		if tag != "" {
			var snapshotMeta types.ProjectSnapshotMeta
			if s, ok := ps.Snapshots[tag]; ok {
				snapshotMeta = s
			}
			if snapshotMeta.SnapshotID == "" {
				// TODO: check if tag is semver and resolve
				rc.L.Warn().Msg("Not implemented semver-check on tags")

			}
			if snapshotMeta.SnapshotID == "" {
				apiErr = NewApiError("Tag not found", http.StatusNotFound, "TagNotFound", ps.Snapshots)
				return
			}
			s, err := ctx.DB.GetSnapshot(snapshotMeta.SnapshotID)
			if err != nil {
				apiErr = NewApiErr(err, http.StatusInternalServerError, string(requestContext.CodeErrSnapshot))
				return
			}
			if s == nil {
				apiErr = NewApiError("The snapshot was not found", http.StatusNotFound, string(CodeInternalServerError))
				return
			}
			ep = s.Project

		} else {
			ep, err = ps.Extend(ctx.DB, types.ExtendOptions{LocaleFilter: locales, ByKeyLike: true, ErrOnNoLocales: true, LocaleFilterFunc: func(locale types.Locale) bool {
				for k, v := range ps.LocaleIDs {
					if locale.ID != k {
						continue
					}
					if v.Publish {
						return true
					}
				}
				return false
			}})

		}
		if err != nil {
			rc.WriteErr(fmt.Errorf("Error extending project '%s' (%s): %w", ep.Title, ep.ID, err), requestContext.CodeErrProject)
			return
		}
		if len(ep.Locales) == 0 {
			apiErr = NewApiError("No locales were published", http.StatusBadGateway, "NoLocalesPublished")
			return
		}
		i18nodes, err := importexport.ExportI18N(ep, importexport.ExportI18NOptions{
			LocaleFilter: locales,
			LocaleKey:    importexport.LocaleKey(localeKey)})
		if err != nil {
			rc.WriteErr(err, requestContext.CodeErrProject)
			return
		}
		i18n, err := importexport.I18NNodeToI18Next(i18nodes)
		if err != nil {
			rc.WriteErr(err, requestContext.CodeErrProject)
			return
		}

		if format == "typescript" {
			var w bytes.Buffer
			err := importexport.ExportByGoTemplate("typescript.tmpl", ep, i18nodes, &w)
			if err != nil {
				rc.WriteErr(err, requestContext.CodeErrTemplating)
				return
			}
			rw.Header().Set("Content-Type", "application/typescript")
			// toWriter = w.String()
			rw.Write(w.Bytes())
			return
		} else {
			if len(locales) == 1 {
				toWriter = i18n[locales[0]]
			} else {
				toWriter = i18n
			}
		}
		return
	}
}
