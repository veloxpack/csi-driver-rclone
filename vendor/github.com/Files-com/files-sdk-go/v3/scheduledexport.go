package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ScheduledExport struct {
	Id                    int64       `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name                  string      `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	ExportType            string      `json:"export_type,omitempty" path:"export_type,omitempty" url:"export_type,omitempty"`
	ReportName            string      `json:"report_name,omitempty" path:"report_name,omitempty" url:"report_name,omitempty"`
	ExportOptions         interface{} `json:"export_options,omitempty" path:"export_options,omitempty" url:"export_options,omitempty"`
	UserId                int64       `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Disabled              *bool       `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	Trigger               string      `json:"trigger,omitempty" path:"trigger,omitempty" url:"trigger,omitempty"`
	Interval              string      `json:"interval,omitempty" path:"interval,omitempty" url:"interval,omitempty"`
	RecurringDay          int64       `json:"recurring_day,omitempty" path:"recurring_day,omitempty" url:"recurring_day,omitempty"`
	ScheduleDaysOfWeek    []int64     `json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week,omitempty" url:"schedule_days_of_week,omitempty"`
	ScheduleTimesOfDay    []string    `json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day,omitempty" url:"schedule_times_of_day,omitempty"`
	ScheduleTimeZone      string      `json:"schedule_time_zone,omitempty" path:"schedule_time_zone,omitempty" url:"schedule_time_zone,omitempty"`
	HolidayRegion         string      `json:"holiday_region,omitempty" path:"holiday_region,omitempty" url:"holiday_region,omitempty"`
	HumanReadableSchedule string      `json:"human_readable_schedule,omitempty" path:"human_readable_schedule,omitempty" url:"human_readable_schedule,omitempty"`
	LastRunAt             *time.Time  `json:"last_run_at,omitempty" path:"last_run_at,omitempty" url:"last_run_at,omitempty"`
	LastExportId          int64       `json:"last_export_id,omitempty" path:"last_export_id,omitempty" url:"last_export_id,omitempty"`
	CreatedAt             *time.Time  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt             *time.Time  `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (s ScheduledExport) Identifier() interface{} {
	return s.Id
}

type ScheduledExportCollection []ScheduledExport

type ScheduledExportTriggerEnum string

func (u ScheduledExportTriggerEnum) String() string {
	return string(u)
}

func (u ScheduledExportTriggerEnum) Enum() map[string]ScheduledExportTriggerEnum {
	return map[string]ScheduledExportTriggerEnum{
		"daily":           ScheduledExportTriggerEnum("daily"),
		"custom_schedule": ScheduledExportTriggerEnum("custom_schedule"),
	}
}

type ScheduledExportListParams struct {
	SortBy       interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	ListParams
}

type ScheduledExportFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ScheduledExportCreateParams struct {
	Name               string                     `url:"name" json:"name" path:"name"`
	ExportType         string                     `url:"export_type" json:"export_type" path:"export_type"`
	ExportOptions      interface{}                `url:"export_options,omitempty" json:"export_options,omitempty" path:"export_options"`
	UserId             int64                      `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Disabled           *bool                      `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	Trigger            ScheduledExportTriggerEnum `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	Interval           string                     `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	RecurringDay       int64                      `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleDaysOfWeek []int64                    `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay []string                   `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	ScheduleTimeZone   string                     `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	HolidayRegion      string                     `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
}

type ScheduledExportUpdateParams struct {
	Id                 int64                      `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name               string                     `url:"name,omitempty" json:"name,omitempty" path:"name"`
	ExportType         string                     `url:"export_type,omitempty" json:"export_type,omitempty" path:"export_type"`
	ExportOptions      interface{}                `url:"export_options,omitempty" json:"export_options,omitempty" path:"export_options"`
	UserId             int64                      `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Disabled           *bool                      `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	Trigger            ScheduledExportTriggerEnum `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	Interval           string                     `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	RecurringDay       int64                      `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleDaysOfWeek []int64                    `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay []string                   `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	ScheduleTimeZone   string                     `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	HolidayRegion      string                     `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
}

type ScheduledExportDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *ScheduledExport) UnmarshalJSON(data []byte) error {
	type scheduledExport ScheduledExport
	var v scheduledExport
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = ScheduledExport(v)
	return nil
}

func (s *ScheduledExportCollection) UnmarshalJSON(data []byte) error {
	type scheduledExports ScheduledExportCollection
	var v scheduledExports
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = ScheduledExportCollection(v)
	return nil
}

func (s *ScheduledExportCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
