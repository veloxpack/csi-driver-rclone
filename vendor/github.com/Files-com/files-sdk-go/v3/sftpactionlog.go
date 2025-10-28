package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SftpActionLog struct {
	Timestamp               *time.Time `json:"timestamp,omitempty" path:"timestamp,omitempty" url:"timestamp,omitempty"`
	RemoteIp                string     `json:"remote_ip,omitempty" path:"remote_ip,omitempty" url:"remote_ip,omitempty"`
	ServerIp                string     `json:"server_ip,omitempty" path:"server_ip,omitempty" url:"server_ip,omitempty"`
	Username                string     `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	SshClientIdentification string     `json:"ssh_client_identification,omitempty" path:"ssh_client_identification,omitempty" url:"ssh_client_identification,omitempty"`
	SessionUuid             string     `json:"session_uuid,omitempty" path:"session_uuid,omitempty" url:"session_uuid,omitempty"`
	SeqId                   int64      `json:"seq_id,omitempty" path:"seq_id,omitempty" url:"seq_id,omitempty"`
	AuthMethod              string     `json:"auth_method,omitempty" path:"auth_method,omitempty" url:"auth_method,omitempty"`
	AuthCiphers             string     `json:"auth_ciphers,omitempty" path:"auth_ciphers,omitempty" url:"auth_ciphers,omitempty"`
	ActionType              string     `json:"action_type,omitempty" path:"action_type,omitempty" url:"action_type,omitempty"`
	Path                    string     `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	TruePath                string     `json:"true_path,omitempty" path:"true_path,omitempty" url:"true_path,omitempty"`
	Name                    string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	SftpResponseCode        string     `json:"sftp_response_code,omitempty" path:"sftp_response_code,omitempty" url:"sftp_response_code,omitempty"`
	SftpResponseMessage     string     `json:"sftp_response_message,omitempty" path:"sftp_response_message,omitempty" url:"sftp_response_message,omitempty"`
	Md5                     string     `json:"md5,omitempty" path:"md5,omitempty" url:"md5,omitempty"`
	Size                    int64      `json:"size,omitempty" path:"size,omitempty" url:"size,omitempty"`
	DataLength              int64      `json:"data_length,omitempty" path:"data_length,omitempty" url:"data_length,omitempty"`
	BytesTransferred        int64      `json:"bytes_transferred,omitempty" path:"bytes_transferred,omitempty" url:"bytes_transferred,omitempty"`
	EntriesReturned         int64      `json:"entries_returned,omitempty" path:"entries_returned,omitempty" url:"entries_returned,omitempty"`
	Success                 *bool      `json:"success,omitempty" path:"success,omitempty" url:"success,omitempty"`
	Status                  string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	DurationMs              int64      `json:"duration_ms,omitempty" path:"duration_ms,omitempty" url:"duration_ms,omitempty"`
	CreatedAt               *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
}

func (s SftpActionLog) Identifier() interface{} {
	return s.Path
}

type SftpActionLogCollection []SftpActionLog

type SftpActionLogListParams struct {
	Filter       SftpActionLog          `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (s *SftpActionLog) UnmarshalJSON(data []byte) error {
	type sftpActionLog SftpActionLog
	var v sftpActionLog
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SftpActionLog(v)
	return nil
}

func (s *SftpActionLogCollection) UnmarshalJSON(data []byte) error {
	type sftpActionLogs SftpActionLogCollection
	var v sftpActionLogs
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SftpActionLogCollection(v)
	return nil
}

func (s *SftpActionLogCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
