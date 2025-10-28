package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Permission struct {
	Id         int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Path       string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	UserId     int64  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username   string `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	GroupId    int64  `json:"group_id,omitempty" path:"group_id,omitempty" url:"group_id,omitempty"`
	GroupName  string `json:"group_name,omitempty" path:"group_name,omitempty" url:"group_name,omitempty"`
	Permission string `json:"permission,omitempty" path:"permission,omitempty" url:"permission,omitempty"`
	Recursive  *bool  `json:"recursive,omitempty" path:"recursive,omitempty" url:"recursive,omitempty"`
	SiteId     int64  `json:"site_id,omitempty" path:"site_id,omitempty" url:"site_id,omitempty"`
}

func (p Permission) Identifier() interface{} {
	return p.Id
}

type PermissionCollection []Permission

type PermissionListParams struct {
	SortBy        map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter        Permission             `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix  map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	Path          string                 `url:"path,omitempty" json:"path,omitempty" path:"path"`
	IncludeGroups *bool                  `url:"include_groups,omitempty" json:"include_groups,omitempty" path:"include_groups"`
	GroupId       string                 `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	UserId        string                 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	ListParams
}

type PermissionCreateParams struct {
	Path       string `url:"path" json:"path" path:"path"`
	GroupId    int64  `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	Permission string `url:"permission,omitempty" json:"permission,omitempty" path:"permission"`
	Recursive  *bool  `url:"recursive,omitempty" json:"recursive,omitempty" path:"recursive"`
	UserId     int64  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Username   string `url:"username,omitempty" json:"username,omitempty" path:"username"`
	GroupName  string `url:"group_name,omitempty" json:"group_name,omitempty" path:"group_name"`
	SiteId     int64  `url:"site_id,omitempty" json:"site_id,omitempty" path:"site_id"`
}

type PermissionDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *Permission) UnmarshalJSON(data []byte) error {
	type permission Permission
	var v permission
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = Permission(v)
	return nil
}

func (p *PermissionCollection) UnmarshalJSON(data []byte) error {
	type permissions PermissionCollection
	var v permissions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PermissionCollection(v)
	return nil
}

func (p *PermissionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
