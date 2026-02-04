package client

import "context"

// v3DirDeletePermanentRequest represents the request structure for permanently deleting a directory.
type v3DirDeletePermanentRequest struct {
	UUID string `json:"uuid"`
}

// PostV3DirDeletePermanent calls /v3/dir/delete/permanent to permanently delete a directory.
// This operation cannot be undone and will remove the directory and all its contents.
func (c *Client) PostV3DirDeletePermanent(ctx context.Context, uuid string) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/dir/delete/permanent"), v3DirDeletePermanentRequest{
		UUID: uuid,
	})
	if err != nil {
		return err
	}
	return nil
}
