package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type AutomationAuthoringSchema struct {
	DefinitionSchema interface{}              `json:"definition_schema,omitempty" path:"definition_schema,omitempty" url:"definition_schema,omitempty"`
	ErrorFamilies    []map[string]interface{} `json:"error_families,omitempty" path:"error_families,omitempty" url:"error_families,omitempty"`
	Nodes            []map[string]interface{} `json:"nodes,omitempty" path:"nodes,omitempty" url:"nodes,omitempty"`
	SchemaUrl        string                   `json:"schema_url,omitempty" path:"schema_url,omitempty" url:"schema_url,omitempty"`
}

// Identifier no path or id

type AutomationAuthoringSchemaCollection []AutomationAuthoringSchema

func (a *AutomationAuthoringSchema) UnmarshalJSON(data []byte) error {
	type automationAuthoringSchema AutomationAuthoringSchema
	var v automationAuthoringSchema
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = AutomationAuthoringSchema(v)
	return nil
}

func (a *AutomationAuthoringSchemaCollection) UnmarshalJSON(data []byte) error {
	type automationAuthoringSchemas AutomationAuthoringSchemaCollection
	var v automationAuthoringSchemas
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AutomationAuthoringSchemaCollection(v)
	return nil
}

func (a *AutomationAuthoringSchemaCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
