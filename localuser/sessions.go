package localuser

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/runar-rkmedia/skiver/types"
)

type UserSessionInMemory struct {
	c *cache.Cache
	t func() string
	UserSessionOptions
}

type Session struct {
	Token     string
	User      types.User
	UserAgent string
	Issued    time.Time
	Expires   time.Time
}

type UserSessionOptions struct {
	TTL time.Duration
}

func NewUserSessionInMemory(options UserSessionOptions, tokenCreator func() string) UserSessionInMemory {
	// We dont want the expiry-check to happen at the same time on the hour, so we add some seconds
	c := cache.New(options.TTL, 10*time.Minute+7717*time.Millisecond)
	return UserSessionInMemory{c, tokenCreator, options}
}

func (us UserSessionInMemory) NewSession(user types.User, userAgent string) (s Session) {
	token := us.t()
	now := time.Now()
	s = Session{
		Token:     token,
		User:      user,
		UserAgent: userAgent,
		Issued:    now,
		Expires:   now.Add(us.TTL),
	}

	us.c.Set(token, s, us.TTL)
	return s
}
func (us UserSessionInMemory) SessionsForUser(userId string) (s []Session) {
	v := us.c.Items()
	for _, val := range v {
		if val.Expired() {
			continue
		}
		if session, ok := val.Object.(Session); ok && session.User.ID == userId {
			s = append(s, session)
		}
	}
	return
}
func (us UserSessionInMemory) GetSession(token string) (s Session, err error) {
	v, exp, found := us.c.GetWithExpiration(token)
	if !found {
		return s, ErrNotFound
	}
	if time.Now().After(exp) {
		return s, ErrSessionExpired
	}

	s, ok := v.(Session)
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
