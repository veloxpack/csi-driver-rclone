package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type FileComment struct {
	Id        int64                    `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Body      string                   `json:"body,omitempty" path:"body,omitempty" url:"body,omitempty"`
	Reactions []map[string]interface{} `json:"reactions,omitempty" path:"reactions,omitempty" url:"reactions,omitempty"`
	Path      string                   `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
}

func (f FileComment) Identifier() interface{} {
	return f.Id
}

type FileCommentCollection []FileComment

type FileCommentListForParams struct {
	Path string `url:"-,omitempty" json:"-,omitempty" path:"path"`
	ListParams
}

type FileCommentCreateParams struct {
	Body string `url:"body" json:"body" path:"body"`
	Path string `url:"path" json:"path" path:"path"`
}

type FileCommentUpdateParams struct {
	Id   int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	Body string `url:"body" json:"body" path:"body"`
}

type FileCommentDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (f *FileComment) UnmarshalJSON(data []byte) error {
	type fileComment FileComment
	var v fileComment
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*f = FileComment(v)
	return nil
}

func (f *FileCommentCollection) UnmarshalJSON(data []byte) error {
	type fileComments FileCommentCollection
	var v fileComments
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*f = FileCommentCollection(v)
	return nil
}

func (f *FileCommentCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*f))
	for i, v := range *f {
		ret[i] = v
	}

	return &ret
}
