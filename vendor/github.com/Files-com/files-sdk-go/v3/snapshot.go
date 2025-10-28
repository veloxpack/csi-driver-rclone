package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Snapshot struct {
	Id          int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" path:"expires_at,omitempty" url:"expires_at,omitempty"`
	FinalizedAt *time.Time `json:"finalized_at,omitempty" path:"finalized_at,omitempty" url:"finalized_at,omitempty"`
	Name        string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	UserId      int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	BundleId    int64      `json:"bundle_id,omitempty" path:"bundle_id,omitempty" url:"bundle_id,omitempty"`
	Paths       []string   `json:"paths,omitempty" path:"paths,omitempty" url:"paths,omitempty"`
}

func (s Snapshot) Identifier() interface{} {
	return s.Id
}

type SnapshotCollection []Snapshot

type SnapshotListParams struct {
	ListParams
}

type SnapshotFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SnapshotCreateParams struct {
	ExpiresAt *time.Time `url:"expires_at,omitempty" json:"expires_at,omitempty" path:"expires_at"`
	Name      string     `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Paths     []string   `url:"paths,omitempty" json:"paths,omitempty" path:"paths"`
}

// Finalize Snapshot
type SnapshotFinalizeParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SnapshotUpdateParams struct {
	Id        int64      `url:"-,omitempty" json:"-,omitempty" path:"id"`
	ExpiresAt *time.Time `url:"expires_at,omitempty" json:"expires_at,omitempty" path:"expires_at"`
	Name      string     `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Paths     []string   `url:"paths,omitempty" json:"paths,omitempty" path:"paths"`
}

type SnapshotDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *Snapshot) UnmarshalJSON(data []byte) error {
	type snapshot Snapshot
	var v snapshot
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Snapshot(v)
	return nil
}

func (s *SnapshotCollection) UnmarshalJSON(data []byte) error {
	type snapshots SnapshotCollection
	var v snapshots
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SnapshotCollection(v)
	return nil
}

func (s *SnapshotCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
