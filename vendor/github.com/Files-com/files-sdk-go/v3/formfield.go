package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FormField struct {
	Id               int64    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Label            string   `json:"label,omitempty" path:"label,omitempty" url:"label,omitempty"`
	Required         *bool    `json:"required,omitempty" path:"required,omitempty" url:"required,omitempty"`
	HelpText         string   `json:"help_text,omitempty" path:"help_text,omitempty" url:"help_text,omitempty"`
	FieldType        string   `json:"field_type,omitempty" path:"field_type,omitempty" url:"field_type,omitempty"`
	OptionsForSelect []string `json:"options_for_select,omitempty" path:"options_for_select,omitempty" url:"options_for_select,omitempty"`
	DefaultOption    string   `json:"default_option,omitempty" path:"default_option,omitempty" url:"default_option,omitempty"`
	FormFieldSetId   int64    `json:"form_field_set_id,omitempty" path:"form_field_set_id,omitempty" url:"form_field_set_id,omitempty"`
}

func (f FormField) Identifier() interface{} {
	return f.Id
}

type FormFieldCollection []FormField

func (f *FormField) UnmarshalJSON(data []byte) error {
	type formField FormField
	var v formField
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FormField(v)
	return nil
}

func (f *FormFieldCollection) UnmarshalJSON(data []byte) error {
	type formFields FormFieldCollection
	var v formFields
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FormFieldCollection(v)
	return nil
}

func (f *FormFieldCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
