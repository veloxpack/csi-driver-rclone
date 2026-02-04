package client

import (
	"context"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// V3UserKeyPairInfoResponse represents the response structure from the keypair info endpoint.
// It contains the user's encrypted private key and public key for asymmetric encryption.
type V3UserKeyPairInfoResponse struct {
	PrivateKey crypto.EncryptedString `json:"privateKey"`
	PublicKey  string                 `json:"publicKey"`
}

// GetV3UserKeyPairInfo calls /v3/user/keyPair/info to retrieve the user's encryption keypair.
// The private key is encrypted with the user's master key and must be decrypted locally.
func (c *Client) GetV3UserKeyPairInfo(ctx context.Context) (*V3UserKeyPairInfoResponse, error) {
	response := &V3UserKeyPairInfoResponse{}
	_, err := c.RequestData(ctx, "GET", GatewayURL("/v3/user/keyPair/info"), nil, response)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return response, nil
}
