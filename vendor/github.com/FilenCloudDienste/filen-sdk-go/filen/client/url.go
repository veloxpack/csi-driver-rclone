// Package client provides the functionality to interact with the Filen API.
package client

import (
	"math/rand"
	"strings"
)

// URL server pools for different Filen service endpoints.
// These provide load balancing and fallback options.
var (
	// gatewayURLs contains the list of available gateway URLs for API requests.
	gatewayURLs = []string{
		"https://gateway.filen.io",
		"https://gateway.filen.net",
		"https://gateway.filen-1.net",
		"https://gateway.filen-2.net",
		"https://gateway.filen-3.net",
		"https://gateway.filen-4.net",
		"https://gateway.filen-5.net",
		"https://gateway.filen-6.net",
	}

	// egestURLs contains the list of available egress URLs for file downloads.
	egestURLs = []string{
		"https://egest.filen.io",
		"https://egest.filen.net",
		"https://egest.filen-1.net",
		"https://egest.filen-2.net",
		"https://egest.filen-3.net",
		"https://egest.filen-4.net",
		"https://egest.filen-5.net",
		"https://egest.filen-6.net",
	}

	// ingestURLs contains the list of available ingress URLs for file uploads.
	ingestURLs = []string{
		"https://ingest.filen.io",
		"https://ingest.filen.net",
		"https://ingest.filen-1.net",
		"https://ingest.filen-2.net",
		"https://ingest.filen-3.net",
		"https://ingest.filen-4.net",
		"https://ingest.filen-5.net",
		"https://ingest.filen-6.net",
	}
)

// URL type constants define the type of Filen service to use.
const (
	// URLTypeIngest represents an upload endpoint URL type
	URLTypeIngest = 1

	// URLTypeEgest represents a download endpoint URL type
	URLTypeEgest = 2

	// URLTypeGateway represents an API endpoint URL type
	URLTypeGateway = 3
)

// FilenURL represents a URL for Filen API or storage operations.
// It handles load balancing by randomly selecting a server from the appropriate pool.
type FilenURL struct {
	Type      int    // The type of URL (ingest, egest, or gateway)
	Path      string // The path component of the URL
	CachedUrl string // The complete URL, cached after first generation
}

// GatewayURL creates a new FilenURL for API gateway operations with the given path.
// This is a convenience function for creating gateway URLs, which are the most common.
func GatewayURL(path string) *FilenURL {
	return &FilenURL{
		Type:      URLTypeGateway,
		Path:      path,
		CachedUrl: "",
	}
}

// String returns the complete URL as a string.
// It randomly selects a server from the appropriate pool on first call,
// then caches and returns the same URL for subsequent calls.
// This implements the fmt.Stringer interface.
func (url *FilenURL) String() string {
	if url.CachedUrl == "" {
		var builder strings.Builder
		switch url.Type {
		case URLTypeIngest:
			builder.WriteString(ingestURLs[rand.Intn(len(ingestURLs))])
		case URLTypeEgest:
			builder.WriteString(egestURLs[rand.Intn(len(egestURLs))])
		case URLTypeGateway:
			builder.WriteString(gatewayURLs[rand.Intn(len(gatewayURLs))])
		}
		builder.WriteString(url.Path)
		url.CachedUrl = builder.String()
	}

	return url.CachedUrl
}
