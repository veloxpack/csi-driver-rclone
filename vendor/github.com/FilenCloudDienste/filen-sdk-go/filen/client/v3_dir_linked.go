package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3DirLinkedRequest represents the request structure for checking directory link status.
type v3DirLinkedRequest struct {
	UUID string `json:"uuid"`
}

// V3DirSharedLink represents a public sharing link for a directory.
type V3DirSharedLink struct {
	UUID string                 `json:"linkUUID"`
	Key  crypto.EncryptedString `json:"linkKey"`
}

// V3DirLinkedResponse represents the response structure from the directory linked endpoint.
type V3DirLinkedResponse struct {
	Linked bool              `json:"link"`
	Links  []V3DirSharedLink `json:"links"`
}

// PostV3DirLinked calls /v3/dir/linked to check if a directory has public sharing links.
// If links exist, it returns their details.
func (c *Client) PostV3DirLinked(ctx context.Context, dirUUID string) (*V3DirLinkedResponse, error) {
	var res V3DirLinkedResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/dir/linked"), v3DirLinkedRequest{
		UUID: dirUUID,
	}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
