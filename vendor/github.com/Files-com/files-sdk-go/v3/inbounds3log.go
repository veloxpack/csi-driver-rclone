package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type InboundS3Log struct {
	Path         string     `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	ClientIp     string     `json:"client_ip,omitempty" path:"client_ip,omitempty" url:"client_ip,omitempty"`
	Operation    string     `json:"operation,omitempty" path:"operation,omitempty" url:"operation,omitempty"`
	Status       string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	AwsAccessKey string     `json:"aws_access_key,omitempty" path:"aws_access_key,omitempty" url:"aws_access_key,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty" path:"error_message,omitempty" url:"error_message,omitempty"`
	ErrorType    string     `json:"error_type,omitempty" path:"error_type,omitempty" url:"error_type,omitempty"`
	DurationMs   int64      `json:"duration_ms,omitempty" path:"duration_ms,omitempty" url:"duration_ms,omitempty"`
	RequestId    string     `json:"request_id,omitempty" path:"request_id,omitempty" url:"request_id,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty" path:"user_agent,omitempty" url:"user_agent,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

func (i InboundS3Log) Identifier() interface{} {
	return i.Path
}

type InboundS3LogCollection []InboundS3Log

type InboundS3LogListParams struct {
	Filter       interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (i *InboundS3Log) UnmarshalJSON(data []byte) error {
	type inboundS3Log InboundS3Log
	var v inboundS3Log
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*i = InboundS3Log(v)
	return nil
}

func (i *InboundS3LogCollection) UnmarshalJSON(data []byte) error {
	type inboundS3Logs InboundS3LogCollection
	var v inboundS3Logs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*i = InboundS3LogCollection(v)
	return nil
}

func (i *InboundS3LogCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*i))
	for i, v := range *i {
		ret[i] = v
	}

	return &ret
}
