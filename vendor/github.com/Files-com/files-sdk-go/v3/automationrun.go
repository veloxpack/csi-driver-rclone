package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type AutomationRun struct {
	Id                   int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	AutomationId         int64      `json:"automation_id,omitempty" path:"automation_id,omitempty" url:"automation_id,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty" path:"completed_at,omitempty" url:"completed_at,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	RetryAt              *time.Time `json:"retry_at,omitempty" path:"retry_at,omitempty" url:"retry_at,omitempty"`
	RetriedAt            *time.Time `json:"retried_at,omitempty" path:"retried_at,omitempty" url:"retried_at,omitempty"`
	RetriedInRunId       int64      `json:"retried_in_run_id,omitempty" path:"retried_in_run_id,omitempty" url:"retried_in_run_id,omitempty"`
	RetryOfRunId         int64      `json:"retry_of_run_id,omitempty" path:"retry_of_run_id,omitempty" url:"retry_of_run_id,omitempty"`
	Runtime              string     `json:"runtime,omitempty" path:"runtime,omitempty" url:"runtime,omitempty"`
	Status               string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	SuccessfulOperations int64      `json:"successful_operations,omitempty" path:"successful_operations,omitempty" url:"successful_operations,omitempty"`
	FailedOperations     int64      `json:"failed_operations,omitempty" path:"failed_operations,omitempty" url:"failed_operations,omitempty"`
	StatusMessagesUrl    string     `json:"status_messages_url,omitempty" path:"status_messages_url,omitempty" url:"status_messages_url,omitempty"`
}

func (a AutomationRun) Identifier() interface{} {
	return a.Id
}

type AutomationRunCollection []AutomationRun

type AutomationRunListParams struct {
	UserId       int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       AutomationRun          `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	AutomationId int64                  `url:"automation_id" json:"automation_id" path:"automation_id"`
	ListParams
}

type AutomationRunFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (a *AutomationRun) UnmarshalJSON(data []byte) error {
	type automationRun AutomationRun
	var v automationRun
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = AutomationRun(v)
	return nil
}

func (a *AutomationRunCollection) UnmarshalJSON(data []byte) error {
	type automationRuns AutomationRunCollection
	var v automationRuns
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AutomationRunCollection(v)
	return nil
}

func (a *AutomationRunCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
