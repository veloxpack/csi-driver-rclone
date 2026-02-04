package client

import "context"

// V3UserInfoResponse represents the response structure from the user info endpoint.
// It contains account details, storage usage, and the base folder UUID.
type V3UserInfoResponse struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	IsPremium   int    `json:"isPremium"`
	MaxStorage  int    `json:"maxStorage"`
	UsedStorage int    `json:"storageUsed"`
	AvatarURL   string `json:"avatarURL"`
	BaseFolder  string `json:"baseFolderUUID"`
}

// GetV3UserInfo calls /v3/user/info to retrieve information about the current user.
// This includes account details, storage quota, and usage statistics.
func (c *Client) GetV3UserInfo(ctx context.Context) (*V3UserInfoResponse, error) {
	var res V3UserInfoResponse
	_, err := c.RequestData(ctx, "GET", GatewayURL("/v3/user/info"), nil, &res)
	return &res, err
}
