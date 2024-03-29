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

// ExtendedCategory extended category
//
// swagger:model ExtendedCategory
type ExtendedCategory struct {

	// Time of which the entity was created in the database
	// Required: true
	// Format: date-time
	CreatedAt *strfmt.DateTime `json:"created_at"`

	// User id refering to the user who created the item
	CreatedBy string `json:"created_by,omitempty"`

	// If set, the item is considered deleted. The item will normally not get deleted from the database,
	// but it may if cleanup is required.
	// Format: date-time
	Deleted strfmt.DateTime `json:"deleted,omitempty"`

	// description
	Description string `json:"description,omitempty"`

	// TODO: change to map
	Exists bool `json:"exists,omitempty"`

	// Unique identifier of the entity
	// Required: true
	ID *string `json:"id"`

	// key
	Key string `json:"key,omitempty"`

	// project ID
	ProjectID string `json:"project_id,omitempty"`

	// title
	Title string `json:"title,omitempty"`

	// translation i ds
	TranslationIDs []string `json:"translation_ids"`

	// translations
	Translations map[string]ExtendedTranslation `json:"translations,omitempty"`

	// Time of which the entity was updated, if any
	// Format: date-time
	UpdatedAt strfmt.DateTime `json:"updated_at,omitempty"`

	// User id refering to who created the item
	UpdatedBy string `json:"updated_by,omitempty"`
}

// Validate validates this extended category
func (m *ExtendedCategory) Validate(formats strfmt.Registry) error {
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

	if err := m.validateTranslations(formats); err != nil {
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

func (m *ExtendedCategory) validateCreatedAt(formats strfmt.Registry) error {

	if err := validate.Required("created_at", "body", m.CreatedAt); err != nil {
		return err
	}

	if err := validate.FormatOf("created_at", "body", "date-time", m.CreatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *ExtendedCategory) validateDeleted(formats strfmt.Registry) error {
	if swag.IsZero(m.Deleted) { // not required
		return nil
	}

	if err := validate.FormatOf("deleted", "body", "date-time", m.Deleted.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *ExtendedCategory) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	return nil
}

func (m *ExtendedCategory) validateTranslations(formats strfmt.Registry) error {
	if swag.IsZero(m.Translations) { // not required
		return nil
	}

	for k := range m.Translations {

		if err := validate.Required("translations"+"."+k, "body", m.Translations[k]); err != nil {
			return err
		}
		if val, ok := m.Translations[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("translations" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("translations" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *ExtendedCategory) validateUpdatedAt(formats strfmt.Registry) error {
	if swag.IsZero(m.UpdatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("updated_at", "body", "date-time", m.UpdatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this extended category based on the context it is used
func (m *ExtendedCategory) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateTranslations(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ExtendedCategory) contextValidateTranslations(ctx context.Context, formats strfmt.Registry) error {

	for k := range m.Translations {

		if val, ok := m.Translations[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *ExtendedCategory) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ExtendedCategory) UnmarshalBinary(b []byte) error {
	var res ExtendedCategory
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
