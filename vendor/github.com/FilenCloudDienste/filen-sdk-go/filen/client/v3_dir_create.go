package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3createDirRequest represents the request structure for creating a directory.
type v3createDirRequest struct {
	UUID       string                 `json:"uuid"`
	Name       crypto.EncryptedString `json:"name"`
	NameHashed string                 `json:"nameHashed"`
	ParentUUID string                 `json:"parent"`
}

// V3CreateDirResponse represents the response structure from the create directory endpoint.
type V3CreateDirResponse struct {
	UUID string `json:"uuid"`
}

// PostV3DirCreate calls /v3/dir/create to create a new directory.
// It requires encrypted metadata for the directory name and a hashed version for lookups.
func (c *Client) PostV3DirCreate(ctx context.Context, uuid string, name crypto.EncryptedString, nameHashed string, parentUUID string) (*V3CreateDirResponse, error) {
	response := &V3CreateDirResponse{}
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/dir/create"), v3createDirRequest{
		UUID:       uuid,
		Name:       name,
		NameHashed: nameHashed,
		ParentUUID: parentUUID,
	}, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
