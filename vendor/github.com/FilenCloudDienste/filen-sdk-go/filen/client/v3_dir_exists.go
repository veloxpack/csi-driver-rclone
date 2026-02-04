package client

import "context"

// V3DirExistsResponse represents the response structure from the directory exists endpoint.
type V3DirExistsResponse struct {
	Exists bool   `json:"exists"`
	UUID   string `json:"uuid"`
}

// v3DirExistsRequest represents the request structure for checking if a directory exists.
type v3DirExistsRequest struct {
	NameHashed string `json:"nameHashed"`
	ParentUUID string `json:"parent"`
}

// PostV3DirExists calls /v3/dir/exists to check if a directory with the given name exists.
// It uses a hashed name for lookup to preserve encryption.
func (c *Client) PostV3DirExists(ctx context.Context, nameHashed string, parentUUID string) (*V3DirExistsResponse, error) {
	var resp V3DirExistsResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/dir/exists"), v3DirExistsRequest{
		NameHashed: nameHashed,
		ParentUUID: parentUUID,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
