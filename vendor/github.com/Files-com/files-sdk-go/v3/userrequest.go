package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type UserRequest struct {
	Id      int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name    string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Email   string `json:"email,omitempty" path:"email,omitempty" url:"email,omitempty"`
	Details string `json:"details,omitempty" path:"details,omitempty" url:"details,omitempty"`
	Company string `json:"company,omitempty" path:"company,omitempty" url:"company,omitempty"`
}

func (u UserRequest) Identifier() interface{} {
	return u.Id
}

type UserRequestCollection []UserRequest

type UserRequestListParams struct {
	ListParams
}

type UserRequestFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type UserRequestCreateParams struct {
	Name    string `url:"name" json:"name" path:"name"`
	Email   string `url:"email" json:"email" path:"email"`
	Details string `url:"details" json:"details" path:"details"`
	Company string `url:"company,omitempty" json:"company,omitempty" path:"company"`
}

type UserRequestDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (u *UserRequest) UnmarshalJSON(data []byte) error {
	type userRequest UserRequest
	var v userRequest
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*u = UserRequest(v)
	return nil
}

func (u *UserRequestCollection) UnmarshalJSON(data []byte) error {
	type userRequests UserRequestCollection
	var v userRequests
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*u = UserRequestCollection(v)
	return nil
}

func (u *UserRequestCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*u))
	for i, v := range *u {
		ret[i] = v
	}

	return &ret
}
