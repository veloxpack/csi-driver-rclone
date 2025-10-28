// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/zeebo/errs"
)

// RedundancyScheme specifies the parameters and the algorithm for redundancy.
type RedundancyScheme struct {
	// Algorithm determines the algorithm to be used for redundancy.
	Algorithm RedundancyAlgorithm

	// ShareSize is the size in bytes for each erasure shares.
	ShareSize int32

	// RequiredShares is the minimum number of shares required to recover a
	// stripe, reed-solomon k.
	RequiredShares int16
	// RepairShares is the minimum number of safe shares that can remain
	// before a repair is triggered.
	RepairShares int16
	// OptimalShares is the desired total number of shares for a segment.
	OptimalShares int16
	// TotalShares is the number of shares to encode. If it is larger than
	// OptimalShares, slower uploads of the excess shares will be aborted in
	// order to improve performance.
	TotalShares int16
}

// IsZero returns true if no field in the struct is set to non-zero value.
func (scheme RedundancyScheme) IsZero() bool {
	return scheme == (RedundancyScheme{})
}

// StripeSize is the number of bytes for a stripe.
// Stripes are erasure encoded and split into n shares, where we need k to
// reconstruct the stripe. Therefore a stripe size is the erasure share size
// times the required shares, k.
func (scheme RedundancyScheme) StripeSize() int32 {
	return scheme.ShareSize * int32(scheme.RequiredShares)
}

// StripeCount returns segment's total number of stripes based on segment's encrypted size.
func (scheme RedundancyScheme) StripeCount(encryptedSegmentSize int32) int32 {
	stripeSize := scheme.StripeSize()
	return (encryptedSegmentSize + stripeSize - 1) / stripeSize
}

// PieceSize calculates piece size for give size.
func (scheme RedundancyScheme) PieceSize(size int64) int64 {
	const uint32Size = 4
	stripeSize := int64(scheme.StripeSize())
	stripes := (size + uint32Size + stripeSize - 1) / stripeSize

	encodedSize := stripes * int64(scheme.StripeSize())
	pieceSize := encodedSize / int64(scheme.RequiredShares)
	return pieceSize
}

// Value implements the driver.Valuer interface.
func (scheme RedundancyScheme) Value() (driver.Value, error) {
	switch {
	case scheme.ShareSize < 0 || scheme.ShareSize >= 1<<24:
		return nil, errs.New("invalid share size %v", scheme.ShareSize)
	case scheme.RequiredShares < 0 || scheme.RequiredShares >= 1<<8:
		return nil, errs.New("invalid required shares %v", scheme.RequiredShares)
	case scheme.RepairShares < 0 || scheme.RepairShares >= 1<<8:
		return nil, errs.New("invalid repair shares %v", scheme.RepairShares)
	case scheme.OptimalShares < 0 || scheme.OptimalShares >= 1<<8:
		return nil, errs.New("invalid optimal shares %v", scheme.OptimalShares)
	case scheme.TotalShares < 0 || scheme.TotalShares >= 1<<8:
		return nil, errs.New("invalid total shares %v", scheme.TotalShares)
	}

	var bytes [8]byte
	bytes[0] = byte(scheme.Algorithm)

	// little endian uint32
	bytes[1] = byte(scheme.ShareSize >> 0)
	bytes[2] = byte(scheme.ShareSize >> 8)
	bytes[3] = byte(scheme.ShareSize >> 16)

	bytes[4] = byte(scheme.RequiredShares)
	bytes[5] = byte(scheme.RepairShares)
	bytes[6] = byte(scheme.OptimalShares)
	bytes[7] = byte(scheme.TotalShares)

	return int64(binary.LittleEndian.Uint64(bytes[:])), nil
}

// Scan implements the sql.Scanner interface.
func (scheme *RedundancyScheme) Scan(value any) error {
	switch value := value.(type) {
	case int64:
		var bytes [8]byte
		binary.LittleEndian.PutUint64(bytes[:], uint64(value))

		scheme.Algorithm = RedundancyAlgorithm(bytes[0])

		// little endian uint32
		scheme.ShareSize = int32(bytes[1]) | int32(bytes[2])<<8 | int32(bytes[3])<<16

		scheme.RequiredShares = int16(bytes[4])
		scheme.RepairShares = int16(bytes[5])
		scheme.OptimalShares = int16(bytes[6])
		scheme.TotalShares = int16(bytes[7])

		return nil
	default:
		return errs.New("unable to scan %T into RedundancyScheme", value)
	}
}

// DecodeSpanner implements spanner.Decoder.
func (scheme *RedundancyScheme) DecodeSpanner(val any) (err error) {
	if v, ok := val.(string); ok {
		val, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return errs.New("unable to scan %T into RedundancyScheme: %v", val, err)
		}
	}
	return scheme.Scan(val)
}

// EncodeSpanner implements spanner.Encoder.
func (scheme RedundancyScheme) EncodeSpanner() (any, error) {
	return scheme.Value()
}

// String returns the string representation of scheme.
// It satisfies the fmt.Stringer interface.
func (scheme RedundancyScheme) String() string {
	var algorithm string
	switch scheme.Algorithm {
	case InvalidRedundancyAlgorithm:
		algorithm = "XX"
	case ReedSolomon:
		algorithm = "RS"
	default:
		algorithm = fmt.Sprintf("unknown RedundancyScheme(%v)", scheme.Algorithm)
	}

	return fmt.Sprintf("%s:%d/%d/%d/%d",
		algorithm, scheme.RequiredShares, scheme.RepairShares, scheme.OptimalShares, scheme.TotalShares,
	)
}

// RedundancyAlgorithm is the algorithm used for redundancy.
type RedundancyAlgorithm byte

// List of supported redundancy algorithms.
const (
	InvalidRedundancyAlgorithm = RedundancyAlgorithm(iota)
	ReedSolomon
)
