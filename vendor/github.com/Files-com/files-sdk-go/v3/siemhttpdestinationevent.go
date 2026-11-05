package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SiemHttpDestinationEvent struct {
	Id                    int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	EventType             string     `json:"event_type,omitempty" path:"event_type,omitempty" url:"event_type,omitempty"`
	Status                string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Body                  string     `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	EventErrors           []string   `json:"event_errors,omitempty" path:"event_errors,omitempty" url:"event_errors,omitempty"`
	CreatedAt             *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	BodyUrl               string     `json:"body_url,omitempty" path:"body_url,omitempty" url:"body_url,omitempty"`
	SiemHttpDestinationId int64      `json:"siem_http_destination_id,omitempty" path:"siem_http_destination_id,omitempty" url:"siem_http_destination_id,omitempty"`
}

func (s SiemHttpDestinationEvent) Identifier() interface{} {
	return s.Id
}

type SiemHttpDestinationEventCollection []SiemHttpDestinationEvent

type SiemHttpDestinationEventListParams struct {
	SortBy     interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type SiemHttpDestinationEventFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *SiemHttpDestinationEvent) UnmarshalJSON(data []byte) error {
	type siemHttpDestinationEvent SiemHttpDestinationEvent
	var v siemHttpDestinationEvent
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SiemHttpDestinationEvent(v)
	return nil
}

func (s *SiemHttpDestinationEventCollection) UnmarshalJSON(data []byte) error {
	type siemHttpDestinationEvents SiemHttpDestinationEventCollection
	var v siemHttpDestinationEvents
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SiemHttpDestinationEventCollection(v)
	return nil
}

func (s *SiemHttpDestinationEventCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
