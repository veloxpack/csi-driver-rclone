// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"storj.io/common/sync2/race2"
	"storj.io/infectious"
)

type unsafeRSScheme struct {
	fc               *infectious.FEC
	erasureShareSize int
}

// NewUnsafeRSScheme returns a Reed-Solomon-based ErasureScheme without error correction.
func NewUnsafeRSScheme(fc *infectious.FEC, erasureShareSize int) ErasureScheme {
	return &unsafeRSScheme{fc: fc, erasureShareSize: erasureShareSize}
}

func (s *unsafeRSScheme) EncodeSingle(input, output []byte, num int) (err error) {
	return s.fc.EncodeSingle(input, output, num)
}

func (s *unsafeRSScheme) Encode(input []byte, output func(num int, data []byte)) (
	err error) {
	return s.fc.Encode(input, func(s infectious.Share) {
		output(s.Number, s.Data)
	})
}

func (s *unsafeRSScheme) Decode(out []byte, in []infectious.Share) ([]byte, error) {
	for _, share := range in {
		race2.ReadSlice(share.Data)
	}

	race2.WriteSlice(out)
	expectedCap := s.RequiredCount() * s.ErasureShareSize()
	if cap(out) < expectedCap {
		out = make([]byte, expectedCap)
	} else {
		out = out[:expectedCap]
	}
	err := s.fc.Rebuild(in, func(share infectious.Share) {
		copy(out[share.Number*s.ErasureShareSize():], share.Data)
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *unsafeRSScheme) Rebuild(in []infectious.Share, out func(infectious.Share)) error {
	for _, v := range in {
		race2.ReadSlice(v.Data)
	}
	return s.fc.Rebuild(in, out)
}

func (s *unsafeRSScheme) ErasureShareSize() int {
	return s.erasureShareSize
}

func (s *unsafeRSScheme) StripeSize() int {
	return s.erasureShareSize * s.fc.Required()
}

func (s *unsafeRSScheme) TotalCount() int {
	return s.fc.Total()
}

func (s *unsafeRSScheme) RequiredCount() int {
	return s.fc.Required()
}
