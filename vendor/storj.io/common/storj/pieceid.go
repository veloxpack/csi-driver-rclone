// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base64"
	"encoding/binary"

	"github.com/zeebo/errs"
)

// ErrPieceID is used when something goes wrong with a piece ID.
var ErrPieceID = errs.Class("piece ID")

// PieceID is the unique identifier for pieces.
type PieceID [32]byte

// NewPieceID creates a piece ID.
func NewPieceID() PieceID {
	var id PieceID

	_, err := rand.Read(id[:])
	if err != nil {
		panic(err)
	}

	return id
}

// PieceIDFromString decodes a hex encoded piece ID string.
func PieceIDFromString(s string) (PieceID, error) {
	idBytes, err := base32Encoding.DecodeString(s)
	if err != nil {
		return PieceID{}, ErrPieceID.Wrap(err)
	}
	return PieceIDFromBytes(idBytes)
}

// PieceIDFromBytes converts a byte slice into a piece ID.
func PieceIDFromBytes(b []byte) (PieceID, error) {
	if len(b) != len(PieceID{}) {
		return PieceID{}, ErrPieceID.New("not enough bytes to make a piece ID; have %d, need %d", len(b), len(PieceID{}))
	}

	var id PieceID
	copy(id[:], b)
	return id, nil
}

// IsZero returns whether piece ID is unassigned.
func (id *PieceID) IsZero() bool {
	// Using `id == PieceID{}` is significantly slower due to more complex generated code.
	return binary.LittleEndian.Uint64(id[0:8])|
		binary.LittleEndian.Uint64(id[8:16])|
		binary.LittleEndian.Uint64(id[16:24])|
		binary.LittleEndian.Uint64(id[24:32]) == 0
}

// String representation of the piece ID.
func (id PieceID) String() string { return base32Encoding.EncodeToString(id.Bytes()) }

// Bytes returns bytes of the piece ID.
func (id PieceID) Bytes() []byte { return id[:] }

// Derive a new PieceID from the current piece ID, the given storage node ID and piece number.
func (id PieceID) Derive(storagenodeID NodeID, pieceNum int32) PieceID {
	return id.Deriver().Derive(storagenodeID, pieceNum)
}

// Marshal serializes a piece ID.
func (id PieceID) Marshal() ([]byte, error) {
	return id.Bytes(), nil
}

// MarshalTo serializes a piece ID into the passed byte slice.
func (id *PieceID) MarshalTo(data []byte) (n int, err error) {
	n = copy(data, id.Bytes())
	return n, nil
}

// Unmarshal deserializes a piece ID.
func (id *PieceID) Unmarshal(data []byte) error {
	var err error
	*id, err = PieceIDFromBytes(data)
	return err
}

// Size returns the length of a piece ID (implements gogo's custom type interface).
func (id *PieceID) Size() int {
	return len(id)
}

// MarshalText serializes a piece ID to a base32 string.
func (id PieceID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText deserializes a base32 string to a piece ID.
func (id *PieceID) UnmarshalText(data []byte) error {
	var err error
	*id, err = PieceIDFromString(string(data))
	if err != nil {
		return err
	}
	return nil
}

// Value set a PieceID to a database field.
func (id PieceID) Value() (driver.Value, error) {
	return id.Bytes(), nil
}

// Scan extracts a PieceID from a database field.
func (id *PieceID) Scan(src any) (err error) {
	b, ok := src.([]byte)
	if !ok {
		return ErrPieceID.New("PieceID Scan expects []byte")
	}
	n, err := PieceIDFromBytes(b)
	*id = n
	return err
}

// DecodeSpanner implements spanner.Decoder.
func (id *PieceID) DecodeSpanner(val any) (err error) {
	if v, ok := val.(string); ok {
		var buffer [32]byte
		b, err := base64.StdEncoding.AppendDecode(buffer[:0], []byte(v))
		if err != nil {
			return err
		}
		x, err := PieceIDFromBytes(b)
		if err != nil {
			return ErrPieceID.Wrap(err)
		}
		*id = x
		return nil
	}
	return id.Scan(val)
}

// EncodeSpanner implements spanner.Encoder.
func (id PieceID) EncodeSpanner() (any, error) {
	return id.Value()
}
