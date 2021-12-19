package models

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
)

func (m *LoginPayload) SuperValidate(formats strfmt.Registry) error {
	err := m.Validate(formats)
	if err == nil {
		return nil
	}
	errs := err.(*errors.CompositeError)
	for i := 0; i < len(errs.Errors); i++ {
		v := errs.Errors[i].(*errors.Validation)
		v.Value = "*REDACTED*"
	}
	return errs
}
