package types

import (
	"time"
)

type UserStorage interface {
	GetUser(userId string) (*User, error)
	FindUsers(max int, filter ...User) (map[string]User, error)
	FindUserByUserName(organizationID, userName string) (*User, error)
	CreateUser(user User) (User, error)
}

type Storage interface {
	UserStorage
	Size() (int64, error)

	GetOrganization(organizationID string) (*Organization, error)
	GetOrganizations() (map[string]Organization, error)
	CreateOrganization(organization Organization) (Organization, error)

	GetLocale(ID string) (Locale, error)
	CreateLocale(locale Locale) (Locale, error)
	GetLocaleFilter(filter ...Locale) (*Locale, error)
	GetLocales() (map[string]Locale, error)
	GetLocaleByIDOrShortName(shortNameOrId string) (*Locale, error)

	GetProject(ID string) (*Project, error)
	CreateProject(project Project) (Project, error)
	UpdateProject(id string, project Project) (Project, error)
	GetProjects() (map[string]Project, error)
	GetProjectByIDOrShortName(shortNameOrId string) (*Project, error)

	GetTranslation(ID string) (*Translation, error)
	CreateTranslation(locale Translation) (Translation, error)
	GetTranslations() (map[string]Translation, error)
	GetTranslationsFilter(max int, filter ...Translation) (map[string]Translation, error)
	UpdateTranslation(id string, paylaod Translation) (Translation, error)

	// These must be added
	GetCategory(ID string) (*Category, error)
	CreateCategory(category Category) (Category, error)
	GetCategories() (map[string]Category, error)
	UpdateCategory(id string, category Category) (Category, error)

	GetTranslationValue(ID string) (*TranslationValue, error)
	CreateTranslationValue(translationValue TranslationValue) (TranslationValue, error)
	// TODO: this should take in a id as first parameter
	UpdateTranslationValue(tv TranslationValue) (TranslationValue, error)
	GetTranslationValues() (map[string]TranslationValue, error)
	GetTranslationValueFilter(filter ...TranslationValue) (*TranslationValue, error)
	GetTranslationValuesFilter(max int, filter ...TranslationValue) (map[string]TranslationValue, error)

	ReportMissing(key MissingTranslation) (*MissingTranslation, error)
	GetMissingKeysFilter(max int, filter ...MissingTranslation) (map[string]MissingTranslation, error)
	UpdateUser(id string, payload User) (User, error)

	GetSnapshot(snapshotId string) (*ProjectSnapshot, error)
	FindSnapshots(max int, filter ...ProjectSnapshot) (map[string]ProjectSnapshot, error)
	CreateSnapshot(snapshot ProjectSnapshot) (ProjectSnapshot, error)
	FindOneSnapshot(filter ...ProjectSnapshot) (*ProjectSnapshot, error)
}

type Entity struct {
	// Time of which the entity was created in the database
	// Required: true
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Time of which the entity was updated, if any
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// Unique identifier of the entity
	// Required: true
	ID string `json:"id,omitempty"`
	// User id refering to the user who created the item
	CreatedBy string `json:"created_by,omitempty"`
	// User id refering to who created the item
	UpdatedBy string `json:"updated_by,omitempty"`
	// If set, the item is considered deleted. The item will normally not get deleted from the database,
	// but it may if cleanup is required.
	Deleted *time.Time `json:"deleted,omitempty"`
	// Organizations are completely seperate from each-other.
	OrganizationID string `json:"-"`
}

func (e Entity) IDString() string {
	return e.ID
}
