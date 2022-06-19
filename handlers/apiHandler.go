package handlers

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

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
	NewSession(user types.User, organization types.Organization, userAgent string, opts ...types.UserSessionOptions) (s types.Session)
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

		case "swagger", "swagger.yaml", "swagger.yml":
			rw.Header().Set("Content-Type", "text/vnd.yaml")
			rw.Header().Set("Content-Disposition", `attachment; filename="swagger-skiver.yaml"`)
			rw.Write(swaggerYml)
			return
		case "logout":
			{
				if isPost {
					logout(session, userSessions, rw)
					rc.WriteOutput(models.OkResponse{Ok: boolPointer(true)}, http.StatusOK)
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
				dry := utils.HasDryRun(r)
				var input map[string]interface{}
				switch kind {
				case "":
					rc.WriteError("empty value for kind, allowed values: i18n, describe, auto", requestContext.CodeErrInputValidation)
					return
				case "i18n", "auto", "describe":
					break
				default:
					rc.WriteError("Invalid value for kind, allowed values: i18n, describe, auto", requestContext.CodeErrInputValidation)
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

				out, Err := ImportIntoProject(ctx.L, ctx.DB, kind, session.User.ID, *project, localeLike, dry, input)
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

				t, err := ctx.DB.GetTranslation(tv.TranslationID)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrTranslation)
					return
				}
				p, err := t.GetProject(ctx.DB)
				if err != nil {
					ctx.L.Error().Err(err).Msg("Project was not found for translation")
					rc.WriteErr(err, requestContext.CodeErrTranslation)
					return
				}
				if t == nil {
					rc.WriteErr(ErrApiNotFound("Translation", tv.TranslationID), "")
				}
				et, err := t.Extend(ctx.DB)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrTranslation)
					return
				}

				translationValue, err := ctx.DB.CreateTranslationValue(tv)
				if err != nil {
					rc.WriteErr(ErrApiDatabase("translation", err), "translation")
					return
				}
				o, err := importexport.CreateInterpolationMapForOrganization(ctx.DB, session.Organization.ID)
				if err != nil {
					ctx.L.Error().Err(err).Msg("Failed during CreateInterpolationMapForOrganization")
				}
				_, err = UpdateTranslationFromInferrence(
					ctx.DB,
					et,
					[]AdditionalValue{
						{Value: tv.Value, LocaleID: tv.LocaleID}},
					o.ByProject(p.ID),
				)
				if err != nil {
					ctx.L.Error().Err(err).Msg("Failed in updateTranslationFromInferrence")
				}

				rc.WriteOutput(translationValue, http.StatusOK)
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
				exTV, err := ctx.DB.GetTranslationValue(id)
				if exTV == nil {
					rc.WriteErr(ErrApiNotFound("TranslationValue", id), "")
				}
				t, err := ctx.DB.GetTranslation(exTV.TranslationID)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrTranslation)
					return
				}
				if t == nil {
					rc.WriteErr(ErrApiNotFound("Translation", exTV.TranslationID), "")
				}
				p, err := t.GetProject(ctx.DB)
				if err != nil {
					ctx.L.Error().Err(err).Msg("Project was not found for translation")
					rc.WriteErr(err, requestContext.CodeErrTranslation)
					return
				}
				et, err := t.Extend(ctx.DB)
				if err != nil {
					rc.WriteErr(err, requestContext.CodeErrTranslation)
					return
				}
				translationValue, err := ctx.DB.UpdateTranslationValue(tv)
				if err != nil {
					rc.WriteErr(ErrApiDatabase("translation", err), "translation")
					return
				}
				o, err := importexport.CreateInterpolationMapForOrganization(ctx.DB, session.Organization.ID)
				if err != nil {
					ctx.L.Error().Err(err).Msg("Failed during CreateInterpolationMapForOrganization")
				}
				_, err = UpdateTranslationFromInferrence(
					ctx.DB,
					et,
					[]AdditionalValue{
						{Value: tv.Value, LocaleID: exTV.LocaleID, Context: j.ContextKey},
					},
					o.ByProject(p.ID),
				)
				if err != nil {
					ctx.L.Error().Err(err).Msg("Failed in updateTranslationFromInferrence")
				}

				rc.WriteOutput(translationValue, http.StatusOK)
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

type AdditionalValue struct {
	Value    string
	LocaleID string
	Context  string
}

// updates a Translation based on updates to its values.
// The additionalValues are meant to be new values to consider for inferrence.
// If the Translation already has a value for the same LocaleID/Context as an additionalValue,
// the existing value will not be considered.
func UpdateTranslationFromInferrence(db types.Storage, et types.ExtendedTranslation, additionalValues []AdditionalValue, interpolationMaps []map[string]interface{}) (*types.Translation, error) {
	var allValues []string

	for _, v := range et.Values {
		found := false
		for _, av := range additionalValues {
			if av.Context != "" {
				continue
			}
			if av.LocaleID != v.LocaleID {
				continue
			}
			found = true
		}
		if !found {
			allValues = append(allValues, v.Value)

		}
		for _, c := range v.Context {
			foundContext := false
			for _, av := range additionalValues {
				if av.Context != c {
					continue
				}
				if av.LocaleID != v.LocaleID {
					continue
				}
				foundContext = true
			}
			if !foundContext {
				allValues = append(allValues, c)
			}
		}
	}

	for _, av := range additionalValues {
		allValues = append(allValues, av.Value)
	}
	// Check if we can infer some more variables/refs, and if can, we may need to update the Translation.
	_, variables, refs := importexport.InferVariablesFromMultiple(allValues, "???", et.ID, interpolationMaps)
	if len(variables) == 0 && len(refs) == 0 {
		return nil, nil
	}

	needsUpdate := false
	// If we have inferred new variables, we should update.
	// We should not remove any variables, since we cannot know if there are more variables available.
	if len(variables) > 0 {
		if et.Variables == nil {
			et.Variables = variables
			needsUpdate = true
		} else {

		outerV:
			for k, v := range variables {
				for tk := range et.Variables {
					if k == tk {
						continue outerV
					}
				}
				et.Variables[k] = v
				needsUpdate = true
			}
		}
	}
	// If the refs are changed, we should also update. This includes if a ref is removed
	if !needsUpdate {
		sort.Strings(et.References)
		sort.Strings(refs)

		if strings.Join(et.References, ";") != strings.Join(refs, ";") {
			et.References = refs
			needsUpdate = true
		}
	}
	if needsUpdate {
		t, err := db.UpdateTranslation(et.ID, et.Translation)
		return &t, err
	}
	return nil, nil
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
