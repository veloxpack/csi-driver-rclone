package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Notification struct {
	Id                       int64    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Path                     string   `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	GroupId                  int64    `json:"group_id,omitempty" path:"group_id,omitempty" url:"group_id,omitempty"`
	GroupName                string   `json:"group_name,omitempty" path:"group_name,omitempty" url:"group_name,omitempty"`
	TriggeringGroupIds       []int64  `json:"triggering_group_ids,omitempty" path:"triggering_group_ids,omitempty" url:"triggering_group_ids,omitempty"`
	TriggeringUserIds        []int64  `json:"triggering_user_ids,omitempty" path:"triggering_user_ids,omitempty" url:"triggering_user_ids,omitempty"`
	TriggerByShareRecipients *bool    `json:"trigger_by_share_recipients,omitempty" path:"trigger_by_share_recipients,omitempty" url:"trigger_by_share_recipients,omitempty"`
	NotifyUserActions        *bool    `json:"notify_user_actions,omitempty" path:"notify_user_actions,omitempty" url:"notify_user_actions,omitempty"`
	NotifyOnCopy             *bool    `json:"notify_on_copy,omitempty" path:"notify_on_copy,omitempty" url:"notify_on_copy,omitempty"`
	NotifyOnDelete           *bool    `json:"notify_on_delete,omitempty" path:"notify_on_delete,omitempty" url:"notify_on_delete,omitempty"`
	NotifyOnDownload         *bool    `json:"notify_on_download,omitempty" path:"notify_on_download,omitempty" url:"notify_on_download,omitempty"`
	NotifyOnMove             *bool    `json:"notify_on_move,omitempty" path:"notify_on_move,omitempty" url:"notify_on_move,omitempty"`
	NotifyOnUpload           *bool    `json:"notify_on_upload,omitempty" path:"notify_on_upload,omitempty" url:"notify_on_upload,omitempty"`
	Recursive                *bool    `json:"recursive,omitempty" path:"recursive,omitempty" url:"recursive,omitempty"`
	SendInterval             string   `json:"send_interval,omitempty" path:"send_interval,omitempty" url:"send_interval,omitempty"`
	Message                  string   `json:"message,omitempty" path:"message,omitempty" url:"message,omitempty"`
	TriggeringFilenames      []string `json:"triggering_filenames,omitempty" path:"triggering_filenames,omitempty" url:"triggering_filenames,omitempty"`
	Unsubscribed             *bool    `json:"unsubscribed,omitempty" path:"unsubscribed,omitempty" url:"unsubscribed,omitempty"`
	UnsubscribedReason       string   `json:"unsubscribed_reason,omitempty" path:"unsubscribed_reason,omitempty" url:"unsubscribed_reason,omitempty"`
	UserId                   int64    `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username                 string   `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	SuppressedEmail          *bool    `json:"suppressed_email,omitempty" path:"suppressed_email,omitempty" url:"suppressed_email,omitempty"`
}

func (n Notification) Identifier() interface{} {
	return n.Id
}

type NotificationCollection []Notification

type NotificationListParams struct {
	SortBy           map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter           Notification           `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterPrefix     map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	Path             string                 `url:"path,omitempty" json:"path,omitempty" path:"path"`
	IncludeAncestors *bool                  `url:"include_ancestors,omitempty" json:"include_ancestors,omitempty" path:"include_ancestors"`
	GroupId          string                 `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	ListParams
}

type NotificationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type NotificationCreateParams struct {
	UserId                   int64    `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	NotifyOnCopy             *bool    `url:"notify_on_copy,omitempty" json:"notify_on_copy,omitempty" path:"notify_on_copy"`
	NotifyOnDelete           *bool    `url:"notify_on_delete,omitempty" json:"notify_on_delete,omitempty" path:"notify_on_delete"`
	NotifyOnDownload         *bool    `url:"notify_on_download,omitempty" json:"notify_on_download,omitempty" path:"notify_on_download"`
	NotifyOnMove             *bool    `url:"notify_on_move,omitempty" json:"notify_on_move,omitempty" path:"notify_on_move"`
	NotifyOnUpload           *bool    `url:"notify_on_upload,omitempty" json:"notify_on_upload,omitempty" path:"notify_on_upload"`
	NotifyUserActions        *bool    `url:"notify_user_actions,omitempty" json:"notify_user_actions,omitempty" path:"notify_user_actions"`
	Recursive                *bool    `url:"recursive,omitempty" json:"recursive,omitempty" path:"recursive"`
	SendInterval             string   `url:"send_interval,omitempty" json:"send_interval,omitempty" path:"send_interval"`
	Message                  string   `url:"message,omitempty" json:"message,omitempty" path:"message"`
	TriggeringFilenames      []string `url:"triggering_filenames,omitempty" json:"triggering_filenames,omitempty" path:"triggering_filenames"`
	TriggeringGroupIds       []int64  `url:"triggering_group_ids,omitempty" json:"triggering_group_ids,omitempty" path:"triggering_group_ids"`
	TriggeringUserIds        []int64  `url:"triggering_user_ids,omitempty" json:"triggering_user_ids,omitempty" path:"triggering_user_ids"`
	TriggerByShareRecipients *bool    `url:"trigger_by_share_recipients,omitempty" json:"trigger_by_share_recipients,omitempty" path:"trigger_by_share_recipients"`
	GroupId                  int64    `url:"group_id,omitempty" json:"group_id,omitempty" path:"group_id"`
	Path                     string   `url:"path,omitempty" json:"path,omitempty" path:"path"`
	Username                 string   `url:"username,omitempty" json:"username,omitempty" path:"username"`
}

