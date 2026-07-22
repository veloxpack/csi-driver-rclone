// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hmacsha512

import _ "unsafe"

// Unfortunately golang.org/x/sys/cpu does not have the correct value set for SHA512,
// so we need to go through this hacky route to set it correctly.
//
// See issue https://github.com/golang/go/issues/76221.
//
//go:linkname sysctlEnabled internal/cpu.sysctlEnabled
func sysctlEnabled(name []byte) bool

func init() {
	if !useSHA512 {
		useSHA512 = sysctlEnabled([]byte("hw.optional.armv8_2_sha512\x00"))
	}
}
