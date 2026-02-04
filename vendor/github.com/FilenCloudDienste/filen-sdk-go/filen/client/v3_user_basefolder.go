package client

import "context"

// V3UserBaseFolderResponse represents the response structure from the user's base folder endpoint.
type V3UserBaseFolderResponse struct {
	UUID string `json:"uuid"`
}

// GetV3UserBaseFolder calls /v3/user/baseFolder to retrieve the UUID of the user's root directory.
// This is the top-level directory in the user's file structure and serves as the starting point for navigation.
func (c *Client) GetV3UserBaseFolder(ctx context.Context) (*V3UserBaseFolderResponse, error) {
	userBaseFolder := &V3UserBaseFolderResponse{}
	_, err := c.RequestData(ctx, "GET", GatewayURL("/v3/user/baseFolder"), nil, userBaseFolder)
	return userBaseFolder, err
}
