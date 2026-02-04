package client

import "context"

// v3SearchAddRequest represents the request structure for adding items to the search index.
type v3SearchAddRequest struct {
	Items []V3SearchAddItem `json:"items"`
}

// V3SearchAddItem represents an item to be indexed for search.
type V3SearchAddItem struct {
	UUID string `json:"uuid"`
	Hash string `json:"hash"`
	Type string `json:"type"`
}

// PostV3SearchAdd calls /v3/search/add to add items to the search index.
// This allows files and directories to be found through the global search functionality
func (c *Client) PostV3SearchAdd(ctx context.Context, items []V3SearchAddItem) error {
	_, err := c.Request(ctx, "POST", GatewayURL("/v3/search/add"), v3SearchAddRequest{
		Items: items,
	})
	return err
}
