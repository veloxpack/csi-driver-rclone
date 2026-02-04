package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3userDekRequest represents the request structure for updating the user's Data Encryption Key.
type v3userDekRequest struct {
	DEK crypto.EncryptedString `json:"dek"`
}

// PostV3UserDEK calls /v3/user/dek to update the user's Data Encryption Key (DEK).
// The DEK should be encrypted with the user's master key before being sent.
func (c *Client) PostV3UserDEK(ctx context.Context, encryptedDEK crypto.EncryptedString) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/user/dek"), v3userDekRequest{
		DEK: encryptedDEK,
	})
	if err != nil {
		return err
	}
	return nil
}

// v3userDEKResponse represents the response structure from the DEK retrieval endpoint.
type v3userDEKResponse struct {
	DEK crypto.EncryptedString `json:"dek"`
}

// GetV3UserDEK calls /v3/user/dek to retrieve the user's encrypted Data Encryption Key.
// The returned DEK is encrypted with the user's master key and must be decrypted locally.
func (c *Client) GetV3UserDEK(ctx context.Context) (crypto.EncryptedString, error) {
	response := &v3userDEKResponse{}
	_, err := c.RequestData(ctx, "GET", GatewayURL("/v3/user/dek"), nil, response)
	if err != nil {
		return "", err
	}
	return response.DEK, nil
}
