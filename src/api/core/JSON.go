package core

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

// JSONSchemaErr describe error in schema validation.
type JSONSchemaErr struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// JSONValidationResult describe result of the schema validation.
type JSONValidationResult struct {
	Valid  bool
	Errors []JSONSchemaErr
}

// ValidateJSON check a JSON string against a JSON schema.
// See: http://json-schema.org/
func ValidateJSON(root, schema, input string) (*JSONValidationResult, error) {
	var f interface{}
	box, err := Assets(root)
	if err != nil {
		return nil, err
	}

	asset, err := box.MustBytes(schema + ".json")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(asset, &f); err != nil {
		return nil, err
	}

	s := f.(map[string]interface{})
	definitions, err := box.MustBytes("definitions.json")
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(definitions, &f); err != nil {
		return nil, err
	}

	defs := f.(map[string]interface{})
	s["definitions"] = defs

	schemaLoader := gojsonschema.NewGoLoader(s)
	objectLoader := gojsonschema.NewStringLoader(input)
	result, err := gojsonschema.Validate(schemaLoader, objectLoader)
	if err != nil {
		return nil, err
	}

	r := JSONValidationResult{}
	r.Valid = result.Valid()

	var errs []JSONSchemaErr
	for _, err := range result.Errors() {
		errs = append(errs, JSONSchemaErr{
			err.Field(),
			err.Type(),
			err.Description(),
		})
	}

	r.Errors = errs
	return &r, nil
}
