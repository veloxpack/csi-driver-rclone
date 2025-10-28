package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PublicHostingRequestLog struct {
	Timestamp        *time.Time `json:"timestamp,omitempty" path:"timestamp,omitempty" url:"timestamp,omitempty"`
	RemoteIp         string     `json:"remote_ip,omitempty" path:"remote_ip,omitempty" url:"remote_ip,omitempty"`
	ServerIp         string     `json:"server_ip,omitempty" path:"server_ip,omitempty" url:"server_ip,omitempty"`
	Hostname         string     `json:"hostname,omitempty" path:"hostname,omitempty" url:"hostname,omitempty"`
	Path             string     `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	ResponseCode     int64      `json:"responseCode,omitempty" path:"responseCode,omitempty" url:"responseCode,omitempty"`
	Success          *bool      `json:"success,omitempty" path:"success,omitempty" url:"success,omitempty"`
	DurationMs       int64      `json:"duration_ms,omitempty" path:"duration_ms,omitempty" url:"duration_ms,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	BytesTransferred int64      `json:"bytes_transferred,omitempty" path:"bytes_transferred,omitempty" url:"bytes_transferred,omitempty"`
	HttpMethod       string     `json:"http_method,omitempty" path:"http_method,omitempty" url:"http_method,omitempty"`
}

func (p PublicHostingRequestLog) Identifier() interface{} {
	return p.Path
}

type PublicHostingRequestLogCollection []PublicHostingRequestLog

type PublicHostingRequestLogListParams struct {
	Filter       PublicHostingRequestLog `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{}  `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{}  `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix map[string]interface{}  `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     map[string]interface{}  `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{}  `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (p *PublicHostingRequestLog) UnmarshalJSON(data []byte) error {
	type publicHostingRequestLog PublicHostingRequestLog
	var v publicHostingRequestLog
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PublicHostingRequestLog(v)
	return nil
}

func (p *PublicHostingRequestLogCollection) UnmarshalJSON(data []byte) error {
	type publicHostingRequestLogs PublicHostingRequestLogCollection
	var v publicHostingRequestLogs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PublicHostingRequestLogCollection(v)
	return nil
}

func (p *PublicHostingRequestLogCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
