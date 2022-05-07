package localuser

import (
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/types"
)

type UserSessionInMemory struct {
	c         *cache.Cache
	t         func() string
	persistor Persistor
	types.UserSessionOptions
}

type Persistor interface {
	GetSessions() (map[string]types.Session, error)
	// CreateSession(key string, session types.Session) (types.Session, error)
	CreateSession(key string, session types.Session) (types.Session, error)
	EvictSession(key string) error
}

func NewUserSessionInMemory(options types.UserSessionOptions, tokenCreator func() string, persistor Persistor) (UserSessionInMemory, error) {
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
				err = persistor.EvictSession(k)
				if err != nil {
					return u, fmt.Errorf("failed to evict prior sessions: %w", err)
				}
			}
			items[k] = cache.Item{
				Object:     v,
				Expiration: v.Expires.UnixNano(),
			}
		}
		u.c = cache.NewFrom(options.TTL, 10*time.Minute+7717*time.Millisecond, items)
		u.c.OnEvicted(func(s string, i interface{}) {
			err := persistor.EvictSession(s)
			if err != nil {
				l := logger.GetLogger("UserSessionInMemory.OnEvicted")
				l.Error().Err(err).Msg("failed to evict prior sessions")
			}
		})
	} else {
		u.c = cache.New(options.TTL, 10*time.Minute+7717*time.Millisecond)
	}

	return u, nil
}

func (us UserSessionInMemory) NewSession(user types.User, organization types.Organization, userAgent string, opts ...types.UserSessionOptions) (s types.Session) {
	token := us.t()
	return us.newSession(token, user, organization, userAgent, opts...)
}
func (us UserSessionInMemory) ReplaceSession(token string, session types.Session) (s types.Session) {
	return us.setSession(token, session)
}
func (us UserSessionInMemory) setSession(token string, session types.Session) types.Session {
	ttl := session.Expires.Sub(time.Now())
	if ttl <= 0 {
		l := logger.GetLogger("UserSessionInMemory.setSession")
		l.Warn().Time("expires", session.Expires).Msg("Session has expired")
	}
	us.c.Set(token, session, ttl)
	if us.persistor != nil {
		s, err := us.persistor.CreateSession(token, session)
		if err != nil {
			l := logger.GetLogger("UserSessionInMemory.setSession")
			l.Error().Err(err).Msg("failed to create session")
		}
		return s
	}
	return session
}
func (us UserSessionInMemory) newSession(token string, user types.User, organization types.Organization, userAgent string, opts ...types.UserSessionOptions) (s types.Session) {
	var options types.UserSessionOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	now := time.Now()
	var ttl time.Duration
	if options.TTL > 0 {
		ttl = options.TTL
	} else {
		ttl = us.UserSessionOptions.TTL
	}
	s = types.Session{
		Token:        token,
		User:         user,
		Organization: organization,
		UserAgent:    userAgent,
		Issued:       now,
		Expires:      now.Add(ttl),
	}

	return us.setSession(token, s)
}
func (us UserSessionInMemory) TTL() time.Duration {
	return us.UserSessionOptions.TTL
}
func (us UserSessionInMemory) ClearSession(token string) error {
	_, err := us.GetSession(token)
	us.c.Delete(token)
	if err != nil && err == ErrNotFound {
		return nil
	}
	return err
}
func (us UserSessionInMemory) ClearAllSessionsForUser(userId string) error {
	s := us.SessionsForUser(userId)
	for _, v := range s {
		us.c.Delete(v.Token)
	}
	return nil
}
func (us UserSessionInMemory) UpdateAllSessionsForUser(userId string, user types.User) error {
	s := us.SessionsForUser(userId)
	for _, v := range s {
		v.User = user
		us.setSession(v.Token, v)
	}
	return nil
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
