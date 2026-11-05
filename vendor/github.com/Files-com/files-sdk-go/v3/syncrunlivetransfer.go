package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SyncRunLiveTransfer struct {
	Path        string  `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Status      string  `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	BytesCopied int64   `json:"bytes_copied,omitempty" path:"bytes_copied,omitempty" url:"bytes_copied,omitempty"`
	BytesTotal  int64   `json:"bytes_total,omitempty" path:"bytes_total,omitempty" url:"bytes_total,omitempty"`
	Percentage  float64 `json:"percentage,omitempty" path:"percentage,omitempty" url:"percentage,omitempty"`
	Eta         string  `json:"eta,omitempty" path:"eta,omitempty" url:"eta,omitempty"`
	StartedAt   string  `json:"started_at,omitempty" path:"started_at,omitempty" url:"started_at,omitempty"`
}

func (s SyncRunLiveTransfer) Identifier() interface{} {
	return s.Path
}

type SyncRunLiveTransferCollection []SyncRunLiveTransfer

func (s *SyncRunLiveTransfer) UnmarshalJSON(data []byte) error {
	type syncRunLiveTransfer SyncRunLiveTransfer
	var v syncRunLiveTransfer
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SyncRunLiveTransfer(v)
	return nil
}

func (s *SyncRunLiveTransferCollection) UnmarshalJSON(data []byte) error {
	type syncRunLiveTransfers SyncRunLiveTransferCollection
	var v syncRunLiveTransfers
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SyncRunLiveTransferCollection(v)
	return nil
}

func (s *SyncRunLiveTransferCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
