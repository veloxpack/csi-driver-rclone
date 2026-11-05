package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type SiteSubdomainRedirect struct {
	Id        int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	Subdomain string     `json:"subdomain,omitempty" path:"subdomain,omitempty" url:"subdomain,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (s SiteSubdomainRedirect) Identifier() interface{} {
	return s.Id
}

type SiteSubdomainRedirectCollection []SiteSubdomainRedirect

type SiteSubdomainRedirectListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type SiteSubdomainRedirectFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type SiteSubdomainRedirectDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (s *SiteSubdomainRedirect) UnmarshalJSON(data []byte) error {
	type siteSubdomainRedirect SiteSubdomainRedirect
	var v siteSubdomainRedirect
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*s = SiteSubdomainRedirect(v)
	return nil
}

func (s *SiteSubdomainRedirectCollection) UnmarshalJSON(data []byte) error {
	type siteSubdomainRedirects SiteSubdomainRedirectCollection
	var v siteSubdomainRedirects
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*s = SiteSubdomainRedirectCollection(v)
	return nil
}

func (s *SiteSubdomainRedirectCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*s))
	for i, v := range *s {
		ret[i] = v
	}

	return &ret
}
