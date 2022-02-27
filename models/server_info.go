// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// ServerInfo server info
//
// swagger:model ServerInfo
type ServerInfo struct {

	// Date of build
	// Format: date-time
	BuildDate strfmt.DateTime `json:"build_date,omitempty"`

	// Size of database.
	DatabaseSize int64 `json:"database_size,omitempty"`

	// database size str
	DatabaseSizeStr string `json:"database_size_str,omitempty"`

	// Short githash for current commit
	GitHash string `json:"git_hash,omitempty"`

	// When the server was started
	// Format: date-time
	ServerStartedAt strfmt.DateTime `json:"server_started_at,omitempty"`

	// Version-number for commit
	Version string `json:"version,omitempty"`
}

// Validate validates this server info
func (m *ServerInfo) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBuildDate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateServerStartedAt(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ServerInfo) validateBuildDate(formats strfmt.Registry) error {
	if swag.IsZero(m.BuildDate) { // not required
		return nil
	}

	if err := validate.FormatOf("build_date", "body", "date-time", m.BuildDate.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *ServerInfo) validateServerStartedAt(formats strfmt.Registry) error {
	if swag.IsZero(m.ServerStartedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("server_started_at", "body", "date-time", m.ServerStartedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this server info based on context it is used
func (m *ServerInfo) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ServerInfo) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ServerInfo) UnmarshalBinary(b []byte) error {
	var res ServerInfo
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
