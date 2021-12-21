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
	// Example: Project-description
	// Max Length: 8000
	Description string `json:"description,omitempty"`

	// If present, any translations with tags matching will also be included in the exported translations
	// If the project contains conflicting translations, the project has presedence.
	// Example: ["actions","general"]
	IncludedTags []string `json:"included_tags"`

	// title
	// Example: My Great Project
	// Required: true
	// Max Length: 400
	// Min Length: 2
	Title *string `json:"title"`
}

// Validate validates this project input
func (m *ProjectInput) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescription(formats); err != nil {
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

	if err := validate.MaxLength("description", "body", m.Description, 8000); err != nil {
		return err
	}

	return nil
}

func (m *ProjectInput) validateTitle(formats strfmt.Registry) error {

	if err := validate.Required("title", "body", m.Title); err != nil {
		return err
	}

	if err := validate.MinLength("title", "body", *m.Title, 2); err != nil {
		return err
	}

	if err := validate.MaxLength("title", "body", *m.Title, 400); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this project input based on context it is used
func (m *ProjectInput) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
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
