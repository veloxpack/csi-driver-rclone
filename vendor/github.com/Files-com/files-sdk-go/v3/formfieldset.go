package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FormFieldSet struct {
	Id          int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Title       string                   `json:"title,omitempty" path:"title,omitempty" url:"title,omitempty"`
	FormLayout  []int64                  `json:"form_layout,omitempty" path:"form_layout,omitempty" url:"form_layout,omitempty"`
	FormFields  []map[string]interface{} `json:"form_fields,omitempty" path:"form_fields,omitempty" url:"form_fields,omitempty"`
	SkipName    *bool                    `json:"skip_name,omitempty" path:"skip_name,omitempty" url:"skip_name,omitempty"`
	SkipEmail   *bool                    `json:"skip_email,omitempty" path:"skip_email,omitempty" url:"skip_email,omitempty"`
	SkipCompany *bool                    `json:"skip_company,omitempty" path:"skip_company,omitempty" url:"skip_company,omitempty"`
	InUse       *bool                    `json:"in_use,omitempty" path:"in_use,omitempty" url:"in_use,omitempty"`
	UserId      int64                    `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
}

func (f FormFieldSet) Identifier() interface{} {
	return f.Id
}

type FormFieldSetCollection []FormFieldSet

type FormFieldSetListParams struct {
	UserId int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	ListParams
}

type FormFieldSetFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type FormFieldSetCreateParams struct {
	UserId      int64                    `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Title       string                   `url:"title,omitempty" json:"title,omitempty" path:"title"`
	SkipEmail   *bool                    `url:"skip_email,omitempty" json:"skip_email,omitempty" path:"skip_email"`
	SkipName    *bool                    `url:"skip_name,omitempty" json:"skip_name,omitempty" path:"skip_name"`
	SkipCompany *bool                    `url:"skip_company,omitempty" json:"skip_company,omitempty" path:"skip_company"`
	FormFields  []map[string]interface{} `url:"form_fields,omitempty" json:"form_fields,omitempty" path:"form_fields"`
}

type FormFieldSetUpdateParams struct {
	Id          int64                    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Title       string                   `url:"title,omitempty" json:"title,omitempty" path:"title"`
	SkipEmail   *bool                    `url:"skip_email,omitempty" json:"skip_email,omitempty" path:"skip_email"`
	SkipName    *bool                    `url:"skip_name,omitempty" json:"skip_name,omitempty" path:"skip_name"`
	SkipCompany *bool                    `url:"skip_company,omitempty" json:"skip_company,omitempty" path:"skip_company"`
	FormFields  []map[string]interface{} `url:"form_fields,omitempty" json:"form_fields,omitempty" path:"form_fields"`
}

type FormFieldSetDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (f *FormFieldSet) UnmarshalJSON(data []byte) error {
	type formFieldSet FormFieldSet
	var v formFieldSet
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FormFieldSet(v)
	return nil
}

func (f *FormFieldSetCollection) UnmarshalJSON(data []byte) error {
	type formFieldSets FormFieldSetCollection
	var v formFieldSets
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FormFieldSetCollection(v)
	return nil
}

func (f *FormFieldSetCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
