package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type WebDavActionLog struct {
	Timestamp        *time.Time `json:"timestamp,omitempty" path:"timestamp,omitempty" url:"timestamp,omitempty"`
	RemoteIp         string     `json:"remote_ip,omitempty" path:"remote_ip,omitempty" url:"remote_ip,omitempty"`
	ServerIp         string     `json:"server_ip,omitempty" path:"server_ip,omitempty" url:"server_ip,omitempty"`
	Username         string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	AuthCiphers      string     `json:"auth_ciphers,omitempty" path:"auth_ciphers,omitempty" url:"auth_ciphers,omitempty"`
	ActionType       string     `json:"action_type,omitempty" path:"action_type,omitempty" url:"action_type,omitempty"`
	Path             string     `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	TruePath         string     `json:"true_path,omitempty" path:"true_path,omitempty" url:"true_path,omitempty"`
	Name             string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	HttpMethod       string     `json:"http_method,omitempty" path:"http_method,omitempty" url:"http_method,omitempty"`
	HttpPath         string     `json:"http_path,omitempty" path:"http_path,omitempty" url:"http_path,omitempty"`
	HttpResponseCode int64      `json:"http_response_code,omitempty" path:"http_response_code,omitempty" url:"http_response_code,omitempty"`
	Size             int64      `json:"size,omitempty" path:"size,omitempty" url:"size,omitempty"`
	EntriesReturned  int64      `json:"entries_returned,omitempty" path:"entries_returned,omitempty" url:"entries_returned,omitempty"`
	Success          *bool      `json:"success,omitempty" path:"success,omitempty" url:"success,omitempty"`
	Status           string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	DurationMs       int64      `json:"duration_ms,omitempty" path:"duration_ms,omitempty" url:"duration_ms,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

func (w WebDavActionLog) Identifier() interface{} {
	return w.Path
}

type WebDavActionLogCollection []WebDavActionLog

type WebDavActionLogListParams struct {
	Filter       WebDavActionLog        `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (w *WebDavActionLog) UnmarshalJSON(data []byte) error {
	type webDavActionLog WebDavActionLog
	var v webDavActionLog
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*w = WebDavActionLog(v)
	return nil
}

func (w *WebDavActionLogCollection) UnmarshalJSON(data []byte) error {
	type webDavActionLogs WebDavActionLogCollection
	var v webDavActionLogs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*w = WebDavActionLogCollection(v)
	return nil
}

func (w *WebDavActionLogCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*w))
	for i, v := range *w {
		ret[i] = v
	}

	return &ret
}
