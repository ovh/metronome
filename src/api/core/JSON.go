package core

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

type JSONSchemaErr struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type JSONValidationResult struct {
	Valid  bool
	Errors []JSONSchemaErr
}

func ValidateJSON(root, schema, input string) *JSONValidationResult {
	var f interface{}
	err := json.Unmarshal(MustAsset(root+"/schema/"+schema+".json"), &f)
	s := f.(map[string]interface{})
	err = json.Unmarshal(MustAsset(root+"/schema/definitions.json"), &f)
	defs := f.(map[string]interface{})
	s["definitions"] = defs

	schemaLoader := gojsonschema.NewGoLoader(s)
	objectLoader := gojsonschema.NewStringLoader(input)

	result, err := gojsonschema.Validate(schemaLoader, objectLoader)
	if err != nil {
		panic(err)
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

	return &r
}
