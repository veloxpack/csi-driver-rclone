// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

package cohorts

import (
	"fmt"
	"math/bits"
)

type bitset struct {
	v [4]uint64
}

func (b bitset) String() string {
	return fmt.Sprintf("%064b%064b%064b%064b", b.v[3], b.v[2], b.v[1], b.v[0])
}

func (b *bitset) set(n byte)      { b.v[n/64] |= 1 << (n % 64) }
func (b *bitset) has(n byte) bool { return b.v[n/64]&(1<<(n%64)) != 0 }
func (b *bitset) unset(n byte)    { b.v[n/64] &^= 1 << (n % 64) }

func (b *bitset) count() int {
	x0 := bits.OnesCount64(b.v[0])
	x1 := bits.OnesCount64(b.v[1])
	x2 := bits.OnesCount64(b.v[2])
	x3 := bits.OnesCount64(b.v[3])
	return x0 + x1 + x2 + x3
}

func (b bitset) iter(cb func(n byte) bool) {
	for i, v := range &b.v {
		for v != 0 {
			if !cb(byte(64*i + bits.TrailingZeros64(v))) {
				return
			}
			v &= v - 1
		}
	}
}
