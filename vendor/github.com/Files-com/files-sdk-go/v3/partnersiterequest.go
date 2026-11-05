package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type PartnerSiteRequest struct {
	Id            int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	HostPartnerId int64      `json:"host_partner_id,omitempty" path:"host_partner_id,omitempty" url:"host_partner_id,omitempty"`
	GuestSiteUrl  string     `json:"guest_site_url,omitempty" path:"guest_site_url,omitempty" url:"guest_site_url,omitempty"`
	Status        string     `json:"status,omitempty" path:"status,omitempty" url:"status,omitempty"`
	HostSiteName  string     `json:"host_site_name,omitempty" path:"host_site_name,omitempty" url:"host_site_name,omitempty"`
	PairingKey    string     `json:"pairing_key,omitempty" path:"pairing_key,omitempty" url:"pairing_key,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty" path:"created_at,omitempty" url:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty" path:"updated_at,omitempty" url:"updated_at,omitempty"`
}

func (p PartnerSiteRequest) Identifier() interface{} {
	return p.Id
}

type PartnerSiteRequestCollection []PartnerSiteRequest

type PartnerSiteRequestListParams struct {
	SortBy interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	Filter interface{} `url:"filter,omitempty" json:"filter,omitempty" path:"filter"`
	ListParams
}

type PartnerSiteRequestFindByPairingKeyParams struct {
	PairingKey string `url:"pairing_key" json:"pairing_key" path:"pairing_key"`
}

type PartnerSiteRequestCreateParams struct {
	HostPartnerId int64  `url:"host_partner_id" json:"host_partner_id" path:"host_partner_id"`
	GuestSiteUrl  string `url:"guest_site_url" json:"guest_site_url" path:"guest_site_url"`
}

type PartnerSiteRequestRejectParams struct {
	PairingKey string `url:"pairing_key" json:"pairing_key" path:"pairing_key"`
}

type PartnerSiteRequestApproveParams struct {
	PairingKey string `url:"pairing_key" json:"pairing_key" path:"pairing_key"`
}

type PartnerSiteRequestDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (p *PartnerSiteRequest) UnmarshalJSON(data []byte) error {
	type partnerSiteRequest PartnerSiteRequest
	var v partnerSiteRequest
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*p = PartnerSiteRequest(v)
	return nil
}

func (p *PartnerSiteRequestCollection) UnmarshalJSON(data []byte) error {
	type partnerSiteRequests PartnerSiteRequestCollection
	var v partnerSiteRequests
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*p = PartnerSiteRequestCollection(v)
	return nil
}

func (p *PartnerSiteRequestCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*p))
	for i, v := range *p {
		ret[i] = v
	}

	return &ret
}
