package handlers

import (
	"fmt"
	"net/http"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

type Error struct {
	models.APIError
	StatusCode int
	Errors     []error
}

func (err Error) Error() string {
	if len(err.Errors) > 0 {
		s := err.APIError.Error.Message
		for i := 0; i < len(err.Errors); i++ {
			s += ", " + err.Errors[i].Error()
		}

		return s
	}
	return err.APIError.Error.Message
}
func (err Error) GetCode() requestContext.ErrorCodes {
	return requestContext.ErrorCodes(err.APIError.Error.Code)
}
func (err Error) GetStatusCode() int {
	if err.StatusCode == 0 {
		// In case we forgot to set the status-code, we assign one which should draw our attention
		return http.StatusTeapot
	}
	return err.StatusCode
}
func (err *Error) GetApiError() models.APIError {
	return err.APIError
}

func NewError(message string, code requestContext.ErrorCodes, details ...interface{}) *Error {
	return &Error{
		Errors: []error{},
		APIError: models.APIError{
			Details: details,
			Error: &models.Error{
				Message: message,
				Code:    models.ErrorCodes(code),
			},
		},
	}
}

func (e *Error) AddError(err error) *Error {
	e.Errors = append(e.Errors, err)
	return e
}

type Updates struct {
	TranslationValueUpdates    map[string]types.TranslationValue
	TranslationsValueCreations map[string]types.TranslationValue
	TranslationCreations       map[string]types.Translation
	CategoryCreations          map[string]types.Category
	// TODO: support category-updates
}

// TODO: Reevaluate this structure.
type ImportResult struct {
	Changes   Updates                      `json:"changes,omitempty"`
	ChangeSet []importexport.ChangeRequest `json:"change_set,omitempty"`
	Imp       importexport.Import          `json:"imp,omitempty"`
	Ex        types.ExtendedProject        `json:"ex,omitempty"`
	Warnings  []importexport.Warning       `json:"warnings,omitempty"`
}

func ImportDescriptionsIntoProject(l logger.AppLogger, db types.Storage, createdBy string, project types.Project, dry bool, input map[string]interface{}) (*ImportResult, *Error) {

	if len(input) == 0 {
		return nil, NewError("Empty map", "ImportDescription:Empty")
	}

	es, err := project.Extend(db)
	if err != nil {
		return nil, NewError("failed during extending project", requestContext.CodeErrProject).AddError(err)
	}

	imp := ImportResult{}
	changeRequests, err := importexport.DescribeProjectContent(es, input)
	if err != nil {
		return nil, NewError("failed during creation of changerequests", requestContext.CodeErrImport).AddError(err)
	}
	imp.ChangeSet = changeRequests

	if len(changeRequests) == 0 {
		return &imp, nil
	}
	if dry {
		return &imp, nil
	}
	for _, v := range changeRequests {
		switch v.Kind {
		case "CategoryTitle":
			payload, ok := v.Payload.(types.Category)
			if !ok {
				return nil, NewError("failed to convert changerequest-payload to Category", "ChangeRequest:To:Category")
			}
			if payload.UpdatedBy == "" {
				payload.UpdatedBy = createdBy
			}
			_, err := db.UpdateCategory(payload.ID, payload)
			if err != nil {
				l.Error().Err(err).Interface("payload", payload).Msg("Failed to update category")
				return nil, NewError(err.Error(), requestContext.CodeErrCategory)
			}
			l.Debug().
				Interface("changerequests", v).
				Msg("Category updated from changerequest")
		case "TranslationTitle":
			payload, ok := v.Payload.(types.Translation)
			if !ok {
				return nil, NewError("failed to convert changerequest-field to string", "ChangeRequest:To:String")
			}
			if payload.UpdatedBy == "" {
				payload.UpdatedBy = createdBy
			}
			payload.UpdatedBy = createdBy
			_, err := db.UpdateTranslation(payload.ID, payload)
			if err != nil {
				l.Error().Err(err).Interface("payload", payload).Msg("Failed to update translation")
				return nil, NewError(err.Error(), requestContext.CodeErrCategory)
			}
			l.Debug().
				Interface("changerequests", v).
				Msg("Translation updated from changerequest")
		}

	}

	return &imp, nil
}
func ImportIntoProject(l logger.AppLogger, db types.Storage, kind string, createdBy string, project types.Project, localeLike string, dry bool, input map[string]interface{}) (*ImportResult, *Error) {
	switch kind {
	case "":
		return nil, NewError("empty value for kind, allowed values: i18n, describe, auto", requestContext.CodeErrInputValidation)
	case "describe":
		return ImportDescriptionsIntoProject(l, db, createdBy, project, dry, input)
	case "i18n", "auto":
		break
	default:
		return nil, NewError("Invalid value for kind, allowed values: i18n, auto", requestContext.CodeErrInputValidation)
	}
	locales, err := db.GetLocales()
	if err != nil {
		return nil, NewError("failed to get locales", requestContext.CodeErrLocale).AddError(err)
	}
	var locale types.Locale
	if localeLike != "" {
		var matches []types.Locale
		for _, l := range locales {
			if localeLike == l.IETF ||
				localeLike == l.Iso639_3 ||
				localeLike == l.Iso639_2 ||
				localeLike == l.Iso639_1 {
				matches = append(matches, l)
			}
		}
		if len(matches) > 1 {
			ietfs := make([]string, len(matches))
			for i, m := range matches {
				ietfs[i] = m.IETF
			}
			return nil, NewError(fmt.Sprintf("The provided locale could not match uniquely against a locale. Try being a bit more specific, for instance %s", ietfs), "Locale:NonUniqueMatch", map[string]interface{}{"possibleMatches": matches, "mostSpecific": ietfs})
		}
		if len(matches) == 0 {
			return nil, NewError("The provided locale did not match any of the known locales", "Locale:NonMatch")
		}
		locale = matches[0]

	}
	localesSlice := make([]types.Locale, len(locales))
	localeKeys := utils.SortedMapKeys(locales)
	for i, k := range localeKeys {
		localesSlice[i] = locales[k]
	}
	base := types.Project{}
	base.ID = project.ID
	base.CreatedBy = createdBy
	base.OrganizationID = project.OrganizationID
	imp, warnings, err := importexport.ImportI18NTranslation(localesSlice, &locale, base, types.CreatorSourceImport, input)
	if err != nil {
		return nil, NewError("failed to import", requestContext.CodeErrImport).AddError(err)
	}
	if imp == nil {
		return nil, NewError("Import resulted in null", requestContext.CodeErrImport)
	}

	extendOptions := types.ExtendOptions{ByKeyLike: true}
	ex, err := project.Extend(db, extendOptions)
	if err != nil {
		return nil, NewError("failed during extending project", requestContext.CodeErrProject).AddError(err)
	}

	updates := Updates{
		map[string]types.TranslationValue{},
		map[string]types.TranslationValue{},
		map[string]types.Translation{},
		map[string]types.Category{},
	}
	// TODO: this should ideally all be done in a single atomic commit.
	// TODO: handle changes to translation-values
	catKeys := utils.SortedMapKeys(imp.Categories)
	for _, cKey := range catKeys {

		cat := imp.Categories[cKey]
		// the ImportI18NTranslation deos not inject subcategories into the category, only the extended categories.
		// injectSubCategories(&cat)
		exCat, catExists := ex.Categories[cat.Key]
		cat.Exists = &catExists
		if !catExists {
			if !dry {

				c := cat.Category
				created, err := db.CreateCategory(c)
				if err != nil {
					return nil, NewError(err.Error(), requestContext.CodeErrCreateCategory, cat)
				}
				esc, err := created.Extend(db, extendOptions)
				if err != nil {
					return nil, NewError("failed to extend category", requestContext.CodeErrCategory).AddError(err)
				}
				exCat = esc
				catExists = true
				updates.CategoryCreations[created.ID] = created
			} else {
				updates.CategoryCreations["_toCreate_"+cKey+""] = cat.Category
			}
		}
		ctKeys := utils.SortedMapKeys(cat.Translations)
		for _, tKey := range ctKeys {
			t := cat.Translations[tKey]
			var exT *types.ExtendedTranslation
			if exCat.ID == "" {
				t.Exists = boolPointer(false)
			} else {
				ex, tExists := exCat.Translations[t.Key]
				t.Exists = &tExists
				t.CategoryID = exCat.ID
				if tExists {
					exT = &ex
				} else {
					if !dry {
						created, err := db.CreateTranslation(t.Translation)
						if err != nil {
							return nil, NewError(err.Error(), requestContext.CodeErrTranslation, t.Translation)
						}
						esc, err := created.Extend(db, extendOptions)
						if err != nil {
							return nil, NewError("faile3d to extend translation", requestContext.CodeErrTranslation).AddError(err)
						}
						ex = esc
						exT = &esc
						tExists = *boolPointer(true)
						updates.TranslationCreations[created.ID] = created
					} else {
						updates.TranslationCreations["_toCreate_in_Category_'"+cKey+"'_"+tKey] = t.Translation
					}
				}
			}
			if exT == nil {
				if dry {
					exT = &t
					exT.Exists = boolPointer(false)
				} else {
					// TODO: Create translationValue
					return nil, NewError("condition not implemented: translation did not resolve", requestContext.CodeErrNotImplemented, map[string]interface{}{"translation": t})
				}
			}
			tvKeys := utils.SortedMapKeys(t.Values)
			for _, k := range tvKeys {
				tv := t.Values[k]
				tv.TranslationID = exT.ID
				exTv, existsTV := exT.Values[tv.LocaleID]
				if existsTV {
					if exTv.Value != tv.Value {
						exTv.Value = tv.Value
						if !dry {
							updated, err := db.UpdateTranslationValue(exTv)
							if err != nil {
								return nil, NewError(err.Error(), requestContext.CodeErrUpdateTranslationValue, tv)
							}
							updates.TranslationValueUpdates[updated.ID] = updated
						} else {
							updates.TranslationValueUpdates[exTv.ID] = exTv
						}
					}
				} else {
					if !dry {
						created, err := db.CreateTranslationValue(tv)
						if err != nil {
							details := struct {
								Input    types.TranslationValue
								Response types.TranslationValue
							}{tv, created}
							return nil, NewError(err.Error(), requestContext.CodeErrCreateTranslationValue, details)
						}
						updates.TranslationsValueCreations[created.ID] = created
					} else {
						updates.TranslationsValueCreations["_toCreate_in_Category_"+cKey+"_"+"under_Translation_"+tKey+"_"+k] = tv
					}
				}
			}
			imp.Categories[cKey].Translations[tKey] = t

		}
		imp.Categories[cKey] = cat
	}

	out := ImportResult{
		Changes:  updates,
		Imp:      *imp,
		Ex:       ex,
		Warnings: warnings,
	}
	return &out, nil

}
