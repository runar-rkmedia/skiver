package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type AppHandler = func(requestContext.ReqContext, http.ResponseWriter, *http.Request) (interface{}, error)

func NewAuthHandler(
	userSessions SessionManager,
) func(http.ResponseWriter, *http.Request) (*http.Request, error) {

	return func(rw http.ResponseWriter, r *http.Request) (*http.Request, error) {
		cookie, err := r.Cookie("token")
		if err != nil {
			return r, err
		}
		if err == nil {
			sess, err := userSessions.GetSession(cookie.Value)
			if err == nil {
				expiresD := sess.Expires.Sub(time.Now())
				rw.Header().Add("session-expires", sess.Expires.String())
				rw.Header().Add("session-expires-in", expiresD.String())
				rw.Header().Add("session-expires-in-seconds", strconv.Itoa(int(expiresD.Seconds())))
				// r = r.WithContext(context.WithValue(r.Context(), ContextKeySession, session))
				r = setValue(r, ContextKeySession, sess)
				return r, nil
			}
		}

		return r, nil
	}
}
func setValue(r *http.Request, key string, val interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}

var (
	ErrApiInternalErrorMissingSession = NewApiError("Missing session", string(CodeInternalServerError))
)

// Returns a valid session or nil
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

var (
	ContextKeySession = "session"
)