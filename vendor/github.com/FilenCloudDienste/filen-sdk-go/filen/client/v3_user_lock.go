package client

import "context"

// V3UserLockRequest represents the request structure for acquiring, refreshing, or releasing a resource lock.
type V3UserLockRequest struct {
	LockUUID string `json:"uuid"`
	Type     string `json:"type"`
	Resource string `json:"resource"`
}

// V3UserLockResponse represents the response structure from the lock management endpoint.
// It indicates whether the requested lock operation was successful.
type V3UserLockResponse struct {
	Acquired  bool   `json:"acquired"`
	Released  bool   `json:"released"`
	Refreshed bool   `json:"refreshed"`
	Resource  string `json:"resource"`
	Status    string `json:"status"`
}

// PostV3UserLock calls /v3/user/lock to manage locks on resources.
// Locks are used to prevent concurrent operations on the same resource.
// The Type field should be one of: "acquire", "refresh", or "release".
func (c *Client) PostV3UserLock(ctx context.Context, req V3UserLockRequest) (*V3UserLockResponse, error) {
	var res V3UserLockResponse
	_, err := c.RequestData(ctx, "POST", GatewayURL("/v3/user/lock"), req, &res)
	return &res, err
}
