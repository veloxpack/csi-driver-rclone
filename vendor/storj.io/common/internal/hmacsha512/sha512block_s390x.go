// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

package hmacsha512

import (
	"golang.org/x/sys/cpu"
)

var useSHA512 = cpu.S390X.HasSHA512

//go:noescape
func blockS390X(dig *digest, p []byte)

func block(dig *digest, p []byte) {
	if useSHA512 {
		blockS390X(dig, p)
	} else {
		blockGeneric(dig, p)
	}
}
