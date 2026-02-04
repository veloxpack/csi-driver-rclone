package client

import "context"

// V3DirSizeResponse represents the response structure from the directory size endpoint.
type V3DirSizeResponse struct {
	Size    int `json:"size"`
	Folders int `json:"folders"`
	Files   int `json:"files"`
}

// v3dirSizeRequest represents the request structure for getting directory size information.
type v3dirSizeRequest struct {
	UUID string `json:"uuid"`
}

// PostV3DirSize calls /v3/dir/size to get the total size and item counts of a directory.
// This recursively calculates size including all nested files and folders.
func (c *Client) PostV3DirSize(ctx context.Context, uuid string) (*V3DirSizeResponse, error) {
	var res V3DirSizeResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/dir/size"), v3dirSizeRequest{UUID: uuid}, &res)
	return &res, err
}
