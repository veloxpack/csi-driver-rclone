package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ExpectationEvaluation struct {
	Id                     int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId            int64                    `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	ExpectationId          int64                    `json:"expectation_id,omitempty" path:"expectation_id,omitempty" url:"expectation_id,omitempty"`
	Status                 string                   `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	OpenedVia              string                   `json:"opened_via,omitempty" path:"opened_via,omitempty" url:"opened_via,omitempty"`
	OpenedAt               *time.Time               `json:"opened_at,omitempty" path:"opened_at,omitempty" url:"opened_at,omitempty"`
	WindowStartAt          *time.Time               `json:"window_start_at,omitempty" path:"window_start_at,omitempty" url:"window_start_at,omitempty"`
	WindowEndAt            *time.Time               `json:"window_end_at,omitempty" path:"window_end_at,omitempty" url:"window_end_at,omitempty"`
	DeadlineAt             *time.Time               `json:"deadline_at,omitempty" path:"deadline_at,omitempty" url:"deadline_at,omitempty"`
	LateAcceptanceCutoffAt *time.Time               `json:"late_acceptance_cutoff_at,omitempty" path:"late_acceptance_cutoff_at,omitempty" url:"late_acceptance_cutoff_at,omitempty"`
	HardCloseAt            *time.Time               `json:"hard_close_at,omitempty" path:"hard_close_at,omitempty" url:"hard_close_at,omitempty"`
	ClosedAt               *time.Time               `json:"closed_at,omitempty" path:"closed_at,omitempty" url:"closed_at,omitempty"`
	MatchedFiles           []map[string]interface{} `json:"matched_files,omitempty" path:"matched_files,omitempty" url:"matched_files,omitempty"`
	MissingFiles           []map[string]interface{} `json:"missing_files,omitempty" path:"missing_files,omitempty" url:"missing_files,omitempty"`
	CriteriaErrors         []map[string]interface{} `json:"criteria_errors,omitempty" path:"criteria_errors,omitempty" url:"criteria_errors,omitempty"`
	Summary                interface{}              `json:"summary,omitempty" path:"summary,omitempty" url:"summary,omitempty"`
	CreatedAt              *time.Time               `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt              *time.Time               `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (e ExpectationEvaluation) Identifier() interface{} {
	return e.Id
}

type ExpectationEvaluationCollection []ExpectationEvaluation

type ExpectationEvaluationListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type ExpectationEvaluationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (e *ExpectationEvaluation) UnmarshalJSON(data []byte) error {
	type expectationEvaluation ExpectationEvaluation
	var v expectationEvaluation
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = ExpectationEvaluation(v)
	return nil
}

func (e *ExpectationEvaluationCollection) UnmarshalJSON(data []byte) error {
	type expectationEvaluations ExpectationEvaluationCollection
	var v expectationEvaluations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = ExpectationEvaluationCollection(v)
	return nil
}

func (e *ExpectationEvaluationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
