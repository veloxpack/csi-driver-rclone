// The MIT License (MIT)
//
// Copyright (C) 2026 Storj Labs, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package infectious

import (
	"errors"
	"fmt"
	"slices"
)

// RebuildPlan holds the part of a rebuild that depends only on which share
// numbers are available, not on the share data itself.
//
// Building a plan performs the O(k^3) decode matrix inversion. Applying it to
// share data is O(k^2) in the number of shares and linear in the share size.
// (*FEC).Rebuild does both on every call, so a caller that rebuilds many
// blocks from the same set of shares -- every stripe of a segment, say --
// pays for the inversion once per block instead of once per share set. For
// typical share sizes the inversion is the more expensive half, because it
// runs addmul over k-byte rows where neither the vectorized kernel nor the
// call overhead amortizes.
//
// A RebuildPlan is immutable once built and is safe for concurrent use.
type RebuildPlan struct {
	k, n int

	// m_dec is the k*k inverted decode matrix, row major. Row i is only
	// meaningful when data share i is missing.
	m_dec []byte

	// slotShare[i] is the number of the share occupying column i of the
	// decode. slotShare[i] == i means data share i was supplied and is passed
	// through untouched; slotShare[i] >= k means data share i is missing and
	// must be reconstructed from row i of m_dec.
	slotShare []int

	// slotOf maps a share number to its column in the decode, or -1 when the
	// plan does not use that share.
	slotOf []int

	// present lists the slots that pass through, and missing lists the slots
	// that must be reconstructed. Both are ascending. Together they cover
	// 0..k-1.
	present []int
	missing []int
}

// PlanRebuild builds a RebuildPlan for the given share numbers. It selects
// which k of them to decode from using the same rule as (*FEC).Rebuild:
// prefer the data share whose number matches the slot, and otherwise fill
// from the highest numbered shares.
//
// shareNumbers must contain at least k entries, each in [0, n), with no
// duplicates. It is not modified.
func (f *FEC) PlanRebuild(shareNumbers []int) (*RebuildPlan, error) {
	k, n := f.k, f.n

	if len(shareNumbers) < k {
		return nil, NotEnoughShares
	}

	sorted := append([]int(nil), shareNumbers...)
	slices.Sort(sorted)

	if sorted[0] < 0 {
		return nil, fmt.Errorf("invalid share id: %d", sorted[0])
	}
	if last := sorted[len(sorted)-1]; last >= n {
		return nil, fmt.Errorf("invalid share id: %d", last)
	}
	for i := 1; i < len(sorted); i++ {
		if sorted[i] == sorted[i-1] {
			return nil, fmt.Errorf("duplicate share id: %d", sorted[i])
		}
	}

	slotOf := make([]int, n)
	for i := range slotOf {
		slotOf[i] = -1
	}

	m_dec := make([]byte, k*k)
	slotShare := make([]int, k)
	var present, missing []int

	// This mirrors the slot selection in (*FEC).Rebuild exactly. Walking the
	// sorted numbers from both ends puts data share i in slot i whenever it
	// was supplied, and backfills the remaining slots with parity shares.
	begin, end := 0, len(sorted)-1
	for i := range k {
		var shareID int
		if sorted[begin] == i {
			shareID = sorted[begin]
			begin++
		} else {
			shareID = sorted[end]
			end--
		}

		if shareID < k {
			// Only reachable via the front pointer, so shareID == i and row i
			// of the decode matrix is the identity row.
			m_dec[i*(k+1)] = 1
			present = append(present, i)
		} else {
			copy(m_dec[i*k:i*k+k], f.enc_matrix[shareID*k:])
			missing = append(missing, i)
		}

		slotShare[i] = shareID
		slotOf[shareID] = i
	}

	if err := invertMatrix(m_dec, k); err != nil {
		return nil, err
	}

	return &RebuildPlan{
		k:         k,
		n:         n,
		m_dec:     m_dec,
		slotShare: slotShare,
		slotOf:    slotOf,
		present:   present,
		missing:   missing,
	}, nil
}

// Required returns the number of shares the plan decodes from.
func (p *RebuildPlan) Required() int { return p.k }

// Shares returns the numbers of the shares the plan decodes from, indexed by
// decode slot. The returned slice must not be modified.
func (p *RebuildPlan) Shares() []int { return p.slotShare }

// Rebuilder applies a RebuildPlan to share data. It owns scratch buffers that
// it reuses across calls, so unlike RebuildPlan it is not safe for concurrent
// use. Create one per goroutine.
type Rebuilder struct {
	plan    *RebuildPlan
	sharesv [][]byte
	buf     []byte
}

// NewRebuilder returns a Rebuilder that applies p.
func (p *RebuildPlan) NewRebuilder() *Rebuilder {
	return &Rebuilder{
		plan:    p,
		sharesv: make([][]byte, p.k),
	}
}

// Plan returns the plan the Rebuilder applies.
func (r *Rebuilder) Plan() *RebuildPlan { return r.plan }

// Rebuild is the planned equivalent of (*FEC).Rebuild: it calls output k
// times, once per original data piece, with the piece number and its data.
//
// shares must supply every share number in the plan; extra shares are
// ignored, and the order does not matter. Unlike (*FEC).Rebuild, shares is
// neither sorted nor otherwise modified.
//
// As with (*FEC).Rebuild, output is not called in piece number order, and the
// byte slices handed to output may be reused once output returns.
//
// Rebuild assumes the shares have already been corrected, or did not need to
// be.
func (r *Rebuilder) Rebuild(shares []Share, output func(Share)) error {
	p := r.plan

	sharesv := r.sharesv
	clear(sharesv)

	shareSize := -1
	for _, share := range shares {
		if share.Number < 0 || share.Number >= p.n {
			return fmt.Errorf("invalid share id: %d", share.Number)
		}
		slot := p.slotOf[share.Number]
		if slot < 0 {
			continue
		}
		if shareSize < 0 {
			shareSize = len(share.Data)
		} else if len(share.Data) != shareSize {
			return errors.New("shares must all be the same length")
		}
		sharesv[slot] = share.Data
	}

	for i, data := range sharesv {
		if data == nil {
			return fmt.Errorf("missing share id: %d", p.slotShare[i])
		}
	}

	for _, i := range p.present {
		output(Share{Number: i, Data: sharesv[i]})
	}

	if len(p.missing) == 0 {
		return nil
	}

	if cap(r.buf) < shareSize {
		r.buf = make([]byte, shareSize)
	}
	buf := r.buf[:shareSize]

	for _, i := range p.missing {
		clear(buf)

		row := p.m_dec[i*p.k : i*p.k+p.k]
		for col := 0; col < p.k; col++ {
			addmul(buf, sharesv[col], row[col])
		}

		output(Share{Number: i, Data: buf})
	}

	return nil
}
