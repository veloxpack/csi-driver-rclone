// Copyright (C) 2023 Storj Labs, Inc.
// See LICENSE for copying information.

package signing

import (
	"context"
	"os"
	"strconv"
)

var (
	disabledGlobally, _ = strconv.ParseBool(os.Getenv("STORJ_DISABLE_SIGNING"))

	// Takes the place of signatures when they are disabled. The presence
	// of a "signature" is required because some components check that the
	// signature field has len > 0.
	disabledSignature = []byte("DISABLED-SIGNATURE")
)

type signaturesDisabledForTestKey struct{}

// Disabled returns true if signatures are disabled. If disabled, signatures
// are set to "DISABLED-SIGNATURE" and are ignored during verification.
func Disabled() bool {
	return disabledGlobally
}

func areSignaturesDisabled(ctx context.Context) bool {
	if disabledGlobally {
		return true
	}
	_, disabledForTest := ctx.Value(signaturesDisabledForTestKey{}).(struct{})
	return disabledForTest
}

func withSignaturesDisabledForTest(ctx context.Context) context.Context {
	return context.WithValue(ctx, signaturesDisabledForTestKey{}, struct{}{})
}
