package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
	"github.com/lpar/date"
)

type UsageDailySnapshot struct {
	Id                           int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Date                         *date.Date               `json:"date,omitempty" path:"date,omitempty" url:"date,omitempty"`
	ApiUsageAvailable            *bool                    `json:"api_usage_available,omitempty" path:"api_usage_available,omitempty" url:"api_usage_available,omitempty"`
	ReadApiUsage                 int64                    `json:"read_api_usage,omitempty" path:"read_api_usage,omitempty" url:"read_api_usage,omitempty"`
	WriteApiUsage                int64                    `json:"write_api_usage,omitempty" path:"write_api_usage,omitempty" url:"write_api_usage,omitempty"`
	UserCount                    int64                    `json:"user_count,omitempty" path:"user_count,omitempty" url:"user_count,omitempty"`
	CurrentStorage               string                   `json:"current_storage,omitempty" path:"current_storage,omitempty" url:"current_storage,omitempty"`
	DeletedFilesStorage          string                   `json:"deleted_files_storage,omitempty" path:"deleted_files_storage,omitempty" url:"deleted_files_storage,omitempty"`
	DeletedFilesCountedInMinimum string                   `json:"deleted_files_counted_in_minimum,omitempty" path:"deleted_files_counted_in_minimum,omitempty" url:"deleted_files_counted_in_minimum,omitempty"`
	RootStorage                  string                   `json:"root_storage,omitempty" path:"root_storage,omitempty" url:"root_storage,omitempty"`
	UsageByTopLevelDir           []map[string]interface{} `json:"usage_by_top_level_dir,omitempty" path:"usage_by_top_level_dir,omitempty" url:"usage_by_top_level_dir,omitempty"`
}

func (u UsageDailySnapshot) Identifier() interface{} {
	return u.Id
}

type UsageDailySnapshotCollection []UsageDailySnapshot

type UsageDailySnapshotListParams struct {
	SortBy     map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     UsageDailySnapshot     `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (u *UsageDailySnapshot) UnmarshalJSON(data []byte) error {
	type usageDailySnapshot UsageDailySnapshot
	var v usageDailySnapshot
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*u = UsageDailySnapshot(v)
	return nil
}

func (u *UsageDailySnapshotCollection) UnmarshalJSON(data []byte) error {
	type usageDailySnapshots UsageDailySnapshotCollection
	var v usageDailySnapshots
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*u = UsageDailySnapshotCollection(v)
	return nil
}

func (u *UsageDailySnapshotCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*u))
	for i, v := range *u {
		ret[i] = v
	}

	return &ret
}
