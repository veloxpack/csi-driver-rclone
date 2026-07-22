package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Workspace struct {
	Id   int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name string `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
}

func (w Workspace) Identifier() interface{} {
	return w.Id
}

type WorkspaceCollection []Workspace

type WorkspaceListParams struct {
	SortBy       interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	ListParams
}

type WorkspaceFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type WorkspaceCreateParams struct {
	Name string `url:"name,omitempty" json:"name,omitempty" path:"name"`
}

type WorkspaceUpdateParams struct {
	Id   int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name string `url:"name,omitempty" json:"name,omitempty" path:"name"`
}

type WorkspaceDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (w *Workspace) UnmarshalJSON(data []byte) error {
	type workspace Workspace
	var v workspace
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*w = Workspace(v)
	return nil
}

func (w *WorkspaceCollection) UnmarshalJSON(data []byte) error {
	type workspaces WorkspaceCollection
	var v workspaces
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*w = WorkspaceCollection(v)
	return nil
}

func (w *WorkspaceCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*w))
	for i, v := range *w {
		ret[i] = v
	}

	return &ret
}
