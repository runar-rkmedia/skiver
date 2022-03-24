package handlers

import (
	"bytes"
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

type ExportOptions struct {
	// FIXME: why do we allow multiple projects here?
	Projects               []string
	Locales                []string
	LocaleKey, Format, Tag string
	NoFlatten              bool
}

func getExport(l logger.AppLogger, exportCache Cache, db types.Storage, opt ExportOptions) (toWriter interface{}, err error) {
	projects := opt.Projects
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
	if len(projects) != 1 {
		err = NewApiError("A project must be selected", http.StatusBadRequest, string(requestContext.CodeErrInputValidation))
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
	if exportCache != nil {
		if v, ok := exportCache.Get(cacheKey); ok {
			return v, nil
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

	ps, err := db.GetProjectByIDOrShortName(projects[0])
	if err != nil {
		err = ErrApiDatabase("Project", err)
		return
	}
	if ps == nil {
		err = ErrApiNotFound("Project", projects[0])
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
			l.Warn().Msg("Not implemented semver-check on tags")

		}
		if snapshotMeta.SnapshotID == "" {
			err = NewApiError("Tag not found", http.StatusNotFound, "TagNotFound", ps.Snapshots)
			return
		}
		s, err := db.GetSnapshot(snapshotMeta.SnapshotID)
		if err != nil {
			err = NewApiErr(err, http.StatusInternalServerError, string(requestContext.CodeErrSnapshot))
			return toWriter, err
		}
		if s == nil {
			err = NewApiError("The snapshot was not found", http.StatusNotFound, string(CodeInternalServerError))
			return toWriter, err
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
			}})

	}
	if err != nil {
		NewApiErr(fmt.Errorf("Error extending project '%s' (%s): %w", ep.Title, ep.ID, err), http.StatusBadGateway, string(requestContext.CodeErrProject))
		return
	}
	if format == "raw" {
		toWriter = ep
		return
	}
	if len(ep.Locales) == 0 {
		err = NewApiError("No locales were published", http.StatusBadGateway, "NoLocalesPublished")
		return
	}
	i18nodes, err := importexport.ExportI18N(ep, importexport.ExportI18NOptions{
		LocaleFilter: locales,
		LocaleKey:    importexport.LocaleKey(localeKey)})
	if err != nil {
		// rc.WriteErr(err, requestContext.CodeErrProject)
		err = NewApiErr(err, http.StatusBadGateway, string(requestContext.CodeErrProject))
		return
	}
	if i18nodes.Nodes == nil {
		return
	}
	i18n, err := importexport.I18NNodeToI18Next(i18nodes)
	if err != nil {
		err = NewApiErr(err, http.StatusBadGateway, string(requestContext.CodeErrProject))
		return
	}

	if format == "typescript" {
		var w bytes.Buffer
		if err := importexport.ExportByGoTemplate("typescript.tmpl", ep, i18nodes, &w); err != nil {

			err = NewApiErr(err, http.StatusBadGateway, string(requestContext.CodeErrTemplating))
			return toWriter, err
		}
		// toWriter = w.String()
		toWriter = w.Bytes()
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

func GetExport(
	exportCache Cache,
) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (toWriter interface{}, apiErr error) {
		AddAccessControl(r, rw)

		q, err := ExtractParams(r)
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

		toWriter, err = getExport(rc.L, exportCache, rc.Context.DB, ExportOptions{
			Projects:  projects,
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
		if format == "typescript" {
			rw.Header().Set("Content-Type", "application/json")
			rw.Write(toWriter.([]byte))
			return nil, nil
		}
		return toWriter, err
	}
}
