// (C) 1996-1998 Luigi Rizzo (luigi@iet.unipi.it)
//     2009-2010 Jack Lloyd (lloyd@randombit.net)
//     2011 Billy Brumley (billy.brumley@aalto.fi)
//     2016-2017 Vivint, Inc. (jeff.wendling@vivint.com)
//
// Portions derived from code by Phil Karn (karn@ka9q.ampr.org),
// Robert Morelos-Zaragoza (robert@spectra.eng.hawaii.edu) and Hari
// Thirumoorthy (harit@spectra.eng.hawaii.edu), Aug 1995
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the
//    distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING
// IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package infectious

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

// FEC represents operations performed on a Reed-Solomon-based
// forward error correction code. Make sure to construct using NewFEC.
type FEC struct {
	k           int
	n           int
	enc_matrix  []byte
	vand_matrix []byte
}

// NewFEC creates a *FEC using k required pieces and n total pieces.
// Encoding data with this *FEC will generate n pieces, and decoding
// data requires k uncorrupted pieces. If during decode more than k pieces
// exist, corrupted data can be detected and recovered from.
func NewFEC(k, n int) (*FEC, error) {
	if k <= 0 || n <= 0 || k > 256 || n > 256 || k > n {
		return nil, errors.New("requires 1 <= k <= n <= 256")
	}

	enc_matrix := make([]byte, n*k)
	temp_matrix := make([]byte, n*k)
	createInvertedVdm(temp_matrix, k)

	for i := k * k; i < len(temp_matrix); i++ {
		temp_matrix[i] = gf_exp[((i/k)*(i%k))%255]
	}

	for i := range k {
		enc_matrix[i*(k+1)] = 1
	}

	for row := k * k; row < n*k; row += k {
		for col := range k {
			pa := temp_matrix[row:]
			pb := temp_matrix[col:]
			acc := byte(0)
			for i := 0; i < k; i, pa, pb = i+1, pa[1:], pb[k:] {
				acc ^= gf_mul_table[pa[0]][pb[0]]
			}
			enc_matrix[row+col] = acc
		}
	}

	// vand_matrix has more columns than rows
	// k rows, n columns.
	vand_matrix := make([]byte, k*n)
	vand_matrix[0] = 1
	g := byte(1)
	for row := range k {
		a := byte(1)
		for col := 1; col < n; col++ {
			vand_matrix[row*n+col] = a // 2.pow(i * j) FIGURE IT OUT
			a = gf_mul_table[g][a]
		}
		g = gf_mul_table[2][g]
	}

	return &FEC{
		k:           k,
		n:           n,
		enc_matrix:  enc_matrix,
		vand_matrix: vand_matrix,
	}, nil
}

// Required returns the number of required pieces for reconstruction. This is
// the k value passed to NewFEC.
func (f *FEC) Required() int {
	return f.k
}

// Total returns the number of total pieces that will be generated during
// encoding. This is the n value passed to NewFEC.
func (f *FEC) Total() int {
	return f.n
}

// Encode will take input data and encode to the total number of pieces n this
// *FEC is configured for. It will call the callback output n times.
//
// The input data must be a multiple of the required number of pieces k.
// Padding to this multiple is up to the caller.
//
// Note that the byte slices in Shares passed to output may be reused when
// output returns.
func (f *FEC) Encode(input []byte, output func(Share)) error {
	size := len(input)

	k := f.k
	n := f.n
	enc_matrix := f.enc_matrix

	if size%k != 0 {
		return fmt.Errorf("input length must be a multiple of %d", k)
	}

	block_size := size / k

	for i := range k {
		output(Share{
			Number: i,
			Data:   input[i*block_size : i*block_size+block_size]})
	}

	fec_buf := make([]byte, block_size)
	for i := k; i < n; i++ {
		clear(fec_buf)

		for j := range k {
			addmul(fec_buf, input[j*block_size:j*block_size+block_size],
				enc_matrix[i*k+j])
		}

		output(Share{
			Number: i,
			Data:   fec_buf})
	}
	return nil
}

