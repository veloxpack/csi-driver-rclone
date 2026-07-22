// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

package hmacsha512

import (
	"golang.org/x/sys/cpu"
)

var useAVX2 = cpu.X86.HasAVX && cpu.X86.HasAVX2 && cpu.X86.HasBMI2

//go:noescape
func blockAVX2(dig *digest, p []byte)

func block(dig *digest, p []byte) {
	if useAVX2 {
		blockAVX2(dig, p)
	} else {
		blockGeneric(dig, p)
	}
}
