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

// ProjectInput project input
//
// swagger:model ProjectInput
type ProjectInput struct {

	// description
	// Max Length: 8000
	// Min Length: 1
	Description string `json:"description,omitempty"`

	// locales
	Locales map[string]LocaleSetting `json:"locales,omitempty"`

	// short name
	// Required: true
	// Max Length: 20
	// Min Length: 1
	// Pattern: ^[a-z1-9]*$
	ShortName *string `json:"short_name"`

	// title
	// Required: true
	// Max Length: 400
	// Min Length: 1
	Title *string `json:"title"`
}

// Validate validates this project input
func (m *ProjectInput) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescription(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLocales(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateShortName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTitle(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ProjectInput) validateDescription(formats strfmt.Registry) error {
	if swag.IsZero(m.Description) { // not required
		return nil
	}

	if err := validate.MinLength("description", "body", m.Description, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("description", "body", m.Description, 8000); err != nil {
		return err
	}

	return nil
}

func (m *ProjectInput) validateLocales(formats strfmt.Registry) error {
	if swag.IsZero(m.Locales) { // not required
		return nil
	}

	for k := range m.Locales {

		if err := validate.Required("locales"+"."+k, "body", m.Locales[k]); err != nil {
			return err
		}
		if val, ok := m.Locales[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("locales" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("locales" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *ProjectInput) validateShortName(formats strfmt.Registry) error {

	if err := validate.Required("short_name", "body", m.ShortName); err != nil {
		return err
	}

	if err := validate.MinLength("short_name", "body", *m.ShortName, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("short_name", "body", *m.ShortName, 20); err != nil {
		return err
	}

	if err := validate.Pattern("short_name", "body", *m.ShortName, `^[a-z1-9]*$`); err != nil {
		return err
	}

	return nil
}

func (m *ProjectInput) validateTitle(formats strfmt.Registry) error {

	if err := validate.Required("title", "body", m.Title); err != nil {
		return err
	}

	if err := validate.MinLength("title", "body", *m.Title, 1); err != nil {
		return err
	}

	if err := validate.MaxLength("title", "body", *m.Title, 400); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this project input based on the context it is used
func (m *ProjectInput) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateLocales(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ProjectInput) contextValidateLocales(ctx context.Context, formats strfmt.Registry) error {

	for k := range m.Locales {

		if val, ok := m.Locales[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *ProjectInput) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ProjectInput) UnmarshalBinary(b []byte) error {
	var res ProjectInput
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
