package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3userMasterKeysRequest represents the request structure for the user's master keys endpoint.
type v3userMasterKeysRequest struct {
	MasterKey crypto.EncryptedString `json:"masterKeys"`
}

// V3UserMasterKeysResponse represents the response structure from the user's master keys endpoint.
type V3UserMasterKeysResponse struct {
	Keys crypto.EncryptedString `json:"keys"`
}

// PostV3UserMasterKeys calls /v3/user/masterKeys to retrieve the user's encrypted master keys.
// It requires an authenticated client with a valid API key.
func (c *Client) PostV3UserMasterKeys(ctx context.Context, encryptedMasterKey crypto.EncryptedString) (*V3UserMasterKeysResponse, error) {
	userMasterKeys := &V3UserMasterKeysResponse{}
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/user/masterKeys"), v3userMasterKeysRequest{
		MasterKey: encryptedMasterKey,
	}, userMasterKeys)
	return userMasterKeys, err
}
