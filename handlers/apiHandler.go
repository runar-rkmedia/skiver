package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	"github.com/runar-rkmedia/skiver/bboltStorage"
	"github.com/runar-rkmedia/skiver/localuser"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

var (
	maxBodySize int64 = 1_000_000 // 1MB
)

func EndpointsHandler(ctx requestContext.Context, pw localuser.PwHasher, serverInfo types.ServerInfo, swaggerYml []byte) http.HandlerFunc {

	p, ok := ctx.DB.(localuser.Persistor)
	if !ok {
		ctx.L.Warn().Str("type", fmt.Sprintf("%T", ctx.DB)).Msg("DB does not implement the localUser.Persistor-interface")
	}
	userSessions, err := localuser.NewUserSessionInMemory(localuser.UserSessionOptions{TTL: time.Hour}, uuid.NewString, p)
	if err != nil {
		ctx.L.Fatal().Err(err).Msg("Failed to set up userSessions")
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		AddAccessControl(r, rw)
		rc := requestContext.NewReqContext(&ctx, r, rw)
		var body []byte
		var err error
		isGet := r.Method == http.MethodGet
		isPost := r.Method == http.MethodPost
		// isDelete := r.Method == http.MethodDelete
		isPut := r.Method == http.MethodPut
		path := r.URL.Path
		paths := strings.Split(strings.TrimSuffix(path, "/"), "/")
		shouldReadBody := rc.ContentKind > 0 && (isPost || isPut)

		if shouldReadBody {
			body, err = io.ReadAll(r.Body)
			if err != nil {
				rc.WriteErr(err, requestContext.CodeErrReadBody)
				return
			}
		}
		// Check login
		var session *types.Session
		if cookie, err := r.Cookie("token"); err == nil {
			sess, err := userSessions.GetSession(cookie.Value)
			if err == nil {
				expiresD := sess.Expires.Sub(time.Now())
				rw.Header().Add("session-expires", sess.Expires.String())
				rw.Header().Add("session-expires-in", expiresD.String())
				rw.Header().Add("session-expires-in-seconds", strconv.Itoa(int(expiresD.Seconds())))
				session = &sess
			}
		}

		switch paths[0] {
		case "missing":
			if isGet {
				r, err := ctx.DB.GetMissingKeysFilter(0)
				rc.WriteAuto(r, err, requestContext.CodeErrReportMissing)
				return
			}
			if isPost {
				project := ""
				locale := ""
				if len(paths) >= 3 {
					project = paths[2]
					locale = paths[1]
				}
				if project == "" {
					rc.WriteErr(fmt.Errorf("project is required"), requestContext.CodeErrInputValidation)
					return
				}
				if locale == "" {
					rc.WriteErr(fmt.Errorf("locale is required"), requestContext.CodeErrInputValidation)
					return
				}
				if body == nil {
					// The default-settings of i18next's AddMissing request does not add the correct Content-Type.
					// Just to be nice, we attempt to read the body anyway...
					rc.ContentKind = requestContext.OutputJson
					body, err = io.ReadAll(r.Body)
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrReadBody)
						return
					}
				}
				var j models.ReportMissingInput
				err = rc.ValidateBytes(body, &j)
				if err != nil {
					return
				}
				var errs []string
				var mts []types.MissingTranslation
				for k := range j {
					splitted := strings.Split(k, ".")
					category := splitted[0]
					var translation string
					if len(splitted) > 0 {
						translation = splitted[1]
					}
					// requires go 1.18
					// category, translation, _ := strings.Cut(k, ".")
					mt := types.MissingTranslation{
						Locale:      locale,
						Project:     project,
						Translation: translation,
						Category:    category,
					}
					if session != nil {
						mt.CreatedBy = session.User.ID
					}
					if mt.CreatedBy == "" {
						mt.CreatedBy = "anonymous"
					}

					mtt, err := ctx.DB.ReportMissing(mt)
					if err != nil {
						ctx.L.Err(err).Interface("mt", mt).Msg("failed to report message")
						errs = append(errs, err.Error())
						continue
					}
					mts = append(mts, *mtt)
				}

				if errs != nil && len(errs) > 0 {

					rc.WriteErr(fmt.Errorf("%d/%d missing translations failed to report: %#v", len(errs), len(j), errs), requestContext.CodeErrProject)
					return
				}
				rc.WriteOutput(mts, http.StatusOK)

				return
			}
		case "export":
			if !isGet {
				break
			}
			q, err := extractParams(r, "export/")
			format := ""
			localeKey := ""
			var projects []string
			var locales []string
			flatten := true
			for k, v := range q {
				switch strings.ToLower(k) {
				case "locale", "l":
					locales = v
				case "format", "f":
					format = v[0]
				case "project", "p":
					projects = v
				case "locale_key":
					localeKey = v[0]
				case "no_flatten":
					flatten = false
				}
			}
			if format == "" {
				format = "i18n"
			} else {

				validFormats := []string{"i18n", "raw"}
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

			ps, err := ctx.DB.GetProjects()
			if err != nil {
				rc.WriteErr(err, requestContext.CodeErrProject)
				return
			}
			projectsLength := len(projects)
			out := map[string]types.ExtendedProject{}
			for _, v := range ps {
				if projectsLength > 0 {
					found := false
					for _, pid := range projects {
						if v.ID == pid || v.ShortName == pid {
							found = true
							break
						}
					}
					if !found {
						break
					}
				}
				ep, err := v.Extend(ctx.DB)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrProject)
					return
				}
				out[v.ID] = ep
			}
			if format == "i18n" {
				out := map[string]types.I18N{}
				for _, p := range ps {
					ep, err := p.Extend(ctx.DB)
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrProject)
						return
					}
					i18n, err := types.ExportI18N(ep, types.ExportI18NOptions{
						LocaleFilter: locales,
						LocaleKey:    types.LocaleKey(localeKey)})
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrProject)
						return
					}
					// If the user requested just a single project, we dont want to return a map
					if flatten && projectsLength == 1 {
						if len(locales) == 1 {
							rc.WriteOutput(i18n[locales[0]], http.StatusOK)
							return
						}
						rc.WriteOutput(i18n, http.StatusOK)
						return
					}
					out[ep.ID] = i18n
				}
				rc.WriteOutput(out, http.StatusOK)
				return
			}
			rc.WriteOutput(out, http.StatusOK)

			return

		case "swagger", "swagger.yaml", "swagger.yml":
			rw.Header().Set("Content-Type", "text/vnd.yaml")
			rw.Header().Set("Content-Disposition", `attachment; filename="swagger-skiver.yaml"`)
			rw.Write(swaggerYml)
			return
		case "serverInfo":
			if isGet && len(paths) == 1 {
				size, sizeErr := ctx.DB.Size()
				if sizeErr != nil {
					ctx.L.Warn().Err(sizeErr).Msg("Failed to retrieve size of database")
				} else {
					serverInfo.DatabaseSize = size
					serverInfo.DatabaseSizeStr = humanize.Bytes(uint64(size))
				}

				rc.WriteAuto(serverInfo, err, "serverInfo")
				return
			}
		case "join":
			if isGet || isPost {
				joinId := getStringSliceIndex(paths, 1)
				if joinId == "" {
					rc.WriteError("Missing join-id", requestContext.CodeErrIDEmpty)
					return
				}

				orgs, err := ctx.DB.GetOrganizations()
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrOrganization)
					return
				}
				var org *types.Organization
				for _, o := range orgs {
					if o.JoinID == joinId {
						org = &o
						break
					}
				}
				if org == nil {
					rc.WriteError("Not found", requestContext.CodeErrOrganizationNotFound)
					return
				}
				if org.JoinIDExpires.Before(time.Now()) {
					rc.WriteError("Not found", requestContext.CodeErrOrganizationNotFound)
					return
				}
				if isPost {
					var joinInput models.JoinInput
					err := rc.ValidateBytes(body, &joinInput)
					if err != nil {
						return
					}

					pass, err := pw.Hash(*joinInput.Password)
					if err != nil {
						rc.L.Error().Err(err).Msg("there was an error with hashing the password")
						rc.WriteError("Failure in password-creation", requestContext.CodeErrPasswordHashing)
						return
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
					}
					existingUsers := false
					{
						orgUsers, err := ctx.DB.GetUsers(1, types.User{Entity: types.Entity{OrganizationID: org.ID}})
						if err != nil {
							rc.WriteErr(err, requestContext.CodeErrNotFoundUser)
							return
						}
						existingUsers = len(orgUsers) > 0
					}
					if existingUsers {
						u.CanUpdateOrganization = true
						// user is the first to join, should have organization-administrative permissions
					}

					user, err := ctx.DB.CreateUser(u)
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrNotFoundUser)
						return
					}
					// TODO: loginUser
					rc.WriteOutput(types.LoginResponse{
						User:         user,
						Organization: *org,
						Ok:           true,
					}, http.StatusOK)
					return
				}

				rc.WriteOutput(org, http.StatusOK)
				return

			}
		case "login":
			if isGet {
				if session == nil {
					rc.WriteError("Not logged in", requestContext.CodeErrAuthenticationRequired)
					return
				}
				expiresD := session.Expires.Sub(time.Now())
				rc.WriteOutput(types.LoginResponse{
					Organization: session.Organization,
					User:         session.User,
					Ok:           true,
					Expires:      session.Expires,
					ExpiresIn:    expiresD.String(),
				}, http.StatusOK)
				return
			}
			if !isPost {
				rc.WriteErr(fmt.Errorf("Only POST is allowed here"), requestContext.CodeErrMethodNotAllowed)
				break
			}
			var j models.LoginInput
			if body == nil {
				rc.WriteErr(fmt.Errorf("Body was empty"), requestContext.CodeErrInputValidation)
				return
			}
			if err := rc.ValidateBytes(body, &j); err != nil {
				return
			}

			err := rc.Unmarshal(body, &j)
			if err != nil {
				rc.WriteErr(err, "err-marshal-user")
				return
			}
			vErrs := models.Validate(&j)
			if vErrs != nil {
				rc.WriteOutput(vErrs, http.StatusBadRequest)
				return
			}

			user, err := ctx.DB.GetUserByUserName(*j.Username)
			if err != nil {
				rc.WriteErr(err, "Err:login")
				return
			}
			if user == nil {
				rc.WriteError("The supplied username/password is incorrect", "incorrect-user-password")
				return
			}

			ok, err := pw.Verify(user.PW, *j.Password)
			if err != nil {
				rc.WriteError("The supplied username/password is incorrect", "incorrect-user-password")
				return
			}
			if !ok {
				rc.WriteError("The supplied username/password is incorrect", "incorrect-user-password")
				return
			}
			userAgent := r.UserAgent() + ";" + rc.RemoteIP

			var session types.Session
			sessions := userSessions.SessionsForUser(user.ID)

			now := time.Now()
			for i := 0; i < len(sessions); i++ {
				// We already have the correct user, we are trying to identify their device,
				// so that sessions are unique per device.
				// This is of course not possible for all devices, because of user-privacy,
				// which we should respect.
				if sessions[i].UserAgent != userAgent {
					continue
				}
				// if the user has a fair amount left in their session, it is not renewed
				d := userSessions.TTL / 6 * 5
				if sessions[i].Expires.Add(-d).Before(now) {
					continue
				}
				session = sessions[i]
			}

			if session.UserAgent == "" {
				organization, err := ctx.DB.GetOrganization(user.OrganizationID)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrOrganization)
					return
				}
				if organization == nil {
					rc.WriteError("Could not find the users organzation. Please contact your administrator", requestContext.CodeErrOrganization)
					return
				}
				session = userSessions.NewSession(*user, *organization, userAgent)
			}

			expiresD := session.Expires.Sub(now)

			cookie := &http.Cookie{
				Name:     "token",
				Path:     "/api/",
				Value:    session.Token,
				MaxAge:   int(expiresD.Seconds()),
				HttpOnly: true,
			}
			rw.Header().Add("session-expires", session.Expires.String())
			rw.Header().Add("session-expires-in", expiresD.String())
			rw.Header().Add("session-expires-in-seconds", strconv.Itoa(int(expiresD.Seconds())))
			http.SetCookie(rw, cookie)
			r := types.LoginResponse{
				User:      session.User,
				Ok:        true,
				Expires:   session.Expires,
				ExpiresIn: expiresD.String(),
			}
			rc.WriteOutput(r, http.StatusOK)
			return

		}

		// Login required beyond this point

		if session == nil {
			rc.WriteError("Not logged in", requestContext.CodeErrAuthenticationRequired)
			return
		}
		rc.L.Debug().
			Str("path", path).
			Str("username", session.User.UserName).
			Msg("User is perorming action on route")

		switch paths[0] {
		case "import":
			if isPost {
				if !session.User.CanCreateTranslations {
					rc.WriteError("You are not authorizatiod to create translations", requestContext.CodeErrAuthoriziation)
					return
				}
				kind := getStringSliceIndex(paths, 1)
				projectLike := getStringSliceIndex(paths, 2)
				localeLike := getStringSliceIndex(paths, 3)
				q := r.URL.Query()
				dry := q.Has("dry")
				var input map[string]interface{}
				switch kind {
				case "":
					rc.WriteError("empty value for kind, allowed values: i18n, auto", requestContext.CodeErrInputValidation)
					return
				case "i18n", "auto":
					break
				default:
					rc.WriteError("Invalid value for kind, allowed values: i18n, auto", requestContext.CodeErrInputValidation)
					return
				}
				if projectLike == "" {
					rc.WriteError("Missing argument for project", requestContext.CodeErrInputValidation)
					return
				}
				if body == nil || len(body) == 0 {
					rc.WriteError("Expected body to be present", requestContext.CodeErrInputValidation)
					return
				}
				err := rc.Unmarshal(body, &input)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrUnmarshal)
					return
				}
				project, err := ctx.DB.GetProjectByIDOrShortName(projectLike)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrProject)
					return
				}
				if project == nil {
					rc.WriteError("Project was not found", requestContext.CodeErrNotFoundProject)
					return
				}
				var locale *types.Locale
				if localeLike != "" {
					if true {

						rc.WriteError("Locale from url is not yet implemented. Please add the locale as the root-key in the body", requestContext.CodeErrNotImplemented)
						return
					}
					locale, err := ctx.DB.GetLocaleByIDOrShortName(localeLike)
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrLocale)
						return
					}
					if locale == nil {
						rc.WriteError("Locale not found", requestContext.CodeErrNotFoundLocale, localeLike)
						return
					}
				}
				locales, err := ctx.DB.GetLocales()
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrLocale)
					return
				}
				localesSlice := make([]types.Locale, len(locales))
				i := 0
				for _, v := range locales {
					localesSlice[i] = v
					i++
				}
				imp, warnings, err := ImportI18NTranslation(localesSlice, locale, project.ID, session.User.ID, types.CreatorSourceImport, input)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrImport)
					return
				}
				if imp == nil {
					rc.WriteError("Import resulted in null", requestContext.CodeErrImport)
					return
				}

				extendOptions := types.ExtendOptions{ByKeyLike: true}
				ex, err := project.Extend(ctx.DB, extendOptions)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrProject)
					return
				}
				type Updates struct {
					TranslationValueUpdates    map[string]types.TranslationValue
					TranslationsValueCreations map[string]types.TranslationValue
					TranslationCreations       map[string]types.Translation
					CategoryCreations          map[string]types.Category
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
							created, err := ctx.DB.CreateCategory(cat.Category)
							if err != nil {
								rc.WriteError(err.Error(), requestContext.CodeErrCreateCategory, cat)
								return
							}
							esc, err := created.Extend(ctx.DB, extendOptions)
							if err != nil {
								rc.WriteErr(err, requestContext.CodeErrCategory)
								return
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
									created, err := ctx.DB.CreateTranslation(t.Translation)
									if err != nil {
										rc.WriteError(err.Error(), requestContext.CodeErrTranslation, t.Translation)
										return
									}
									esc, err := created.Extend(ctx.DB, extendOptions)
									if err != nil {
										rc.WriteErr(err, requestContext.CodeErrTranslation)
										return
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
								rc.WriteError("condition not implemented: translation did not resolve", requestContext.CodeErrNotImplemented, map[string]interface{}{"translation": t})
								return
							}
						}
						for k, tv := range t.Values {
							tv.TranslationID = exT.ID
							exTv, existsTV := exT.Values[tv.LocaleID]
							if existsTV {
								if exTv.Value != tv.Value {
									exTv.Value = tv.Value
									if !dry {
										updated, err := ctx.DB.UpdateTranslationValue(exTv)
										if err != nil {
											rc.WriteError(err.Error(), requestContext.CodeErrUpdateTranslationValue, tv)
											return
										}
										updates.TranslationValueUpdates[updated.ID] = updated
									} else {
										updates.TranslationValueUpdates[exTv.ID] = exTv
									}
								}
							} else {
								if !dry {
									created, err := ctx.DB.CreateTranslationValue(tv)
									if err != nil {
										details := struct {
											Input    types.TranslationValue
											Response types.TranslationValue
										}{tv, created}
										rc.WriteError(err.Error(), requestContext.CodeErrCreateTranslationValue, details)
										return
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

				out := struct {
					Changes  Updates
					Imp      Import
					Ex       types.ExtendedProject
					Warnings []Warning
				}{

					Changes:  updates,
					Imp:      *imp,
					Ex:       ex,
					Warnings: warnings,
				}
				rc.WriteOutput(out, http.StatusOK)
				return

			} else {
				rc.WriteError("Only post is allowed here", requestContext.CodeErrMethodNotAllowed)
				return
			}
		case "organization":
			if isGet {
				orgs, err := ctx.DB.GetOrganizations()
				rc.WriteAuto(orgs, err, requestContext.CodeErrProject)
				return
			}
			if isPost {
				if !session.User.CanCreateOrganization {
					rc.WriteError("You are not authorizatiod to create organizations", requestContext.CodeErrAuthoriziation)
					return
				}
				var j models.OrganizationInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				l := types.Organization{
					Title: *j.Title,
					// Initially set to expire within 30 days.
					JoinIDExpires: time.Now().Add(30 * 24 * time.Hour),
				}
				l.JoinID = utils.GetRandomName()
				l.CreatedBy = session.User.ID
				org, err := ctx.DB.CreateOrganization(l)
				rc.WriteAuto(org, err, requestContext.CodeErrCreateProject)
				return
			}
		case "project":
			if isGet {
				projects, err := ctx.DB.GetProjects()
				rc.WriteAuto(projects, err, requestContext.CodeErrProject)
				return
			}
			if isPost {
				if !session.User.CanCreateProjects {
					rc.WriteError("You are not authorizatiod to create projects", requestContext.CodeErrAuthoriziation)
					return
				}
				var j models.ProjectInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				l := types.Project{
					Title:       *j.Title,
					Description: j.Description,
					ShortName:   *j.ShortName,
				}
				l.CreatedBy = session.User.ID
				l.OrganizationID = session.Organization.ID
				locale, err := ctx.DB.CreateProject(l)
				rc.WriteAuto(locale, err, requestContext.CodeErrCreateProject)
				return
			}
		case "translation":
			if isGet {
				translations, err := ctx.DB.GetTranslations()
				rc.WriteAuto(translations, err, requestContext.CodeErrTranslation)
				return
			}
			if isPost {
				if !session.User.CanCreateTranslations {
					rc.WriteError("You are not authorizatiod to create translations", requestContext.CodeErrAuthoriziation)
					return
				}
				var j models.TranslationInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				t := types.Translation{
					// TranslationInput: j,
					CategoryID:  *j.CategoryID,
					Key:         *j.Key,
					Description: j.Description,
					Title:       j.Title,
					Variables:   j.Variables,
				}
				t.CreatedBy = session.User.ID
				translation, err := ctx.DB.CreateTranslation(t)
				rc.WriteAuto(translation, err, requestContext.CodeErrCreateTranslation)
				return
			}
		case "category":
			if isGet {
				categories, err := ctx.DB.GetCategories()
				rc.WriteAuto(categories, err, requestContext.CodeErrCategory)
				return
			}
			if isPost {
				if !session.User.CanCreateTranslations {
					rc.WriteError("You are not authorizatiod to create categories", requestContext.CodeErrAuthoriziation)
					return
				}
				var j models.CategoryInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				c := types.Category{
					// TranslationInput: j,
					ProjectID:   *j.ProjectID,
					Key:         *j.Key,
					Description: j.Description,
					Title:       *j.Title,
				}
				c.CreatedBy = session.User.ID
				category, err := ctx.DB.CreateCategory(c)
				rc.WriteAuto(category, err, requestContext.CodeErrCategory)
				return
			}
		case "translationValue":
			if isGet {
				tvs, err := ctx.DB.GetTranslationValues()
				rc.WriteAuto(tvs, err, requestContext.CodeErrCategory)
				return
			}
			if isPost {
				if !session.User.CanCreateTranslations {
					rc.WriteError("You are not authorizatiod to create translation-values", requestContext.CodeErrAuthoriziation)
					return
				}
				var j models.TranslationValueInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				tv := types.TranslationValue{
					LocaleID:      *j.LocaleID,
					TranslationID: *j.TranslationID,
					Context:       j.Context,
					Value:         *j.Value,
				}
				tv.CreatedBy = session.User.ID
				translationValue, err := ctx.DB.CreateTranslationValue(tv)
				rc.WriteAuto(translationValue, err, requestContext.CodeErrCreateTranslationValue)
				return
			}
			if isPut {
				if !session.User.CanCreateTranslations {
					rc.WriteError("You are not authorizatiod to update translation-values", requestContext.CodeErrAuthoriziation)
					return
				}

				id := getStringSliceIndex(paths, 1)
				if id == "" {
					rc.WriteError("Missing id", requestContext.CodeErrIDEmpty)
					return
				}
				var j models.UpdateTranslationValueInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}
				tv := types.TranslationValue{}
				tv.ID = id
				tv.Value = *j.Value
				tv.UpdatedBy = session.User.ID
				translationValue, err := ctx.DB.UpdateTranslationValue(tv)
				rc.WriteAuto(translationValue, err, requestContext.CodeErrUpdateTranslationValue)
				return

			}
		case "locale":
			if isGet {
				locales, err := ctx.DB.GetLocales()
				if err != nil {

					if err == bboltStorage.ErrNotFound {
						rc.WriteErr(err, requestContext.CodeErrNotFoundLocale)
						return
					}
					rc.WriteErr(err, requestContext.CodeErrLocale)
					return
				}
				rc.WriteOutput(locales, http.StatusOK)
				return
			}
			if isPost {
				if !session.User.CanCreateTranslations {
					rc.WriteError("You are not authorizatiod to create locales", requestContext.CodeErrAuthoriziation)
					return
				}
				var j models.LocaleInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}
				l := types.Locale{
					Iso639_1: *j.Iso6391,
					Iso639_2: *j.Iso6392,
					Iso639_3: *j.Iso6393,
					IETF:     *j.IetfTag,
					Title:    *j.Title,
				}
				l.CreatedBy = session.User.ID
				locale, err := ctx.DB.CreateLocale(l)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrDBCreateLocale)
					return
				}
				rc.WriteOutput(locale, http.StatusCreated)
				return
			}
		}
		rc.WriteError(fmt.Sprintf("No route registerd for: %s %s", r.Method, r.URL.Path), requestContext.CodeErrNoRoute)
	}
}

func getStringSliceIndex(slice []string, index int) string {
	if len(slice) <= index {
		return ""
	}
	return slice[index]
}
func boolPointer(v bool) *bool {
	return &v
}
