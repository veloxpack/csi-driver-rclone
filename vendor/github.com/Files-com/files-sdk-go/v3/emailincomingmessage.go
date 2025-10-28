package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type EmailIncomingMessage struct {
	Id         int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	InboxId    int64      `json:"inbox_id,omitempty" path:"inbox_id,omitempty" url:"inbox_id,omitempty"`
	Sender     string     `json:"sender,omitempty" path:"sender,omitempty" url:"sender,omitempty"`
	SenderName string     `json:"sender_name,omitempty" path:"sender_name,omitempty" url:"sender_name,omitempty"`
	Status     string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Body       string     `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	Message    string     `json:"message,omitempty" path:"message,omitempty" url:"message,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	InboxTitle string     `json:"inbox_title,omitempty" path:"inbox_title,omitempty" url:"inbox_title,omitempty"`
}

func (e EmailIncomingMessage) Identifier() interface{} {
	return e.Id
}

type EmailIncomingMessageCollection []EmailIncomingMessage

type EmailIncomingMessageListParams struct {
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       EmailIncomingMessage   `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (e *EmailIncomingMessage) UnmarshalJSON(data []byte) error {
	type emailIncomingMessage EmailIncomingMessage
	var v emailIncomingMessage
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = EmailIncomingMessage(v)
	return nil
}

func (e *EmailIncomingMessageCollection) UnmarshalJSON(data []byte) error {
	type emailIncomingMessages EmailIncomingMessageCollection
	var v emailIncomingMessages
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = EmailIncomingMessageCollection(v)
	return nil
}

func (e *EmailIncomingMessageCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
