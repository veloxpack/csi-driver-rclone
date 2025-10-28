package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FileAction struct {
	Status          string `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	FileMigrationId int64  `json:"file_migration_id,omitempty" path:"file_migration_id,omitempty" url:"file_migration_id,omitempty"`
}

// Identifier no path or id

type FileActionCollection []FileAction

func (f *FileAction) UnmarshalJSON(data []byte) error {
	type fileAction FileAction
	var v fileAction
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FileAction(v)
	return nil
}

func (f *FileActionCollection) UnmarshalJSON(data []byte) error {
	type fileActions FileActionCollection
	var v fileActions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FileActionCollection(v)
	return nil
}

func (f *FileActionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
