// Package client provides the functionality to interact with the Filen API.
package client

const gatewayUrl = "https://gateway.filen.io"
const ingestUrl = "https://ingest.filen.io"
const egestUrl = "https://egest.filen.io"

func GatewayURL(path string) string {
	return gatewayUrl + path
}

func IngestURL(path string) string {
	return ingestUrl + path
}

func EgestURL(path string) string {
	return egestUrl + path
}
