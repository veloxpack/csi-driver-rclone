package files_sdk

import (
	"encoding/json"
	"time"

	lib "github.com/Files-com/files-sdk-go/v3/lib"
)

type GpgKey struct {
	Id                    int64      `json:"id,omitempty" path:"id,omitempty" url:"id,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty" path:"expires_at,omitempty" url:"expires_at,omitempty"`
	Name                  string     `json:"name,omitempty" path:"name,omitempty" url:"name,omitempty"`
	PartnerId             int64      `json:"partner_id,omitempty" path:"partner_id,omitempty" url:"partner_id,omitempty"`
	PartnerName           string     `json:"partner_name,omitempty" path:"partner_name,omitempty" url:"partner_name,omitempty"`
	UserId                int64      `json:"user_id,omitempty" path:"user_id,omitempty" url:"user_id,omitempty"`
	PublicKeyMd5          string     `json:"public_key_md5,omitempty" path:"public_key_md5,omitempty" url:"public_key_md5,omitempty"`
	PrivateKeyMd5         string     `json:"private_key_md5,omitempty" path:"private_key_md5,omitempty" url:"private_key_md5,omitempty"`
	GeneratedPublicKey    string     `json:"generated_public_key,omitempty" path:"generated_public_key,omitempty" url:"generated_public_key,omitempty"`
	GeneratedPrivateKey   string     `json:"generated_private_key,omitempty" path:"generated_private_key,omitempty" url:"generated_private_key,omitempty"`
	PrivateKeyPasswordMd5 string     `json:"private_key_password_md5,omitempty" path:"private_key_password_md5,omitempty" url:"private_key_password_md5,omitempty"`
	PublicKey             string     `json:"public_key,omitempty" path:"public_key,omitempty" url:"public_key,omitempty"`
	PrivateKey            string     `json:"private_key,omitempty" path:"private_key,omitempty" url:"private_key,omitempty"`
	PrivateKeyPassword    string     `json:"private_key_password,omitempty" path:"private_key_password,omitempty" url:"private_key_password,omitempty"`
	GenerateExpiresAt     string     `json:"generate_expires_at,omitempty" path:"generate_expires_at,omitempty" url:"generate_expires_at,omitempty"`
	GenerateKeypair       *bool      `json:"generate_keypair,omitempty" path:"generate_keypair,omitempty" url:"generate_keypair,omitempty"`
	GenerateFullName      string     `json:"generate_full_name,omitempty" path:"generate_full_name,omitempty" url:"generate_full_name,omitempty"`
	GenerateEmail         string     `json:"generate_email,omitempty" path:"generate_email,omitempty" url:"generate_email,omitempty"`
}

func (g GpgKey) Identifier() interface{} {
	return g.Id
}

type GpgKeyCollection []GpgKey

type GpgKeyListParams struct {
	UserId int64                  `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	SortBy map[string]interface{} `url:"sort_by,omitempty" json:"sort_by,omitempty" path:"sort_by"`
	ListParams
}

type GpgKeyFindParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

type GpgKeyCreateParams struct {
	UserId             int64      `url:"user_id,omitempty" json:"user_id,omitempty" path:"user_id"`
	PartnerId          int64      `url:"partner_id,omitempty" json:"partner_id,omitempty" path:"partner_id"`
	PublicKey          string     `url:"public_key,omitempty" json:"public_key,omitempty" path:"public_key"`
	PrivateKey         string     `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassword string     `url:"private_key_password,omitempty" json:"private_key_password,omitempty" path:"private_key_password"`
	Name               string     `url:"name" json:"name" path:"name"`
	GenerateExpiresAt  *time.Time `url:"generate_expires_at,omitempty" json:"generate_expires_at,omitempty" path:"generate_expires_at"`
	GenerateKeypair    *bool      `url:"generate_keypair,omitempty" json:"generate_keypair,omitempty" path:"generate_keypair"`
	GenerateFullName   string     `url:"generate_full_name,omitempty" json:"generate_full_name,omitempty" path:"generate_full_name"`
	GenerateEmail      string     `url:"generate_email,omitempty" json:"generate_email,omitempty" path:"generate_email"`
}

type GpgKeyUpdateParams struct {
	Id                 int64  `url:"-,omitempty" json:"-,omitempty" path:"id"`
	PartnerId          int64  `url:"partner_id,omitempty" json:"partner_id,omitempty" path:"partner_id"`
	PublicKey          string `url:"public_key,omitempty" json:"public_key,omitempty" path:"public_key"`
	PrivateKey         string `url:"private_key,omitempty" json:"private_key,omitempty" path:"private_key"`
	PrivateKeyPassword string `url:"private_key_password,omitempty" json:"private_key_password,omitempty" path:"private_key_password"`
	Name               string `url:"name,omitempty" json:"name,omitempty" path:"name"`
}

type GpgKeyDeleteParams struct {
	Id int64 `url:"-,omitempty" json:"-,omitempty" path:"id"`
}

func (g *GpgKey) UnmarshalJSON(data []byte) error {
	type gpgKey GpgKey
	var v gpgKey
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, map[string]interface{}{})
	}

	*g = GpgKey(v)
	return nil
}

func (g *GpgKeyCollection) UnmarshalJSON(data []byte) error {
	type gpgKeys GpgKeyCollection
	var v gpgKeys
	if err := json.Unmarshal(data, &v); err != nil {
		return lib.ErrorWithOriginalResponse{}.ProcessError(data, err, []map[string]interface{}{})
	}

	*g = GpgKeyCollection(v)
	return nil
}

func (g *GpgKeyCollection) ToSlice() *[]interface{} {
	ret := make([]interface{}, len(*g))
	for i, v := range *g {
		ret[i] = v
	}

	return &ret
}
