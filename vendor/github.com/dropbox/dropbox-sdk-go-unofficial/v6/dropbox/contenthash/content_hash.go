// Copyright (c) Dropbox, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package contenthash computes Dropbox API content_hash values.
package contenthash

import (
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"hash"
	"io"
)

// BlockSize is the block size used by the Dropbox API content_hash algorithm.
const BlockSize = 4 * 1024 * 1024

// Hasher computes hashes using the same algorithm as the Dropbox API
// content_hash metadata and upload argument fields.
type Hasher struct {
	block   []byte
	overall hash.Hash
}

var _ hash.Hash = (*Hasher)(nil)

// New returns a hasher for Dropbox API content_hash values.
func New() *Hasher {
	return &Hasher{}
}

// Compute reads r and returns its Dropbox API content_hash in hexadecimal form.
func Compute(r io.Reader) (string, error) {
	h := New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return h.HexDigest(), nil
}

// Write adds p to the running content hash.
func (h *Hasher) Write(p []byte) (int, error) {
	written := len(p)
	for len(p) > 0 {
		if len(h.block) == 0 && len(p) >= BlockSize {
			h.addBlockHash(p[:BlockSize])
			p = p[BlockSize:]
			continue
		}

		space := BlockSize - len(h.block)
		if space > len(p) {
			space = len(p)
		}
		h.block = append(h.block, p[:space]...)
		p = p[space:]

		if len(h.block) == BlockSize {
			h.addBlockHash(h.block)
			h.block = h.block[:0]
		}
	}
	return written, nil
}

// Sum appends the current binary Dropbox API content hash to b and returns the
// resulting slice. It does not change the underlying hash state.
func (h *Hasher) Sum(b []byte) []byte {
	overall := cloneSHA256(h.overall)
	if len(h.block) > 0 {
		blockHash := sha256.Sum256(h.block)
		_, _ = overall.Write(blockHash[:])
	}
	return overall.Sum(b)
}

// Reset resets h to its initial state.
func (h *Hasher) Reset() {
	h.block = h.block[:0]
	if h.overall != nil {
		h.overall.Reset()
	}
}

// Size returns the number of bytes in a Dropbox API content hash.
func (h *Hasher) Size() int {
	return sha256.Size
}

// BlockSize returns the Dropbox API content_hash block size.
func (h *Hasher) BlockSize() int {
	return BlockSize
}

// HexDigest returns the current Dropbox API content hash in hexadecimal form.
func (h *Hasher) HexDigest() string {
	return hex.EncodeToString(h.Sum(nil))
}

func (h *Hasher) addBlockHash(block []byte) {
	blockHash := sha256.Sum256(block)
	_, _ = h.overallHash().Write(blockHash[:])
}

func cloneSHA256(h hash.Hash) hash.Hash {
	if h == nil {
		return sha256.New()
	}
	state, err := h.(encoding.BinaryMarshaler).MarshalBinary()
	if err != nil {
		panic("contenthash: failed to marshal SHA-256 state: " + err.Error())
	}
	clone := sha256.New()
	if err := clone.(encoding.BinaryUnmarshaler).UnmarshalBinary(state); err != nil {
		panic("contenthash: failed to unmarshal SHA-256 state: " + err.Error())
	}
	return clone
}

func (h *Hasher) overallHash() hash.Hash {
	if h.overall == nil {
		h.overall = sha256.New()
	}
	return h.overall
}
