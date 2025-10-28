package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type BundleRegistration struct {
	Code              string                 `json:"code,omitempty" path:"code,omitempty" url:"code,omitempty"`
	Name              string                 `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Company           string                 `json:"company,omitempty" path:"company,omitempty" url:"company,omitempty"`
	Email             string                 `json:"email,omitempty" path:"email,omitempty" url:"email,omitempty"`
	Ip                string                 `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	InboxCode         string                 `json:"inbox_code,omitempty" path:"inbox_code,omitempty" url:"inbox_code,omitempty"`
	ClickwrapBody     string                 `json:"clickwrap_body,omitempty" path:"clickwrap_body,omitempty" url:"clickwrap_body,omitempty"`
	FormFieldSetId    int64                  `json:"form_field_set_id,omitempty" path:"form_field_set_id,omitempty" url:"form_field_set_id,omitempty"`
	FormFieldData     map[string]interface{} `json:"form_field_data,omitempty" path:"form_field_data,omitempty" url:"form_field_data,omitempty"`
	BundleCode        string                 `json:"bundle_code,omitempty" path:"bundle_code,omitempty" url:"bundle_code,omitempty"`
	BundleId          int64                  `json:"bundle_id,omitempty" path:"bundle_id,omitempty" url:"bundle_id,omitempty"`
	BundleRecipientId int64                  `json:"bundle_recipient_id,omitempty" path:"bundle_recipient_id,omitempty" url:"bundle_recipient_id,omitempty"`
	CreatedAt         *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

// Identifier no path or id

type BundleRegistrationCollection []BundleRegistration

type BundleRegistrationListParams struct {
	UserId   int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy   map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	BundleId int64                  `url:"bundle_id,omitempty" json:"bundle_id,omitempty" path:"bundle_id"`
	ListParams
}

func (b *BundleRegistration) UnmarshalJSON(data []byte) error {
	type bundleRegistration BundleRegistration
	var v bundleRegistration
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*b = BundleRegistration(v)
	return nil
}

func (b *BundleRegistrationCollection) UnmarshalJSON(data []byte) error {
	type bundleRegistrations BundleRegistrationCollection
	var v bundleRegistrations
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*b = BundleRegistrationCollection(v)
	return nil
}

func (b *BundleRegistrationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*b))
	for i, v := range *b {
		ret[i] = v
	}

	return &ret
}
