package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type UsageSnapshot struct {
	Id                           int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	StartAt                      *time.Time               `json:"start_at,omitempty" path:"start_at,omitempty" url:"start_at,omitempty"`
	EndAt                        *time.Time               `json:"end_at,omitempty" path:"end_at,omitempty" url:"end_at,omitempty"`
	HighWaterUserCount           int64                    `json:"high_water_user_count,omitempty" path:"high_water_user_count,omitempty" url:"high_water_user_count,omitempty"`
	CurrentStorage               string                   `json:"current_storage,omitempty" path:"current_storage,omitempty" url:"current_storage,omitempty"`
	HighWaterStorage             string                   `json:"high_water_storage,omitempty" path:"high_water_storage,omitempty" url:"high_water_storage,omitempty"`
	RootStorage                  string                   `json:"root_storage,omitempty" path:"root_storage,omitempty" url:"root_storage,omitempty"`
	DeletedFilesCountedInMinimum string                   `json:"deleted_files_counted_in_minimum,omitempty" path:"deleted_files_counted_in_minimum,omitempty" url:"deleted_files_counted_in_minimum,omitempty"`
	DeletedFilesStorage          string                   `json:"deleted_files_storage,omitempty" path:"deleted_files_storage,omitempty" url:"deleted_files_storage,omitempty"`
	TotalBillableUsage           string                   `json:"total_billable_usage,omitempty" path:"total_billable_usage,omitempty" url:"total_billable_usage,omitempty"`
	TotalBillableTransferUsage   string                   `json:"total_billable_transfer_usage,omitempty" path:"total_billable_transfer_usage,omitempty" url:"total_billable_transfer_usage,omitempty"`
	BytesSent                    string                   `json:"bytes_sent,omitempty" path:"bytes_sent,omitempty" url:"bytes_sent,omitempty"`
	SyncBytesReceived            string                   `json:"sync_bytes_received,omitempty" path:"sync_bytes_received,omitempty" url:"sync_bytes_received,omitempty"`
	SyncBytesSent                string                   `json:"sync_bytes_sent,omitempty" path:"sync_bytes_sent,omitempty" url:"sync_bytes_sent,omitempty"`
	UsageByTopLevelDir           []map[string]interface{} `json:"usage_by_top_level_dir,omitempty" path:"usage_by_top_level_dir,omitempty" url:"usage_by_top_level_dir,omitempty"`
}

func (u UsageSnapshot) Identifier() interface{} {
	return u.Id
}

type UsageSnapshotCollection []UsageSnapshot

type UsageSnapshotListParams struct {
	ListParams
}

func (u *UsageSnapshot) UnmarshalJSON(data []byte) error {
	type usageSnapshot UsageSnapshot
	var v usageSnapshot
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*u = UsageSnapshot(v)
	return nil
}

func (u *UsageSnapshotCollection) UnmarshalJSON(data []byte) error {
	type usageSnapshots UsageSnapshotCollection
	var v usageSnapshots
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*u = UsageSnapshotCollection(v)
	return nil
}

func (u *UsageSnapshotCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*u))
	for i, v := range *u {
		ret[i] = v
	}

	return &ret
}
