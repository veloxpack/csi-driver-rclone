package client

import (
	"context"
	"fmt"
)

// v3DirSharedRequest represents the request structure for checking directory sharing status.
type v3DirSharedRequest struct {
	UUID string `json:"uuid"`
}

// V3DirSharedUser represents a user with whom a directory is shared.
type V3DirSharedUser struct {
	Email     string `json:"email"`
	PublicKey string `json:"publicKey"`
}

// V3DirSharedResponse represents the response structure from the directory shared endpoint.
type V3DirSharedResponse struct {
	Shared bool              `json:"shared"`
	Users  []V3DirSharedUser `json:"users"`
}

// PostV3DirShared calls /v3/dir/shared to check if a directory is shared with other users.
// If shared, it returns the list of users with access.
func (c *Client) PostV3DirShared(ctx context.Context, dirUUID string) (*V3DirSharedResponse, error) {
	var res V3DirSharedResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/dir/shared"), v3DirSharedRequest{
		UUID: dirUUID,
	}, &res)
	if err != nil {
		return nil, fmt.Errorf("PostV3DirShared: %w", err)
	}
	return &res, nil
}
