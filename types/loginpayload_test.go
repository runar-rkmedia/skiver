package types

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/qri-io/jsonschema"
)

func TestLoginPayload_Validate(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *[]jsonschema.KeyError
		wantErr bool
	}{
		{
			"no keyError on valid input",
			fields{"joe_johnson", "so_secret_so_great"},
			nil,
			false,
		},
		{
			"keyError on short username",
			fields{"a", "xxxxxxxxxxxxx"},
			&[]jsonschema.KeyError{
				{
					PropertyPath: "/username",
					InvalidValue: "a",
					Message:      "min length of 3 characters required: a"},
			},
			false,
		},
		{
			"keyError on username with spaces",
			fields{"barney foo", "xxxxxxxxxxxxx"},
			&[]jsonschema.KeyError{
				{
					PropertyPath: "/username",
					InvalidValue: "barney foo",
					Message:      "regexp pattern ^[^\\s]*$ mismatch on string: barney foo"},
			},
			false,
		},
		{
			"keyError on short password should not output the password",
			fields{"abc", "b"},
			&[]jsonschema.KeyError{
				{
					PropertyPath: "/password",
					InvalidValue: "*REDACTED*",
					Message:      "min length of 3 characters required: *REDACTED*"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LoginPayload{
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			got, err := l.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoginPayload.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			testza.AssertEqual(t, tt.want, got)

		})
	}
}
