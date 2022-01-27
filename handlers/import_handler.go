package handlers

import (
	"net/http"

	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type Error struct {
	models.APIError
	StatusCode int
	Errors     []error
}

func (err Error) Error() string {
	if len(err.Errors) > 0 {
		s := err.Message
		for i := 0; i < len(err.Errors); i++ {
			s += ", " + err.Errors[i].Error()
		}

		return s
	}
	return err.Message
}
func (err Error) GetCode() requestContext.ErrorCodes {
	return requestContext.ErrorCodes(err.Code)
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
			Message: message,
			Code:    models.ErrorCodes(code),
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
type ImportResult struct {
	Changes  Updates
	Imp      importexport.Import
	Ex       types.ExtendedProject
	Warnings []importexport.Warning
}

func ImportIntoProject(db types.Storage, kind string, createdBy string, project types.Project, localeLike string, dry bool, input map[string]interface{}) (*ImportResult, *Error) {
	switch kind {
	case "":
		return nil, NewError("empty value for kind, allowed values: i18n, auto", requestContext.CodeErrInputValidation)
	case "i18n", "auto":
		break
	default:
		return nil, NewError("Invalid value for kind, allowed values: i18n, auto", requestContext.CodeErrInputValidation)
	}
	var locale *types.Locale
	if localeLike != "" {
		if true {

			return nil, NewError("Locale from url is not yet implemented. Please add the locale as the root-key in the body", requestContext.CodeErrNotImplemented)
		}
		locale, err := db.GetLocaleByIDOrShortName(localeLike)
		if err != nil {
			return nil, NewError("Failed ot get locale", requestContext.CodeErrLocale).AddError(err)
		}
		if locale == nil {
			return nil, NewError("Locale not found", requestContext.CodeErrNotFoundLocale, localeLike)
		}
	}
	locales, err := db.GetLocales()
	if err != nil {
		return nil, NewError("failed to get locales", requestContext.CodeErrLocale).AddError(err)
	}
	localesSlice := make([]types.Locale, len(locales))
	i := 0
	for _, v := range locales {
		localesSlice[i] = v
		i++
	}
	base := types.Project{}
	base.ID = project.ID
	base.CreatedBy = createdBy
	base.OrganizationID = project.OrganizationID
	imp, warnings, err := importexport.ImportI18NTranslation(localesSlice, locale, base, types.CreatorSourceImport, input)
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
	for cKey, cat := range imp.Categories {
		exCat, catExists := ex.Categories[cat.Key]
		cat.Exists = &catExists
		if !catExists {
			if !dry {
				created, err := db.CreateCategory(cat.Category)
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
		for tKey, t := range cat.Translations {
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
			for k, tv := range t.Values {
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
