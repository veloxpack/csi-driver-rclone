package files_sdk

import (
	"encoding/json"
	"io"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Bundle struct {
	Code                            string                   `json:"code,omitempty" path:"code,omitempty" url:"code,omitempty"`
	ColorLeft                       string                   `json:"color_left,omitempty" path:"color_left,omitempty" url:"color_left,omitempty"`
	ColorLink                       string                   `json:"color_link,omitempty" path:"color_link,omitempty" url:"color_link,omitempty"`
	ColorText                       string                   `json:"color_text,omitempty" path:"color_text,omitempty" url:"color_text,omitempty"`
	ColorTop                        string                   `json:"color_top,omitempty" path:"color_top,omitempty" url:"color_top,omitempty"`
	ColorTopText                    string                   `json:"color_top_text,omitempty" path:"color_top_text,omitempty" url:"color_top_text,omitempty"`
	Url                             string                   `json:"url,omitempty" path:"url,omitempty" url:"url,omitempty"`
	Description                     string                   `json:"description,omitempty" path:"description,omitempty" url:"description,omitempty"`
	ExpiresAt                       *time.Time               `json:"expires_at,omitempty" path:"expires_at,omitempty" url:"expires_at,omitempty"`
	PasswordProtected               *bool                    `json:"password_protected,omitempty" path:"password_protected,omitempty" url:"password_protected,omitempty"`
	Permissions                     string                   `json:"permissions,omitempty" path:"permissions,omitempty" url:"permissions,omitempty"`
	PreviewOnly                     *bool                    `json:"preview_only,omitempty" path:"preview_only,omitempty" url:"preview_only,omitempty"`
	RequireRegistration             *bool                    `json:"require_registration,omitempty" path:"require_registration,omitempty" url:"require_registration,omitempty"`
	RequireShareRecipient           *bool                    `json:"require_share_recipient,omitempty" path:"require_share_recipient,omitempty" url:"require_share_recipient,omitempty"`
	RequireLogout                   *bool                    `json:"require_logout,omitempty" path:"require_logout,omitempty" url:"require_logout,omitempty"`
	ClickwrapBody                   string                   `json:"clickwrap_body,omitempty" path:"clickwrap_body,omitempty" url:"clickwrap_body,omitempty"`
	FormFieldSet                    FormFieldSet             `json:"form_field_set,omitempty" path:"form_field_set,omitempty" url:"form_field_set,omitempty"`
	SkipName                        *bool                    `json:"skip_name,omitempty" path:"skip_name,omitempty" url:"skip_name,omitempty"`
	SkipEmail                       *bool                    `json:"skip_email,omitempty" path:"skip_email,omitempty" url:"skip_email,omitempty"`
	StartAccessOnDate               *time.Time               `json:"start_access_on_date,omitempty" path:"start_access_on_date,omitempty" url:"start_access_on_date,omitempty"`
	SkipCompany                     *bool                    `json:"skip_company,omitempty" path:"skip_company,omitempty" url:"skip_company,omitempty"`
	Id                              int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	CreatedAt                       *time.Time               `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	DontSeparateSubmissionsByFolder *bool                    `json:"dont_separate_submissions_by_folder,omitempty" path:"dont_separate_submissions_by_folder,omitempty" url:"dont_separate_submissions_by_folder,omitempty"`
	MaxUses                         int64                    `json:"max_uses,omitempty" path:"max_uses,omitempty" url:"max_uses,omitempty"`
	Note                            string                   `json:"note,omitempty" path:"note,omitempty" url:"note,omitempty"`
	PathTemplate                    string                   `json:"path_template,omitempty" path:"path_template,omitempty" url:"path_template,omitempty"`
	PathTemplateTimeZone            string                   `json:"path_template_time_zone,omitempty" path:"path_template_time_zone,omitempty" url:"path_template_time_zone,omitempty"`
	SendEmailReceiptToUploader      *bool                    `json:"send_email_receipt_to_uploader,omitempty" path:"send_email_receipt_to_uploader,omitempty" url:"send_email_receipt_to_uploader,omitempty"`
	SnapshotId                      int64                    `json:"snapshot_id,omitempty" path:"snapshot_id,omitempty" url:"snapshot_id,omitempty"`
	UserId                          int64                    `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	Username                        string                   `json:"username,omitempty" path:"username,omitempty" url:"username,omitempty"`
	ClickwrapId                     int64                    `json:"clickwrap_id,omitempty" path:"clickwrap_id,omitempty" url:"clickwrap_id,omitempty"`
	InboxId                         int64                    `json:"inbox_id,omitempty" path:"inbox_id,omitempty" url:"inbox_id,omitempty"`
	WatermarkAttachment             Image                    `json:"watermark_attachment,omitempty" path:"watermark_attachment,omitempty" url:"watermark_attachment,omitempty"`
	WatermarkValue                  map[string]interface{}   `json:"watermark_value,omitempty" path:"watermark_value,omitempty" url:"watermark_value,omitempty"`
	HasInbox                        *bool                    `json:"has_inbox,omitempty" path:"has_inbox,omitempty" url:"has_inbox,omitempty"`
	DontAllowFoldersInUploads       *bool                    `json:"dont_allow_folders_in_uploads,omitempty" path:"dont_allow_folders_in_uploads,omitempty" url:"dont_allow_folders_in_uploads,omitempty"`
	Paths                           []string                 `json:"paths,omitempty" path:"paths,omitempty" url:"paths,omitempty"`
	Bundlepaths                     []map[string]interface{} `json:"bundlepaths,omitempty" path:"bundlepaths,omitempty" url:"bundlepaths,omitempty"`
	Password                        string                   `json:"password,omitempty" path:"password,omitempty" url:"password,omitempty"`
	FormFieldSetId                  int64                    `json:"form_field_set_id,omitempty" path:"form_field_set_id,omitempty" url:"form_field_set_id,omitempty"`
	CreateSnapshot                  *bool                    `json:"create_snapshot,omitempty" path:"create_snapshot,omitempty" url:"create_snapshot,omitempty"`
	FinalizeSnapshot                *bool                    `json:"finalize_snapshot,omitempty" path:"finalize_snapshot,omitempty" url:"finalize_snapshot,omitempty"`
	WatermarkAttachmentFile         io.Reader                `json:"watermark_attachment_file,omitempty" path:"watermark_attachment_file,omitempty" url:"watermark_attachment_file,omitempty"`
	WatermarkAttachmentDelete       *bool                    `json:"watermark_attachment_delete,omitempty" path:"watermark_attachment_delete,omitempty" url:"watermark_attachment_delete,omitempty"`
}

func (b Bundle) Identifier() interface{} {
	return b.Id
}

type BundleCollection []Bundle

type BundlePermissionsEnum string

func (u BundlePermissionsEnum) String() string {
	return string(u)
}

func (u BundlePermissionsEnum) Enum() map[string]BundlePermissionsEnum {
	return map[string]BundlePermissionsEnum{
		"read":         BundlePermissionsEnum("read"),
		"write":        BundlePermissionsEnum("write"),
		"read_write":   BundlePermissionsEnum("read_write"),
		"full":         BundlePermissionsEnum("full"),
		"none":         BundlePermissionsEnum("none"),
		"preview_only": BundlePermissionsEnum("preview_only"),
	}
}

type BundleListParams struct {
	UserId       int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy       map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter       Bundle                 `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	FilterGt     map[string]interface{} `url:"filter_gt,omitempty" json:"filter_gt,omitempty" path:"filter_gt"`
	FilterGteq   map[string]interface{} `url:"filter_gteq,omitempty" json:"filter_gteq,omitempty" path:"filter_gteq"`
	FilterPrefix map[string]interface{} `url:"filter_prefix,omitempty" json:"filter_prefix,omitempty" path:"filter_prefix"`
	FilterLt     map[string]interface{} `url:"filter_lt,omitempty" json:"filter_lt,omitempty" path:"filter_lt"`
	FilterLteq   map[string]interface{} `url:"filter_lteq,omitempty" json:"filter_lteq,omitempty" path:"filter_lteq"`
	ListParams
}

type BundleFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type BundleCreateParams struct {
	UserId                          int64                 `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	Paths                           []string              `url:"paths" json:"paths" path:"paths"`
	Password                        string                `url:"password,omitempty" json:"password,omitempty" path:"password"`
	FormFieldSetId                  int64                 `url:"form_field_set_id,omitempty" json:"form_field_set_id,omitempty" path:"form_field_set_id"`
	CreateSnapshot                  *bool                 `url:"create_snapshot,omitempty" json:"create_snapshot,omitempty" path:"create_snapshot"`
	DontSeparateSubmissionsByFolder *bool                 `url:"dont_separate_submissions_by_folder,omitempty" json:"dont_separate_submissions_by_folder,omitempty" path:"dont_separate_submissions_by_folder"`
	ExpiresAt                       *time.Time            `url:"expires_at,omitempty" json:"expires_at,omitempty" path:"expires_at"`
	FinalizeSnapshot                *bool                 `url:"finalize_snapshot,omitempty" json:"finalize_snapshot,omitempty" path:"finalize_snapshot"`
	MaxUses                         int64                 `url:"max_uses,omitempty" json:"max_uses,omitempty" path:"max_uses"`
	Description                     string                `url:"description,omitempty" json:"description,omitempty" path:"description"`
	Note                            string                `url:"note,omitempty" json:"note,omitempty" path:"note"`
	Code                            string                `url:"code,omitempty" json:"code,omitempty" path:"code"`
	PathTemplate                    string                `url:"path_template,omitempty" json:"path_template,omitempty" path:"path_template"`
	PathTemplateTimeZone            string                `url:"path_template_time_zone,omitempty" json:"path_template_time_zone,omitempty" path:"path_template_time_zone"`
	Permissions                     BundlePermissionsEnum `url:"permissions,omitempty" json:"permissions,omitempty" path:"permissions"`
	RequireRegistration             *bool                 `url:"require_registration,omitempty" json:"require_registration,omitempty" path:"require_registration"`
	ClickwrapId                     int64                 `url:"clickwrap_id,omitempty" json:"clickwrap_id,omitempty" path:"clickwrap_id"`
	InboxId                         int64                 `url:"inbox_id,omitempty" json:"inbox_id,omitempty" path:"inbox_id"`
	RequireShareRecipient           *bool                 `url:"require_share_recipient,omitempty" json:"require_share_recipient,omitempty" path:"require_share_recipient"`
	SendEmailReceiptToUploader      *bool                 `url:"send_email_receipt_to_uploader,omitempty" json:"send_email_receipt_to_uploader,omitempty" path:"send_email_receipt_to_uploader"`
	SkipEmail                       *bool                 `url:"skip_email,omitempty" json:"skip_email,omitempty" path:"skip_email"`
	SkipName                        *bool                 `url:"skip_name,omitempty" json:"skip_name,omitempty" path:"skip_name"`
	SkipCompany                     *bool                 `url:"skip_company,omitempty" json:"skip_company,omitempty" path:"skip_company"`
	StartAccessOnDate               *time.Time            `url:"start_access_on_date,omitempty" json:"start_access_on_date,omitempty" path:"start_access_on_date"`
	SnapshotId                      int64                 `url:"snapshot_id,omitempty" json:"snapshot_id,omitempty" path:"snapshot_id"`
	WatermarkAttachmentFile         io.Writer             `url:"watermark_attachment_file,omitempty" json:"watermark_attachment_file,omitempty" path:"watermark_attachment_file"`
}

// Send email(s) with a link to bundle
type BundleShareParams struct {
	Id         int64                    `url:"-,omitempty" json:"-,omitempty" path:"id"`
	To         []string                 `url:"to,omitempty" json:"to,omitempty" path:"to"`
	Note       string                   `url:"note,omitempty" json:"note,omitempty" path:"note"`
	Recipients []map[string]interface{} `url:"recipients,omitempty" json:"recipients,omitempty" path:"recipients"`
}

type BundleUpdateParams struct {
	Id                              int64                 `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Paths                           []string              `url:"paths,omitempty" json:"paths,omitempty" path:"paths"`
	Password                        string                `url:"password,omitempty" json:"password,omitempty" path:"password"`
	FormFieldSetId                  int64                 `url:"form_field_set_id,omitempty" json:"form_field_set_id,omitempty" path:"form_field_set_id"`
	ClickwrapId                     int64                 `url:"clickwrap_id,omitempty" json:"clickwrap_id,omitempty" path:"clickwrap_id"`
	Code                            string                `url:"code,omitempty" json:"code,omitempty" path:"code"`
	CreateSnapshot                  *bool                 `url:"create_snapshot,omitempty" json:"create_snapshot,omitempty" path:"create_snapshot"`
	Description                     string                `url:"description,omitempty" json:"description,omitempty" path:"description"`
	DontSeparateSubmissionsByFolder *bool                 `url:"dont_separate_submissions_by_folder,omitempty" json:"dont_separate_submissions_by_folder,omitempty" path:"dont_separate_submissions_by_folder"`
	ExpiresAt                       *time.Time            `url:"expires_at,omitempty" json:"expires_at,omitempty" path:"expires_at"`
	FinalizeSnapshot                *bool                 `url:"finalize_snapshot,omitempty" json:"finalize_snapshot,omitempty" path:"finalize_snapshot"`
	InboxId                         int64                 `url:"inbox_id,omitempty" json:"inbox_id,omitempty" path:"inbox_id"`
	MaxUses                         int64                 `url:"max_uses,omitempty" json:"max_uses,omitempty" path:"max_uses"`
	Note                            string                `url:"note,omitempty" json:"note,omitempty" path:"note"`
	PathTemplate                    string                `url:"path_template,omitempty" json:"path_template,omitempty" path:"path_template"`
	PathTemplateTimeZone            string                `url:"path_template_time_zone,omitempty" json:"path_template_time_zone,omitempty" path:"path_template_time_zone"`
	Permissions                     BundlePermissionsEnum `url:"permissions,omitempty" json:"permissions,omitempty" path:"permissions"`
	RequireRegistration             *bool                 `url:"require_registration,omitempty" json:"require_registration,omitempty" path:"require_registration"`
	RequireShareRecipient           *bool                 `url:"require_share_recipient,omitempty" json:"require_share_recipient,omitempty" path:"require_share_recipient"`
	SendEmailReceiptToUploader      *bool                 `url:"send_email_receipt_to_uploader,omitempty" json:"send_email_receipt_to_uploader,omitempty" path:"send_email_receipt_to_uploader"`
	SkipCompany                     *bool                 `url:"skip_company,omitempty" json:"skip_company,omitempty" path:"skip_company"`
	StartAccessOnDate               *time.Time            `url:"start_access_on_date,omitempty" json:"start_access_on_date,omitempty" path:"start_access_on_date"`
	SkipEmail                       *bool                 `url:"skip_email,omitempty" json:"skip_email,omitempty" path:"skip_email"`
	SkipName                        *bool                 `url:"skip_name,omitempty" json:"skip_name,omitempty" path:"skip_name"`
	WatermarkAttachmentDelete       *bool                 `url:"watermark_attachment_delete,omitempty" json:"watermark_attachment_delete,omitempty" path:"watermark_attachment_delete"`
	WatermarkAttachmentFile         io.Writer             `url:"watermark_attachment_file,omitempty" json:"watermark_attachment_file,omitempty" path:"watermark_attachment_file"`
}

type BundleDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (b *Bundle) UnmarshalJSON(data []byte) error {
	type bundle Bundle
	var v bundle
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*b = Bundle(v)
	return nil
}

func (b *BundleCollection) UnmarshalJSON(data []byte) error {
	type bundles BundleCollection
	var v bundles
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*b = BundleCollection(v)
	return nil
}

func (b *BundleCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*b))
	for i, v := range *b {
		ret[i] = v
	}

	return &ret
}
