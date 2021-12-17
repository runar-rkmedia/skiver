package bboltStorage

import (
	"time"

	"github.com/runar-rkmedia/gabyoall/utils"
)

type PubType string
type PubVerb string

const (
	PubTypeEndpoint PubType = "endpoint"
	PubTypeRequest  PubType = "request"
	PubTypeSchedule PubType = "schedule"
	PubTypeStat     PubType = "stat"

	PubVerbCreate PubVerb = "create"
	PubVerbUpdate PubVerb = "update"
	// Marks the item as deleted in the database, but does not delete it
	PubVerbSoftDelete PubVerb = "soft-delete"
	// Removes all items permanently
	PubVerbClean PubVerb = "clean"
)

type Entity struct {
	// Time of which the entity was created in the database
	// Required: true
	CreatedAt time.Time `json:"createdAt,omitempty"`
	// Time of which the entity was updated, if any
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	// Unique identifier of the entity
	// Required: true
	ID string `json:"id,omitempty"`
	// If set, the item is considered deleted. The item will normally not get deleted from the database,
	// but it may if cleanup is required.
	Deleted *time.Time `json:"deleted,omitempty"`
}

// Returns an entity for use by database, with id set and createdAt to current time.
// It is guaranteeed to be created correctly, if if it errors.
// The error should be logged.
func ForceNewEntity() (Entity, error) {
	id, err := utils.ForceCreateUniqueId()

	return Entity{
		ID:        id,
		CreatedAt: time.Now(),
	}, err
}
