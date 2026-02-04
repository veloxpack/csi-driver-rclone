package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3authInfoRequest represents the request structure for the auth info endpoint.
type v3authInfoRequest struct {
	Email string `json:"email"`
}

// V3AuthInfoResponse represents the response structure from the auth info endpoint.
type V3AuthInfoResponse struct {
	AuthVersion crypto.AuthVersion `json:"authVersion"`
	Salt        string             `json:"salt"`
}

// PostV3AuthInfo calls /v3/auth/info to retrieve authentication information for a user.
// This endpoint doesn't require authentication and can be used before login.
func (uc *UnauthorizedClient) PostV3AuthInfo(ctx context.Context, email string) (*V3AuthInfoResponse, error) {
	authInfo := &V3AuthInfoResponse{}
	_, err := uc.RequestData(ctx, "POST", GatewayURL("/v3/auth/info"), v3authInfoRequest{
		Email: email,
	}, authInfo)
	return authInfo, err
}
