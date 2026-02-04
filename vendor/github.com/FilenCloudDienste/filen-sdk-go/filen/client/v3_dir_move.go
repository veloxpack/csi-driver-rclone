package client

import "context"

// v3DirMoveRequest represents the request structure for moving a directory.
type v3DirMoveRequest struct {
	UUID          string `json:"uuid"`
	NewParentUUID string `json:"to"`
}

// PostV3DirMove calls /v3/dir/move to move a directory to a new parent location.
// This changes the directory's location in the filesystem hierarchy.
func (c *Client) PostV3DirMove(ctx context.Context, uuid string, newParentUUID string) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/dir/move"), v3DirMoveRequest{
		UUID:          uuid,
		NewParentUUID: newParentUUID,
	})
	return err
}
