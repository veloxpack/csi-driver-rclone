package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ChildSiteManagementPolicy struct {
	Id                  int64                  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	PolicyType          string                 `json:"policy_type,omitempty" path:"policy_type,omitempty" url:"policy_type,omitempty"`
	Name                string                 `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Description         string                 `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	Value               map[string]interface{} `json:"value,omitempty" path:"value,omitempty" url:"value,omitempty"`
	AppliedChildSiteIds []int64                `json:"applied_child_site_ids,omitempty" path:"applied_child_site_ids,omitempty" url:"applied_child_site_ids,omitempty"`
	SkipChildSiteIds    []int64                `json:"skip_child_site_ids,omitempty" path:"skip_child_site_ids,omitempty" url:"skip_child_site_ids,omitempty"`
	CreatedAt           *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt           *time.Time             `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (c ChildSiteManagementPolicy) Identifier() interface{} {
	return c.Id
}

type ChildSiteManagementPolicyCollection []ChildSiteManagementPolicy

type ChildSiteManagementPolicyPolicyTypeEnum string

func (u ChildSiteManagementPolicyPolicyTypeEnum) String() string {
	return string(u)
}

func (u ChildSiteManagementPolicyPolicyTypeEnum) Enum() map[string]ChildSiteManagementPolicyPolicyTypeEnum {
	return map[string]ChildSiteManagementPolicyPolicyTypeEnum{
		"settings": ChildSiteManagementPolicyPolicyTypeEnum("settings"),
	}
}

type ChildSiteManagementPolicyListParams struct {
	ListParams
}

type ChildSiteManagementPolicyFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ChildSiteManagementPolicyCreateParams struct {
	Value            string                                  `url:"value,omitempty" json:"value,omitempty" path:"value"`
	SkipChildSiteIds []int64                                 `url:"skip_child_site_ids,omitempty" json:"skip_child_site_ids,omitempty" path:"skip_child_site_ids"`
	PolicyType       ChildSiteManagementPolicyPolicyTypeEnum `url:"policy_type" json:"policy_type" path:"policy_type"`
	Name             string                                  `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description      string                                  `url:"description,omitempty" json:"description,omitempty" path:"description"`
}

type ChildSiteManagementPolicyUpdateParams struct {
	Id               int64                                   `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Value            string                                  `url:"value,omitempty" json:"value,omitempty" path:"value"`
	SkipChildSiteIds []int64                                 `url:"skip_child_site_ids,omitempty" json:"skip_child_site_ids,omitempty" path:"skip_child_site_ids"`
	PolicyType       ChildSiteManagementPolicyPolicyTypeEnum `url:"policy_type,omitempty" json:"policy_type,omitempty" path:"policy_type"`
	Name             string                                  `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description      string                                  `url:"description,omitempty" json:"description,omitempty" path:"description"`
}

type ChildSiteManagementPolicyDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (c *ChildSiteManagementPolicy) UnmarshalJSON(data []byte) error {
	type childSiteManagementPolicy ChildSiteManagementPolicy
	var v childSiteManagementPolicy
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*c = ChildSiteManagementPolicy(v)
	return nil
}

func (c *ChildSiteManagementPolicyCollection) UnmarshalJSON(data []byte) error {
	type childSiteManagementPolicys ChildSiteManagementPolicyCollection
	var v childSiteManagementPolicys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*c = ChildSiteManagementPolicyCollection(v)
	return nil
}

func (c *ChildSiteManagementPolicyCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*c))
	for i, v := range *c {
		ret[i] = v
	}

	return &ret
}
