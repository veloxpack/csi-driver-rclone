package client

import "context"

// V3UserPublicKeyResponse represents the response structure from the public key retrieval endpoint.
type V3UserPublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
}

// v3UserPublicKeyRequest represents the request structure for retrieving another user's public key.
type v3UserPublicKeyRequest struct {
	Email string `json:"email"`
}

// PostV3UserPublicKey calls /v3/user/publicKey to retrieve the public key of another Filen user.
// This key is needed for securely sharing files with the user through end-to-end encryption.
func (c *Client) PostV3UserPublicKey(ctx context.Context, email string) (*V3UserPublicKeyResponse, error) {
	var resp V3UserPublicKeyResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/user/publicKey"), v3UserPublicKeyRequest{
		Email: email,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
