package types

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/qri-io/jsonschema"
)

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func ValidateJsonSchema(payload interface{}, schema *jsonschema.Schema) (*[]jsonschema.KeyError, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	errs, err := schema.ValidateBytes(context.Background(), b)
	if err != nil {
		return nil, err
	}
	if errs == nil {
		return nil, nil
	}
	if len(errs) == 0 {
		return nil, err
	}
	return &errs, err
}

func MustCreateJsonSchema(raw string) *jsonschema.Schema {
	rs := &jsonschema.Schema{}
	if err := json.Unmarshal([]byte(raw), rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
	return rs
}

var loginSchema = MustCreateJsonSchema(`
{
  "type": "object",
  "properties": {
    "username": {
      "type": "string",
			"minLength": 3,
			"pattern": "^[^\\s]*$",
			"maxLength": 100
    },
    "password": {
      "type": "string",
			"minLength": 3,
			"maxLength": 400
    }
  },
  "required": [
    "username",
    "password"
  ]
}
`)

func (l LoginPayload) Validate() (*[]jsonschema.KeyError, error) {
	errs, err := ValidateJsonSchema(l, loginSchema)
	if errs != nil {
		for i := 0; i < len(*errs); i++ {
			if (*errs)[i].PropertyPath != "/password" {
				continue
			}
			if s, ok := (*errs)[i].InvalidValue.(string); ok {
				(*errs)[i].InvalidValue = strings.ReplaceAll(s, l.Password, "*REDACTED*")
			}
			(*errs)[i].Message = strings.ReplaceAll((*errs)[i].Message, l.Password, "*REDACTED*")
		}
	}
	return errs, err
}
