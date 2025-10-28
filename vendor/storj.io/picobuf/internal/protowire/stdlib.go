// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protowire

// PutUvarint encodes a uint64 into buf and returns the number of bytes written.
// If the buffer is too small, PutUvarint will panic.
//
// This copied from encoding/binary package.
func PutUvarint(buf []byte, x uint64) int {
	i := 0
	for x >= 0x80 {
		buf[i] = byte(x) | 0x80
		x >>= 7
		i++
	}
	buf[i] = byte(x)
	return i + 1
}
