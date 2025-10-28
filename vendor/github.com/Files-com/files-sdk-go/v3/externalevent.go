package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ExternalEvent struct {
	Id                    int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	EventType             string     `json:"event_type,omitempty" path:"event_type,omitempty" url:"event_type,omitempty"`
	Status                string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Body                  string     `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	CreatedAt             *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	BodyUrl               string     `json:"body_url,omitempty" path:"body_url,omitempty" url:"body_url,omitempty"`
	FolderBehaviorId      int64      `json:"folder_behavior_id,omitempty" path:"folder_behavior_id,omitempty" url:"folder_behavior_id,omitempty"`
	SiemHttpDestinationId int64      `json:"siem_http_destination_id,omitempty" path:"siem_http_destination_id,omitempty" url:"siem_http_destination_id,omitempty"`
	SuccessfulFiles       int64      `json:"successful_files,omitempty" path:"successful_files,omitempty" url:"successful_files,omitempty"`
	ErroredFiles          int64      `json:"errored_files,omitempty" path:"errored_files,omitempty" url:"errored_files,omitempty"`
	BytesSynced           int64      `json:"bytes_synced,omitempty" path:"bytes_synced,omitempty" url:"bytes_synced,omitempty"`
	ComparedFiles         int64      `json:"compared_files,omitempty" path:"compared_files,omitempty" url:"compared_files,omitempty"`
	ComparedFolders       int64      `json:"compared_folders,omitempty" path:"compared_folders,omitempty" url:"compared_folders,omitempty"`
	RemoteServerType      string     `json:"remote_server_type,omitempty" path:"remote_server_type,omitempty" url:"remote_server_type,omitempty"`
}

func (e ExternalEvent) Identifier() interface{} {
	return e.Id
}

type ExternalEventCollection []ExternalEvent

type ExternalEventStatusEnum string

func (u ExternalEventStatusEnum) String() string {
	return string(u)
}

func (u ExternalEventStatusEnum) Enum() map[string]ExternalEventStatusEnum {
	return map[string]ExternalEventStatusEnum{
		"success":         ExternalEventStatusEnum("success"),
		"failure":         ExternalEventStatusEnum("failure"),
		"partial_failure": ExternalEventStatusEnum("partial_failure"),
		"in_progress":     ExternalEventStatusEnum("in_progress"),
		"skipped":         ExternalEventStatusEnum("skipped"),
	}
}

type ExternalEventListParams struct {
	SortBy     map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     ExternalEvent          `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type ExternalEventFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ExternalEventCreateParams struct {
	Status ExternalEventStatusEnum `url:"status" json:"status" path:"status"`
	Body   string                  `url:"body" json:"body" path:"body"`
}

func (e *ExternalEvent) UnmarshalJSON(data []byte) error {
	type externalEvent ExternalEvent
	var v externalEvent
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = ExternalEvent(v)
	return nil
}

func (e *ExternalEventCollection) UnmarshalJSON(data []byte) error {
	type externalEvents ExternalEventCollection
	var v externalEvents
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = ExternalEventCollection(v)
	return nil
}

func (e *ExternalEventCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
