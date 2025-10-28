package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type HistoryExportResult struct {
	Id                     int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	CreatedAt              int64  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	CreatedAtIso8601       string `json:"created_at_iso8601,omitempty" path:"created_at_iso8601,omitempty" url:"created_at_iso8601,omitempty"`
	UserId                 int64  `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	FileId                 int64  `json:"file_id,omitempty" path:"file_id,omitempty" url:"file_id,omitempty"`
	ParentId               int64  `json:"parent_id,omitempty" path:"parent_id,omitempty" url:"parent_id,omitempty"`
	Path                   string `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Folder                 string `json:"folder,omitempty" path:"folder,omitempty" url:"folder,omitempty"`
	Src                    string `json:"src,omitempty" path:"src,omitempty" url:"src,omitempty"`
	Destination            string `json:"destination,omitempty" path:"destination,omitempty" url:"destination,omitempty"`
	Ip                     string `json:"ip,omitempty" path:"ip,omitempty" url:"ip,omitempty"`
	Username               string `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	UserIsFromParentSite   *bool  `json:"user_is_from_parent_site,omitempty" path:"user_is_from_parent_site,omitempty" url:"user_is_from_parent_site,omitempty"`
	Action                 string `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	FailureType            string `json:"failure_type,omitempty" path:"failure_type,omitempty" url:"failure_type,omitempty"`
	Interface              string `json:"interface,omitempty" path:"interface,omitempty" url:"interface,omitempty"`
	TargetId               int64  `json:"target_id,omitempty" path:"target_id,omitempty" url:"target_id,omitempty"`
	TargetName             string `json:"target_name,omitempty" path:"target_name,omitempty" url:"target_name,omitempty"`
	TargetPermission       string `json:"target_permission,omitempty" path:"target_permission,omitempty" url:"target_permission,omitempty"`
	TargetRecursive        *bool  `json:"target_recursive,omitempty" path:"target_recursive,omitempty" url:"target_recursive,omitempty"`
	TargetExpiresAt        int64  `json:"target_expires_at,omitempty" path:"target_expires_at,omitempty" url:"target_expires_at,omitempty"`
	TargetExpiresAtIso8601 string `json:"target_expires_at_iso8601,omitempty" path:"target_expires_at_iso8601,omitempty" url:"target_expires_at_iso8601,omitempty"`
	TargetPermissionSet    string `json:"target_permission_set,omitempty" path:"target_permission_set,omitempty" url:"target_permission_set,omitempty"`
	TargetPlatform         string `json:"target_platform,omitempty" path:"target_platform,omitempty" url:"target_platform,omitempty"`
	TargetUsername         string `json:"target_username,omitempty" path:"target_username,omitempty" url:"target_username,omitempty"`
	TargetUserId           int64  `json:"target_user_id,omitempty" path:"target_user_id,omitempty" url:"target_user_id,omitempty"`
}

func (h HistoryExportResult) Identifier() interface{} {
	return h.Id
}

type HistoryExportResultCollection []HistoryExportResult

type HistoryExportResultListParams struct {
	UserId          int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	HistoryExportId int64 `url:"history_export_id" json:"history_export_id" path:"history_export_id"`
	ListParams
}

func (h *HistoryExportResult) UnmarshalJSON(data []byte) error {
	type historyExportResult HistoryExportResult
	var v historyExportResult
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*h = HistoryExportResult(v)
	return nil
}

func (h *HistoryExportResultCollection) UnmarshalJSON(data []byte) error {
	type historyExportResults HistoryExportResultCollection
	var v historyExportResults
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*h = HistoryExportResultCollection(v)
	return nil
}

func (h *HistoryExportResultCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*h))
	for i, v := range *h {
		ret[i] = v
	}

	return &ret
}
