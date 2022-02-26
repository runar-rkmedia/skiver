package types

type PubType string
type PubVerb string

const (
	PubTypeUser               PubType = "user"
	PubTypeTranslation        PubType = "translation"
	PubTypeMissingTranslation PubType = "missingTranslation"
	PubTypeTranslationValue   PubType = "translationValue"
	PubTypeCategory           PubType = "category"
	PubTypeSnapshot           PubType = "snapshot"
	PubTypeLocale             PubType = "locale"
	PubTypeProject            PubType = "project"
	PubTypeOrganization       PubType = "organization"

	PubVerbCreate PubVerb = "create"
	PubVerbUpdate PubVerb = "update"
	// Marks the item as deleted in the database, but does not delete it
	PubVerbSoftDelete PubVerb = "soft-delete"
	// Removes all items permanently
	PubVerbClean       PubVerb = "clean"
	PubVerbConnectItem PubVerb = "connect"
)
