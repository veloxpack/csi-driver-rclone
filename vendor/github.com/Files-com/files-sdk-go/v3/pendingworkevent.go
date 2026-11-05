package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PendingWorkEvent struct {
	Id               int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	EventType        string     `json:"event_type,omitempty" path:"event_type,omitempty" url:"event_type,omitempty"`
	Status           string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Body             string     `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	EventErrors      []string   `json:"event_errors,omitempty" path:"event_errors,omitempty" url:"event_errors,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	BodyUrl          string     `json:"body_url,omitempty" path:"body_url,omitempty" url:"body_url,omitempty"`
	FolderBehaviorId int64      `json:"folder_behavior_id,omitempty" path:"folder_behavior_id,omitempty" url:"folder_behavior_id,omitempty"`
}

func (p PendingWorkEvent) Identifier() interface{} {
	return p.Id
}

type PendingWorkEventCollection []PendingWorkEvent

type PendingWorkEventListParams struct {
	SortBy     interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type PendingWorkEventFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *PendingWorkEvent) UnmarshalJSON(data []byte) error {
	type pendingWorkEvent PendingWorkEvent
	var v pendingWorkEvent
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PendingWorkEvent(v)
	return nil
}

func (p *PendingWorkEventCollection) UnmarshalJSON(data []byte) error {
	type pendingWorkEvents PendingWorkEventCollection
	var v pendingWorkEvents
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PendingWorkEventCollection(v)
	return nil
}

func (p *PendingWorkEventCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
