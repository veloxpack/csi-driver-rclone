package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ExpectationIncident struct {
	Id                     int64       `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId            int64       `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	ExpectationId          int64       `json:"expectation_id,omitempty" path:"expectation_id,omitempty" url:"expectation_id,omitempty"`
	Status                 string      `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	OpenedAt               *time.Time  `json:"opened_at,omitempty" path:"opened_at,omitempty" url:"opened_at,omitempty"`
	LastFailedAt           *time.Time  `json:"last_failed_at,omitempty" path:"last_failed_at,omitempty" url:"last_failed_at,omitempty"`
	AcknowledgedAt         *time.Time  `json:"acknowledged_at,omitempty" path:"acknowledged_at,omitempty" url:"acknowledged_at,omitempty"`
	SnoozedUntil           *time.Time  `json:"snoozed_until,omitempty" path:"snoozed_until,omitempty" url:"snoozed_until,omitempty"`
	ResolvedAt             *time.Time  `json:"resolved_at,omitempty" path:"resolved_at,omitempty" url:"resolved_at,omitempty"`
	OpenedByEvaluationId   int64       `json:"opened_by_evaluation_id,omitempty" path:"opened_by_evaluation_id,omitempty" url:"opened_by_evaluation_id,omitempty"`
	LastEvaluationId       int64       `json:"last_evaluation_id,omitempty" path:"last_evaluation_id,omitempty" url:"last_evaluation_id,omitempty"`
	ResolvedByEvaluationId int64       `json:"resolved_by_evaluation_id,omitempty" path:"resolved_by_evaluation_id,omitempty" url:"resolved_by_evaluation_id,omitempty"`
	Summary                interface{} `json:"summary,omitempty" path:"summary,omitempty" url:"summary,omitempty"`
	CreatedAt              *time.Time  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt              *time.Time  `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (e ExpectationIncident) Identifier() interface{} {
	return e.Id
}

type ExpectationIncidentCollection []ExpectationIncident

type ExpectationIncidentListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type ExpectationIncidentFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

// Resolve an expectation incident
type ExpectationIncidentResolveParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

// Snooze an expectation incident until a specified time
type ExpectationIncidentSnoozeParams struct {
	Id           int64      `url:"-,omitempty" json:"-,omitempty" path:"id"`
	SnoozedUntil *time.Time `url:"snoozed_until" json:"snoozed_until" path:"snoozed_until"`
}

// Acknowledge an expectation incident
type ExpectationIncidentAcknowledgeParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (e *ExpectationIncident) UnmarshalJSON(data []byte) error {
	type expectationIncident ExpectationIncident
	var v expectationIncident
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = ExpectationIncident(v)
	return nil
}

func (e *ExpectationIncidentCollection) UnmarshalJSON(data []byte) error {
	type expectationIncidents ExpectationIncidentCollection
	var v expectationIncidents
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = ExpectationIncidentCollection(v)
	return nil
}

func (e *ExpectationIncidentCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
