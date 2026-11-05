package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type EventSubscription struct {
	Id                   int64       `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	EventChannelId       int64       `json:"event_channel_id,omitempty" path:"event_channel_id,omitempty" url:"event_channel_id,omitempty"`
	WorkspaceId          int64       `json:"workspace_id,omitempty" path:"workspace_id,omitempty" url:"workspace_id,omitempty"`
	ApplyToAllWorkspaces *bool       `json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces,omitempty" url:"apply_to_all_workspaces,omitempty"`
	Name                 string      `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	Enabled              *bool       `json:"enabled,omitempty" path:"enabled,omitempty" url:"enabled,omitempty"`
	EventTypes           []string    `json:"event_types,omitempty" path:"event_types,omitempty" url:"event_types,omitempty"`
	Filter               interface{} `json:"filter,omitempty" path:"filter,omitempty" url:"filter,omitempty"`
	DeliveryPolicy       interface{} `json:"delivery_policy,omitempty" path:"delivery_policy,omitempty" url:"delivery_policy,omitempty"`
	EventTargetIds       []int64     `json:"event_target_ids,omitempty" path:"event_target_ids,omitempty" url:"event_target_ids,omitempty"`
	CreatedAt            *time.Time  `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt            *time.Time  `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (e EventSubscription) Identifier() interface{} {
	return e.Id
}

type EventSubscriptionCollection []EventSubscription

type EventSubscriptionListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type EventSubscriptionFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type EventSubscriptionCreateParams struct {
	EventChannelId       int64       `url:"event_channel_id,omitempty" json:"event_channel_id,omitempty" path:"event_channel_id"`
	WorkspaceId          int64       `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	ApplyToAllWorkspaces *bool       `url:"apply_to_all_workspaces,omitempty" json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces"`
	Name                 string      `url:"name" json:"name" path:"name"`
	Enabled              *bool       `url:"enabled,omitempty" json:"enabled,omitempty" path:"enabled"`
	EventTypes           []string    `url:"event_types,omitempty" json:"event_types,omitempty" path:"event_types"`
	Filter               interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	DeliveryPolicy       interface{} `url:"delivery_policy,omitempty" json:"delivery_policy,omitempty" path:"delivery_policy"`
	EventTargetIds       []int64     `url:"event_target_ids,omitempty" json:"event_target_ids,omitempty" path:"event_target_ids"`
}

type EventSubscriptionUpdateParams struct {
	Id                   int64       `url:"-,omitempty" json:"-,omitempty" path:"id"`
	EventChannelId       int64       `url:"event_channel_id,omitempty" json:"event_channel_id,omitempty" path:"event_channel_id"`
	WorkspaceId          int64       `url:"workspace_id,omitempty" json:"workspace_id,omitempty" path:"workspace_id"`
	ApplyToAllWorkspaces *bool       `url:"apply_to_all_workspaces,omitempty" json:"apply_to_all_workspaces,omitempty" path:"apply_to_all_workspaces"`
	Name                 string      `url:"name,omitempty" json:"name,omitempty" path:"name"`
	Enabled              *bool       `url:"enabled,omitempty" json:"enabled,omitempty" path:"enabled"`
	EventTypes           []string    `url:"event_types,omitempty" json:"event_types,omitempty" path:"event_types"`
	Filter               interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	DeliveryPolicy       interface{} `url:"delivery_policy,omitempty" json:"delivery_policy,omitempty" path:"delivery_policy"`
	EventTargetIds       []int64     `url:"event_target_ids,omitempty" json:"event_target_ids,omitempty" path:"event_target_ids"`
}

type EventSubscriptionDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (e *EventSubscription) UnmarshalJSON(data []byte) error {
	type eventSubscription EventSubscription
	var v eventSubscription
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*e = EventSubscription(v)
	return nil
}

func (e *EventSubscriptionCollection) UnmarshalJSON(data []byte) error {
	type eventSubscriptions EventSubscriptionCollection
	var v eventSubscriptions
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*e = EventSubscriptionCollection(v)
	return nil
}

func (e *EventSubscriptionCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*e))
	for i, v := range *e {
		ret[i] = v
	}

	return &ret
}
