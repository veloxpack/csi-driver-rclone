package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FileUploadPart struct {
	Send               map[string]interface{} `json:"send,omitempty" path:"send,omitempty" url:"send,omitempty"`
	Action             string                 `json:"action,omitempty" path:"action,omitempty" url:"action,omitempty"`
	AskAboutOverwrites *bool                  `json:"ask_about_overwrites,omitempty" path:"ask_about_overwrites,omitempty" url:"ask_about_overwrites,omitempty"`
	AvailableParts     int64                  `json:"available_parts,omitempty" path:"available_parts,omitempty" url:"available_parts,omitempty"`
	Expires            string                 `json:"expires,omitempty" path:"expires,omitempty" url:"expires,omitempty"`
	Headers            map[string]interface{} `json:"headers,omitempty" path:"headers,omitempty" url:"headers,omitempty"`
	HttpMethod         string                 `json:"http_method,omitempty" path:"http_method,omitempty" url:"http_method,omitempty"`
	NextPartsize       int64                  `json:"next_partsize,omitempty" path:"next_partsize,omitempty" url:"next_partsize,omitempty"`
	ParallelParts      *bool                  `json:"parallel_parts,omitempty" path:"parallel_parts,omitempty" url:"parallel_parts,omitempty"`
	RetryParts         *bool                  `json:"retry_parts,omitempty" path:"retry_parts,omitempty" url:"retry_parts,omitempty"`
	Parameters         map[string]interface{} `json:"parameters,omitempty" path:"parameters,omitempty" url:"parameters,omitempty"`
	PartNumber         int64                  `json:"part_number,omitempty" path:"part_number,omitempty" url:"part_number,omitempty"`
	Partsize           int64                  `json:"partsize,omitempty" path:"partsize,omitempty" url:"partsize,omitempty"`
	Path               string                 `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Ref                string                 `json:"ref,omitempty" path:"ref,omitempty" url:"ref,omitempty"`
	UploadUri          string                 `json:"upload_uri,omitempty" path:"upload_uri,omitempty" url:"upload_uri,omitempty"`
}

func (f FileUploadPart) Identifier() interface{} {
	return f.Path
}

type FileUploadPartCollection []FileUploadPart

const UploadPartExpires = time.Minute * 15
const UploadObjectExpires = time.Minute * (24 * 3)

func (f FileUploadPart) ExpiresTime() time.Time {
	partExpires, _ := time.Parse(time.RFC3339, f.Expires)
	return partExpires
}

// UploadExpires only valid on first part request
func (f FileUploadPart) UploadExpires() time.Time {
	if f.ExpiresTime().IsZero() {
		return time.Time{}
	}
	return f.ExpiresTime().Add(-(UploadPartExpires)).Add(UploadObjectExpires)
}
func (f *FileUploadPart) UnmarshalJSON(data []byte) error {
	type fileUploadPart FileUploadPart
	var v fileUploadPart
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FileUploadPart(v)
	return nil
}

func (f *FileUploadPartCollection) UnmarshalJSON(data []byte) error {
	type fileUploadParts FileUploadPartCollection
	var v fileUploadParts
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FileUploadPartCollection(v)
	return nil
}

func (f *FileUploadPartCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
