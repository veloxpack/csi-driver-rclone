package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ShareGroupMember struct {
	Name    string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Company string `json:"company,omitempty" path:"company,omitempty" url:"company,omitempty"`
	Email   string `json:"email,omitempty" path:"email,omitempty" url:"email,omitempty"`
}

// Identifier no path or id

type ShareGroupMemberCollection []ShareGroupMember

func (s *ShareGroupMember) UnmarshalJSON(data []byte) error {
	type shareGroupMember ShareGroupMember
	var v shareGroupMember
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = ShareGroupMember(v)
	return nil
}

func (s *ShareGroupMemberCollection) UnmarshalJSON(data []byte) error {
	type shareGroupMembers ShareGroupMemberCollection
	var v shareGroupMembers
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = ShareGroupMemberCollection(v)
	return nil
}

func (s *ShareGroupMemberCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
