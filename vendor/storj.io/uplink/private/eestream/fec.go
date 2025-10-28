// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import "storj.io/infectious"

// A Share represents a piece of the FEC-encoded data.
type Share = infectious.Share

// NewFEC creates a *FEC using k required pieces and n total pieces.
// Encoding data with this *FEC will generate n pieces, and decoding
// data requires k uncorrupted pieces. During decode, when more than
// k pieces exist, corrupted data can be detected and recovered from.
func NewFEC(k, n int) (*infectious.FEC, error) {
	return infectious.NewFEC(k, n)
}
