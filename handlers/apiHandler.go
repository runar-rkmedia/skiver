package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/skiver/bboltStorage"
	"github.com/runar-rkmedia/skiver/importexport"
	"github.com/runar-rkmedia/skiver/localuser"
	"github.com/runar-rkmedia/skiver/models"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
)

var (
	maxBodySize int64 = 1_000_000 // 1MB
)

type Cache interface {
	// Get
	Get(k string) (interface{}, bool)
	Set(k string, x interface{}, d time.Duration)
	SetDefault(k string, x interface{})
}
type SessionManager interface {
	NewSession(user types.User, organization types.Organization, userAgent string) (s types.Session)
	GetSession(token string) (s types.Session, err error)
	SessionsForUser(userId string) (s []types.Session)
	ClearAllSessionsForUser(userId string) error
	TTL() time.Duration
}

// Deprecated. Migrating to using httproutermux
func EndpointsHandler(
	ctx requestContext.Context,
	userSessions SessionManager,
	pw localuser.PwHasher,
	serverInfo func() *types.ServerInfo,
	swaggerYml []byte,
) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			return
		}
		AddAccessControl(r, rw)

		path := r.URL.Path
		paths := strings.Split(strings.TrimSuffix(path, "/"), "/")
		// We are finally migrating to using a mux, but only a few routes have been migrated this far.
		rc := requestContext.NewReqContext(&ctx, r, rw)
		var body []byte
		var err error
		isGet := r.Method == http.MethodGet

		isPost := r.Method == http.MethodPost
		isPut := r.Method == http.MethodPut
		shouldReadBody := rc.ContentKind > 0 && (isPost || isPut)

		if shouldReadBody {
			body, err = io.ReadAll(r.Body)
			if err != nil {
				rc.WriteErr(err, requestContext.CodeErrReadBody)
				return
			}
		}
		// Check login
		authSource := ""
		var session *types.Session
		var authValue string
		if v := r.Header.Get("Authorization"); v != "" {
			authSource = "header"
			authValue = v
		}
		if authValue == "" {
			if cookie, err := r.Cookie("token"); err == nil {
				authSource = "cookie"
				authValue = cookie.Value
			}
		}
		if authValue != "" {
			sess, err := userSessions.GetSession(authValue)
			if err == nil {
				expiresD := sess.Expires.Sub(time.Now())
				rw.Header().Add("session-expires", sess.Expires.String())
				rw.Header().Add("session-expires-in", expiresD.String())
				rw.Header().Add("session-expires-in-seconds", strconv.Itoa(int(expiresD.Seconds())))
				session = &sess
			} else {
				details := map[string]any{"authSource": authSource}
				if authSource == "cookie" {
					// Authentication failed, logout the user
					err := logout(session, userSessions, rw)
					if err != nil {
						rc.WriteError(err.Error(), requestContext.CodeErrAuthoriziationFailed, details)
						return
					}

				}
				rc.WriteError("The authorization provided was invalid", requestContext.CodeErrAuthoriziationFailed, details)
				return
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

		case "swagger", "swagger.yaml", "swagger.yml":
			rw.Header().Set("Content-Type", "text/vnd.yaml")
			rw.Header().Set("Content-Disposition", `attachment; filename="swagger-skiver.yaml"`)
			rw.Write(swaggerYml)
			return
		case "serverInfo":
			if isGet && len(paths) == 1 {
				size, sizeErr := ctx.DB.Size()
				info := serverInfo()
				if sizeErr != nil {
					ctx.L.Warn().Err(sizeErr).Msg("Failed to retrieve size of database")
				} else {
					info.DatabaseSize = size
					info.DatabaseSizeStr = humanize.Bytes(uint64(size))
				}
				rc.WriteAuto(info, err, "serverInfo")
				return
			}
		case "join":
			if isPost || isGet {
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
						CanManageSnapshots:    true,
					}
					existingUsers := false
					{
						orgUsers, err := ctx.DB.FindUsers(1, types.User{Entity: types.Entity{OrganizationID: org.ID}})
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
		case "logout":
			{
				if isPost {
					logout(session, userSessions, rw)
					rc.WriteOutput(models.LogoutResponse{Ok: boolPointer(true)}, http.StatusOK)
					return
				}
			}
		case "login":
			if isGet {
				if session == nil {
					rc.WriteError("Not logged in", requestContext.CodeErrAuthenticationRequired)
					return
				}
				expiresD := session.Expires.Sub(time.Now())
				rc.WriteOutput(types.LoginResponse{
					// TODO:
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

			user, err := ctx.DB.FindUserByUserName("", *j.Username)
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
				d := userSessions.TTL() / 6 * 5
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

			// TODO: move some of these settings to global config, organization settings and/or project settings.
			cookie := &http.Cookie{
				Name: "token",
				// TODO: when the server is behind a subpath (e.g.
				// exmaple.com/skiver/), the reverse-proxy in front may not return our
				// path, and we probably need to get it from the config
				Path:     "/",
				Value:    session.Token,
				MaxAge:   int(expiresD.Seconds()),
				Secure:   r.TLS != nil,
				HttpOnly: true,
			}
			xproto := r.Header.Get("X-Forwarded-Proto")
			switch xproto {
			case "http":
				cookie.Secure = false
				// does not work with http
			case "https":
				cookie.Secure = true
				cookie.SameSite = http.SameSiteNoneMode
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
			if authSource == "" {
				rc.WriteError("No authentication-method provided", requestContext.CodeErrAuthenticationRequired)
			} else {
				rc.WriteError("Authentication provided, but failed", requestContext.CodeErrAuthoriziationFailed, map[string]interface{}{"authSource": authSource})
			}
			return
		}
		orgId := session.Organization.ID

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
				if project.OrganizationID != orgId {
					rc.WriteErr(err, requestContext.CodeErrProject)
					return
				}

				out, Err := ImportIntoProject(ctx.DB, kind, session.User.ID, *project, localeLike, dry, input)
				if Err != nil {
					rc.WriteErr(Err, Err.GetCode())
					return
				}
				rc.WriteOutput(out, http.StatusOK)
				return

			} else {
				rc.WriteError("Only post is allowed here", requestContext.CodeErrMethodNotAllowed)
				return
			}
		case "organization":
			if isGet {
				if session.User.CanCreateOrganization {
					orgs, err := ctx.DB.GetOrganizations()
					rc.WriteAuto(orgs, err, requestContext.CodeErrProject)
					return

				}
				org, err := ctx.DB.GetOrganization(session.User.OrganizationID)
				rc.WriteAuto(map[string]*types.Organization{org.ID: org}, err, requestContext.CodeErrProject)
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
				l.JoinID, err = utils.GetRandomName()
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrOrganization)
					return
				}
				l.CreatedBy = session.User.ID
				org, err := ctx.DB.CreateOrganization(l)
				rc.WriteAuto(org, err, requestContext.CodeErrCreateProject)
				return
			}
		case "project":
			if getStringSliceIndex(paths, 1) == "" {
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

					p := types.Project{
						Title:       *j.Title,
						Description: j.Description,
						ShortName:   *j.ShortName,
						LocaleIDs:   map[string]types.LocaleSetting{},
					}
					if len(j.Locales) > 0 {
						for lID, ls := range j.Locales {
							p.LocaleIDs[lID] = types.LocaleSetting{
								Enabled:         ls.Enabled,
								Publish:         ls.Publish,
								AutoTranslation: ls.AutoTranslation,
							}
						}
					}

					p.CreatedBy = session.User.ID
					p.OrganizationID = session.Organization.ID
					locale, err := ctx.DB.CreateProject(p)
					rc.WriteAuto(locale, err, requestContext.CodeErrCreateProject)
					return
				}
				if isPut {
					if !session.User.CanUpdateProjects {
						rc.WriteError("You are not authorizatiod to update projects", requestContext.CodeErrAuthoriziation)
						return
					}
					var j models.UpdateProjectInput
					if err := rc.ValidateBytes(body, &j); err != nil {
						return
					}

					p, err := ctx.DB.GetProject(*j.ID)
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrProject)
						return
					}
					if p == nil || session.User.OrganizationID != p.OrganizationID {
						rc.WriteError("Could not find this project, or you do not have access", requestContext.CodeErrNotFoundProject)
						return
					}
					payload := types.Project{
						Title:       j.Title,
						Description: j.Description,
						ShortName:   j.ShortName,
					}
					payload.UpdatedBy = session.User.ID
					if len(j.Locales) > 0 {
						payload.LocaleIDs = map[string]types.LocaleSetting{}
						for lID, ls := range j.Locales {
							payload.LocaleIDs[lID] = types.LocaleSetting{
								Enabled:         ls.Enabled,
								Publish:         ls.Publish,
								AutoTranslation: ls.AutoTranslation,
							}

						}
					}
					project, err := ctx.DB.UpdateProject(*j.ID, payload)

					rc.WriteAuto(project, err, requestContext.CodeErrProject)

					return
				}
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
				c.OrganizationID = session.Organization.ID
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
					Value:         *j.Value,
				}
				tv.CreatedBy = session.User.ID
				tv.OrganizationID = session.Organization.ID

				_, variables := importexport.InferVariables(tv.Value, "???", tv.TranslationID)
				if len(variables) > 0 {
					t, err := ctx.DB.GetTranslation(tv.TranslationID)
					if err != nil {
						rc.WriteErr(err, requestContext.CodeErrTranslation)
						return
					}
					needsUpdate := false
					if t.Variables == nil {
						t.Variables = variables
						needsUpdate = true
					} else {

					outerV:
						for k, v := range variables {
							for tk := range t.Variables {
								if k == tk {
									continue outerV
								}
							}
							t.Variables[k] = v
							needsUpdate = true
						}
					}
					if needsUpdate {
						ctx.DB.UpdateTranslation(t.ID, *t)
					}
				}

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
				var j models.UpdateTranslationValueInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}
				if id == "" {
					id = *j.ID
				}
				if id == "" {
					rc.WriteError("Missing id", requestContext.CodeErrIDEmpty)
					return
				}
				tv := types.TranslationValue{}
				tv.ID = id
				if j.ContextKey != "" {
					tv.Context = map[string]string{j.ContextKey: *j.Value}
				} else {
					tv.Value = *j.Value
				}
				tv.Source = types.CreatorSourceUser
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
				l.OrganizationID = session.Organization.ID
				locale, err := ctx.DB.CreateLocale(l)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrDBCreateLocale)
					return
				}
				rc.WriteOutput(locale, http.StatusCreated)
				return
			}
		}
		rc.WriteError(fmt.Sprintf("No route registered for: %s %s", r.Method, r.URL.Path), requestContext.CodeErrNoRoute)
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
