// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.

//go:build !amd64 && !arm64 && !loong64 && !ppc64 && !ppc64le && !riscv64 && !s390x

package hmacsha512

func block(dig *digest, p []byte) {
	blockGeneric(dig, p)
}
