// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GO_LICENSE file.


#include "textflag.h"

// func blockS390X(dig *digest, p []byte)
TEXT ·blockS390X(SB), NOSPLIT|NOFRAME, $0-32
	LMG    dig+0(FP), R1, R3            // R2 = &p[0], R3 = len(p)
	MOVBZ  $3, R0                       // SHA-512 function code

loop:
	KIMD R0, R2      // compute intermediate message digest (KIMD)
	BVS  loop        // continue if interrupted
	RET
