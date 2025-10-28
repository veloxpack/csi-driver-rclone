package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type ApiRequestLog struct {
	Timestamp            *time.Time `json:"timestamp,omitempty" path:"timestamp,omitempty" url:"timestamp,omitempty"`
	ApiKeyId             int64      `json:"api_key_id,omitempty" path:"api_key_id,omitempty" url:"api_key_id,omitempty"`
	ApiKeyPrefix         string     `json:"api_key_prefix,omitempty" path:"api_key_prefix,omitempty" url:"api_key_prefix,omitempty"`
	UserId               int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username             string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	UserIsFromParentSite *bool      `json:"user_is_from_parent_site,omitempty" path:"user_is_from_parent_site,omitempty" url:"user_is_from_parent_site,omitempty"`
	Interface            string     `json:"interface,omitempty" path:"interface,omitempty" url:"interface,omitempty"`
	RequestMethod        string     `json:"request_method,omitempty" path:"request_method,omitempty" url:"request_method,omitempty"`
	RequestPath          string     `json:"request_path,omitempty" path:"request_path,omitempty" url:"request_path,omitempty"`
	RequestIp            string     `json:"request_ip,omitempty" path:"request_ip,omitempty" url:"request_ip,omitempty"`
	RequestHost          string     `json:"request_host,omitempty" path:"request_host,omitempty" url:"request_host,omitempty"`
	RequestId            string     `json:"request_id,omitempty" path:"request_id,omitempty" url:"request_id,omitempty"`
	ApiName              string     `json:"api_name,omitempty" path:"api_name,omitempty" url:"api_name,omitempty"`
	UserAgent            string     `json:"user_agent,omitempty" path:"user_agent,omitempty" url:"user_agent,omitempty"`
	ErrorType            string     `json:"error_type,omitempty" path:"error_type,omitempty" url:"error_type,omitempty"`
	ErrorMessage         string     `json:"error_message,omitempty" path:"error_message,omitempty" url:"error_message,omitempty"`
	ResponseCode         int64      `json:"response_code,omitempty" path:"response_code,omitempty" url:"response_code,omitempty"`
	Success              *bool      `json:"success,omitempty" path:"success,omitempty" url:"success,omitempty"`
	DurationMs           int64      `json:"duration_ms,omitempty" path:"duration_ms,omitempty" url:"duration_ms,omitempty"`
	ImpersonatorUserId   int64      `json:"impersonator_user_id,omitempty" path:"impersonator_user_id,omitempty" url:"impersonator_user_id,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

// Identifier no path or id

type ApiRequestLogCollection []ApiRequestLog

type ApiRequestLogListParams struct {
	Filter       ApiRequestLog          `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (a *ApiRequestLog) UnmarshalJSON(data []byte) error {
	type apiRequestLog ApiRequestLog
	var v apiRequestLog
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*a = ApiRequestLog(v)
	return nil
}

func (a *ApiRequestLogCollection) UnmarshalJSON(data []byte) error {
	type apiRequestLogs ApiRequestLogCollection
	var v apiRequestLogs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*a = ApiRequestLogCollection(v)
	return nil
}

func (a *ApiRequestLogCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*a))
	for i, v := range *a {
		ret[i] = v
	}

	return &ret
}
