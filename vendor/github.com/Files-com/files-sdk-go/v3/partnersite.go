package files_sdk

import (
	"encoding/json"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PartnerSite struct {
}

// Identifier no path or id

type PartnerSiteCollection []PartnerSite

type PartnerSiteDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
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
