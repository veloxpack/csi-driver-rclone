package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ActionNotificationExportResult struct {
	Id             int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	Status         int64  `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	Message        string `json:"message,omitempty" path:"message,omitempty" url:"message,omitempty"`
	Success        *bool  `json:"success,omitempty" path:"success,omitempty" url:"success,omitempty"`
	RequestHeaders string `json:"request_headers,omitempty" path:"request_headers,omitempty" url:"request_headers,omitempty"`
	RequestMethod  string `json:"request_method,omitempty" path:"request_method,omitempty" url:"request_method,omitempty"`
	RequestUrl     string `json:"request_url,omitempty" path:"request_url,omitempty" url:"request_url,omitempty"`
	Path           string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Folder         string `json:"folder,omitempty" path:"folder,omitempty" url:"folder,omitempty"`
}

func (a ActionNotificationExportResult) Identifier() interface{} {
	return a.Id
}

type ActionNotificationExportResultCollection []ActionNotificationExportResult

type ActionNotificationExportResultListParams struct {
	UserId                     int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	ActionNotificationExportId int64 `url:"action_notification_export_id" json:"action_notification_export_id" path:"action_notification_export_id"`
	ListParams
}

func (a *ActionNotificationExportResult) UnmarshalJSON(data []byte) error {
	type actionNotificationExportResult ActionNotificationExportResult
	var v actionNotificationExportResult
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = ActionNotificationExportResult(v)
	return nil
}

func (a *ActionNotificationExportResultCollection) UnmarshalJSON(data []byte) error {
	type actionNotificationExportResults ActionNotificationExportResultCollection
	var v actionNotificationExportResults
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = ActionNotificationExportResultCollection(v)
	return nil
}

func (a *ActionNotificationExportResultCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
