package client

import "context"

// V3FileExistsResponse represents the response structure from the file exists endpoint.
type V3FileExistsResponse struct {
	Exists bool   `json:"exists"`
	UUID   string `json:"uuid"`
}

// v3FileExistsRequest represents the request structure for checking if a file exists.
type v3FileExistsRequest struct {
	NameHashed string `json:"nameHashed"`
	ParentUUID string `json:"parent"`
}

// PostV3FileExists calls /v3/file/exists to check if a file with the given name exists.
// It uses a hashed name for lookup to preserve encryption.
func (c *Client) PostV3FileExists(ctx context.Context, nameHashed string, parentUUID string) (*V3FileExistsResponse, error) {
	var resp V3FileExistsResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/file/exists"), v3FileExistsRequest{
		NameHashed: nameHashed,
		ParentUUID: parentUUID,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