type NotificationUpdateParams struct {
	Id                       int64    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	NotifyOnCopy             *bool    `url:"notify_on_copy,omitempty" json:"notify_on_copy,omitempty" path:"notify_on_copy"`
	NotifyOnDelete           *bool    `url:"notify_on_delete,omitempty" json:"notify_on_delete,omitempty" path:"notify_on_delete"`
	NotifyOnDownload         *bool    `url:"notify_on_download,omitempty" json:"notify_on_download,omitempty" path:"notify_on_download"`
	NotifyOnMove             *bool    `url:"notify_on_move,omitempty" json:"notify_on_move,omitempty" path:"notify_on_move"`
	NotifyOnUpload           *bool    `url:"notify_on_upload,omitempty" json:"notify_on_upload,omitempty" path:"notify_on_upload"`
	NotifyUserActions        *bool    `url:"notify_user_actions,omitempty" json:"notify_user_actions,omitempty" path:"notify_user_actions"`
	Recursive                *bool    `url:"recursive,omitempty" json:"recursive,omitempty" path:"recursive"`
	SendInterval             string   `url:"send_interval,omitempty" json:"send_interval,omitempty" path:"send_interval"`
	Message                  string   `url:"message,omitempty" json:"message,omitempty" path:"message"`
	TriggeringFilenames      []string `url:"triggering_filenames,omitempty" json:"triggering_filenames,omitempty" path:"triggering_filenames"`
	TriggeringGroupIds       []int64  `url:"triggering_group_ids,omitempty" json:"triggering_group_ids,omitempty" path:"triggering_group_ids"`
	TriggeringUserIds        []int64  `url:"triggering_user_ids,omitempty" json:"triggering_user_ids,omitempty" path:"triggering_user_ids"`
	TriggerByShareRecipients *bool    `url:"trigger_by_share_recipients,omitempty" json:"trigger_by_share_recipients,omitempty" path:"trigger_by_share_recipients"`
}

type NotificationDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (n *Notification) UnmarshalJSON(data []byte) error {
	type notification Notification
	var v notification
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*n = Notification(v)
	return nil
}

func (n *NotificationCollection) UnmarshalJSON(data []byte) error {
	type notifications NotificationCollection
	var v notifications
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*n = NotificationCollection(v)
	return nil
}

func (n *NotificationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*n))
	for i, v := range *n {
		ret[i] = v
	}

	return &ret
}
