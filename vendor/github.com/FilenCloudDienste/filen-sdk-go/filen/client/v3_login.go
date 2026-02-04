package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3loginRequest represents the request structure for user authentication.
type v3loginRequest struct {
	Email         string             `json:"email"`
	Password      string             `json:"password"`
	TwoFactorCode string             `json:"twoFactorCode"`
	AuthVersion   crypto.AuthVersion `json:"authVersion"`
}

// V3LoginResponse represents the response structure from the login endpoint.
// It contains the API key and encrypted keys needed for further operations.
type V3LoginResponse struct {
	APIKey     string                 `json:"apiKey"`
	MasterKeys crypto.EncryptedString `json:"masterKeys"`
	PublicKey  string                 `json:"publicKey"`
	PrivateKey crypto.EncryptedString `json:"privateKey"`
	DEK        crypto.EncryptedString `json:"dek"`
}

// PostV3Login calls /v3/login to authenticate a user and obtain an API key.
// The password should be derived according to Filen's password derivation scheme.
func (uc *UnauthorizedClient) PostV3Login(ctx context.Context, email string, password crypto.DerivedPassword, authVersion crypto.AuthVersion, twoFactorCode string) (*V3LoginResponse, error) {
	response := &V3LoginResponse{}
	_, err := uc.RequestData(ctx, "POST", GatewayURL("/v3/login"), v3loginRequest{
		Email:         email,
		Password:      string(password),
		TwoFactorCode: twoFactorCode,
		AuthVersion:   authVersion,
	}, response)
	return response, err
}
