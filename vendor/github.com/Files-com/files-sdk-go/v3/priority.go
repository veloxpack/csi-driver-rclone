package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Priority struct {
	Path  string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Color string `json:"color,omitempty" path:"color,omitempty" url:"color,omitempty"`
}

func (p Priority) Identifier() interface{} {
	return p.Path
}

type PriorityCollection []Priority

type PriorityListParams struct {
	Path string `url:"path" json:"path" path:"path"`
	ListParams
}

func (p *Priority) UnmarshalJSON(data []byte) error {
	type priority Priority
	var v priority
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = Priority(v)
	return nil
}

func (p *PriorityCollection) UnmarshalJSON(data []byte) error {
	type prioritys PriorityCollection
	var v prioritys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PriorityCollection(v)
	return nil
}

func (p *PriorityCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
