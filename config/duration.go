package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"
)

// Duration that marshals and unmarshal yaml, toml and json.
type Duration time.Duration

func (d Duration) String() string {
	return d.String()
}
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
func (d *Duration) UnmarshalText(b []byte) error {
	fmt.Println("fooish", string(b))
	return d.UnmarshalJSON(b)
	// x, err := time.ParseDuration(string(b))
	// if err != nil {
	// 	return err
	// }
	// *d = Duration(x)
	// return nil
}
func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(time.Duration(d).String()), nil
}
func (duration *Duration) UnmarshalJSON(b []byte) error {

	var unmarshalledJson interface{}

	err := json.Unmarshal(b, &unmarshalledJson)
	if err != nil {
		return err
	}

	d, err := ParseDuration(unmarshalledJson)
	if err != nil {
		return err
	}
	*duration = d
	return nil
}
func ParseDuration(data interface{}) (d Duration, err error) {

	var dur time.Duration
	switch value := data.(type) {
	case float64:
		dur = time.Duration(value)
	case string:
		dur, err = time.ParseDuration(value)
		if err != nil {
			return d, err
		}
	default:
		return d, fmt.Errorf("invalid duration: %#v", data)
	}
	d = Duration(dur)

	return d, err
}
func DurationViperHookFunc() mapstructure.DecodeHookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		// Check that the data is string
		if f.Kind() != reflect.String {
			return data, nil
		}

		// Check that the target type is our custom type
		if t != reflect.TypeOf(Duration(0)) {
			return data, nil
		}

		// Return the parsed value
		// return Duration(8), nil
		return ParseDuration(data)
	}
}

func (Duration) JSONSchema() *jsonschema.Schema {
	Examples := []interface{}{
		(90 * time.Second).String(),
		(10 * time.Second).String(),
		(150 * time.Minute).String(),
		(150 * time.Millisecond).String(),
	}
	return &jsonschema.Schema{
		Type:        "string",
		Title:       "Duration-type",
		Description: fmt.Sprintf("Textual representation of a duration", Examples...),
		Examples:    Examples,
	}
}
