package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type BundleNotification struct {
	BundleId             int64 `json:"bundle_id,omitempty" path:"bundle_id,omitempty" url:"bundle_id,omitempty"`
	Id                   int64 `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	NotifyOnRegistration *bool `json:"notify_on_registration,omitempty" path:"notify_on_registration,omitempty" url:"notify_on_registration,omitempty"`
	NotifyOnUpload       *bool `json:"notify_on_upload,omitempty" path:"notify_on_upload,omitempty" url:"notify_on_upload,omitempty"`
	NotifyUserId         int64 `json:"notify_user_id,omitempty" path:"notify_user_id,omitempty" url:"notify_user_id,omitempty"`
	UserId               int64 `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
}

func (b BundleNotification) Identifier() interface{} {
	return b.Id
}

type BundleNotificationCollection []BundleNotification

type BundleNotificationListParams struct {
	UserId int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter BundleNotification     `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type BundleNotificationFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type BundleNotificationCreateParams struct {
	UserId               int64 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	BundleId             int64 `url:"bundle_id" json:"bundle_id" path:"bundle_id"`
	NotifyUserId         int64 `url:"notify_user_id,omitempty" json:"notify_user_id,omitempty" path:"notify_user_id"`
	NotifyOnRegistration *bool `url:"notify_on_registration,omitempty" json:"notify_on_registration,omitempty" path:"notify_on_registration"`
	NotifyOnUpload       *bool `url:"notify_on_upload,omitempty" json:"notify_on_upload,omitempty" path:"notify_on_upload"`
}

type BundleNotificationUpdateParams struct {
	Id                   int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
	NotifyOnRegistration *bool `url:"notify_on_registration,omitempty" json:"notify_on_registration,omitempty" path:"notify_on_registration"`
	NotifyOnUpload       *bool `url:"notify_on_upload,omitempty" json:"notify_on_upload,omitempty" path:"notify_on_upload"`
}

type BundleNotificationDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (b *BundleNotification) UnmarshalJSON(data []byte) error {
	type bundleNotification BundleNotification
	var v bundleNotification
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*b = BundleNotification(v)
	return nil
}

func (b *BundleNotificationCollection) UnmarshalJSON(data []byte) error {
	type bundleNotifications BundleNotificationCollection
	var v bundleNotifications
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*b = BundleNotificationCollection(v)
	return nil
}

func (b *BundleNotificationCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*b))
	for i, v := range *b {
		ret[i] = v
	}

	return &ret
}
