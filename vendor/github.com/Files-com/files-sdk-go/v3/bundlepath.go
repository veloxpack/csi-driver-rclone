package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type BundlePath struct {
	Recursive *bool  `json:"recursive,omitempty" path:"recursive,omitempty" url:"recursive,omitempty"`
	Path      string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
}

func (b BundlePath) Identifier() interface{} {
	return b.Path
}

type BundlePathCollection []BundlePath

func (b *BundlePath) UnmarshalJSON(data []byte) error {
	type bundlePath BundlePath
	var v bundlePath
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*b = BundlePath(v)
	return nil
}

func (b *BundlePathCollection) UnmarshalJSON(data []byte) error {
	type bundlePaths BundlePathCollection
	var v bundlePaths
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*b = BundlePathCollection(v)
	return nil
}

func (b *BundlePathCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*b))
	for i, v := range *b {
		ret[i] = v
	}

	return &ret
}
