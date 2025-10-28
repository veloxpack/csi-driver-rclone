// Copyright (C) 2021 Storj Labs, Inc.
// See LICENSE for copying information.

package uplink

import (
	"context"
	"sync"
	"sync/atomic"

	"storj.io/common/identity"
	"storj.io/common/peertls/tlsopts"
)

var processTLSOptions struct {
	mu         sync.Mutex
	tlsOptions atomic.Pointer[tlsopts.Options]
}

var processTLSOptionsByPEM sync.Map

func getProcessTLSOptionsFromPEM(chainPEM, keyPEM []byte) (*tlsopts.Options, error) {
	if len(chainPEM) == 0 || len(keyPEM) == 0 {
		return nil, packageError.New("both chain and key PEM must be provided")
	}
	lookupKey := [2]*byte{&chainPEM[0], &keyPEM[0]}

	// first check the memoized value as a fast path with no locks
	tlsOptionsI, ok := processTLSOptionsByPEM.Load(lookupKey)
	if ok {
		return tlsOptionsI.(*tlsopts.Options), nil
	}

	// it was a miss so create an identity from the pem data and make the tls options
	ident, err := identity.FullIdentityFromPEM(chainPEM, keyPEM)
	if err != nil {
		return nil, packageError.Wrap(err)
	}
	tlsOptions, err := tlsOptionsFromIdentity(ident)
	if err != nil {
		return nil, packageError.Wrap(err)
	}

	tlsOptionsI, _ = processTLSOptionsByPEM.LoadOrStore(lookupKey, tlsOptions)
	return tlsOptionsI.(*tlsopts.Options), nil
}

func getProcessTLSOptions(ctx context.Context) (*tlsopts.Options, error) {
	if tlsOptions := processTLSOptions.tlsOptions.Load(); tlsOptions != nil {
		return tlsOptions, nil
	}

	processTLSOptions.mu.Lock()
	defer processTLSOptions.mu.Unlock()

	if tlsOptions := processTLSOptions.tlsOptions.Load(); tlsOptions != nil {
		return tlsOptions, nil
	}

	ident, err := identity.NewFullIdentity(ctx, identity.NewCAOptions{
		Difficulty:  0,
		Concurrency: 1,
	})
	if err != nil {
		return nil, packageError.Wrap(err)
	}

	tlsOptions, err := tlsOptionsFromIdentity(ident)
	if err != nil {
		return nil, packageError.Wrap(err)
	}

	processTLSOptions.tlsOptions.Store(tlsOptions)

	return tlsOptions, nil
}

func tlsOptionsFromIdentity(ident *identity.FullIdentity) (*tlsopts.Options, error) {
	tlsConfig := tlsopts.Config{
		UsePeerCAWhitelist: false,
		PeerIDVersions:     "0",
	}
	tlsOptions, err := tlsopts.NewOptions(ident, tlsConfig, nil)
	if err != nil {
		return nil, packageError.Wrap(err)
	}
	return tlsOptions, nil
}
