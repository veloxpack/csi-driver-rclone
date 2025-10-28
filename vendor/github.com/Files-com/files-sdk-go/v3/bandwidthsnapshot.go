package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type BandwidthSnapshot struct {
	Id                int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	BytesReceived     string     `json:"bytes_received,omitempty" path:"bytes_received,omitempty" url:"bytes_received,omitempty"`
	BytesSent         string     `json:"bytes_sent,omitempty" path:"bytes_sent,omitempty" url:"bytes_sent,omitempty"`
	SyncBytesReceived string     `json:"sync_bytes_received,omitempty" path:"sync_bytes_received,omitempty" url:"sync_bytes_received,omitempty"`
	SyncBytesSent     string     `json:"sync_bytes_sent,omitempty" path:"sync_bytes_sent,omitempty" url:"sync_bytes_sent,omitempty"`
	RequestsGet       string     `json:"requests_get,omitempty" path:"requests_get,omitempty" url:"requests_get,omitempty"`
	RequestsPut       string     `json:"requests_put,omitempty" path:"requests_put,omitempty" url:"requests_put,omitempty"`
	RequestsOther     string     `json:"requests_other,omitempty" path:"requests_other,omitempty" url:"requests_other,omitempty"`
	LoggedAt          *time.Time `json:"logged_at,omitempty" path:"logged_at,omitempty" url:"logged_at,omitempty"`
}

func (b BandwidthSnapshot) Identifier() interface{} {
	return b.Id
}

type BandwidthSnapshotCollection []BandwidthSnapshot

type BandwidthSnapshotListParams struct {
	SortBy     map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter     BandwidthSnapshot      `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt   map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterLt   map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

func (b *BandwidthSnapshot) UnmarshalJSON(data []byte) error {
	type bandwidthSnapshot BandwidthSnapshot
	var v bandwidthSnapshot
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*b = BandwidthSnapshot(v)
	return nil
}

func (b *BandwidthSnapshotCollection) UnmarshalJSON(data []byte) error {
	type bandwidthSnapshots BandwidthSnapshotCollection
	var v bandwidthSnapshots
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*b = BandwidthSnapshotCollection(v)
	return nil
}

func (b *BandwidthSnapshotCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*b))
	for i, v := range *b {
		ret[i] = v
	}

	return &ret
}
