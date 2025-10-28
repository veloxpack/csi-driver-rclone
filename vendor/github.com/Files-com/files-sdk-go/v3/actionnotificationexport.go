package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ActionNotificationExport struct {
	Id                 int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	ExportVersion      string     `json:"export_version,omitempty" path:"export_version,omitempty" url:"export_version,omitempty"`
	StartAt            *time.Time `json:"start_at,omitempty" path:"start_at,omitempty" url:"start_at,omitempty"`
	EndAt              *time.Time `json:"end_at,omitempty" path:"end_at,omitempty" url:"end_at,omitempty"`
	Status             string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	QueryPath          string     `json:"query_path,omitempty" path:"query_path,omitempty" url:"query_path,omitempty"`
	QueryFolder        string     `json:"query_folder,omitempty" path:"query_folder,omitempty" url:"query_folder,omitempty"`
	QueryMessage       string     `json:"query_message,omitempty" path:"query_message,omitempty" url:"query_message,omitempty"`
	QueryRequestMethod string     `json:"query_request_method,omitempty" path:"query_request_method,omitempty" url:"query_request_method,omitempty"`
	QueryRequestUrl    string     `json:"query_request_url,omitempty" path:"query_request_url,omitempty" url:"query_request_url,omitempty"`
	QueryStatus        string     `json:"query_status,omitempty" path:"query_status,omitempty" url:"query_status,omitempty"`
	QuerySuccess       *bool      `json:"query_success,omitempty" path:"query_success,omitempty" url:"query_success,omitempty"`
	ResultsUrl         string     `json:"results_url,omitempty" path:"results_url,omitempty" url:"results_url,omitempty"`
	UserId             int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
}

func (a ActionNotificationExport) Identifier() interface{} {
	return a.Id
}

type ActionNotificationExportCollection []ActionNotificationExport

type ActionNotificationExportFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type ActionNotificationExportCreateParams struct {
	UserId             int64      `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	StartAt            *time.Time `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt              *time.Time `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	QueryMessage       string     `url:"query_message,omitempty" json:"query_message,omitempty" path:"query_message"`
	QueryRequestMethod string     `url:"query_request_method,omitempty" json:"query_request_method,omitempty" path:"query_request_method"`
	QueryRequestUrl    string     `url:"query_request_url,omitempty" json:"query_request_url,omitempty" path:"query_request_url"`
	QueryStatus        string     `url:"query_status,omitempty" json:"query_status,omitempty" path:"query_status"`
	QuerySuccess       *bool      `url:"query_success,omitempty" json:"query_success,omitempty" path:"query_success"`
	QueryPath          string     `url:"query_path,omitempty" json:"query_path,omitempty" path:"query_path"`
	QueryFolder        string     `url:"query_folder,omitempty" json:"query_folder,omitempty" path:"query_folder"`
}

func (a *ActionNotificationExport) UnmarshalJSON(data []byte) error {
	type actionNotificationExport ActionNotificationExport
	var v actionNotificationExport
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = ActionNotificationExport(v)
	return nil
}

func (a *ActionNotificationExportCollection) UnmarshalJSON(data []byte) error {
	type actionNotificationExports ActionNotificationExportCollection
	var v actionNotificationExports
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = ActionNotificationExportCollection(v)
	return nil
}

func (a *ActionNotificationExportCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
