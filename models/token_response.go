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

// TokenResponse token response
//
// swagger:model TokenResponse
type TokenResponse struct {

	// Description of user-generated-token, or for login-tokens, this will be the last User-Agent used
	Description string `json:"description,omitempty"`

	// expires
	// Format: date-time
	Expires strfmt.DateTime `json:"expires,omitempty"`

	// issued
	// Format: date-time
	Issued strfmt.DateTime `json:"issued,omitempty"`

	// token
	Token string `json:"token,omitempty"`
}

// Validate validates this token response
func (m *TokenResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateExpires(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIssued(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *TokenResponse) validateExpires(formats strfmt.Registry) error {
	if swag.IsZero(m.Expires) { // not required
		return nil
	}

	if err := validate.FormatOf("expires", "body", "date-time", m.Expires.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *TokenResponse) validateIssued(formats strfmt.Registry) error {
	if swag.IsZero(m.Issued) { // not required
		return nil
	}

	if err := validate.FormatOf("issued", "body", "date-time", m.Issued.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this token response based on context it is used
func (m *TokenResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *TokenResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TokenResponse) UnmarshalBinary(b []byte) error {
	var res TokenResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
