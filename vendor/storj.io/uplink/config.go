// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package uplink

import (
	"context"
	"net"
	"time"
	_ "unsafe" // for go:linkname

	"storj.io/common/peertls/tlsopts"
	"storj.io/common/rpc"
	"storj.io/common/rpc/rpcpool"
	"storj.io/common/useragent"
)

const defaultDialTimeout = 10 * time.Second

// Config defines configuration for using uplink library.
type Config struct {
	// UserAgent defines a registered partner's Value Attribution Code, and is used by the satellite to associate
	// a bucket with the partner at the time of bucket creation.
	// See https://docs.storj.io/dcs/how-tos/configure-tools-for-the-partner-program for info on the Partner Program.
	// UserAgent should follow https://tools.ietf.org/html/rfc7231#section-5.5.3.
	UserAgent string

	// DialTimeout defines how long client should wait for establishing
	// a connection to peers.
	// No explicit value or 0 means default 20s will be used. Value lower than 0 means there is no timeout.
	// DialTimeout is ignored if DialContext is provided.
	//
	// Deprecated: with the advent of Noise and TCP_FASTOPEN use, traditional dialing
	// doesn't necessarily happen anymore. This is already ignored for certain
	// connections and will be removed in a future release.
	DialTimeout time.Duration

	// DialContext is an extremely low level concern. It should almost certainly
	// remain unset so that this library can make informed choices about how to
	// talk to each node.
	// DialContext is how sockets are opened to nodes of all kinds and is called to
	// establish a connection. If DialContext is nil, it'll try to use the implementation
	// best suited for each node.
	//
	// Deprecated: this will be removed in a future release. All analyzed uses of
	// setting this value in open source projects are attempting to solve some more
	// nuanced problem (like QoS) which can only be handled for some types of
	// connections. This value is a hammer where we need a scalpel.
	DialContext func(ctx context.Context, network, address string) (net.Conn, error)

	// ChainPEM and KeyPEM are optional fields that specify the tls identity used by
	// the uplink while talking to other peers on the network. Don't set just one.
	// It is expected that generally these will be left unset and a new tls identity
	// will be generated.
	ChainPEM, KeyPEM []byte

	// satellitePool is a connection pool dedicated for satellite connections.
	// If not set, the normal pool / default will be used.
	satellitePool *rpcpool.Pool

	// pool is a connection pool for everything else (mainly for storagenode). Or everything if satellitePool is not set.
	// If nil, a default pool will be created.
	pool *rpcpool.Pool

	// maximumBufferSize is used to set the maximum buffer size for DRPC
	// connections/streams.
	maximumBufferSize int

	// disableObjectKeyEncryption disables the encryption of object keys for newly
	// uploaded objects.
	//
	// Disabling the encryption of object keys means that the object keys are
	// stored in plain text in the satellite database. This allows object listings
	// to be returned in lexicographically sorted order.
	//
	// Object content is still encrypted as usual.
	disableObjectKeyEncryption bool

	// disableBackgroundQoS tells the uplink library to not try setting background
	// QoS flags on the network sockets. This will impact the congestion control
	// profile as well.
	disableBackgroundQoS bool
}

// getDialer returns a new rpc.Dialer corresponding to the config.
func (config Config) getDialer(ctx context.Context) (_ rpc.Dialer, err error) {
	return config.getDialerForPool(ctx, nil)
}

// getDialerForPool returns a new rpc.Dialer corresponding to the config, using the chosen pool (or config.pool if pool is nil).
func (config Config) getDialerForPool(ctx context.Context, pool *rpcpool.Pool) (_ rpc.Dialer, err error) {
	var tlsOptions *tlsopts.Options
	if len(config.ChainPEM) > 0 && len(config.KeyPEM) > 0 {
		tlsOptions, err = getProcessTLSOptionsFromPEM(config.ChainPEM, config.KeyPEM)
	} else {
		tlsOptions, err = getProcessTLSOptions(ctx)
	}
	if err != nil {
		return rpc.Dialer{}, packageError.Wrap(err)
	}

	dialer := rpc.NewDefaultDialer(tlsOptions)
	if pool != nil {
		dialer.Pool = pool
	} else if config.pool != nil {
		dialer.Pool = config.pool
	} else {
		dialer.Pool = rpc.NewDefaultConnectionPool()
	}

	dialer.DialTimeout = config.DialTimeout
	dialer.AttemptBackgroundQoS = !config.disableBackgroundQoS

	if config.DialContext != nil {
		// N.B.: It is okay to use NewDefaultTCPConnector here because we explicitly don't want
		// NewHybridConnector. NewHybridConnector would not be able to use the user-provided
		// DialContext.
		//lint:ignore SA1019 deprecated okay,
		//nolint:staticcheck // deprecated okay.
		dialer.Connector = rpc.NewDefaultTCPConnector(config.DialContext)
	}

	dialer.ConnectionOptions.Manager.Stream.MaximumBufferSize = config.maximumBufferSize

	return dialer, nil
}

// NB: this is used with linkname in internal/expose.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:unused,revive
//go:linkname config_getDialer
func config_getDialer(config Config, ctx context.Context) (_ rpc.Dialer, err error) {
	return config.getDialer(ctx)
}

// setConnectionPool exposes setting connection pool.
//
// NB: this is used with linkname in internal/expose.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:unused
//go:linkname config_setConnectionPool
func config_setConnectionPool(config *Config, pool *rpcpool.Pool) { config.pool = pool }

// setSatelliteConnectionPool exposes setting connection pool for satellite.
//
// NB: this is used with linkname in internal/expose.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:unused
//go:linkname config_setSatelliteConnectionPool
func config_setSatelliteConnectionPool(config *Config, pool *rpcpool.Pool) {
	config.satellitePool = pool
}

// setMaximumBufferSize exposes setting maximumBufferSize.
//
// NB: this is used with linkname in internal/expose.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:unused
//go:linkname config_setMaximumBufferSize
func config_setMaximumBufferSize(config *Config, maximumBufferSize int) {
	config.maximumBufferSize = maximumBufferSize
}

// disableObjectKeyEncryption exposes setting disableObjectKeyEncryption.
//
// NB: this is used with linkname in internal/expose.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:unused
//go:linkname config_disableObjectKeyEncryption
func config_disableObjectKeyEncryption(config *Config) {
	config.disableObjectKeyEncryption = true
}

func (config Config) validateUserAgent(ctx context.Context) error {
	if len(config.UserAgent) == 0 {
		return nil
	}

	if _, err := useragent.ParseEntries([]byte(config.UserAgent)); err != nil {
		return err
	}

	return nil
}

// disableBackgroundQoS exposes setting Config.disableBackgroundQoS.
//
// NB: this is used with linkname in internal/expose.
// It needs to be updated when this is updated.
//
//lint:ignore U1000, used with linkname
//nolint:unused
//go:linkname config_disableBackgroundQoS
func config_disableBackgroundQoS(config *Config, disabled bool) {
	config.disableBackgroundQoS = disabled
}
