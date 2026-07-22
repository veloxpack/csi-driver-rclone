package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type KeyLifecycleRule struct {
	Id                   int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	KeyType              string `json:"key_type,omitempty" path:"key_type,omitempty" url:"key_type,omitempty"`
	InactivityDays       int64  `json:"inactivity_days,omitempty" path:"inactivity_days,omitempty" url:"inactivity_days,omitempty"`
	ApplyToAllWorkspaces *bool  `json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces,omitempty" url:"apply_to_all_workspaces,omitempty"`
	Name                 string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	WorkspaceId          int64  `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
}

func (k KeyLifecycleRule) Identifier() interface{} {
	return k.Id
}

type KeyLifecycleRuleCollection []KeyLifecycleRule

type KeyLifecycleRuleKeyTypeEnum string

func (u KeyLifecycleRuleKeyTypeEnum) String() string {
	return string(u)
}

func (u KeyLifecycleRuleKeyTypeEnum) Enum() map[string]KeyLifecycleRuleKeyTypeEnum {
	return map[string]KeyLifecycleRuleKeyTypeEnum{
		"gpg": KeyLifecycleRuleKeyTypeEnum("gpg"),
		"ssh": KeyLifecycleRuleKeyTypeEnum("ssh"),
	}
}

type KeyLifecycleRuleListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type KeyLifecycleRuleFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type KeyLifecycleRuleCreateParams struct {
	ApplyToAllWorkspaces *bool                       `url:"apply_to_all_workspaces,omitempty" json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces"`
	KeyType              KeyLifecycleRuleKeyTypeEnum `url:"key_type,omitempty" json:"key_type,omitempty" path:"key_type"`
	InactivityDays       int64                       `url:"inactivity_days,omitempty" json:"inactivity_days,omitempty" path:"inactivity_days"`
	Name                 string                      `url:"name,omitempty" json:"name,omitempty" path:"name"`
	WorkspaceId          int64                       `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type KeyLifecycleRuleUpdateParams struct {
	Id                   int64                       `url:"-,omitempty" json:"-,omitempty" path:"id"`
	ApplyToAllWorkspaces *bool                       `url:"apply_to_all_workspaces,omitempty" json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces"`
	KeyType              KeyLifecycleRuleKeyTypeEnum `url:"key_type,omitempty" json:"key_type,omitempty" path:"key_type"`
	InactivityDays       int64                       `url:"inactivity_days,omitempty" json:"inactivity_days,omitempty" path:"inactivity_days"`
	Name                 string                      `url:"name,omitempty" json:"name,omitempty" path:"name"`
	WorkspaceId          int64                       `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type KeyLifecycleRuleDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (k *KeyLifecycleRule) UnmarshalJSON(data []byte) error {
	type keyLifecycleRule KeyLifecycleRule
	var v keyLifecycleRule
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*k = KeyLifecycleRule(v)
	return nil
}

func (k *KeyLifecycleRuleCollection) UnmarshalJSON(data []byte) error {
	type keyLifecycleRules KeyLifecycleRuleCollection
	var v keyLifecycleRules
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*k = KeyLifecycleRuleCollection(v)
	return nil
}

func (k *KeyLifecycleRuleCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*k))
	for i, v := range *k {
		ret[i] = v
	}

	return &ret
}
