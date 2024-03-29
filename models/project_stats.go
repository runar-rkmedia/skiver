// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ProjectStats project stats
//
// swagger:model ProjectStats
type ProjectStats struct {

	// hash
	Hash string `json:"hash,omitempty"`

	// identi hash
	IdentiHash []uint8 `json:"identi_hash"`

	// project ID
	ProjectID string `json:"project_id,omitempty"`

	// size
	Size uint64 `json:"size,omitempty"`

	// size humanized
	SizeHumanized string `json:"size_humanized,omitempty"`

	// tag
	Tag string `json:"tag,omitempty"`
}

// Validate validates this project stats
func (m *ProjectStats) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this project stats based on context it is used
func (m *ProjectStats) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ProjectStats) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProjectStats) UnmarshalBinary(b []byte) error {
	var res ProjectStats
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
