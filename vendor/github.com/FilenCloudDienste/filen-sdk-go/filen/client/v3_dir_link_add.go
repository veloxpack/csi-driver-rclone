package client

import (
	"context"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// V3DirLinkAddRequest represents the request structure for creating a public link to a directory.
type V3DirLinkAddRequest struct {
	UUID       string                 `json:"uuid"`
	ParentUUID string                 `json:"parent"`
	LinkUUID   string                 `json:"linkUUID"`
	ItemType   string                 `json:"type"`
	Metadata   crypto.EncryptedString `json:"metadata"`
	LinkKey    crypto.EncryptedString `json:"key"`
	Expiration string                 `json:"expiration"`
}

// PostV3DirLinkAdd calls /v3/dir/link/add to create a public sharing link for a directory.
// This allows anonymous access to the directory via a link.
func (c *Client) PostV3DirLinkAdd(ctx context.Context, request V3DirLinkAddRequest) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/dir/link/add"), request)
	if err != nil {
		return fmt.Errorf("PostV3DirLinkAdd: %w", err)
	}
	return err
}
