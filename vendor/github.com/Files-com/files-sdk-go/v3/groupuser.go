package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type GroupUser struct {
	GroupName string `json:"group_name,omitempty" path:"group_name,omitempty" url:"group_name,omitempty"`
	GroupId   int64  `json:"group_id,omitempty" path:"group_id,omitempty" url:"group_id,omitempty"`
	UserId    int64  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Admin     *bool  `json:"admin,omitempty" path:"admin,omitempty" url:"admin,omitempty"`
	Usernames string `json:"usernames,omitempty" path:"usernames,omitempty" url:"usernames,omitempty"`
	Id        int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
}

func (g GroupUser) Identifier() interface{} {
	return g.Id
}

type GroupUserCollection []GroupUser

type GroupUserListParams struct {
	UserId  int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	GroupId int64 `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	ListParams
}

type GroupUserCreateParams struct {
	GroupId int64 `url:"group_id" json:"group_id" path:"group_id"`
	UserId  int64 `url:"user_id" json:"user_id" path:"user_id"`
	Admin   *bool `url:"admin,omitempty" json:"admin,omitempty" path:"admin"`
}

type GroupUserUpdateParams struct {
	Id      int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
	GroupId int64 `url:"group_id" json:"group_id" path:"group_id"`
	UserId  int64 `url:"user_id" json:"user_id" path:"user_id"`
	Admin   *bool `url:"admin,omitempty" json:"admin,omitempty" path:"admin"`
}

type GroupUserDeleteParams struct {
	Id      int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
	GroupId int64 `url:"group_id" json:"group_id" path:"group_id"`
	UserId  int64 `url:"user_id" json:"user_id" path:"user_id"`
}

func (g *GroupUser) UnmarshalJSON(data []byte) error {
	type groupUser GroupUser
	var v groupUser
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*g = GroupUser(v)
	return nil
}

func (g *GroupUserCollection) UnmarshalJSON(data []byte) error {
	type groupUsers GroupUserCollection
	var v groupUsers
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*g = GroupUserCollection(v)
	return nil
}

func (g *GroupUserCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*g))
	for i, v := range *g {
		ret[i] = v
	}

	return &ret
}
