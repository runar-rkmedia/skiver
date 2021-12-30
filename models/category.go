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

// Category category
//
// swagger:model Category
type Category struct {

	// Time of which the entity was created in the database
	// Required: true
	// Format: date-time
	CreatedAt *strfmt.DateTime `json:"createdAt"`

	// User id refering to the user who created the item
	CreatedBy string `json:"createdBy,omitempty"`

	// If set, the item is considered deleted. The item will normally not get deleted from the database,
	// but it may if cleanup is required.
	// Format: date-time
	Deleted strfmt.DateTime `json:"deleted,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// Unique identifier of the entity
	// Required: true
	ID *string `json:"id"`

	// key
	Key string `json:"key,omitempty"`

	// project ID
	ProjectID string `json:"project_id,omitempty"`

	// title
	Title string `json:"title,omitempty"`

	// Time of which the entity was updated, if any
	// Format: date-time
	UpdatedAt strfmt.DateTime `json:"updatedAt,omitempty"`

	// User id refering to who created the item
	UpdatedBy string `json:"updatedBy,omitempty"`
}

// Validate validates this category
func (m *Category) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCreatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDeleted(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUpdatedAt(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Category) validateCreatedAt(formats strfmt.Registry) error {

	if err := validate.Required("createdAt", "body", m.CreatedAt); err != nil {
		return err
	}

	if err := validate.FormatOf("createdAt", "body", "date-time", m.CreatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Category) validateDeleted(formats strfmt.Registry) error {
	if swag.IsZero(m.Deleted) { // not required
		return nil
	}

	if err := validate.FormatOf("deleted", "body", "date-time", m.Deleted.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Category) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	return nil
}

func (m *Category) validateUpdatedAt(formats strfmt.Registry) error {
	if swag.IsZero(m.UpdatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("updatedAt", "body", "date-time", m.UpdatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this category based on context it is used
func (m *Category) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Category) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Category) UnmarshalBinary(b []byte) error {
	var res Category
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
