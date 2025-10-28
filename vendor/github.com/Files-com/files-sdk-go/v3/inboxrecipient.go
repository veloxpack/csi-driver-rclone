package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type InboxRecipient struct {
	Company          string     `json:"company,omitempty" path:"company,omitempty" url:"company,omitempty"`
	Name             string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Note             string     `json:"note,omitempty" path:"note,omitempty" url:"note,omitempty"`
	Recipient        string     `json:"recipient,omitempty" path:"recipient,omitempty" url:"recipient,omitempty"`
	SentAt           *time.Time `json:"sent_at,omitempty" path:"sent_at,omitempty" url:"sent_at,omitempty"`
	InboxId          int64      `json:"inbox_id,omitempty" path:"inbox_id,omitempty" url:"inbox_id,omitempty"`
	ShareAfterCreate *bool      `json:"share_after_create,omitempty" path:"share_after_create,omitempty" url:"share_after_create,omitempty"`
}

// Identifier no path or id

type InboxRecipientCollection []InboxRecipient

type InboxRecipientListParams struct {
	SortBy  map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter  InboxRecipient         `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	InboxId int64                  `url:"inbox_id" json:"inbox_id" path:"inbox_id"`
	ListParams
}

type InboxRecipientCreateParams struct {
	InboxId          int64  `url:"inbox_id" json:"inbox_id" path:"inbox_id"`
	Recipient        string `url:"recipient" json:"recipient" path:"recipient"`
	Name             string `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Company          string `url:"company,omitempty" json:"company,omitempty" path:"company"`
	Note             string `url:"note,omitempty" json:"note,omitempty" path:"note"`
	ShareAfterCreate *bool  `url:"share_after_create,omitempty" json:"share_after_create,omitempty" path:"share_after_create"`
}

func (i *InboxRecipient) UnmarshalJSON(data []byte) error {
	type inboxRecipient InboxRecipient
	var v inboxRecipient
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*i = InboxRecipient(v)
	return nil
}

func (i *InboxRecipientCollection) UnmarshalJSON(data []byte) error {
	type inboxRecipients InboxRecipientCollection
	var v inboxRecipients
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*i = InboxRecipientCollection(v)
	return nil
}

func (i *InboxRecipientCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*i))
	for i, v := range *i {
		ret[i] = v
	}

	return &ret
}
