package client

import (
	"context"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// V3ItemShareRequest represents the request structure for sharing an item with another user.
type V3ItemShareRequest struct {
	UUID       string                 `json:"uuid"`
	ParentUUID string                 `json:"parent"`
	Email      string                 `json:"email"`
	Type       string                 `json:"type"`
	Metadata   crypto.EncryptedString `json:"metadata"`
}

// PostV3ItemShare calls /v3/item/share to share an item (file or directory) with another Filen user.
// This enables secure file sharing between Filen accounts while maintaining end-to-end encryption.
func (c *Client) PostV3ItemShare(ctx context.Context, req V3ItemShareRequest) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/item/share"), req)
	if err != nil {
		return fmt.Errorf("PostV3ItemShare: %w", err)
	}
	return nil
}
