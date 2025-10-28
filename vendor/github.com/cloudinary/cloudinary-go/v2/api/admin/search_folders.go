package admin

// The Cloudinary API search folders method allows you fine control on filtering and retrieving information on all the
// folders in your account with the help of query expressions in a Lucene-like query language.
//
// https://cloudinary.com/documentation/search_api
import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin/search"
)

const searchFoldersEndPoint api.EndPoint = "folders/search"

// SearchFolders executes the search folders API request.
func (a *API) SearchFolders(ctx context.Context, searchQuery search.Query) (*SearchFoldersResult, error) {
	res := &SearchFoldersResult{}
	_, err := a.post(ctx, api.BuildPath(searchFoldersEndPoint), searchQuery, res)

	return res, err
}

// SearchFoldersResult is the result of SearchFolders.
type SearchFoldersResult struct {
	TotalCount int            `json:"total_count"`
	Time       int            `json:"time"`
	Folders    []SearchFolder `json:"folders"`
	Error      api.ErrorResp  `json:"error,omitempty"`
	NextCursor string         `json:"next_cursor,omitempty"`
	Response   interface{}
}

// SearchFolder represents the details of a single folder that was found.
type SearchFolder struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	CreatedAt  time.Time `json:"created_at"`
	ExternalID string    `json:"external_id"`
}
