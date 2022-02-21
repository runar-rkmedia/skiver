// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// LocaleSettingInput locale setting input
//
// swagger:model LocaleSettingInput
type LocaleSettingInput struct {

	// If set, will allow registered translation-services to translate from other languages to this locale.
	// This might help speed up translations for new locales.
	// See the Config or Organization-settings for instructions on how to set up translation-services.
	//
	// Organization-settings are not yet available.
	//
	// TODO: implement organization-settings
	AutoTranslation bool `json:"auto_translation,omitempty"`

	// If set, the locale will be visible for editing.
	Enabled bool `json:"enabled,omitempty"`

	// If set, the associated translations will be published in releases.
	// This is useful for when adding new locales, and one don't want to publish it to users until it is complete
	Publish bool `json:"publish,omitempty"`
}

// Validate validates this locale setting input
func (m *LocaleSettingInput) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this locale setting input based on context it is used
func (m *LocaleSettingInput) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *LocaleSettingInput) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *LocaleSettingInput) UnmarshalBinary(b []byte) error {
	var res LocaleSettingInput
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
