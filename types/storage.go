package types

import "time"

type Storage interface {
	// Endpoints() (es map[string]EndpointEntity, err error)
	// Endpoint(id string) (EndpointEntity, error)
	// CreateEndpoint(e EndpointPayload) (EndpointEntity, error)
	// UpdateEndpoint(id string, p EndpointPayload) (EndpointEntity, error)
	// SoftDeleteEndpoint(id string) (EndpointEntity, error)
	Size() (int64, error)
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
	// If set, the item is considered deleted. The item will normally not get deleted from the database,
	// but it may if cleanup is required.
	Deleted *time.Time `json:"deleted,omitempty"`
}
