package client

import "context"

// v3ItemSharedRequest represents the request structure for checking item sharing status.
type v3ItemSharedRequest struct {
	UUID string `json:"uuid"`
}

// V3ItemSharedUser represents a user with whom an item is shared.
type V3ItemSharedUser struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	PublicKey string `json:"publicKey"`
}

// V3ItemSharedResponse represents the response structure from the item shared endpoint.
type V3ItemSharedResponse struct {
	Shared bool               `json:"sharing"`
	Users  []V3ItemSharedUser `json:"users"`
}

// PostV3ItemShared calls /v3/item/shared to check if an item is shared with other users.
// If shared, it returns the list of users with access to the item.
func (c *Client) PostV3ItemShared(ctx context.Context, uuid string) (*V3ItemSharedResponse, error) {
	var res V3ItemSharedResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/item/shared"), v3ItemSharedRequest{
		UUID: uuid,
	}, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
