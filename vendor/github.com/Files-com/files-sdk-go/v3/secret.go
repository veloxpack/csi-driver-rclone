package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Secret struct {
	Id              int64       `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId     int64       `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name            string      `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Description     string      `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	SecretType      string      `json:"secret_type,omitempty" path:"secret_type,omitempty" url:"secret_type,omitempty"`
	Metadata        interface{} `json:"metadata,omitempty" path:"metadata,omitempty" url:"metadata,omitempty"`
	ValueFieldNames []string    `json:"value_field_names,omitempty" path:"value_field_names,omitempty" url:"value_field_names,omitempty"`
	CreatedAt       *time.Time  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt       *time.Time  `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (s Secret) Identifier() interface{} {
	return s.Id
}

type SecretCollection []Secret

type SecretSecretTypeEnum string

func (u SecretSecretTypeEnum) String() string {
	return string(u)
}

func (u SecretSecretTypeEnum) Enum() map[string]SecretSecretTypeEnum {
	return map[string]SecretSecretTypeEnum{
		"basic":       SecretSecretTypeEnum("basic"),
		"token":       SecretSecretTypeEnum("token"),
		"headers":     SecretSecretTypeEnum("headers"),
		"certificate": SecretSecretTypeEnum("certificate"),
		"key_value":   SecretSecretTypeEnum("key_value"),
	}
}

type SecretListParams struct {
	SortBy       interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	ListParams
}

type SecretFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SecretCreateParams struct {
	Name        string               `url:"name" json:"name" path:"name"`
	Description string               `url:"description,omitempty" json:"description,omitempty" path:"description"`
	SecretType  SecretSecretTypeEnum `url:"secret_type" json:"secret_type" path:"secret_type"`
	Metadata    interface{}          `url:"metadata,omitempty" json:"metadata,omitempty" path:"metadata"`
	WorkspaceId int64                `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type SecretUpdateParams struct {
	Id          int64                `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name        string               `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description string               `url:"description,omitempty" json:"description,omitempty" path:"description"`
	SecretType  SecretSecretTypeEnum `url:"secret_type,omitempty" json:"secret_type,omitempty" path:"secret_type"`
	Metadata    interface{}          `url:"metadata,omitempty" json:"metadata,omitempty" path:"metadata"`
}

type SecretDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *Secret) UnmarshalJSON(data []byte) error {
	type secret Secret
	var v secret
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Secret(v)
	return nil
}

func (s *SecretCollection) UnmarshalJSON(data []byte) error {
	type secrets SecretCollection
	var v secrets
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SecretCollection(v)
	return nil
}

func (s *SecretCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
