package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type AutomationExecutionNode struct {
	NodeId               string                   `json:"node_id,omitempty" path:"node_id,omitempty" url:"node_id,omitempty"`
	NodeType             string                   `json:"node_type,omitempty" path:"node_type,omitempty" url:"node_type,omitempty"`
	Status               string                   `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	RunStage             string                   `json:"run_stage,omitempty" path:"run_stage,omitempty" url:"run_stage,omitempty"`
	Reused               *bool                    `json:"reused,omitempty" path:"reused,omitempty" url:"reused,omitempty"`
	SuccessfulOperations int64                    `json:"successful_operations,omitempty" path:"successful_operations,omitempty" url:"successful_operations,omitempty"`
	FailedOperations     int64                    `json:"failed_operations,omitempty" path:"failed_operations,omitempty" url:"failed_operations,omitempty"`
	StartedAt            *time.Time               `json:"started_at,omitempty" path:"started_at,omitempty" url:"started_at,omitempty"`
	CompletedAt          *time.Time               `json:"completed_at,omitempty" path:"completed_at,omitempty" url:"completed_at,omitempty"`
	DurationMs           int64                    `json:"duration_ms,omitempty" path:"duration_ms,omitempty" url:"duration_ms,omitempty"`
	Inputs               []map[string]interface{} `json:"inputs,omitempty" path:"inputs,omitempty" url:"inputs,omitempty"`
	Outputs              interface{}              `json:"outputs,omitempty" path:"outputs,omitempty" url:"outputs,omitempty"`
	InputItems           interface{}              `json:"input_items,omitempty" path:"input_items,omitempty" url:"input_items,omitempty"`
	OutputItems          interface{}              `json:"output_items,omitempty" path:"output_items,omitempty" url:"output_items,omitempty"`
}

// Identifier no path or id

type AutomationExecutionNodeCollection []AutomationExecutionNode

func (a *AutomationExecutionNode) UnmarshalJSON(data []byte) error {
	type automationExecutionNode AutomationExecutionNode
	var v automationExecutionNode
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = AutomationExecutionNode(v)
	return nil
}

func (a *AutomationExecutionNodeCollection) UnmarshalJSON(data []byte) error {
	type automationExecutionNodes AutomationExecutionNodeCollection
	var v automationExecutionNodes
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AutomationExecutionNodeCollection(v)
	return nil
}

func (a *AutomationExecutionNodeCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
