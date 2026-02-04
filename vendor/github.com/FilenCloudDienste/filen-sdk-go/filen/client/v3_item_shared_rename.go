package client

import (
	"context"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3ItemSharedRenameRequest represents the request structure for renaming a shared item.
type v3ItemSharedRenameRequest struct {
	Uuid       string                 `json:"uuid"`
	ReceiverId int                    `json:"receiverId"`
	Metadata   crypto.EncryptedString `json:"metadata"`
}

// PostV3ItemSharedRename calls /v3/item/shared/rename to update the metadata for a specific sharing recipient.
// This allows customizing how shared items appear to different recipients.
func (c *Client) PostV3ItemSharedRename(ctx context.Context, uuid string, receiverId int, metadata crypto.EncryptedString) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/item/shared/rename"), v3ItemSharedRenameRequest{
		Uuid:       uuid,
		ReceiverId: receiverId,
		Metadata:   metadata,
	})
	if err != nil {
		return fmt.Errorf("post v3 item shared rename: %w", err)
	}
	return nil
}
