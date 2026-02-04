package client

import (
	"context"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3DirMetadataRequest represents the request structure for updating directory metadata.
type v3DirMetadataRequest struct {
	UUID       string                 `json:"uuid"`
	NameHashed string                 `json:"nameHashed"`
	Metadata   crypto.EncryptedString `json:"name"`
}

// PostV3DirMetadata calls /v3/dir/metadata to update a directory's metadata.
// This is typically used to rename a directory while preserving encryption.
func (c *Client) PostV3DirMetadata(ctx context.Context, uuid string, nameHashed string, metadata crypto.EncryptedString) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/dir/metadata"), v3DirMetadataRequest{
		UUID:       uuid,
		NameHashed: nameHashed,
		Metadata:   metadata,
	})
	if err != nil {
		return fmt.Errorf("post v3 dir metadata: %w", err)
	}
	return nil
}
