package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Lock struct {
	Path                 string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Timeout              int64  `json:"timeout,omitempty" path:"timeout,omitempty" url:"timeout,omitempty"`
	Depth                string `json:"depth,omitempty" path:"depth,omitempty" url:"depth,omitempty"`
	Recursive            *bool  `json:"recursive,omitempty" path:"recursive,omitempty" url:"recursive,omitempty"`
	Owner                string `json:"owner,omitempty" path:"owner,omitempty" url:"owner,omitempty"`
	Scope                string `json:"scope,omitempty" path:"scope,omitempty" url:"scope,omitempty"`
	Exclusive            *bool  `json:"exclusive,omitempty" path:"exclusive,omitempty" url:"exclusive,omitempty"`
	Token                string `json:"token,omitempty" path:"token,omitempty" url:"token,omitempty"`
	Type                 string `json:"type,omitempty" path:"type,omitempty" url:"type,omitempty"`
	AllowAccessByAnyUser *bool  `json:"allow_access_by_any_user,omitempty" path:"allow_access_by_any_user,omitempty" url:"allow_access_by_any_user,omitempty"`
	UserId               int64  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username             string `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
}

func (l Lock) Identifier() interface{} {
	return l.Path
}

type LockCollection []Lock

type LockListForParams struct {
	Path            string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	IncludeChildren *bool  `url:"include_children,omitempty" json:"include_children,omitempty" path:"include_children"`
	ListParams
}

type LockCreateParams struct {
	Path                 string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	AllowAccessByAnyUser *bool  `url:"allow_access_by_any_user,omitempty" json:"allow_access_by_any_user,omitempty" path:"allow_access_by_any_user"`
	Exclusive            *bool  `url:"exclusive,omitempty" json:"exclusive,omitempty" path:"exclusive"`
	Recursive            *bool  `url:"recursive,omitempty" json:"recursive,omitempty" path:"recursive"`
	Timeout              int64  `url:"timeout,omitempty" json:"timeout,omitempty" path:"timeout"`
}

type LockDeleteParams struct {
	Path  string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	Token string `url:"token" json:"token" path:"token"`
}

func (l *Lock) UnmarshalJSON(data []byte) error {
	type lock Lock
	var v lock
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*l = Lock(v)
	return nil
}

func (l *LockCollection) UnmarshalJSON(data []byte) error {
	type locks LockCollection
	var v locks
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*l = LockCollection(v)
	return nil
}

func (l *LockCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*l))
	for i, v := range *l {
		ret[i] = v
	}

	return &ret
}
