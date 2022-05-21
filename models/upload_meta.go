// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// UploadMeta upload meta
//
// swagger:model UploadMeta
type UploadMeta struct {

	// ID
	ID string `json:"id,omitempty"`

	// locale
	Locale string `json:"locale,omitempty"`

	// locale key
	LocaleKey string `json:"locale_key,omitempty"`

	// parent
	Parent string `json:"parent,omitempty"`

	// provider ID
	ProviderID string `json:"provider_id,omitempty"`

	// size
	Size int64 `json:"size,omitempty"`

	// tag
	Tag string `json:"tag,omitempty"`

	// URL
	URL string `json:"url,omitempty"`
}

// Validate validates this upload meta
func (m *UploadMeta) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this upload meta based on context it is used
func (m *UploadMeta) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UploadMeta) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UploadMeta) UnmarshalBinary(b []byte) error {
	var res UploadMeta
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
