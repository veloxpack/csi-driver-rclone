package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ChildSiteManagementPolicy struct {
	Id               int64   `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	SiteId           int64   `json:"site_id,omitempty" path:"site_id,omitempty" url:"site_id,omitempty"`
	SiteSettingName  string  `json:"site_setting_name,omitempty" path:"site_setting_name,omitempty" url:"site_setting_name,omitempty"`
	ManagedValue     string  `json:"managed_value,omitempty" path:"managed_value,omitempty" url:"managed_value,omitempty"`
	SkipChildSiteIds []int64 `json:"skip_child_site_ids,omitempty" path:"skip_child_site_ids,omitempty" url:"skip_child_site_ids,omitempty"`
}

func (c ChildSiteManagementPolicy) Identifier() interface{} {
	return c.Id
}

type ChildSiteManagementPolicyCollection []ChildSiteManagementPolicy

type ChildSiteManagementPolicyListParams struct {
	ListParams
}

type ChildSiteManagementPolicyFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ChildSiteManagementPolicyCreateParams struct {
	SiteSettingName  string  `url:"site_setting_name" json:"site_setting_name" path:"site_setting_name"`
	ManagedValue     string  `url:"managed_value" json:"managed_value" path:"managed_value"`
	SkipChildSiteIds []int64 `url:"skip_child_site_ids,omitempty" json:"skip_child_site_ids,omitempty" path:"skip_child_site_ids"`
}

type ChildSiteManagementPolicyUpdateParams struct {
	Id               int64   `url:"-,omitempty" json:"-,omitempty" path:"id"`
	SiteSettingName  string  `url:"site_setting_name" json:"site_setting_name" path:"site_setting_name"`
	ManagedValue     string  `url:"managed_value" json:"managed_value" path:"managed_value"`
	SkipChildSiteIds []int64 `url:"skip_child_site_ids,omitempty" json:"skip_child_site_ids,omitempty" path:"skip_child_site_ids"`
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
