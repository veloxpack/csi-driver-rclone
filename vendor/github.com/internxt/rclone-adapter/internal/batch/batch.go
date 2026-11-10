// Package batch splits requests to endpoints that limit how large a request
// may be.
package batch

import (
	"encoding/json"
	"fmt"
)

// MaxBodyBytes is the largest JSON request body the drive server accepts.
// Larger bodies are rejected with a 500 error.
const MaxBodyBytes = 100 * 1024

// Split splits items into consecutive batches of at most maxItems items
// whose request body, as built by body and encoded as JSON, is at most
// maxBytes. body must encode the items it is given as a single JSON array.
func Split[T any](items []T, maxItems, maxBytes int, body func([]T) any) ([][]T, error) {
	envelope, err := json.Marshal(body([]T{}))
	if err != nil {
		return nil, err
	}
	empty := len(envelope) - 1
	var batches [][]T
	start, size := 0, empty
	for i, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		n := len(b) + 1
		if i > start && (i-start == maxItems || size+n > maxBytes) {
			batches = append(batches, items[start:i])
			start, size = i, empty
		}
		if size+n > maxBytes {
			return nil, fmt.Errorf("item %d is too large for one request: %d bytes, at most %d allowed", i, size+n, maxBytes)
		}
		size += n
	}
	if start < len(items) {
		batches = append(batches, items[start:])
	}
	return batches, nil
}
