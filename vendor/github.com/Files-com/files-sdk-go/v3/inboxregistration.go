package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type InboxRegistration struct {
	Code             string                 `json:"code,omitempty" path:"code,omitempty" url:"code,omitempty"`
	Name             string                 `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Company          string                 `json:"company,omitempty" path:"company,omitempty" url:"company,omitempty"`
	Email            string                 `json:"email,omitempty" path:"email,omitempty" url:"email,omitempty"`
	Ip               string                 `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	ClickwrapBody    string                 `json:"clickwrap_body,omitempty" path:"clickwrap_body,omitempty" url:"clickwrap_body,omitempty"`
	FormFieldSetId   int64                  `json:"form_field_set_id,omitempty" path:"form_field_set_id,omitempty" url:"form_field_set_id,omitempty"`
	FormFieldData    map[string]interface{} `json:"form_field_data,omitempty" path:"form_field_data,omitempty" url:"form_field_data,omitempty"`
	InboxId          int64                  `json:"inbox_id,omitempty" path:"inbox_id,omitempty" url:"inbox_id,omitempty"`
	InboxRecipientId int64                  `json:"inbox_recipient_id,omitempty" path:"inbox_recipient_id,omitempty" url:"inbox_recipient_id,omitempty"`
	InboxTitle       string                 `json:"inbox_title,omitempty" path:"inbox_title,omitempty" url:"inbox_title,omitempty"`
	CreatedAt        *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

// Identifier no path or id

type InboxRegistrationCollection []InboxRegistration

type InboxRegistrationListParams struct {
	FolderBehaviorId int64 `url:"folder_behavior_id,omitempty" json:"folder_behavior_id,omitempty" path:"folder_behavior_id"`
	ListParams
}

func (i *InboxRegistration) UnmarshalJSON(data []byte) error {
	type inboxRegistration InboxRegistration
	var v inboxRegistration
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*i = InboxRegistration(v)
	return nil
}

func (i *InboxRegistrationCollection) UnmarshalJSON(data []byte) error {
	type inboxRegistrations InboxRegistrationCollection
	var v inboxRegistrations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*i = InboxRegistrationCollection(v)
	return nil
}

func (i *InboxRegistrationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*i))
	for i, v := range *i {
		ret[i] = v
	}

	return &ret
}
