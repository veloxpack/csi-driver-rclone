package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Sync struct {
	Id                  int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name                string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Description         string     `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	SiteId              int64      `json:"site_id,omitempty" path:"site_id,omitempty" url:"site_id,omitempty"`
	UserId              int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	SrcPath             string     `json:"src_path,omitempty" path:"src_path,omitempty" url:"src_path,omitempty"`
	DestPath            string     `json:"dest_path,omitempty" path:"dest_path,omitempty" url:"dest_path,omitempty"`
	SrcRemoteServerId   int64      `json:"src_remote_server_id,omitempty" path:"src_remote_server_id,omitempty" url:"src_remote_server_id,omitempty"`
	DestRemoteServerId  int64      `json:"dest_remote_server_id,omitempty" path:"dest_remote_server_id,omitempty" url:"dest_remote_server_id,omitempty"`
	TwoWay              *bool      `json:"two_way,omitempty" path:"two_way,omitempty" url:"two_way,omitempty"`
	KeepAfterCopy       *bool      `json:"keep_after_copy,omitempty" path:"keep_after_copy,omitempty" url:"keep_after_copy,omitempty"`
	DeleteEmptyFolders  *bool      `json:"delete_empty_folders,omitempty" path:"delete_empty_folders,omitempty" url:"delete_empty_folders,omitempty"`
	Disabled            *bool      `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	Trigger             string     `json:"trigger,omitempty" path:"trigger,omitempty" url:"trigger,omitempty"`
	TriggerFile         string     `json:"trigger_file,omitempty" path:"trigger_file,omitempty" url:"trigger_file,omitempty"`
	IncludePatterns     []string   `json:"include_patterns,omitempty" path:"include_patterns,omitempty" url:"include_patterns,omitempty"`
	ExcludePatterns     []string   `json:"exclude_patterns,omitempty" path:"exclude_patterns,omitempty" url:"exclude_patterns,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
	SyncIntervalMinutes int64      `json:"sync_interval_minutes,omitempty" path:"sync_interval_minutes,omitempty" url:"sync_interval_minutes,omitempty"`
	Interval            string     `json:"interval,omitempty" path:"interval,omitempty" url:"interval,omitempty"`
	RecurringDay        int64      `json:"recurring_day,omitempty" path:"recurring_day,omitempty" url:"recurring_day,omitempty"`
	ScheduleDaysOfWeek  []int64    `json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week,omitempty" url:"schedule_days_of_week,omitempty"`
	ScheduleTimesOfDay  []string   `json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day,omitempty" url:"schedule_times_of_day,omitempty"`
	ScheduleTimeZone    string     `json:"schedule_time_zone,omitempty" path:"schedule_time_zone,omitempty" url:"schedule_time_zone,omitempty"`
	HolidayRegion       string     `json:"holiday_region,omitempty" path:"holiday_region,omitempty" url:"holiday_region,omitempty"`
	LatestSyncRun       SyncRun    `json:"latest_sync_run,omitempty" path:"latest_sync_run,omitempty" url:"latest_sync_run,omitempty"`
}

func (s Sync) Identifier() interface{} {
	return s.Id
}

type SyncCollection []Sync

type SyncListParams struct {
	ListParams
}

type SyncFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SyncCreateParams struct {
	Name                string   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description         string   `url:"description,omitempty" json:"description,omitempty" path:"description"`
	SrcPath             string   `url:"src_path,omitempty" json:"src_path,omitempty" path:"src_path"`
	DestPath            string   `url:"dest_path,omitempty" json:"dest_path,omitempty" path:"dest_path"`
	SrcRemoteServerId   int64    `url:"src_remote_server_id,omitempty" json:"src_remote_server_id,omitempty" path:"src_remote_server_id"`
	DestRemoteServerId  int64    `url:"dest_remote_server_id,omitempty" json:"dest_remote_server_id,omitempty" path:"dest_remote_server_id"`
	KeepAfterCopy       *bool    `url:"keep_after_copy,omitempty" json:"keep_after_copy,omitempty" path:"keep_after_copy"`
	DeleteEmptyFolders  *bool    `url:"delete_empty_folders,omitempty" json:"delete_empty_folders,omitempty" path:"delete_empty_folders"`
	Disabled            *bool    `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	Interval            string   `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	Trigger             string   `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	TriggerFile         string   `url:"trigger_file,omitempty" json:"trigger_file,omitempty" path:"trigger_file"`
	HolidayRegion       string   `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	SyncIntervalMinutes int64    `url:"sync_interval_minutes,omitempty" json:"sync_interval_minutes,omitempty" path:"sync_interval_minutes"`
	RecurringDay        int64    `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleTimeZone    string   `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	ScheduleDaysOfWeek  []int64  `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay  []string `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
}

// Dry Run Sync
type SyncDryRunParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

// Manually Run Sync
type SyncManualRunParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SyncUpdateParams struct {
	Id                  int64    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name                string   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description         string   `url:"description,omitempty" json:"description,omitempty" path:"description"`
	SrcPath             string   `url:"src_path,omitempty" json:"src_path,omitempty" path:"src_path"`
	DestPath            string   `url:"dest_path,omitempty" json:"dest_path,omitempty" path:"dest_path"`
	SrcRemoteServerId   int64    `url:"src_remote_server_id,omitempty" json:"src_remote_server_id,omitempty" path:"src_remote_server_id"`
	DestRemoteServerId  int64    `url:"dest_remote_server_id,omitempty" json:"dest_remote_server_id,omitempty" path:"dest_remote_server_id"`
	KeepAfterCopy       *bool    `url:"keep_after_copy,omitempty" json:"keep_after_copy,omitempty" path:"keep_after_copy"`
	DeleteEmptyFolders  *bool    `url:"delete_empty_folders,omitempty" json:"delete_empty_folders,omitempty" path:"delete_empty_folders"`
	Disabled            *bool    `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	Interval            string   `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	Trigger             string   `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	TriggerFile         string   `url:"trigger_file,omitempty" json:"trigger_file,omitempty" path:"trigger_file"`
	HolidayRegion       string   `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	SyncIntervalMinutes int64    `url:"sync_interval_minutes,omitempty" json:"sync_interval_minutes,omitempty" path:"sync_interval_minutes"`
	RecurringDay        int64    `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleTimeZone    string   `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	ScheduleDaysOfWeek  []int64  `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay  []string `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
}

type SyncDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *Sync) UnmarshalJSON(data []byte) error {
	type sync Sync
	var v sync
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Sync(v)
	return nil
}

func (s *SyncCollection) UnmarshalJSON(data []byte) error {
	type syncs SyncCollection
	var v syncs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SyncCollection(v)
	return nil
}

func (s *SyncCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
