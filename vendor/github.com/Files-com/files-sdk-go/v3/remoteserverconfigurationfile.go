package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type RemoteServerConfigurationFile struct {
	Id                        int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	PermissionSet             string `json:"permission_set,omitempty" path:"permission_set,omitempty" url:"permission_set,omitempty"`
	PrivateKey                string `json:"private_key,omitempty" path:"private_key,omitempty" url:"private_key,omitempty"`
	Subdomain                 string `json:"subdomain,omitempty" path:"subdomain,omitempty" url:"subdomain,omitempty"`
	Root                      string `json:"root,omitempty" path:"root,omitempty" url:"root,omitempty"`
	FollowLinks               *bool  `json:"follow_links,omitempty" path:"follow_links,omitempty" url:"follow_links,omitempty"`
	PreferProtocol            string `json:"prefer_protocol,omitempty" path:"prefer_protocol,omitempty" url:"prefer_protocol,omitempty"`
	Dns                       string `json:"dns,omitempty" path:"dns,omitempty" url:"dns,omitempty"`
	ProxyAllOutbound          *bool  `json:"proxy_all_outbound,omitempty" path:"proxy_all_outbound,omitempty" url:"proxy_all_outbound,omitempty"`
	EndpointOverride          string `json:"endpoint_override,omitempty" path:"endpoint_override,omitempty" url:"endpoint_override,omitempty"`
	LogFile                   string `json:"log_file,omitempty" path:"log_file,omitempty" url:"log_file,omitempty"`
	LogLevel                  string `json:"log_level,omitempty" path:"log_level,omitempty" url:"log_level,omitempty"`
	LogRotateNum              int64  `json:"log_rotate_num,omitempty" path:"log_rotate_num,omitempty" url:"log_rotate_num,omitempty"`
	LogRotateSize             int64  `json:"log_rotate_size,omitempty" path:"log_rotate_size,omitempty" url:"log_rotate_size,omitempty"`
	OverrideMaxConcurrentJobs int64  `json:"override_max_concurrent_jobs,omitempty" path:"override_max_concurrent_jobs,omitempty" url:"override_max_concurrent_jobs,omitempty"`
	GracefulShutdownTimeout   int64  `json:"graceful_shutdown_timeout,omitempty" path:"graceful_shutdown_timeout,omitempty" url:"graceful_shutdown_timeout,omitempty"`
	TransferRateLimit         string `json:"transfer_rate_limit,omitempty" path:"transfer_rate_limit,omitempty" url:"transfer_rate_limit,omitempty"`
	AutoUpdatePolicy          string `json:"auto_update_policy,omitempty" path:"auto_update_policy,omitempty" url:"auto_update_policy,omitempty"`
	ApiToken                  string `json:"api_token,omitempty" path:"api_token,omitempty" url:"api_token,omitempty"`
	Port                      int64  `json:"port,omitempty" path:"port,omitempty" url:"port,omitempty"`
	Hostname                  string `json:"hostname,omitempty" path:"hostname,omitempty" url:"hostname,omitempty"`
	PublicKey                 string `json:"public_key,omitempty" path:"public_key,omitempty" url:"public_key,omitempty"`
	Status                    string `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	ServerHostKey             string `json:"server_host_key,omitempty" path:"server_host_key,omitempty" url:"server_host_key,omitempty"`
	ConfigVersion             string `json:"config_version,omitempty" path:"config_version,omitempty" url:"config_version,omitempty"`
}

func (r RemoteServerConfigurationFile) Identifier() interface{} {
	return r.Id
}

type RemoteServerConfigurationFileCollection []RemoteServerConfigurationFile

func (r *RemoteServerConfigurationFile) UnmarshalJSON(data []byte) error {
	type remoteServerConfigurationFile RemoteServerConfigurationFile
	var v remoteServerConfigurationFile
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = RemoteServerConfigurationFile(v)
	return nil
}

func (r *RemoteServerConfigurationFileCollection) UnmarshalJSON(data []byte) error {
	type remoteServerConfigurationFiles RemoteServerConfigurationFileCollection
	var v remoteServerConfigurationFiles
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RemoteServerConfigurationFileCollection(v)
	return nil
}

func (r *RemoteServerConfigurationFileCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
