package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type Preview struct {
	Id          int64  `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Status      string `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	DownloadUri string `json:"download_uri,omitempty" path:"download_uri,omitempty" url:"download_uri,omitempty"`
	Type        string `json:"type,omitempty" path:"type,omitempty" url:"type,omitempty"`
	Size        string `json:"size,omitempty" path:"size,omitempty" url:"size,omitempty"`
}

func (p Preview) Identifier() interface{} {
	return p.Id
}

type PreviewCollection []Preview

func (p *Preview) UnmarshalJSON(data []byte) error {
	type preview Preview
	var v preview
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = Preview(v)
	return nil
}

func (p *PreviewCollection) UnmarshalJSON(data []byte) error {
	type previews PreviewCollection
	var v previews
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PreviewCollection(v)
	return nil
}

func (p *PreviewCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
