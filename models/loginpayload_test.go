package models

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/ghodss/yaml"
)

func TestLoginInput_Validate(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"no fail on valid input",
			fields{"joe_johnson", "so_secret_so_great"},
			false,
		},
		{
			"fail on empty input",
			fields{},
			true,
		},
		{
			"fail on short username",
			fields{"a", "xxxxxxxxxxxxx"},
			true,
		},
		{
			"fail on username with spaces",
			fields{"barney foo", "xxxxxxxxxxxxx"},
			true,
		},
		{
			"fail on short password",
			fields{"abc", "b"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LoginInput{
				Username: &tt.fields.Username,
				Password: &tt.fields.Password,
			}
			got := Validate(&l)

			if !tt.wantErr {
				testza.AssertNoError(t, got)
			} else if got == nil {
				t.Error("expected error, but none was returned")
			}
		})
	}
}
func TestLoginInput_ValidatePassword(t *testing.T) {
	type fields struct {
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantStr string
	}{
		{
			"fail should not output the password",
			fields{"abc", "å"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LoginInput{
				Username: &tt.fields.Username,
				Password: &tt.fields.Password,
			}
			got := Validate(&l)
			if got == nil {
				t.Fatal("expected to fail on password-input")
			}
			y, err := yaml.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(y), tt.fields.Password) {
				t.Errorf("password (%s) should not be included in the output, but was:\n%s", tt.fields.Password, string(y))
			}

		})
	}
}
