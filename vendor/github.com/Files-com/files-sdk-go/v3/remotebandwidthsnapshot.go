package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type RemoteBandwidthSnapshot struct {
	Id                int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	SyncBytesReceived string     `json:"sync_bytes_received,omitempty" path:"sync_bytes_received,omitempty" url:"sync_bytes_received,omitempty"`
	SyncBytesSent     string     `json:"sync_bytes_sent,omitempty" path:"sync_bytes_sent,omitempty" url:"sync_bytes_sent,omitempty"`
	LoggedAt          *time.Time `json:"logged_at,omitempty" path:"logged_at,omitempty" url:"logged_at,omitempty"`
	RemoteServerId    int64      `json:"remote_server_id,omitempty" path:"remote_server_id,omitempty" url:"remote_server_id,omitempty"`
}

func (r RemoteBandwidthSnapshot) Identifier() interface{} {
	return r.Id
}

type RemoteBandwidthSnapshotCollection []RemoteBandwidthSnapshot

type RemoteBandwidthSnapshotListParams struct {
	SortBy     map[string]interface{}  `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     RemoteBandwidthSnapshot `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   map[string]interface{}  `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq map[string]interface{}  `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   map[string]interface{}  `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq map[string]interface{}  `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (r *RemoteBandwidthSnapshot) UnmarshalJSON(data []byte) error {
	type remoteBandwidthSnapshot RemoteBandwidthSnapshot
	var v remoteBandwidthSnapshot
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*r = RemoteBandwidthSnapshot(v)
	return nil
}

func (r *RemoteBandwidthSnapshotCollection) UnmarshalJSON(data []byte) error {
	type remoteBandwidthSnapshots RemoteBandwidthSnapshotCollection
	var v remoteBandwidthSnapshots
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*r = RemoteBandwidthSnapshotCollection(v)
	return nil
}

func (r *RemoteBandwidthSnapshotCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*r))
	for i, v := range *r {
		ret[i] = v
	}

	return &ret
}
