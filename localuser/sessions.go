package localuser

import (
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/runar-rkmedia/skiver/types"
)

type UserSessionInMemory struct {
	c         *cache.Cache
	t         func() string
	persistor Persistor
	UserSessionOptions
}

type UserSessionOptions struct {
	TTL time.Duration
}

type Persistor interface {
	GetSessions() (map[string]types.Session, error)
	// CreateSession(key string, session types.Session) (types.Session, error)
	CreateSession(key string, session types.Session) (types.Session, error)
	EvictSession(key string) error
}

func NewUserSessionInMemory(options UserSessionOptions, tokenCreator func() string, persistor Persistor) (UserSessionInMemory, error) {
	// We dont want the expiry-check to happen at the same time on the hour, so we add some seconds
	u := UserSessionInMemory{t: tokenCreator, UserSessionOptions: options}
	if persistor != nil {
		u.persistor = persistor
		sessions, err := persistor.GetSessions()
		if err != nil {
			return UserSessionInMemory{}, fmt.Errorf("failed to retrieve sessions from persistance: %w", err)
		}
		items := make(map[string]cache.Item, len(sessions))
		now := time.Now()
		for k, v := range sessions {
			if v.Expires.Before(now) {
				persistor.EvictSession(k)
			}
			items[k] = cache.Item{
				Object:     v,
				Expiration: v.Expires.UnixNano(),
			}
		}
		u.c = cache.NewFrom(options.TTL, 10*time.Minute+7717*time.Millisecond, items)
		u.c.OnEvicted(func(s string, i interface{}) { persistor.EvictSession(s) })
	} else {
		u.c = cache.New(options.TTL, 10*time.Minute+7717*time.Millisecond)
	}

	return u, nil
}

func (us UserSessionInMemory) NewSession(user types.User, organization types.Organization, userAgent string) (s types.Session) {
	token := us.t()
	now := time.Now()
	s = types.Session{
		Token:        token,
		User:         user,
		Organization: organization,
		UserAgent:    userAgent,
		Issued:       now,
		Expires:      now.Add(us.TTL),
	}

	us.c.Set(token, s, us.TTL)
	if us.persistor != nil {
		us.persistor.CreateSession(token, s)
	}
	return s
}
func (us UserSessionInMemory) SessionsForUser(userId string) (s []types.Session) {
	v := us.c.Items()
	for _, val := range v {
		if val.Expired() {
			continue
		}
		if session, ok := val.Object.(types.Session); ok && session.User.ID == userId {
			s = append(s, session)
		}
	}
	return
}
func (us UserSessionInMemory) GetSession(token string) (s types.Session, err error) {
	v, exp, found := us.c.GetWithExpiration(token)
	if !found {
		return s, ErrNotFound
	}
	if time.Now().After(exp) {
		return s, ErrSessionExpired
	}

	s, ok := v.(types.Session)
	if !ok {
		return s, errors.New("Interface is not a session")
	}
	s.Expires = exp
	return s, nil
}

var (
	ErrNotFound       = errors.New("Session not found")
	ErrSessionExpired = errors.New("Session is expired")
)
