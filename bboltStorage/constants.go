package bboltStorage

type PubType string
type PubVerb string

// TODO: move these to types.
const (
	// PubTypeUser               PubType = "user"
	PubTypeTranslation        PubType = "translation"
	PubTypeMissingTranslation PubType = "missingTranslation"
	PubTypeTranslationValue   PubType = "translationValue"
	PubTypeCategory           PubType = "category"
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
