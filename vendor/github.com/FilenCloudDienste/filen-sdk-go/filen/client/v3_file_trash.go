package client

import "context"

// v3fileTrashRequest represents the request structure for moving a file to trash.
type v3fileTrashRequest struct {
	UUID string `json:"uuid"`
}

// PostV3FileTrash calls /v3/file/trash to move a file to the trash.
// This is a soft delete operation that can be reversed later.
func (c *Client) PostV3FileTrash(ctx context.Context, uuid string) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/file/trash"), v3fileTrashRequest{
		UUID: uuid,
	})
	if err != nil {
		return err
	}
	return nil
}
