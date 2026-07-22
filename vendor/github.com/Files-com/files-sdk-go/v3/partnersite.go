package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PartnerSite struct {
	PartnerId      int64  `json:"partner_id,omitempty" path:"partner_id,omitempty" url:"partner_id,omitempty"`
	PartnerName    string `json:"partner_name,omitempty" path:"partner_name,omitempty" url:"partner_name,omitempty"`
	LinkedSiteId   int64  `json:"linked_site_id,omitempty" path:"linked_site_id,omitempty" url:"linked_site_id,omitempty"`
	LinkedSiteName string `json:"linked_site_name,omitempty" path:"linked_site_name,omitempty" url:"linked_site_name,omitempty"`
	MainSiteId     int64  `json:"main_site_id,omitempty" path:"main_site_id,omitempty" url:"main_site_id,omitempty"`
	MainSiteName   string `json:"main_site_name,omitempty" path:"main_site_name,omitempty" url:"main_site_name,omitempty"`
}

// Identifier no path or id

type PartnerSiteCollection []PartnerSite

type PartnerSiteListParams struct {
	ListParams
}

func (p *PartnerSite) UnmarshalJSON(data []byte) error {
	type partnerSite PartnerSite
	var v partnerSite
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PartnerSite(v)
	return nil
}

func (p *PartnerSiteCollection) UnmarshalJSON(data []byte) error {
	type partnerSites PartnerSiteCollection
	var v partnerSites
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PartnerSiteCollection(v)
	return nil
}

func (p *PartnerSiteCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
