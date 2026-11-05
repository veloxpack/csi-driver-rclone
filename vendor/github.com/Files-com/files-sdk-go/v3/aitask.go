package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type AiTask struct {
	Id                    int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId           int64      `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name                  string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Description           string     `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	Prompt                string     `json:"prompt,omitempty" path:"prompt,omitempty" url:"prompt,omitempty"`
	PermissionSet         string     `json:"permission_set,omitempty" path:"permission_set,omitempty" url:"permission_set,omitempty"`
	Path                  string     `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Source                string     `json:"source,omitempty" path:"source,omitempty" url:"source,omitempty"`
	Disabled              *bool      `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	Trigger               string     `json:"trigger,omitempty" path:"trigger,omitempty" url:"trigger,omitempty"`
	TriggerActions        []string   `json:"trigger_actions,omitempty" path:"trigger_actions,omitempty" url:"trigger_actions,omitempty"`
	Interval              string     `json:"interval,omitempty" path:"interval,omitempty" url:"interval,omitempty"`
	RecurringDay          int64      `json:"recurring_day,omitempty" path:"recurring_day,omitempty" url:"recurring_day,omitempty"`
	ScheduleDaysOfWeek    []int64    `json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week,omitempty" url:"schedule_days_of_week,omitempty"`
	ScheduleTimesOfDay    []string   `json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day,omitempty" url:"schedule_times_of_day,omitempty"`
	ScheduleTimeZone      string     `json:"schedule_time_zone,omitempty" path:"schedule_time_zone,omitempty" url:"schedule_time_zone,omitempty"`
	HolidayRegion         string     `json:"holiday_region,omitempty" path:"holiday_region,omitempty" url:"holiday_region,omitempty"`
	HumanReadableSchedule string     `json:"human_readable_schedule,omitempty" path:"human_readable_schedule,omitempty" url:"human_readable_schedule,omitempty"`
	LastRunAt             *time.Time `json:"last_run_at,omitempty" path:"last_run_at,omitempty" url:"last_run_at,omitempty"`
	MasterAdminUserId     int64      `json:"master_admin_user_id,omitempty" path:"master_admin_user_id,omitempty" url:"master_admin_user_id,omitempty"`
	CreatedAt             *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (a AiTask) Identifier() interface{} {
	return a.Id
}

type AiTaskCollection []AiTask

type AiTaskPermissionSetEnum string

func (u AiTaskPermissionSetEnum) String() string {
	return string(u)
}

func (u AiTaskPermissionSetEnum) Enum() map[string]AiTaskPermissionSetEnum {
	return map[string]AiTaskPermissionSetEnum{
		"full":       AiTaskPermissionSetEnum("full"),
		"files_only": AiTaskPermissionSetEnum("files_only"),
	}
}

type AiTaskTriggerEnum string

func (u AiTaskTriggerEnum) String() string {
	return string(u)
}

func (u AiTaskTriggerEnum) Enum() map[string]AiTaskTriggerEnum {
	return map[string]AiTaskTriggerEnum{
		"manual":          AiTaskTriggerEnum("manual"),
		"daily":           AiTaskTriggerEnum("daily"),
		"custom_schedule": AiTaskTriggerEnum("custom_schedule"),
		"action":          AiTaskTriggerEnum("action"),
	}
}

type AiTaskListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type AiTaskFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type AiTaskCreateParams struct {
	Description        string                  `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Disabled           *bool                   `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	HolidayRegion      string                  `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	Interval           string                  `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	Name               string                  `url:"name" json:"name" path:"name"`
	Path               string                  `url:"path,omitempty" json:"path,omitempty" path:"path"`
	PermissionSet      AiTaskPermissionSetEnum `url:"permission_set,omitempty" json:"permission_set,omitempty" path:"permission_set"`
	Prompt             string                  `url:"prompt" json:"prompt" path:"prompt"`
	RecurringDay       int64                   `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleDaysOfWeek []int64                 `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimeZone   string                  `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	ScheduleTimesOfDay []string                `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	Source             string                  `url:"source,omitempty" json:"source,omitempty" path:"source"`
	Trigger            AiTaskTriggerEnum       `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	TriggerActions     []string                `url:"trigger_actions,omitempty" json:"trigger_actions,omitempty" path:"trigger_actions"`
	WorkspaceId        int64                   `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

// Manually Run AI Task
type AiTaskManualRunParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type AiTaskUpdateParams struct {
	Id                 int64                   `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Description        string                  `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Disabled           *bool                   `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	HolidayRegion      string                  `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	Interval           string                  `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	Name               string                  `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Path               string                  `url:"path,omitempty" json:"path,omitempty" path:"path"`
	PermissionSet      AiTaskPermissionSetEnum `url:"permission_set,omitempty" json:"permission_set,omitempty" path:"permission_set"`
	Prompt             string                  `url:"prompt,omitempty" json:"prompt,omitempty" path:"prompt"`
	RecurringDay       int64                   `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleDaysOfWeek []int64                 `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimeZone   string                  `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	ScheduleTimesOfDay []string                `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	Source             string                  `url:"source,omitempty" json:"source,omitempty" path:"source"`
	Trigger            AiTaskTriggerEnum       `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	TriggerActions     []string                `url:"trigger_actions,omitempty" json:"trigger_actions,omitempty" path:"trigger_actions"`
	WorkspaceId        int64                   `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type AiTaskDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (a *AiTask) UnmarshalJSON(data []byte) error {
	type aiTask AiTask
	var v aiTask
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = AiTask(v)
	return nil
}

func (a *AiTaskCollection) UnmarshalJSON(data []byte) error {
	type aiTasks AiTaskCollection
	var v aiTasks
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AiTaskCollection(v)
	return nil
}

func (a *AiTaskCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
