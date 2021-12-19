package models

import "github.com/go-openapi/strfmt"

var (
	formats = strfmt.NewFormats()
)

type Validator interface {
	Validate(formats strfmt.Registry) error
}
type SuperValidator interface {
	SuperValidate(formats strfmt.Registry) error
}

func Validate(m Validator) error {
	if sv, ok := m.(SuperValidator); ok {
		return sv.SuperValidate(formats)
	}
	return m.Validate(formats)
}
