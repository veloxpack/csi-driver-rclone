package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Automation struct {
	Id                               int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	AlwaysSerializeJobs              *bool                    `json:"always_serialize_jobs,omitempty" path:"always_serialize_jobs,omitempty" url:"always_serialize_jobs,omitempty"`
	AlwaysOverwriteSizeMatchingFiles *bool                    `json:"always_overwrite_size_matching_files,omitempty" path:"always_overwrite_size_matching_files,omitempty" url:"always_overwrite_size_matching_files,omitempty"`
	Automation                       string                   `json:"automation,omitempty" path:"automation,omitempty" url:"automation,omitempty"`
	Deleted                          *bool                    `json:"deleted,omitempty" path:"deleted,omitempty" url:"deleted,omitempty"`
	Description                      string                   `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	DestinationReplaceFrom           string                   `json:"destination_replace_from,omitempty" path:"destination_replace_from,omitempty" url:"destination_replace_from,omitempty"`
	DestinationReplaceTo             string                   `json:"destination_replace_to,omitempty" path:"destination_replace_to,omitempty" url:"destination_replace_to,omitempty"`
	Destinations                     []string                 `json:"destinations,omitempty" path:"destinations,omitempty" url:"destinations,omitempty"`
	Disabled                         *bool                    `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	ExcludePattern                   string                   `json:"exclude_pattern,omitempty" path:"exclude_pattern,omitempty" url:"exclude_pattern,omitempty"`
	ImportUrls                       []map[string]interface{} `json:"import_urls,omitempty" path:"import_urls,omitempty" url:"import_urls,omitempty"`
	FlattenDestinationStructure      *bool                    `json:"flatten_destination_structure,omitempty" path:"flatten_destination_structure,omitempty" url:"flatten_destination_structure,omitempty"`
	GroupIds                         []int64                  `json:"group_ids,omitempty" path:"group_ids,omitempty" url:"group_ids,omitempty"`
	IgnoreLockedFolders              *bool                    `json:"ignore_locked_folders,omitempty" path:"ignore_locked_folders,omitempty" url:"ignore_locked_folders,omitempty"`
	Interval                         string                   `json:"interval,omitempty" path:"interval,omitempty" url:"interval,omitempty"`
	LastModifiedAt                   *time.Time               `json:"last_modified_at,omitempty" path:"last_modified_at,omitempty" url:"last_modified_at,omitempty"`
	LegacyFolderMatching             *bool                    `json:"legacy_folder_matching,omitempty" path:"legacy_folder_matching,omitempty" url:"legacy_folder_matching,omitempty"`
	Name                             string                   `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	OverwriteFiles                   *bool                    `json:"overwrite_files,omitempty" path:"overwrite_files,omitempty" url:"overwrite_files,omitempty"`
	Path                             string                   `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	PathTimeZone                     string                   `json:"path_time_zone,omitempty" path:"path_time_zone,omitempty" url:"path_time_zone,omitempty"`
	RecurringDay                     int64                    `json:"recurring_day,omitempty" path:"recurring_day,omitempty" url:"recurring_day,omitempty"`
	RetryOnFailureIntervalInMinutes  int64                    `json:"retry_on_failure_interval_in_minutes,omitempty" path:"retry_on_failure_interval_in_minutes,omitempty" url:"retry_on_failure_interval_in_minutes,omitempty"`
	RetryOnFailureNumberOfAttempts   int64                    `json:"retry_on_failure_number_of_attempts,omitempty" path:"retry_on_failure_number_of_attempts,omitempty" url:"retry_on_failure_number_of_attempts,omitempty"`
	Schedule                         map[string]interface{}   `json:"schedule,omitempty" path:"schedule,omitempty" url:"schedule,omitempty"`
	HumanReadableSchedule            string                   `json:"human_readable_schedule,omitempty" path:"human_readable_schedule,omitempty" url:"human_readable_schedule,omitempty"`
	ScheduleDaysOfWeek               []int64                  `json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week,omitempty" url:"schedule_days_of_week,omitempty"`
	ScheduleTimesOfDay               []string                 `json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day,omitempty" url:"schedule_times_of_day,omitempty"`
	ScheduleTimeZone                 string                   `json:"schedule_time_zone,omitempty" path:"schedule_time_zone,omitempty" url:"schedule_time_zone,omitempty"`
	Source                           string                   `json:"source,omitempty" path:"source,omitempty" url:"source,omitempty"`
	LegacySyncIds                    []int64                  `json:"legacy_sync_ids,omitempty" path:"legacy_sync_ids,omitempty" url:"legacy_sync_ids,omitempty"`
	SyncIds                          []int64                  `json:"sync_ids,omitempty" path:"sync_ids,omitempty" url:"sync_ids,omitempty"`
	TriggerActions                   []string                 `json:"trigger_actions,omitempty" path:"trigger_actions,omitempty" url:"trigger_actions,omitempty"`
	Trigger                          string                   `json:"trigger,omitempty" path:"trigger,omitempty" url:"trigger,omitempty"`
	UserId                           int64                    `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	UserIds                          []int64                  `json:"user_ids,omitempty" path:"user_ids,omitempty" url:"user_ids,omitempty"`
	Value                            map[string]interface{}   `json:"value,omitempty" path:"value,omitempty" url:"value,omitempty"`
	WebhookUrl                       string                   `json:"webhook_url,omitempty" path:"webhook_url,omitempty" url:"webhook_url,omitempty"`
	HolidayRegion                    string                   `json:"holiday_region,omitempty" path:"holiday_region,omitempty" url:"holiday_region,omitempty"`
}

func (a Automation) Identifier() interface{} {
	return a.Id
}

type AutomationCollection []Automation

type AutomationTriggerEnum string

func (u AutomationTriggerEnum) String() string {
	return string(u)
}

func (u AutomationTriggerEnum) Enum() map[string]AutomationTriggerEnum {
	return map[string]AutomationTriggerEnum{
		"daily":           AutomationTriggerEnum("daily"),
		"custom_schedule": AutomationTriggerEnum("custom_schedule"),
		"webhook":         AutomationTriggerEnum("webhook"),
		"email":           AutomationTriggerEnum("email"),
		"action":          AutomationTriggerEnum("action"),
	}
}

type AutomationEnum string

func (u AutomationEnum) String() string {
	return string(u)
}

func (u AutomationEnum) Enum() map[string]AutomationEnum {
	return map[string]AutomationEnum{
		"create_folder": AutomationEnum("create_folder"),
		"delete_file":   AutomationEnum("delete_file"),
		"copy_file":     AutomationEnum("copy_file"),
		"move_file":     AutomationEnum("move_file"),
		"as2_send":      AutomationEnum("as2_send"),
		"run_sync":      AutomationEnum("run_sync"),
		"import_file":   AutomationEnum("import_file"),
	}
}

type AutomationListParams struct {
	SortBy     map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     Automation             `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type AutomationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type AutomationCreateParams struct {
	Source                           string                   `url:"source,omitempty" json:"source,omitempty" path:"source"`
	Destinations                     []string                 `url:"destinations,omitempty" json:"destinations,omitempty" path:"destinations"`
	DestinationReplaceFrom           string                   `url:"destination_replace_from,omitempty" json:"destination_replace_from,omitempty" path:"destination_replace_from"`
	DestinationReplaceTo             string                   `url:"destination_replace_to,omitempty" json:"destination_replace_to,omitempty" path:"destination_replace_to"`
	Interval                         string                   `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	Path                             string                   `url:"path,omitempty" json:"path,omitempty" path:"path"`
	LegacySyncIds                    string                   `url:"legacy_sync_ids,omitempty" json:"legacy_sync_ids,omitempty" path:"legacy_sync_ids"`
	SyncIds                          string                   `url:"sync_ids,omitempty" json:"sync_ids,omitempty" path:"sync_ids"`
	UserIds                          string                   `url:"user_ids,omitempty" json:"user_ids,omitempty" path:"user_ids"`
	GroupIds                         string                   `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
	ScheduleDaysOfWeek               []int64                  `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay               []string                 `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	ScheduleTimeZone                 string                   `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	HolidayRegion                    string                   `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	AlwaysOverwriteSizeMatchingFiles *bool                    `url:"always_overwrite_size_matching_files,omitempty" json:"always_overwrite_size_matching_files,omitempty" path:"always_overwrite_size_matching_files"`
	AlwaysSerializeJobs              *bool                    `url:"always_serialize_jobs,omitempty" json:"always_serialize_jobs,omitempty" path:"always_serialize_jobs"`
	Description                      string                   `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Disabled                         *bool                    `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	ExcludePattern                   string                   `url:"exclude_pattern,omitempty" json:"exclude_pattern,omitempty" path:"exclude_pattern"`
	ImportUrls                       []map[string]interface{} `url:"import_urls,omitempty" json:"import_urls,omitempty" path:"import_urls"`
	FlattenDestinationStructure      *bool                    `url:"flatten_destination_structure,omitempty" json:"flatten_destination_structure,omitempty" path:"flatten_destination_structure"`
	IgnoreLockedFolders              *bool                    `url:"ignore_locked_folders,omitempty" json:"ignore_locked_folders,omitempty" path:"ignore_locked_folders"`
	LegacyFolderMatching             *bool                    `url:"legacy_folder_matching,omitempty" json:"legacy_folder_matching,omitempty" path:"legacy_folder_matching"`
	Name                             string                   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	OverwriteFiles                   *bool                    `url:"overwrite_files,omitempty" json:"overwrite_files,omitempty" path:"overwrite_files"`
	PathTimeZone                     string                   `url:"path_time_zone,omitempty" json:"path_time_zone,omitempty" path:"path_time_zone"`
	RetryOnFailureIntervalInMinutes  int64                    `url:"retry_on_failure_interval_in_minutes,omitempty" json:"retry_on_failure_interval_in_minutes,omitempty" path:"retry_on_failure_interval_in_minutes"`
	RetryOnFailureNumberOfAttempts   int64                    `url:"retry_on_failure_number_of_attempts,omitempty" json:"retry_on_failure_number_of_attempts,omitempty" path:"retry_on_failure_number_of_attempts"`
	Trigger                          AutomationTriggerEnum    `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	TriggerActions                   []string                 `url:"trigger_actions,omitempty" json:"trigger_actions,omitempty" path:"trigger_actions"`
	Value                            map[string]interface{}   `url:"value,omitempty" json:"value,omitempty" path:"value"`
	RecurringDay                     int64                    `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	Automation                       AutomationEnum           `url:"automation" json:"automation" path:"automation"`
}

// Manually Run Automation
type AutomationManualRunParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type AutomationUpdateParams struct {
	Id                               int64                    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Source                           string                   `url:"source,omitempty" json:"source,omitempty" path:"source"`
	Destinations                     []string                 `url:"destinations,omitempty" json:"destinations,omitempty" path:"destinations"`
	DestinationReplaceFrom           string                   `url:"destination_replace_from,omitempty" json:"destination_replace_from,omitempty" path:"destination_replace_from"`
	DestinationReplaceTo             string                   `url:"destination_replace_to,omitempty" json:"destination_replace_to,omitempty" path:"destination_replace_to"`
	Interval                         string                   `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	Path                             string                   `url:"path,omitempty" json:"path,omitempty" path:"path"`
	LegacySyncIds                    string                   `url:"legacy_sync_ids,omitempty" json:"legacy_sync_ids,omitempty" path:"legacy_sync_ids"`
	SyncIds                          string                   `url:"sync_ids,omitempty" json:"sync_ids,omitempty" path:"sync_ids"`
	UserIds                          string                   `url:"user_ids,omitempty" json:"user_ids,omitempty" path:"user_ids"`
	GroupIds                         string                   `url:"group_ids,omitempty" json:"group_ids,omitempty" path:"group_ids"`
	ScheduleDaysOfWeek               []int64                  `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay               []string                 `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	ScheduleTimeZone                 string                   `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	HolidayRegion                    string                   `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	AlwaysOverwriteSizeMatchingFiles *bool                    `url:"always_overwrite_size_matching_files,omitempty" json:"always_overwrite_size_matching_files,omitempty" path:"always_overwrite_size_matching_files"`
	AlwaysSerializeJobs              *bool                    `url:"always_serialize_jobs,omitempty" json:"always_serialize_jobs,omitempty" path:"always_serialize_jobs"`
	Description                      string                   `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Disabled                         *bool                    `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	ExcludePattern                   string                   `url:"exclude_pattern,omitempty" json:"exclude_pattern,omitempty" path:"exclude_pattern"`
	ImportUrls                       []map[string]interface{} `url:"import_urls,omitempty" json:"import_urls,omitempty" path:"import_urls"`
	FlattenDestinationStructure      *bool                    `url:"flatten_destination_structure,omitempty" json:"flatten_destination_structure,omitempty" path:"flatten_destination_structure"`
	IgnoreLockedFolders              *bool                    `url:"ignore_locked_folders,omitempty" json:"ignore_locked_folders,omitempty" path:"ignore_locked_folders"`
	LegacyFolderMatching             *bool                    `url:"legacy_folder_matching,omitempty" json:"legacy_folder_matching,omitempty" path:"legacy_folder_matching"`
	Name                             string                   `url:"name,omitempty" json:"name,omitempty" path:"name"`
	OverwriteFiles                   *bool                    `url:"overwrite_files,omitempty" json:"overwrite_files,omitempty" path:"overwrite_files"`
	PathTimeZone                     string                   `url:"path_time_zone,omitempty" json:"path_time_zone,omitempty" path:"path_time_zone"`
	RetryOnFailureIntervalInMinutes  int64                    `url:"retry_on_failure_interval_in_minutes,omitempty" json:"retry_on_failure_interval_in_minutes,omitempty" path:"retry_on_failure_interval_in_minutes"`
	RetryOnFailureNumberOfAttempts   int64                    `url:"retry_on_failure_number_of_attempts,omitempty" json:"retry_on_failure_number_of_attempts,omitempty" path:"retry_on_failure_number_of_attempts"`
	Trigger                          AutomationTriggerEnum    `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	TriggerActions                   []string                 `url:"trigger_actions,omitempty" json:"trigger_actions,omitempty" path:"trigger_actions"`
	Value                            map[string]interface{}   `url:"value,omitempty" json:"value,omitempty" path:"value"`
	RecurringDay                     int64                    `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	Automation                       AutomationEnum           `url:"automation,omitempty" json:"automation,omitempty" path:"automation"`
}

type AutomationDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (a *Automation) UnmarshalJSON(data []byte) error {
	type automation Automation
	var v automation
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = Automation(v)
	return nil
}

func (a *AutomationCollection) UnmarshalJSON(data []byte) error {
	type automations AutomationCollection
	var v automations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AutomationCollection(v)
	return nil
}

func (a *AutomationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
