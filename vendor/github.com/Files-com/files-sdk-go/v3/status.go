package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Status struct {
	Code          int64                    `json:"code,omitempty" path:"code,omitempty" url:"code,omitempty"`
	Message       string                   `json:"message,omitempty" path:"message,omitempty" url:"message,omitempty"`
	Status        string                   `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Data          Auto                     `json:"data,omitempty" path:"data,omitempty" url:"data,omitempty"`
	Errors        []map[string]interface{} `json:"errors,omitempty" path:"errors,omitempty" url:"errors,omitempty"`
	ClickwrapId   int64                    `json:"clickwrap_id,omitempty" path:"clickwrap_id,omitempty" url:"clickwrap_id,omitempty"`
	ClickwrapBody string                   `json:"clickwrap_body,omitempty" path:"clickwrap_body,omitempty" url:"clickwrap_body,omitempty"`
}

// Identifier no path or id

type StatusCollection []Status

func (s *Status) UnmarshalJSON(data []byte) error {
	type status Status
	var v status
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Status(v)
	return nil
}

func (s *StatusCollection) UnmarshalJSON(data []byte) error {
	type statuss StatusCollection
	var v statuss
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = StatusCollection(v)
	return nil
}

func (s *StatusCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
