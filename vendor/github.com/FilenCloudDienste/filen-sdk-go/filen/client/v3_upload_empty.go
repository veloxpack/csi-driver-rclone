package client

import (
	"context"

	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// V3UploadEmptyRequest represents the request structure for creating an empty file.
type V3UploadEmptyRequest struct {
	UUID       string                       `json:"uuid"`
	Name       crypto.EncryptedString       `json:"name"`
	NameHashed string                       `json:"nameHashed"`
	Size       crypto.EncryptedString       `json:"size"`
	Parent     string                       `json:"parent"`
	MimeType   crypto.EncryptedString       `json:"mime"`
	Metadata   crypto.EncryptedString       `json:"metadata"`
	Version    crypto.FileEncryptionVersion `json:"version"`
}

// PostV3UploadEmpty calls /v3/upload/empty to create an empty file.
// This can be used to create placeholder files or zero-byte files.
func (c *Client) PostV3UploadEmpty(ctx context.Context, request V3UploadEmptyRequest) (*V3UploadDoneResponse, error) {
	response := &V3UploadDoneResponse{}
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/upload/empty"), request, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
