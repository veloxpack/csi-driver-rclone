package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type RemoteMountBackend struct {
	CanaryFilePath        string `json:"canary_file_path,omitempty" path:"canary_file_path,omitempty" url:"canary_file_path,omitempty"`
	Enabled               *bool  `json:"enabled,omitempty" path:"enabled,omitempty" url:"enabled,omitempty"`
	Fall                  int64  `json:"fall,omitempty" path:"fall,omitempty" url:"fall,omitempty"`
	HealthCheckEnabled    *bool  `json:"health_check_enabled,omitempty" path:"health_check_enabled,omitempty" url:"health_check_enabled,omitempty"`
	HealthCheckType       string `json:"health_check_type,omitempty" path:"health_check_type,omitempty" url:"health_check_type,omitempty"`
	Id                    int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Interval              int64  `json:"interval,omitempty" path:"interval,omitempty" url:"interval,omitempty"`
	MinFreeCpu            string `json:"min_free_cpu,omitempty" path:"min_free_cpu,omitempty" url:"min_free_cpu,omitempty"`
	MinFreeMem            string `json:"min_free_mem,omitempty" path:"min_free_mem,omitempty" url:"min_free_mem,omitempty"`
	Priority              int64  `json:"priority,omitempty" path:"priority,omitempty" url:"priority,omitempty"`
	RemotePath            string `json:"remote_path,omitempty" path:"remote_path,omitempty" url:"remote_path,omitempty"`
	RemoteServerId        int64  `json:"remote_server_id,omitempty" path:"remote_server_id,omitempty" url:"remote_server_id,omitempty"`
	RemoteServerMountId   int64  `json:"remote_server_mount_id,omitempty" path:"remote_server_mount_id,omitempty" url:"remote_server_mount_id,omitempty"`
	Rise                  int64  `json:"rise,omitempty" path:"rise,omitempty" url:"rise,omitempty"`
	Status                string `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	UndergoingMaintenance *bool  `json:"undergoing_maintenance,omitempty" path:"undergoing_maintenance,omitempty" url:"undergoing_maintenance,omitempty"`
}

func (r RemoteMountBackend) Identifier() interface{} {
	return r.Id
}

type RemoteMountBackendCollection []RemoteMountBackend

type RemoteMountBackendHealthCheckTypeEnum string

func (u RemoteMountBackendHealthCheckTypeEnum) String() string {
	return string(u)
}

func (u RemoteMountBackendHealthCheckTypeEnum) Enum() map[string]RemoteMountBackendHealthCheckTypeEnum {
	return map[string]RemoteMountBackendHealthCheckTypeEnum{
		"active":  RemoteMountBackendHealthCheckTypeEnum("active"),
		"passive": RemoteMountBackendHealthCheckTypeEnum("passive"),
	}
}

type RemoteMountBackendListParams struct {
	Filter RemoteMountBackend `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type RemoteMountBackendFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type RemoteMountBackendCreateParams struct {
	Enabled             *bool                                 `url:"enabled,omitempty" json:"enabled,omitempty" path:"enabled"`
	Fall                int64                                 `url:"fall,omitempty" json:"fall,omitempty" path:"fall"`
	HealthCheckEnabled  *bool                                 `url:"health_check_enabled,omitempty" json:"health_check_enabled,omitempty" path:"health_check_enabled"`
	HealthCheckType     RemoteMountBackendHealthCheckTypeEnum `url:"health_check_type,omitempty" json:"health_check_type,omitempty" path:"health_check_type"`
	Interval            int64                                 `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	MinFreeCpu          string                                `url:"min_free_cpu,omitempty" json:"min_free_cpu,omitempty" path:"min_free_cpu"`
	MinFreeMem          string                                `url:"min_free_mem,omitempty" json:"min_free_mem,omitempty" path:"min_free_mem"`
	Priority            int64                                 `url:"priority,omitempty" json:"priority,omitempty" path:"priority"`
	RemotePath          string                                `url:"remote_path,omitempty" json:"remote_path,omitempty" path:"remote_path"`
	Rise                int64                                 `url:"rise,omitempty" json:"rise,omitempty" path:"rise"`
	CanaryFilePath      string                                `url:"canary_file_path" json:"canary_file_path" path:"canary_file_path"`
	RemoteServerMountId int64                                 `url:"remote_server_mount_id" json:"remote_server_mount_id" path:"remote_server_mount_id"`
	RemoteServerId      int64                                 `url:"remote_server_id" json:"remote_server_id" path:"remote_server_id"`
}

// Reset backend status to healthy
type RemoteMountBackendResetStatusParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type RemoteMountBackendUpdateParams struct {
	Id                 int64                                 `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Enabled            *bool                                 `url:"enabled,omitempty" json:"enabled,omitempty" path:"enabled"`
	Fall               int64                                 `url:"fall,omitempty" json:"fall,omitempty" path:"fall"`
	HealthCheckEnabled *bool                                 `url:"health_check_enabled,omitempty" json:"health_check_enabled,omitempty" path:"health_check_enabled"`
	HealthCheckType    RemoteMountBackendHealthCheckTypeEnum `url:"health_check_type,omitempty" json:"health_check_type,omitempty" path:"health_check_type"`
	Interval           int64                                 `url:"interval,omitempty" json:"interval,omitempty" path:"interval"`
	MinFreeCpu         string                                `url:"min_free_cpu,omitempty" json:"min_free_cpu,omitempty" path:"min_free_cpu"`
	MinFreeMem         string                                `url:"min_free_mem,omitempty" json:"min_free_mem,omitempty" path:"min_free_mem"`
	Priority           int64                                 `url:"priority,omitempty" json:"priority,omitempty" path:"priority"`
	RemotePath         string                                `url:"remote_path,omitempty" json:"remote_path,omitempty" path:"remote_path"`
	Rise               int64                                 `url:"rise,omitempty" json:"rise,omitempty" path:"rise"`
	CanaryFilePath     string                                `url:"canary_file_path,omitempty" json:"canary_file_path,omitempty" path:"canary_file_path"`
	RemoteServerId     int64                                 `url:"remote_server_id,omitempty" json:"remote_server_id,omitempty" path:"remote_server_id"`
}

type RemoteMountBackendDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (r *RemoteMountBackend) UnmarshalJSON(data []byte) error {
	type remoteMountBackend RemoteMountBackend
	var v remoteMountBackend
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = RemoteMountBackend(v)
	return nil
}

func (r *RemoteMountBackendCollection) UnmarshalJSON(data []byte) error {
	type remoteMountBackends RemoteMountBackendCollection
	var v remoteMountBackends
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RemoteMountBackendCollection(v)
	return nil
}

func (r *RemoteMountBackendCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
