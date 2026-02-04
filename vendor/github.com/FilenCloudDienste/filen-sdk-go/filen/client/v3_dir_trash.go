package client

import "context"

// v3DirTrashRequest represents the request structure for moving a directory to trash.
type v3DirTrashRequest struct {
	UUID string `json:"uuid"`
}

// PostV3DirTrash calls /v3/dir/trash to move a directory to the trash.
// This is a soft delete operation that can be reversed later.
func (c *Client) PostV3DirTrash(ctx context.Context, uuid string) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/dir/trash"), v3DirTrashRequest{
		UUID: uuid,
	})
	if err != nil {
		return err
	}
	return nil
}