// EncodeSingle will take input data and encode it to output only for the num
// piece.
//
// The input data must be a multiple of the required number of pieces k.
// Padding to this multiple is up to the caller.
//
// The output must be exactly len(input) / k bytes.
//
// The num must be 0 <= num < n.
func (f *FEC) EncodeSingle(input, output []byte, num int) error {
	size := len(input)

	k := f.k
	n := f.n
	enc_matrix := f.enc_matrix

	if num < 0 {
		return errors.New("num must be non-negative")
	}

	if num >= n {
		return fmt.Errorf("num must be less than %d", n)
	}

	if size%k != 0 {
		return fmt.Errorf("input length must be a multiple of %d", k)
	}

	block_size := size / k

	if len(output) != block_size {
		return fmt.Errorf("output length must be %d", block_size)
	}

	if num < k {
		copy(output, input[num*block_size:])
		return nil
	}

	clear(output)

	for i := range k {
		addmul(output, input[i*block_size:i*block_size+block_size],
			enc_matrix[num*k+i])
	}

	return nil
}

// A Share represents a piece of the FEC-encoded data.
// Both fields are required.
type Share struct {
	Number int
	Data   []byte
}

// DeepCopy makes getting a deep copy of a Share easier. It will return an
// identical Share that uses all new memory locations.
func (s *Share) DeepCopy() (c Share) {
	c.Number = s.Number
	c.Data = append([]byte(nil), s.Data...)
	return c
}

func sortByNumber(shares []Share) {
	slices.SortFunc(shares, func(a, b Share) int {
		return cmp.Compare(a.Number, b.Number)
	})
}

// checkShares validates shares that have already been sorted by number.
func checkShares(shares []Share, n int) error {
	for i := range shares {
		if shares[i].Number < 0 || shares[i].Number >= n {
			return fmt.Errorf("invalid share id: %d", shares[i].Number)
		}
		if i > 0 && shares[i].Number == shares[i-1].Number {
			return fmt.Errorf("duplicate share id: %d", shares[i].Number)
		}
		if len(shares[i].Data) != len(shares[0].Data) {
			return errors.New("shares must all be the same length")
		}
	}
	return nil
}

// Rebuild will take a list of corrected shares (pieces) and a callback output.
// output will be called k times ((*FEC).Required() times) with 1/k of the
// original data each time and the index of that data piece.
// Decode is usually preferred.
//
// Note that the data is not necessarily sent to output ordered by the piece
// number.
//
// Note that the byte slices in Shares passed to output may be reused when
// output returns.
//
// Note that Rebuild sorts the shares slice, reordering it in place.
//
// Rebuild assumes that you have already called Correct or did not need to.
func (f *FEC) Rebuild(shares []Share, output func(Share)) error {
	k := f.k
	n := f.n
	enc_matrix := f.enc_matrix

	if len(shares) < k {
		return NotEnoughShares
	}

	share_size := len(shares[0].Data)
	sortByNumber(shares)

	if err := checkShares(shares, n); err != nil {
		return err
	}

	m_dec := make([]byte, k*k)
	indexes := make([]int, k)
	sharesv := make([][]byte, k)

	shares_b_iter := 0
	shares_e_iter := len(shares) - 1

	for i := range k {
		var share_id int
		var share_data []byte

		if share := shares[shares_b_iter]; share.Number == i {
			share_id = share.Number
			share_data = share.Data
			shares_b_iter++
		} else {
			share := shares[shares_e_iter]
			share_id = share.Number
			share_data = share.Data
			shares_e_iter--
		}

		if share_id < k {
			m_dec[i*(k+1)] = 1
			if output != nil {
				output(Share{
					Number: share_id,
					Data:   share_data})
			}
		} else {
			copy(m_dec[i*k:i*k+k], enc_matrix[share_id*k:])
		}

		sharesv[i] = share_data
		indexes[i] = share_id
	}

	if err := invertMatrix(m_dec, k); err != nil {
		return err
	}

	buf := make([]byte, share_size)
	for i := range indexes {
		if indexes[i] >= k {
			clear(buf)

			for col := range k {
				addmul(buf, sharesv[col], m_dec[i*k+col])
			}

			if output != nil {
				output(Share{
					Number: i,
					Data:   buf})
			}
		}
	}
	return nil
}
