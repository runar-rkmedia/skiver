package types

import (
	"time"
)

type User struct {
	Entity
	UserName string `json:"userName"`
	// If not active, the account cannot be used until any issues are resolved.
	Active bool
	Store  UserStore `json:"-"`
	// If set, the user must change the password before the account can be used
	TemporaryPassword bool   `json:"temporary_password,omitempty"`
	PW                []byte `json:"-"`

	CanCreateOrganization bool `json:"can_create_organization,omitempty"`
	CanCreateUsers        bool `json:"can_create_users,omitempty"`
	CanCreateProjects     bool `json:"can_create_projects,omitempty"`
	CanCreateTranslations bool `json:"can_create_translations,omitempty"`
	CanCreateLocales      bool `json:"can_create_locales,omitempty"`

	CanUpdateOrganization bool `json:"can_update_organization,omitempty"`
	CanUpdateUsers        bool `json:"can_update_users,omitempty"`
	CanUpdateProjects     bool `json:"can_update_projects,omitempty"`
	CanUpdateTranslations bool `json:"can_update_translations,omitempty"`
	CanUpdateLocales      bool `json:"can_update_locales,omitempty"`
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
	User
	Organization Organization `json:"organization"`
	Ok           bool         `json:"ok"`
	Expires      time.Time    `json:"expires"`
	ExpiresIn    string       `json:"expires_in"`
}

type Session struct {
	Token        string
	User         User
	Organization Organization
	UserAgent    string
	Issued       time.Time
	Expires      time.Time
}

type Organization struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	CreatedBy string     `json:"createdBy"`
	UpdatedBy string     `json:"updatedBy,omitempty"`
	Deleted   *time.Time `json:"deleted,omitempty"`

	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// This will allow anybody with the id to create a standard user, and join the organization
	// The first user to join, gets priviliges to administer the organization.
	JoinID        string    `json:"join_id,omitempty"`
	JoinIDExpires time.Time `json:"join_id_expires"`
}

func (e User) Namespace() string {
	return e.Kind()
}
func (e User) Kind() string {
	return string(PubTypeUser)
}
