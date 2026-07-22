package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Expectation struct {
	Id                     int64       `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId            int64       `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Name                   string      `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Description            string      `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	Path                   string      `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Source                 string      `json:"source,omitempty" path:"source,omitempty" url:"source,omitempty"`
	ExcludePattern         string      `json:"exclude_pattern,omitempty" path:"exclude_pattern,omitempty" url:"exclude_pattern,omitempty"`
	Disabled               *bool       `json:"disabled,omitempty" path:"disabled,omitempty" url:"disabled,omitempty"`
	ExpectationsVersion    int64       `json:"expectations_version,omitempty" path:"expectations_version,omitempty" url:"expectations_version,omitempty"`
	Trigger                string      `json:"trigger,omitempty" path:"trigger,omitempty" url:"trigger,omitempty"`
	Interval               string      `json:"interval,omitempty" path:"interval,omitempty" url:"interval,omitempty"`
	RecurringDay           int64       `json:"recurring_day,omitempty" path:"recurring_day,omitempty" url:"recurring_day,omitempty"`
	ScheduleDaysOfWeek     []int64     `json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week,omitempty" url:"schedule_days_of_week,omitempty"`
	ScheduleTimesOfDay     []string    `json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day,omitempty" url:"schedule_times_of_day,omitempty"`
	ScheduleTimeZone       string      `json:"schedule_time_zone,omitempty" path:"schedule_time_zone,omitempty" url:"schedule_time_zone,omitempty"`
	HolidayRegion          string      `json:"holiday_region,omitempty" path:"holiday_region,omitempty" url:"holiday_region,omitempty"`
	LookbackInterval       int64       `json:"lookback_interval,omitempty" path:"lookback_interval,omitempty" url:"lookback_interval,omitempty"`
	LateAcceptanceInterval int64       `json:"late_acceptance_interval,omitempty" path:"late_acceptance_interval,omitempty" url:"late_acceptance_interval,omitempty"`
	InactivityInterval     int64       `json:"inactivity_interval,omitempty" path:"inactivity_interval,omitempty" url:"inactivity_interval,omitempty"`
	MaxOpenInterval        int64       `json:"max_open_interval,omitempty" path:"max_open_interval,omitempty" url:"max_open_interval,omitempty"`
	Criteria               interface{} `json:"criteria,omitempty" path:"criteria,omitempty" url:"criteria,omitempty"`
	LastEvaluatedAt        *time.Time  `json:"last_evaluated_at,omitempty" path:"last_evaluated_at,omitempty" url:"last_evaluated_at,omitempty"`
	LastSuccessAt          *time.Time  `json:"last_success_at,omitempty" path:"last_success_at,omitempty" url:"last_success_at,omitempty"`
	LastFailureAt          *time.Time  `json:"last_failure_at,omitempty" path:"last_failure_at,omitempty" url:"last_failure_at,omitempty"`
	LastResult             string      `json:"last_result,omitempty" path:"last_result,omitempty" url:"last_result,omitempty"`
	CreatedAt              *time.Time  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt              *time.Time  `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (e Expectation) Identifier() interface{} {
	return e.Id
}

type ExpectationCollection []Expectation

type ExpectationTriggerEnum string

func (u ExpectationTriggerEnum) String() string {
	return string(u)
}

func (u ExpectationTriggerEnum) Enum() map[string]ExpectationTriggerEnum {
	return map[string]ExpectationTriggerEnum{
		"manual":          ExpectationTriggerEnum("manual"),
		"upload":          ExpectationTriggerEnum("upload"),
		"daily":           ExpectationTriggerEnum("daily"),
		"custom_schedule": ExpectationTriggerEnum("custom_schedule"),
	}
}

type ExpectationListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type ExpectationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ExpectationCreateParams struct {
	Name                   string                 `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description            string                 `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Path                   string                 `url:"path,omitempty" json:"path,omitempty" path:"path"`
	Source                 string                 `url:"source,omitempty" json:"source,omitempty" path:"source"`
	ExcludePattern         string                 `url:"exclude_pattern,omitempty" json:"exclude_pattern,omitempty" path:"exclude_pattern"`
	Disabled               *bool                  `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	Trigger                ExpectationTriggerEnum `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	Interval               string                 `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	RecurringDay           int64                  `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleDaysOfWeek     []int64                `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay     []string               `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	ScheduleTimeZone       string                 `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	HolidayRegion          string                 `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	LookbackInterval       int64                  `url:"lookback_interval,omitempty" json:"lookback_interval,omitempty" path:"lookback_interval"`
	LateAcceptanceInterval int64                  `url:"late_acceptance_interval,omitempty" json:"late_acceptance_interval,omitempty" path:"late_acceptance_interval"`
	InactivityInterval     int64                  `url:"inactivity_interval,omitempty" json:"inactivity_interval,omitempty" path:"inactivity_interval"`
	MaxOpenInterval        int64                  `url:"max_open_interval,omitempty" json:"max_open_interval,omitempty" path:"max_open_interval"`
	Criteria               interface{}            `url:"criteria,omitempty" json:"criteria,omitempty" path:"criteria"`
	WorkspaceId            int64                  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

// Manually open an Expectation window
type ExpectationTriggerEvaluationParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ExpectationUpdateParams struct {
	Id                     int64                  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name                   string                 `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Description            string                 `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Path                   string                 `url:"path,omitempty" json:"path,omitempty" path:"path"`
	Source                 string                 `url:"source,omitempty" json:"source,omitempty" path:"source"`
	ExcludePattern         string                 `url:"exclude_pattern,omitempty" json:"exclude_pattern,omitempty" path:"exclude_pattern"`
	Disabled               *bool                  `url:"disabled,omitempty" json:"disabled,omitempty" path:"disabled"`
	Trigger                ExpectationTriggerEnum `url:"trigger,omitempty" json:"trigger,omitempty" path:"trigger"`
	Interval               string                 `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	RecurringDay           int64                  `url:"recurring_day,omitempty" json:"recurring_day,omitempty" path:"recurring_day"`
	ScheduleDaysOfWeek     []int64                `url:"schedule_days_of_week,omitempty" json:"schedule_days_of_week,omitempty" path:"schedule_days_of_week"`
	ScheduleTimesOfDay     []string               `url:"schedule_times_of_day,omitempty" json:"schedule_times_of_day,omitempty" path:"schedule_times_of_day"`
	ScheduleTimeZone       string                 `url:"schedule_time_zone,omitempty" json:"schedule_time_zone,omitempty" path:"schedule_time_zone"`
	HolidayRegion          string                 `url:"holiday_region,omitempty" json:"holiday_region,omitempty" path:"holiday_region"`
	LookbackInterval       int64                  `url:"lookback_interval,omitempty" json:"lookback_interval,omitempty" path:"lookback_interval"`
	LateAcceptanceInterval int64                  `url:"late_acceptance_interval,omitempty" json:"late_acceptance_interval,omitempty" path:"late_acceptance_interval"`
	InactivityInterval     int64                  `url:"inactivity_interval,omitempty" json:"inactivity_interval,omitempty" path:"inactivity_interval"`
	MaxOpenInterval        int64                  `url:"max_open_interval,omitempty" json:"max_open_interval,omitempty" path:"max_open_interval"`
	Criteria               interface{}            `url:"criteria,omitempty" json:"criteria,omitempty" path:"criteria"`
	WorkspaceId            int64                  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
}

type ExpectationDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (e *Expectation) UnmarshalJSON(data []byte) error {
	type expectation Expectation
	var v expectation
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = Expectation(v)
	return nil
}

func (e *ExpectationCollection) UnmarshalJSON(data []byte) error {
	type expectations ExpectationCollection
	var v expectations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = ExpectationCollection(v)
	return nil
}

func (e *ExpectationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
