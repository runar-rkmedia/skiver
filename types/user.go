package types

import "time"

type User struct {
	Entity
	UserName string `json:"userName"`
	// If not active, the account cannot be used until any issues are resolved.
	Active bool
	Store  UserStore `json:"-"`
	// If set, the user must change the password before the account can be used
	TemporaryPassword bool   `json:"temporary_password,omitempty"`
	PW                []byte `json:"-"`
}

type UserStore int

const (
	// A local user, with password stored in the database.
	UserStoreLocal UserStore = iota + 1
)

// Locale represents a language, dialect etc.
// swagger:response
type loginResponse struct {
	// in:body
	Body LoginResponse
}

type LoginResponse struct {
	Ok        bool
	Expires   time.Time
	ExpiresIn string
}
