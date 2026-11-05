package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type EventRecord struct {
	Id           int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	WorkspaceId  int64                    `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	EventUuid    string                   `json:"event_uuid,omitempty" path:"event_uuid,omitempty" url:"event_uuid,omitempty"`
	EventType    string                   `json:"event_type,omitempty" path:"event_type,omitempty" url:"event_type,omitempty"`
	Severity     string                   `json:"severity,omitempty" path:"severity,omitempty" url:"severity,omitempty"`
	SourceType   string                   `json:"source_type,omitempty" path:"source_type,omitempty" url:"source_type,omitempty"`
	SourceId     int64                    `json:"source_id,omitempty" path:"source_id,omitempty" url:"source_id,omitempty"`
	OccurredAt   *time.Time               `json:"occurred_at,omitempty" path:"occurred_at,omitempty" url:"occurred_at,omitempty"`
	HumanTitle   string                   `json:"human_title,omitempty" path:"human_title,omitempty" url:"human_title,omitempty"`
	HumanSummary string                   `json:"human_summary,omitempty" path:"human_summary,omitempty" url:"human_summary,omitempty"`
	HumanFields  []map[string]interface{} `json:"human_fields,omitempty" path:"human_fields,omitempty" url:"human_fields,omitempty"`
	Actor        interface{}              `json:"actor,omitempty" path:"actor,omitempty" url:"actor,omitempty"`
	Resources    []map[string]interface{} `json:"resources,omitempty" path:"resources,omitempty" url:"resources,omitempty"`
	Payload      interface{}              `json:"payload,omitempty" path:"payload,omitempty" url:"payload,omitempty"`
	CreatedAt    *time.Time               `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

func (e EventRecord) Identifier() interface{} {
	return e.Id
}

type EventRecordCollection []EventRecord

type EventRecordListParams struct {
	SortBy       interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type EventRecordFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (e *EventRecord) UnmarshalJSON(data []byte) error {
	type eventRecord EventRecord
	var v eventRecord
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = EventRecord(v)
	return nil
}

func (e *EventRecordCollection) UnmarshalJSON(data []byte) error {
	type eventRecords EventRecordCollection
	var v eventRecords
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = EventRecordCollection(v)
	return nil
}

func (e *EventRecordCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
