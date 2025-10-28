package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type UsageByTopLevelDir struct {
	Dir   string `json:"dir,omitempty" path:"dir,omitempty" url:"dir,omitempty"`
	Size  int64  `json:"size,omitempty" path:"size,omitempty" url:"size,omitempty"`
	Count int64  `json:"count,omitempty" path:"count,omitempty" url:"count,omitempty"`
}

// Identifier no path or id

type UsageByTopLevelDirCollection []UsageByTopLevelDir

func (u *UsageByTopLevelDir) UnmarshalJSON(data []byte) error {
	type usageByTopLevelDir UsageByTopLevelDir
	var v usageByTopLevelDir
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*u = UsageByTopLevelDir(v)
	return nil
}

func (u *UsageByTopLevelDirCollection) UnmarshalJSON(data []byte) error {
	type usageByTopLevelDirs UsageByTopLevelDirCollection
	var v usageByTopLevelDirs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*u = UsageByTopLevelDirCollection(v)
	return nil
}

func (u *UsageByTopLevelDirCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*u))
	for i, v := range *u {
		ret[i] = v
	}

	return &ret
}
