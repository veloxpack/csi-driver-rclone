package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Request struct {
	Id              int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Path            string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Source          string `json:"source,omitempty" path:"source,omitempty" url:"source,omitempty"`
	Destination     string `json:"destination,omitempty" path:"destination,omitempty" url:"destination,omitempty"`
	AutomationId    int64  `json:"automation_id,omitempty" path:"automation_id,omitempty" url:"automation_id,omitempty"`
	UserDisplayName string `json:"user_display_name,omitempty" path:"user_display_name,omitempty" url:"user_display_name,omitempty"`
	UserIds         string `json:"user_ids,omitempty" path:"user_ids,omitempty" url:"user_ids,omitempty"`
	GroupIds        string `json:"group_ids,omitempty" path:"group_ids,omitempty" url:"group_ids,omitempty"`
}

func (r Request) Identifier() interface{} {
	return r.Id
}

type RequestCollection []Request

type RequestListParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Mine   *bool                  `url:"mine,omitempty" json:"mine,omitempty" path:"mine"`
	Path   string                 `url:"path,omitempty" json:"path,omitempty" path:"path"`
	ListParams
}

type RequestGetFolderParams struct {
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Mine   *bool                  `url:"mine,omitempty" json:"mine,omitempty" path:"mine"`
	Path   string                 `url:"-,omitempty" json:"-,omitempty" path:"path"`
	ListParams
}

type RequestCreateParams struct {
	Path        string `url:"path" json:"path" path:"path"`
	Destination string `url:"destination" json:"destination" path:"destination"`
	UserIds     string `url:"user_ids,omitempty" json:"user_ids,omitempty" path:"user_ids"`
	GroupIds    string `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
}

type RequestDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (r *Request) UnmarshalJSON(data []byte) error {
	type request Request
	var v request
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = Request(v)
	return nil
}

func (r *RequestCollection) UnmarshalJSON(data []byte) error {
	type requests RequestCollection
	var v requests
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RequestCollection(v)
	return nil
}

func (r *RequestCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
