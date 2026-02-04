package client

import "context"

// v3FileMoveRequest represents the request structure for moving a file.
type v3FileMoveRequest struct {
	UUID          string `json:"uuid"`
	NewParentUUID string `json:"to"`
}

// PostV3FileMove calls /v3/file/move to move a file to a new parent directory.
// This changes the file's location in the filesystem hierarchy.
func (c *Client) PostV3FileMove(ctx context.Context, uuid string, newParentUUID string) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/file/move"), v3FileMoveRequest{
		UUID:          uuid,
		NewParentUUID: newParentUUID,
	})
	return err
}
