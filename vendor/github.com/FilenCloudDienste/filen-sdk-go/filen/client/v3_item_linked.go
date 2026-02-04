package client

import (
	"context"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
)

// v3ItemLinkedRequest represents the request structure for checking item link status.
type v3ItemLinkedRequest struct {
	UUID string `json:"uuid"`
}

// V3ItemLinkedLink represents a public sharing link for an item (file or directory).
type V3ItemLinkedLink struct {
	LinkUUID string                 `json:"linkUUID"`
	Key      crypto.EncryptedString `json:"linkKey"`
}

// V3ItemLinkedResponse represents the response structure from the item linked endpoint.
type V3ItemLinkedResponse struct {
	Linked bool               `json:"link"`
	Links  []V3ItemLinkedLink `json:"links"`
}

// PostV3ItemLinked calls /v3/item/linked to check if an item (file or directory) has public sharing links.
// This is a generic version that works with both files and directories.
func (c *Client) PostV3ItemLinked(ctx context.Context, uuid string) (*V3ItemLinkedResponse, error) {
	var res V3ItemLinkedResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/item/linked"), v3ItemLinkedRequest{
		UUID: uuid,
	}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
