package types

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
	// A local
	UserStoreLocal UserStore = iota + 1
)
