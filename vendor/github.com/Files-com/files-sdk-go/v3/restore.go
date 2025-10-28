package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Restore struct {
	EarliestDate              *time.Time `json:"earliest_date,omitempty" path:"earliest_date,omitempty" url:"earliest_date,omitempty"`
	Id                        int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	DirsRestored              int64      `json:"dirs_restored,omitempty" path:"dirs_restored,omitempty" url:"dirs_restored,omitempty"`
	DirsErrored               int64      `json:"dirs_errored,omitempty" path:"dirs_errored,omitempty" url:"dirs_errored,omitempty"`
	DirsTotal                 int64      `json:"dirs_total,omitempty" path:"dirs_total,omitempty" url:"dirs_total,omitempty"`
	FilesRestored             int64      `json:"files_restored,omitempty" path:"files_restored,omitempty" url:"files_restored,omitempty"`
	FilesErrored              int64      `json:"files_errored,omitempty" path:"files_errored,omitempty" url:"files_errored,omitempty"`
	FilesTotal                int64      `json:"files_total,omitempty" path:"files_total,omitempty" url:"files_total,omitempty"`
	Prefix                    string     `json:"prefix,omitempty" path:"prefix,omitempty" url:"prefix,omitempty"`
	RestoreInPlace            *bool      `json:"restore_in_place,omitempty" path:"restore_in_place,omitempty" url:"restore_in_place,omitempty"`
	RestoreDeletedPermissions *bool      `json:"restore_deleted_permissions,omitempty" path:"restore_deleted_permissions,omitempty" url:"restore_deleted_permissions,omitempty"`
	Status                    string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	UpdateTimestamps          *bool      `json:"update_timestamps,omitempty" path:"update_timestamps,omitempty" url:"update_timestamps,omitempty"`
	ErrorMessages             []string   `json:"error_messages,omitempty" path:"error_messages,omitempty" url:"error_messages,omitempty"`
}

func (r Restore) Identifier() interface{} {
	return r.Id
}

type RestoreCollection []Restore

type RestoreListParams struct {
	ListParams
}

type RestoreCreateParams struct {
	EarliestDate              *time.Time `url:"earliest_date" json:"earliest_date" path:"earliest_date"`
	Prefix                    string     `url:"prefix,omitempty" json:"prefix,omitempty" path:"prefix"`
	RestoreDeletedPermissions *bool      `url:"restore_deleted_permissions,omitempty" json:"restore_deleted_permissions,omitempty" path:"restore_deleted_permissions"`
	RestoreInPlace            *bool      `url:"restore_in_place,omitempty" json:"restore_in_place,omitempty" path:"restore_in_place"`
	UpdateTimestamps          *bool      `url:"update_timestamps,omitempty" json:"update_timestamps,omitempty" path:"update_timestamps"`
}

func (r *Restore) UnmarshalJSON(data []byte) error {
	type restore Restore
	var v restore
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = Restore(v)
	return nil
}

func (r *RestoreCollection) UnmarshalJSON(data []byte) error {
	type restores RestoreCollection
	var v restores
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RestoreCollection(v)
	return nil
}

func (r *RestoreCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
