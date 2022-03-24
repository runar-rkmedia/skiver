package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type AppHandler = func(requestContext.ReqContext, http.ResponseWriter, *http.Request) (interface{}, error)

func NewAuthHandler(
	userSessions SessionManager,
) func(http.ResponseWriter, *http.Request) (*http.Request, error) {

	return func(rw http.ResponseWriter, r *http.Request) (*http.Request, error) {

		token := r.Header.Get("Authorization")
		if token == "" {
			cookie, err := r.Cookie("token")
			if err != nil {
				if !errors.Is(err, http.ErrNoCookie) {
					return r, err
				}
			}
			if cookie != nil {
				token = cookie.Value
			}

		}
		if token == "" {
			// Using tokens within the url's query-paramaters is not recommended, but Skivers api does not restrict the usage.
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			return r, nil
		}
		sess, err := userSessions.GetSession(token)
		if err == nil {
			expiresD := sess.Expires.Sub(time.Now())
			rw.Header().Add("session-expires", sess.Expires.String())
			rw.Header().Add("session-expires-in", expiresD.String())
			rw.Header().Add("session-expires-in-seconds", strconv.Itoa(int(expiresD.Seconds())))
			// r = r.WithContext(context.WithValue(r.Context(), ContextKeySession, session))
			r = setValue(r, ContextKeySession, sess)
			return r, nil
		}

		return r, nil
	}
}
func setValue(r *http.Request, key string, val interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}

var (
	ErrApiInternalErrorMissingSession = NewApiError("Missing session", http.StatusBadGateway, string(CodeInternalServerError))
)

func ErrApiNotFound(key string, input string) error {
	return NewApiError(fmt.Sprintf("%s not found for '%s'", key, input), http.StatusNotFound, "NotFound:"+key)
}
func ErrApiNotAuthorized(key, verb string) error {
	return NewApiError(fmt.Sprintf("You are not authorized to %s: on %s", verb, key), http.StatusNotFound, "Auth:"+key+"+"+verb)
}
func ErrApiMissingArgument(key string) error {
	return NewApiError("Missign argument: "+key, http.StatusBadRequest, "Missing:"+key)
}
func ErrApiInputValidation(msg, key string) error {
	return NewApiError(msg, http.StatusBadRequest, "InputValidation:"+key)
}
func ErrApiDatabase(key string, err error) error {
	if err == nil {
		return nil
	}
	// TOOD: check the err in a switch-case
	e := NewApiError(err.Error(), http.StatusBadGateway, "Database:"+key)
	return &e
}

func GetParams(r *http.Request) httprouter.Params {
	return httprouter.ParamsFromContext(r.Context())
}

func GetRequestSession(r *http.Request) (session types.Session, err error) {
	err = ErrApiInternalErrorMissingSession
	si := r.Context().Value(ContextKeySession)
	if si == nil {
		return
	}
	s, ok := si.(types.Session)
	if !ok {
		return
	}
	if s.User.ID == "" {
		return
	}
	err = nil
	session = s
	return
}

func writeLogoutCookie(rw http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "token",
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
	}
	http.SetCookie(rw, cookie)
}
func logout(session *types.Session, userSessions SessionManager, rw http.ResponseWriter) error {

	writeLogoutCookie(rw)
	if session == nil || session.User.ID == "" {
		return NewApiError("Not logged in", http.StatusBadRequest, string(requestContext.CodeErrAuthenticationRequired))
	}
	err := userSessions.ClearAllSessionsForUser(session.User.ID)
	if err != nil {
		return NewApiError("Logout failed", http.StatusBadGateway, string(CodeInternalServerError))
	}
	return nil
}

var (
	ContextKeySession = "session"
)
