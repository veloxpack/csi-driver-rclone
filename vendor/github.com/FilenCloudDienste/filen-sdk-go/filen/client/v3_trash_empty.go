package client

import "context"

// PostV3TrashEmpty calls /v3/trash/empty to permanently delete all items in the trash.
// This operation cannot be undone and will remove all trashed files and directories.
func (c *Client) PostV3TrashEmpty(ctx context.Context) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/trash/empty"), nil)
	return err
}
