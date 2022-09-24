package config

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

func CreateJsonSchemaForConfig() ([]byte, error) {
	r := new(jsonschema.Reflector)
	r.AddGoComments("github.com/runar-rkmedia/skiver", "./config")
	r.RequiredFromJSONSchemaTags = true
	schema := r.Reflect(Config{})
	b, err := json.MarshalIndent(schema, "", "  ")
	return b, err
}
