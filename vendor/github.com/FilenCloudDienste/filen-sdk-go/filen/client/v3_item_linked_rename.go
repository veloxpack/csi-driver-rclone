package client

import (
	"context"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3ItemLinkedRenameRequest represents the request structure for renaming an item in a public link.
type v3ItemLinkedRenameRequest struct {
	UUID     string                 `json:"uuid"`
	LinkUUID string                 `json:"linkUUID"`
	Metadata crypto.EncryptedString `json:"metadata"`
}

// PostV3ItemLinkedRename calls /v3/item/linked/rename to update the display name of an item in a public link.
// This allows renaming the item in the shared link without affecting the original.
func (c *Client) PostV3ItemLinkedRename(ctx context.Context, uuid string, linkUUID string, metadata crypto.EncryptedString) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/item/linked/rename"), v3ItemLinkedRenameRequest{
		UUID:     uuid,
		LinkUUID: linkUUID,
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("post v3 item linked rename: %w", err)
	}
	return nil
}
