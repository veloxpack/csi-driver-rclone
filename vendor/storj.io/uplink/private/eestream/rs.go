// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package eestream

import (
	"storj.io/common/sync2/race2"
	"storj.io/infectious"
)

type rsScheme struct {
	fc               *infectious.FEC
	erasureShareSize int
}

// NewRSScheme returns a Reed-Solomon-based ErasureScheme.
func NewRSScheme(fc *infectious.FEC, erasureShareSize int) ErasureScheme {
	return &rsScheme{fc: fc, erasureShareSize: erasureShareSize}
}

func (s *rsScheme) EncodeSingle(input, output []byte, num int) (err error) {
	return s.fc.EncodeSingle(input, output, num)
}

func (s *rsScheme) Encode(input []byte, output func(num int, data []byte)) (
	err error) {
	return s.fc.Encode(input, func(s infectious.Share) {
		output(s.Number, s.Data)
	})
}

func (s *rsScheme) Decode(out []byte, in []infectious.Share) ([]byte, error) {
	for _, share := range in {
		race2.ReadSlice(share.Data)
	}
	race2.WriteSlice(out)
	return s.fc.Decode(out, in)
}

func (s *rsScheme) Rebuild(in []infectious.Share, out func(infectious.Share)) error {
	for _, v := range in {
		race2.ReadSlice(v.Data)
	}
	return s.fc.Rebuild(in, out)
}

func (s *rsScheme) ErasureShareSize() int {
	return s.erasureShareSize
}

func (s *rsScheme) StripeSize() int {
	return s.erasureShareSize * s.fc.Required()
}

func (s *rsScheme) TotalCount() int {
	return s.fc.Total()
}

func (s *rsScheme) RequiredCount() int {
	return s.fc.Required()
}
