package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type HistoryExport struct {
	Id                       int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	HistoryVersion           string     `json:"history_version,omitempty" path:"history_version,omitempty" url:"history_version,omitempty"`
	StartAt                  *time.Time `json:"start_at,omitempty" path:"start_at,omitempty" url:"start_at,omitempty"`
	EndAt                    *time.Time `json:"end_at,omitempty" path:"end_at,omitempty" url:"end_at,omitempty"`
	Status                   string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	QueryAction              string     `json:"query_action,omitempty" path:"query_action,omitempty" url:"query_action,omitempty"`
	QueryInterface           string     `json:"query_interface,omitempty" path:"query_interface,omitempty" url:"query_interface,omitempty"`
	QueryUserId              string     `json:"query_user_id,omitempty" path:"query_user_id,omitempty" url:"query_user_id,omitempty"`
	QueryFileId              string     `json:"query_file_id,omitempty" path:"query_file_id,omitempty" url:"query_file_id,omitempty"`
	QueryParentId            string     `json:"query_parent_id,omitempty" path:"query_parent_id,omitempty" url:"query_parent_id,omitempty"`
	QueryPath                string     `json:"query_path,omitempty" path:"query_path,omitempty" url:"query_path,omitempty"`
	QueryFolder              string     `json:"query_folder,omitempty" path:"query_folder,omitempty" url:"query_folder,omitempty"`
	QuerySrc                 string     `json:"query_src,omitempty" path:"query_src,omitempty" url:"query_src,omitempty"`
	QueryDestination         string     `json:"query_destination,omitempty" path:"query_destination,omitempty" url:"query_destination,omitempty"`
	QueryIp                  string     `json:"query_ip,omitempty" path:"query_ip,omitempty" url:"query_ip,omitempty"`
	QueryUsername            string     `json:"query_username,omitempty" path:"query_username,omitempty" url:"query_username,omitempty"`
	QueryFailureType         string     `json:"query_failure_type,omitempty" path:"query_failure_type,omitempty" url:"query_failure_type,omitempty"`
	QueryTargetId            string     `json:"query_target_id,omitempty" path:"query_target_id,omitempty" url:"query_target_id,omitempty"`
	QueryTargetName          string     `json:"query_target_name,omitempty" path:"query_target_name,omitempty" url:"query_target_name,omitempty"`
	QueryTargetPermission    string     `json:"query_target_permission,omitempty" path:"query_target_permission,omitempty" url:"query_target_permission,omitempty"`
	QueryTargetUserId        string     `json:"query_target_user_id,omitempty" path:"query_target_user_id,omitempty" url:"query_target_user_id,omitempty"`
	QueryTargetUsername      string     `json:"query_target_username,omitempty" path:"query_target_username,omitempty" url:"query_target_username,omitempty"`
	QueryTargetPlatform      string     `json:"query_target_platform,omitempty" path:"query_target_platform,omitempty" url:"query_target_platform,omitempty"`
	QueryTargetPermissionSet string     `json:"query_target_permission_set,omitempty" path:"query_target_permission_set,omitempty" url:"query_target_permission_set,omitempty"`
	ResultsUrl               string     `json:"results_url,omitempty" path:"results_url,omitempty" url:"results_url,omitempty"`
	UserId                   int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
}

func (h HistoryExport) Identifier() interface{} {
	return h.Id
}

type HistoryExportCollection []HistoryExport

type HistoryExportFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type HistoryExportCreateParams struct {
	UserId                   int64      `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	StartAt                  *time.Time `url:"start_at,omitempty" json:"start_at,omitempty" path:"start_at"`
	EndAt                    *time.Time `url:"end_at,omitempty" json:"end_at,omitempty" path:"end_at"`
	QueryAction              string     `url:"query_action,omitempty" json:"query_action,omitempty" path:"query_action"`
	QueryInterface           string     `url:"query_interface,omitempty" json:"query_interface,omitempty" path:"query_interface"`
	QueryUserId              string     `url:"query_user_id,omitempty" json:"query_user_id,omitempty" path:"query_user_id"`
	QueryFileId              string     `url:"query_file_id,omitempty" json:"query_file_id,omitempty" path:"query_file_id"`
	QueryParentId            string     `url:"query_parent_id,omitempty" json:"query_parent_id,omitempty" path:"query_parent_id"`
	QueryPath                string     `url:"query_path,omitempty" json:"query_path,omitempty" path:"query_path"`
	QueryFolder              string     `url:"query_folder,omitempty" json:"query_folder,omitempty" path:"query_folder"`
	QuerySrc                 string     `url:"query_src,omitempty" json:"query_src,omitempty" path:"query_src"`
	QueryDestination         string     `url:"query_destination,omitempty" json:"query_destination,omitempty" path:"query_destination"`
	QueryIp                  string     `url:"query_ip,omitempty" json:"query_ip,omitempty" path:"query_ip"`
	QueryUsername            string     `url:"query_username,omitempty" json:"query_username,omitempty" path:"query_username"`
	QueryFailureType         string     `url:"query_failure_type,omitempty" json:"query_failure_type,omitempty" path:"query_failure_type"`
	QueryTargetId            string     `url:"query_target_id,omitempty" json:"query_target_id,omitempty" path:"query_target_id"`
	QueryTargetName          string     `url:"query_target_name,omitempty" json:"query_target_name,omitempty" path:"query_target_name"`
	QueryTargetPermission    string     `url:"query_target_permission,omitempty" json:"query_target_permission,omitempty" path:"query_target_permission"`
	QueryTargetUserId        string     `url:"query_target_user_id,omitempty" json:"query_target_user_id,omitempty" path:"query_target_user_id"`
	QueryTargetUsername      string     `url:"query_target_username,omitempty" json:"query_target_username,omitempty" path:"query_target_username"`
	QueryTargetPlatform      string     `url:"query_target_platform,omitempty" json:"query_target_platform,omitempty" path:"query_target_platform"`
	QueryTargetPermissionSet string     `url:"query_target_permission_set,omitempty" json:"query_target_permission_set,omitempty" path:"query_target_permission_set"`
}

func (h *HistoryExport) UnmarshalJSON(data []byte) error {
	type historyExport HistoryExport
	var v historyExport
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*h = HistoryExport(v)
	return nil
}

func (h *HistoryExportCollection) UnmarshalJSON(data []byte) error {
	type historyExports HistoryExportCollection
	var v historyExports
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*h = HistoryExportCollection(v)
	return nil
}

func (h *HistoryExportCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*h))
	for i, v := range *h {
		ret[i] = v
	}

	return &ret
}
