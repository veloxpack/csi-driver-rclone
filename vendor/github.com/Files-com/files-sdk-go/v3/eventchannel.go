package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type EventChannel struct {
	Id             int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Name           string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	WorkspaceId    int64      `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	Description    string     `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	Enabled        *bool      `json:"enabled,omitempty" path:"enabled,omitempty" url:"enabled,omitempty"`
	DefaultChannel *bool      `json:"default_channel,omitempty" path:"default_channel,omitempty" url:"default_channel,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (e EventChannel) Identifier() interface{} {
	return e.Id
}

type EventChannelCollection []EventChannel

type EventChannelListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type EventChannelFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type EventChannelCreateParams struct {
	Name           string `url:"name" json:"name" path:"name"`
	WorkspaceId    int64  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	Description    string `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Enabled        *bool  `url:"enabled,omitempty" json:"enabled,omitempty" path:"enabled"`
	DefaultChannel *bool  `url:"default_channel,omitempty" json:"default_channel,omitempty" path:"default_channel"`
}

type EventChannelUpdateParams struct {
	Id             int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Name           string `url:"name,omitempty" json:"name,omitempty" path:"name"`
	WorkspaceId    int64  `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	Description    string `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Enabled        *bool  `url:"enabled,omitempty" json:"enabled,omitempty" path:"enabled"`
	DefaultChannel *bool  `url:"default_channel,omitempty" json:"default_channel,omitempty" path:"default_channel"`
}

type EventChannelDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (e *EventChannel) UnmarshalJSON(data []byte) error {
	type eventChannel EventChannel
	var v eventChannel
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = EventChannel(v)
	return nil
}

func (e *EventChannelCollection) UnmarshalJSON(data []byte) error {
	type eventChannels EventChannelCollection
	var v eventChannels
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = EventChannelCollection(v)
	return nil
}

func (e *EventChannelCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
