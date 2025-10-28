package files_sdk

import (
	"encoding/json"
	"io"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Style struct {
	Id        int64     `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Path      string    `json:"path,omitempty" path:"path,omitempty" url:"path,omitempty"`
	Logo      Image     `json:"logo,omitempty" path:"logo,omitempty" url:"logo,omitempty"`
	Thumbnail Image     `json:"thumbnail,omitempty" path:"thumbnail,omitempty" url:"thumbnail,omitempty"`
	File      io.Reader `json:"file,omitempty" path:"file,omitempty" url:"file,omitempty"`
}

func (s Style) Identifier() interface{} {
	return s.Id
}

type StyleCollection []Style

type StyleFindParams struct {
	Path string `url:"-,omitempty" json:"-,omitempty" path:"path"`
}

type StyleUpdateParams struct {
	Path string    `url:"-,omitempty" json:"-,omitempty" path:"path"`
	File io.Writer `url:"file" json:"file" path:"file"`
}

type StyleDeleteParams struct {
	Path string `url:"-,omitempty" json:"-,omitempty" path:"path"`
}

func (s *Style) UnmarshalJSON(data []byte) error {
	type style Style
	var v style
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = Style(v)
	return nil
}

func (s *StyleCollection) UnmarshalJSON(data []byte) error {
	type styles StyleCollection
	var v styles
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = StyleCollection(v)
	return nil
}

func (s *StyleCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
