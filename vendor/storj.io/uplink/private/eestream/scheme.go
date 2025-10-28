// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"storj.io/infectious"
)

// ErasureScheme represents the general format of any erasure scheme algorithm.
// If this interface can be implemented, the rest of this library will work
// with it.
type ErasureScheme interface {
	// Encode will take 'in' and call 'out' with erasure coded pieces.
	Encode(in []byte, out func(num int, data []byte)) error

	// EncodeSingle will take 'in' with the stripe and fill 'out' with the erasure share for piece 'num'.
	EncodeSingle(in, out []byte, num int) error

	// Decode will take a mapping of available erasure coded piece num -> data,
	// 'in', and append the combined data to 'out', returning it.
	Decode(out []byte, in []infectious.Share) ([]byte, error)

	// Rebuild is a direct call to infectious.Rebuild, which does no error
	// detection and is faster.
	Rebuild(in []infectious.Share, out func(infectious.Share)) error

	// ErasureShareSize is the size of the erasure shares that come from Encode
	// and are passed to Decode.
	ErasureShareSize() int

	// StripeSize is the size the stripes that are passed to Encode and come
	// from Decode.
	StripeSize() int

	// Encode will generate this many erasure shares and therefore this many pieces.
	TotalCount() int

	// Decode requires at least this many pieces.
	RequiredCount() int
}
