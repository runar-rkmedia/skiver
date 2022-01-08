package types

import (
	"time"
)

type Storage interface {
	Size() (int64, error)

	GetUser(userId string) (User, error)
	GetUserByUserName(userName string) (*User, error)
	CreateUser(user User) (User, error)

	GetLocale(ID string) (Locale, error)
	CreateLocale(locale Locale) (Locale, error)
	GetLocaleFilter(filter ...Locale) (*Locale, error)
	GetLocales() (map[string]Locale, error)

	GetProject(ID string) (*Project, error)
	CreateProject(locale Project) (Project, error)
	GetProjects() (map[string]Project, error)

	GetTranslation(ID string) (*Translation, error)
	CreateTranslation(locale Translation) (Translation, error)
	GetTranslations() (map[string]Translation, error)

	// These must be added
	GetCategory(ID string) (*Category, error)
	CreateCategory(category Category) (Category, error)
	GetCategories() (map[string]Category, error)

	GetTranslationValue(ID string) (*TranslationValue, error)
	CreateTranslationValue(translationValue TranslationValue) (TranslationValue, error)
	UpdateTranslationValue(tv TranslationValue) (TranslationValue, error)
	GetTranslationValues() (map[string]TranslationValue, error)
	GetTranslationValueFilter(filter ...TranslationValue) (*TranslationValue, error)
	GetTranslationValuesFilter(max int, filter ...TranslationValue) (map[string]TranslationValue, error)

	ReportMissing(key MissingTranslation) (*MissingTranslation, error)
	GetMissingKeysFilter(max int, filter ...MissingTranslation) (map[string]MissingTranslation, error)
}

type Entity struct {
	// Time of which the entity was created in the database
	// Required: true
	CreatedAt time.Time `json:"createdAt,omitempty"`
	// Time of which the entity was updated, if any
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	// Unique identifier of the entity
	// Required: true
	ID string `json:"id,omitempty"`
	// User id refering to the user who created the item
	CreatedBy string `json:"createdBy,omitempty"`
	// User id refering to who created the item
	UpdatedBy string `json:"updatedBy,omitempty"`
	// If set, the item is considered deleted. The item will normally not get deleted from the database,
	// but it may if cleanup is required.
	Deleted *time.Time `json:"deleted,omitempty"`
}
