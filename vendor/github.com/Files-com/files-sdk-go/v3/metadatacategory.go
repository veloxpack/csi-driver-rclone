package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type MetadataCategory struct {
	Id             int64               `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name           string              `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Definitions    map[string][]string `json:"definitions,omitempty" path:"definitions,omitempty" url:"definitions,omitempty"`
	DefaultColumns []string            `json:"default_columns,omitempty" path:"default_columns,omitempty" url:"default_columns,omitempty"`
}

func (m MetadataCategory) Identifier() interface{} {
	return m.Id
}

type MetadataCategoryCollection []MetadataCategory

type MetadataCategoryListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type MetadataCategoryFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type MetadataCategoryListForParams struct {
	Path string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	ListParams
}

type MetadataCategoryCreateParams struct {
	Name           string   `url:"name" json:"name" path:"name"`
	DefaultColumns []string `url:"default_columns,omitempty" json:"default_columns,omitempty" path:"default_columns"`
}

type MetadataCategoryUpdateParams struct {
	Id             int64    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name           string   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	DefaultColumns []string `url:"default_columns,omitempty" json:"default_columns,omitempty" path:"default_columns"`
}

type MetadataCategoryDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (m *MetadataCategory) UnmarshalJSON(data []byte) error {
	type metadataCategory MetadataCategory
	var v metadataCategory
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*m = MetadataCategory(v)
	return nil
}

func (m *MetadataCategoryCollection) UnmarshalJSON(data []byte) error {
	type metadataCategorys MetadataCategoryCollection
	var v metadataCategorys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*m = MetadataCategoryCollection(v)
	return nil
}

func (m *MetadataCategoryCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*m))
	for i, v := range *m {
		ret[i] = v
	}

	return &ret
}
