package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Project struct {
	Id           int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	GlobalAccess string `json:"global_access,omitempty" path:"global_access,omitempty" url:"global_access,omitempty"`
}

func (p Project) Identifier() interface{} {
	return p.Id
}

type ProjectCollection []Project

type ProjectListParams struct {
	ListParams
}

type ProjectFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ProjectCreateParams struct {
	GlobalAccess string `url:"global_access" json:"global_access" path:"global_access"`
}

type ProjectUpdateParams struct {
	Id           int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	GlobalAccess string `url:"global_access" json:"global_access" path:"global_access"`
}

type ProjectDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *Project) UnmarshalJSON(data []byte) error {
	type project Project
	var v project
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = Project(v)
	return nil
}

func (p *ProjectCollection) UnmarshalJSON(data []byte) error {
	type projects ProjectCollection
	var v projects
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = ProjectCollection(v)
	return nil
}

func (p *ProjectCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
