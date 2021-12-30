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

		if rc.ContentKind > 0 && (isPost || isPut) {
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
		case "login":
			if isGet {
				if session == nil {
					rc.WriteError("Not logged in", requestContext.CodeErrAuthenticationRequired)
					return
				}
				expiresD := session.Expires.Sub(time.Now())
				rc.WriteOutput(types.LoginResponse{
					User:      session.User,
					Ok:        true,
					Expires:   session.Expires,
					ExpiresIn: expiresD.String(),
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
				session = userSessions.NewSession(user, userAgent)
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
		case "project":
			if isGet {
				projects, err := ctx.DB.GetProjects()
				rc.WriteAuto(projects, err, requestContext.CodeErrProject)
				return
			}
			if isPost {
				var j models.ProjectInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				l := types.Project{
					Title:       *j.Title,
					Description: j.Description,
				}
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
				var j models.TranslationInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				t := types.Translation{
					// TranslationInput: j,
					CategoryID:  *j.CategoryID,
					Key:         *j.Key,
					Context:     j.Context,
					Description: j.Description,
					Title:       j.Title,
				}
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
				var j models.TranslationValueInput
				if err := rc.ValidateBytes(body, &j); err != nil {
					return
				}

				tv := types.TranslationValue{
					LocaleID:      *j.LocaleID,
					TranslationID: *j.TranslationID,
					Value:         *j.Value,
				}
				translationValue, err := ctx.DB.CreateTranslationValue(tv)
				rc.WriteAuto(translationValue, err, requestContext.CodeErrCreateTranslationValue)
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
