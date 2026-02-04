package client

import "context"

// v3fileDeletePermanentRequest represents the request structure for permanently deleting a file.
type v3fileDeletePermanentRequest struct {
	UUID string `json:"uuid"`
}

// PostV3FileDeletePermanent calls /v3/file/delete/permanent to permanently delete a file.
// This operation cannot be undone.
func (c *Client) PostV3FileDeletePermanent(ctx context.Context, uuid string) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/file/delete/permanent"), v3fileDeletePermanentRequest{
		UUID: uuid,
	})
	if err != nil {
		return err
	}
	return nil
}
