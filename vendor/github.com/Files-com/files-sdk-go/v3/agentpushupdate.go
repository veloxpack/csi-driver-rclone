package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type AgentPushUpdate struct {
	Version        string `json:"version,omitempty" path:"version,omitempty" url:"version,omitempty"`
	Message        string `json:"message,omitempty" path:"message,omitempty" url:"message,omitempty"`
	CurrentVersion string `json:"current_version,omitempty" path:"current_version,omitempty" url:"current_version,omitempty"`
	PendingVersion string `json:"pending_version,omitempty" path:"pending_version,omitempty" url:"pending_version,omitempty"`
	LastError      string `json:"last_error,omitempty" path:"last_error,omitempty" url:"last_error,omitempty"`
	Error          string `json:"error,omitempty" path:"error,omitempty" url:"error,omitempty"`
}

// Identifier no path or id

type AgentPushUpdateCollection []AgentPushUpdate

func (a *AgentPushUpdate) UnmarshalJSON(data []byte) error {
	type agentPushUpdate AgentPushUpdate
	var v agentPushUpdate
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = AgentPushUpdate(v)
	return nil
}

func (a *AgentPushUpdateCollection) UnmarshalJSON(data []byte) error {
	type agentPushUpdates AgentPushUpdateCollection
	var v agentPushUpdates
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = AgentPushUpdateCollection(v)
	return nil
}

func (a *AgentPushUpdateCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
