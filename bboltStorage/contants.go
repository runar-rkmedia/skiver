package bboltStorage

import (
	"time"

	"github.com/runar-rkmedia/gabyoall/utils"
	"github.com/runar-rkmedia/skiver/types"
)

type PubType string
type PubVerb string

const (
	PubTypeUser        PubType = "user"
	PubTypeTranslation PubType = "translation"
	PubTypeLocale      PubType = "locale"
	PubTypeProject     PubType = "project"

	PubVerbCreate PubVerb = "create"
	PubVerbUpdate PubVerb = "update"
	// Marks the item as deleted in the database, but does not delete it
	PubVerbSoftDelete PubVerb = "soft-delete"
	// Removes all items permanently
	PubVerbClean PubVerb = "clean"
)

// Returns an entity for use by database, with id set and createdAt to current time.
// It is guaranteeed to be created correctly, if if it errors.
// The error should be logged.
func ForceNewEntity() (types.Entity, error) {
	id, err := utils.ForceCreateUniqueId()

	return types.Entity{
		ID:        id,
		CreatedAt: time.Now(),
	}, err
}
