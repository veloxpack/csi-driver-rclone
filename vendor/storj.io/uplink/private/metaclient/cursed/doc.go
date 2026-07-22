// Copyright (C) 2025 Storj Labs, Inc.
// See LICENSE for copying information.

//go:generate go run gen.go

// Package cursed copies the code from the storj module necessary to unpack
// and repack internal satellite segment ids.
package cursed

import (
	"github.com/spacemonkeygo/monkit/v3"

	"storj.io/common/storj"
)

var mon = monkit.Package()

// PieceID is an alias to storj.PieceID for use in generated protobuf code.
type PieceID = storj.PieceID
