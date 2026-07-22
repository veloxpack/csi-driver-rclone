package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type DesktopConfigurationProfile struct {
	Id             int64       `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name           string      `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	WorkspaceId    int64       `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	UseForAllUsers *bool       `json:"use_for_all_users,omitempty" path:"use_for_all_users,omitempty" url:"use_for_all_users,omitempty"`
	MountMappings  interface{} `json:"mount_mappings,omitempty" path:"mount_mappings,omitempty" url:"mount_mappings,omitempty"`
}

func (d DesktopConfigurationProfile) Identifier() interface{} {
	return d.Id
}

type DesktopConfigurationProfileCollection []DesktopConfigurationProfile

type DesktopConfigurationProfileListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type DesktopConfigurationProfileFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type DesktopConfigurationProfileCreateParams struct {
	Name           string      `url:"name" json:"name" path:"name"`
	MountMappings  interface{} `url:"mount_mappings" json:"mount_mappings" path:"mount_mappings"`
	WorkspaceId    int64       `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	UseForAllUsers *bool       `url:"use_for_all_users,omitempty" json:"use_for_all_users,omitempty" path:"use_for_all_users"`
}

type DesktopConfigurationProfileUpdateParams struct {
	Id             int64       `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name           string      `url:"name,omitempty" json:"name,omitempty" path:"name"`
	WorkspaceId    int64       `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	MountMappings  interface{} `url:"mount_mappings,omitempty" json:"mount_mappings,omitempty" path:"mount_mappings"`
	UseForAllUsers *bool       `url:"use_for_all_users,omitempty" json:"use_for_all_users,omitempty" path:"use_for_all_users"`
}

type DesktopConfigurationProfileDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (d *DesktopConfigurationProfile) UnmarshalJSON(data []byte) error {
	type desktopConfigurationProfile DesktopConfigurationProfile
	var v desktopConfigurationProfile
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*d = DesktopConfigurationProfile(v)
	return nil
}

func (d *DesktopConfigurationProfileCollection) UnmarshalJSON(data []byte) error {
	type desktopConfigurationProfiles DesktopConfigurationProfileCollection
	var v desktopConfigurationProfiles
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*d = DesktopConfigurationProfileCollection(v)
	return nil
}

func (d *DesktopConfigurationProfileCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*d))
	for i, v := range *d {
		ret[i] = v
	}

	return &ret
}
