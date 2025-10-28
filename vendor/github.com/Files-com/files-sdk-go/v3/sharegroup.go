package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ShareGroup struct {
	Id      int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name    string                   `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Notes   string                   `json:"notes,omitempty" path:"notes,omitempty" url:"notes,omitempty"`
	UserId  int64                    `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Members []map[string]interface{} `json:"members,omitempty" path:"members,omitempty" url:"members,omitempty"`
}

func (s ShareGroup) Identifier() interface{} {
	return s.Id
}

type ShareGroupCollection []ShareGroup

type ShareGroupListParams struct {
	UserId int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	ListParams
}

type ShareGroupFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ShareGroupCreateParams struct {
	UserId  int64                    `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Notes   string                   `url:"notes,omitempty" json:"notes,omitempty" path:"notes"`
	Name    string                   `url:"name" json:"name" path:"name"`
	Members []map[string]interface{} `url:"members" json:"members" path:"members"`
}

type ShareGroupUpdateParams struct {
	Id      int64                    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Notes   string                   `url:"notes,omitempty" json:"notes,omitempty" path:"notes"`
	Name    string                   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Members []map[string]interface{} `url:"members,omitempty" json:"members,omitempty" path:"members"`
}

type ShareGroupDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *ShareGroup) UnmarshalJSON(data []byte) error {
	type shareGroup ShareGroup
	var v shareGroup
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = ShareGroup(v)
	return nil
}

func (s *ShareGroupCollection) UnmarshalJSON(data []byte) error {
	type shareGroups ShareGroupCollection
	var v shareGroups
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = ShareGroupCollection(v)
	return nil
}

func (s *ShareGroupCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
