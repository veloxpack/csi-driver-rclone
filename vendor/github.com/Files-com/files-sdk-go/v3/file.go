package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type File struct {
	Path                               string                 `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	CreatedById                        int64                  `json:"created_by_id,omitempty" path:"created_by_id,omitempty" url:"created_by_id,omitempty"`
	CreatedByApiKeyId                  int64                  `json:"created_by_api_key_id,omitempty" path:"created_by_api_key_id,omitempty" url:"created_by_api_key_id,omitempty"`
	CreatedByAs2IncomingMessageId      int64                  `json:"created_by_as2_incoming_message_id,omitempty" path:"created_by_as2_incoming_message_id,omitempty" url:"created_by_as2_incoming_message_id,omitempty"`
	CreatedByAutomationId              int64                  `json:"created_by_automation_id,omitempty" path:"created_by_automation_id,omitempty" url:"created_by_automation_id,omitempty"`
	CreatedByBundleRegistrationId      int64                  `json:"created_by_bundle_registration_id,omitempty" path:"created_by_bundle_registration_id,omitempty" url:"created_by_bundle_registration_id,omitempty"`
	CreatedByInboxId                   int64                  `json:"created_by_inbox_id,omitempty" path:"created_by_inbox_id,omitempty" url:"created_by_inbox_id,omitempty"`
	CreatedByRemoteServerId            int64                  `json:"created_by_remote_server_id,omitempty" path:"created_by_remote_server_id,omitempty" url:"created_by_remote_server_id,omitempty"`
	CreatedByRemoteServerSyncId        int64                  `json:"created_by_remote_server_sync_id,omitempty" path:"created_by_remote_server_sync_id,omitempty" url:"created_by_remote_server_sync_id,omitempty"`
	CustomMetadata                     map[string]interface{} `json:"custom_metadata,omitempty" path:"custom_metadata,omitempty" url:"custom_metadata,omitempty"`
	DisplayName                        string                 `json:"display_name,omitempty" path:"display_name,omitempty" url:"display_name,omitempty"`
	Type                               string                 `json:"type,omitempty" path:"type,omitempty" url:"type,omitempty"`
	Size                               int64                  `json:"size,omitempty" path:"size,omitempty" url:"size,omitempty"`
	CreatedAt                          *time.Time             `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	LastModifiedById                   int64                  `json:"last_modified_by_id,omitempty" path:"last_modified_by_id,omitempty" url:"last_modified_by_id,omitempty"`
	LastModifiedByApiKeyId             int64                  `json:"last_modified_by_api_key_id,omitempty" path:"last_modified_by_api_key_id,omitempty" url:"last_modified_by_api_key_id,omitempty"`
	LastModifiedByAutomationId         int64                  `json:"last_modified_by_automation_id,omitempty" path:"last_modified_by_automation_id,omitempty" url:"last_modified_by_automation_id,omitempty"`
	LastModifiedByBundleRegistrationId int64                  `json:"last_modified_by_bundle_registration_id,omitempty" path:"last_modified_by_bundle_registration_id,omitempty" url:"last_modified_by_bundle_registration_id,omitempty"`
	LastModifiedByRemoteServerId       int64                  `json:"last_modified_by_remote_server_id,omitempty" path:"last_modified_by_remote_server_id,omitempty" url:"last_modified_by_remote_server_id,omitempty"`
	LastModifiedByRemoteServerSyncId   int64                  `json:"last_modified_by_remote_server_sync_id,omitempty" path:"last_modified_by_remote_server_sync_id,omitempty" url:"last_modified_by_remote_server_sync_id,omitempty"`
	Mtime                              *time.Time             `json:"mtime,omitempty" path:"mtime,omitempty" url:"mtime,omitempty"`
	ProvidedMtime                      *time.Time             `json:"provided_mtime,omitempty" path:"provided_mtime,omitempty" url:"provided_mtime,omitempty"`
	Crc32                              string                 `json:"crc32,omitempty" path:"crc32,omitempty" url:"crc32,omitempty"`
	Md5                                string                 `json:"md5,omitempty" path:"md5,omitempty" url:"md5,omitempty"`
	Sha1                               string                 `json:"sha1,omitempty" path:"sha1,omitempty" url:"sha1,omitempty"`
	Sha256                             string                 `json:"sha256,omitempty" path:"sha256,omitempty" url:"sha256,omitempty"`
	MimeType                           string                 `json:"mime_type,omitempty" path:"mime_type,omitempty" url:"mime_type,omitempty"`
	Region                             string                 `json:"region,omitempty" path:"region,omitempty" url:"region,omitempty"`
	Permissions                        string                 `json:"permissions,omitempty" path:"permissions,omitempty" url:"permissions,omitempty"`
	SubfoldersLocked                   *bool                  `json:"subfolders_locked?,omitempty" path:"subfolders_locked?,omitempty" url:"subfolders_locked?,omitempty"`
	IsLocked                           *bool                  `json:"is_locked,omitempty" path:"is_locked,omitempty" url:"is_locked,omitempty"`
	DownloadUri                        string                 `json:"download_uri,omitempty" path:"download_uri,omitempty" url:"download_uri,omitempty"`
	PriorityColor                      string                 `json:"priority_color,omitempty" path:"priority_color,omitempty" url:"priority_color,omitempty"`
	PreviewId                          int64                  `json:"preview_id,omitempty" path:"preview_id,omitempty" url:"preview_id,omitempty"`
	Preview                            Preview                `json:"preview,omitempty" path:"preview,omitempty" url:"preview,omitempty"`
	Action                             string                 `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	Length                             int64                  `json:"length,omitempty" path:"length,omitempty" url:"length,omitempty"`
	MkdirParents                       *bool                  `json:"mkdir_parents,omitempty" path:"mkdir_parents,omitempty" url:"mkdir_parents,omitempty"`
	Part                               int64                  `json:"part,omitempty" path:"part,omitempty" url:"part,omitempty"`
	Parts                              int64                  `json:"parts,omitempty" path:"parts,omitempty" url:"parts,omitempty"`
	Ref                                string                 `json:"ref,omitempty" path:"ref,omitempty" url:"ref,omitempty"`
	Restart                            int64                  `json:"restart,omitempty" path:"restart,omitempty" url:"restart,omitempty"`
	Structure                          string                 `json:"structure,omitempty" path:"structure,omitempty" url:"structure,omitempty"`
	WithRename                         *bool                  `json:"with_rename,omitempty" path:"with_rename,omitempty" url:"with_rename,omitempty"`
	BufferedUpload                     *bool                  `json:"buffered_upload,omitempty" path:"buffered_upload,omitempty" url:"buffered_upload,omitempty"`
}

func (f File) Identifier() interface{} {
	return f.Path
}

type FileCollection []File

type EtagsParam struct {
	Etag string `url:"etag,omitempty" json:"etag,omitempty" path:"etag"`
	Part string `url:"part,omitempty" json:"part,omitempty" path:"part"`
}

// Download File
type FileDownloadParams struct {
	Path              string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	Action            string `url:"action,omitempty" json:"action,omitempty" path:"action"`
	PreviewSize       string `url:"preview_size,omitempty" json:"preview_size,omitempty" path:"preview_size"`
	WithPreviews      *bool  `url:"with_previews,omitempty" json:"with_previews,omitempty" path:"with_previews"`
	WithPriorityColor *bool  `url:"with_priority_color,omitempty" json:"with_priority_color,omitempty" path:"with_priority_color"`
	File              File   `url:"-,omitempty" required:"false" json:"-,omitempty"`
}

type FileCreateParams struct {
	Path             string         `url:"-,omitempty" json:"-,omitempty" path:"path"`
	Action           string         `url:"action,omitempty" json:"action,omitempty" path:"action"`
	EtagsParam       []EtagsParam   `url:"etags,omitempty" json:"etags,omitempty" path:"etags"`
	Length           int64          `url:"length,omitempty" json:"length,omitempty" path:"length"`
	MkdirParents     *bool          `url:"mkdir_parents,omitempty" json:"mkdir_parents,omitempty" path:"mkdir_parents"`
	Part             int64          `url:"part,omitempty" json:"part,omitempty" path:"part"`
	Parts            int64          `url:"parts,omitempty" json:"parts,omitempty" path:"parts"`
	ProvidedMtime    *time.Time     `url:"provided_mtime,omitempty" json:"provided_mtime,omitempty" path:"provided_mtime"`
	Ref              string         `url:"ref,omitempty" json:"ref,omitempty" path:"ref"`
	Restart          int64          `url:"restart,omitempty" json:"restart,omitempty" path:"restart"`
	Size             int64          `url:"size,omitempty" json:"size,omitempty" path:"size"`
	Structure        string         `url:"structure,omitempty" json:"structure,omitempty" path:"structure"`
	WithRename       *bool          `url:"with_rename,omitempty" json:"with_rename,omitempty" path:"with_rename"`
	BufferedUpload   *bool          `url:"buffered_upload,omitempty" json:"buffered_upload,omitempty" path:"buffered_upload"`
	ActionAttributes map[string]any `url:"action_attributes,omitempty" json:"action_attributes,omitempty" path:"action_attributes"`
}

type FileUpdateParams struct {
	Path           string                 `url:"-,omitempty" json:"-,omitempty" path:"path"`
	CustomMetadata map[string]interface{} `url:"custom_metadata,omitempty" json:"custom_metadata,omitempty" path:"custom_metadata"`
	ProvidedMtime  *time.Time             `url:"provided_mtime,omitempty" json:"provided_mtime,omitempty" path:"provided_mtime"`
	PriorityColor  string                 `url:"priority_color,omitempty" json:"priority_color,omitempty" path:"priority_color"`
}

type FileDeleteParams struct {
	Path      string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	Recursive *bool  `url:"recursive,omitempty" json:"recursive,omitempty" path:"recursive"`
}

type FileFindParams struct {
	Path              string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	PreviewSize       string `url:"preview_size,omitempty" json:"preview_size,omitempty" path:"preview_size"`
	WithPreviews      *bool  `url:"with_previews,omitempty" json:"with_previews,omitempty" path:"with_previews"`
	WithPriorityColor *bool  `url:"with_priority_color,omitempty" json:"with_priority_color,omitempty" path:"with_priority_color"`
}

// Copy File/Folder
type FileCopyParams struct {
	Path        string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	Destination string `url:"destination" json:"destination" path:"destination"`
	Structure   *bool  `url:"structure,omitempty" json:"structure,omitempty" path:"structure"`
	Overwrite   *bool  `url:"overwrite,omitempty" json:"overwrite,omitempty" path:"overwrite"`
}

// Move File/Folder
type FileMoveParams struct {
	Path        string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	Destination string `url:"destination" json:"destination" path:"destination"`
	Overwrite   *bool  `url:"overwrite,omitempty" json:"overwrite,omitempty" path:"overwrite"`
}

// Begin File Upload
type FileBeginUploadParams struct {
	Path         string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	MkdirParents *bool  `url:"mkdir_parents,omitempty" json:"mkdir_parents,omitempty" path:"mkdir_parents"`
	Part         int64  `url:"part,omitempty" json:"part,omitempty" path:"part"`
	Parts        int64  `url:"parts,omitempty" json:"parts,omitempty" path:"parts"`
	Ref          string `url:"ref,omitempty" json:"ref,omitempty" path:"ref"`
	Restart      int64  `url:"restart,omitempty" json:"restart,omitempty" path:"restart"`
	Size         int64  `url:"size,omitempty" json:"size,omitempty" path:"size"`
	WithRename   *bool  `url:"with_rename,omitempty" json:"with_rename,omitempty" path:"with_rename"`
}

func (f File) ToFolder() (Folder, error) {
	bodyBytes, err := json.Marshal(f)
	if err != nil {
		return Folder{}, err
	}
	folder := Folder{}
	folder.UnmarshalJSON(bodyBytes)
	return folder, nil
}

func (f File) String() string {
	return f.Path
}

func (f File) Iterable() bool {
	return f.IsDir()
}

func (f File) IsDir() bool {
	return f.Type == "directory"
}

func (f File) ModTime() time.Time {
	if f.ProvidedMtime != nil {
		return *f.ProvidedMtime
	}
	if f.Mtime != nil {
		return *f.Mtime
	}
	return time.Time{}
}

func (f File) CreationTime() time.Time {
	if f.CreatedAt != nil {
		return *f.CreatedAt
	}
	return time.Time{}
}
func (f *File) UnmarshalJSON(data []byte) error {
	type file File
	var v file
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = File(v)
	return nil
}

func (f *FileCollection) UnmarshalJSON(data []byte) error {
	type files FileCollection
	var v files
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FileCollection(v)
	return nil
}

func (f *FileCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
