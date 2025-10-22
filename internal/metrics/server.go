/*
Copyright 2025 Veloxpack.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package metrics implements the HTTP endpoint to serve Prometheus metrics
package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Options contains configuration for the metrics server
type Options struct {
	Enabled      bool
	MetricsPath  string
	MetricsPort  int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Validate checks if the options are valid
func (o *Options) Validate() error {
	if o.MetricsPort < 1 || o.MetricsPort > 65535 {
		return fmt.Errorf("metrics port must be between 1 and 65535, got %d", o.MetricsPort)
	}
	if !strings.HasPrefix(o.MetricsPath, "/") {
		return fmt.Errorf("metrics path must start with '/', got %q", o.MetricsPath)
	}
	if o.ReadTimeout < 0 {
		return fmt.Errorf("metrics read timeout cannot be negative")
	}
	if o.WriteTimeout < 0 {
		return fmt.Errorf("metrics write timeout cannot be negative")
	}
	if o.IdleTimeout < 0 {
		return fmt.Errorf("metrics idle timeout cannot be negative")
	}

	return nil
}

// NewOptions returns sensible defaults
func NewOptions() *Options {
	return &Options{
		MetricsPort:  9090,
		MetricsPath:  "/metrics",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// metricsServer represents a running metrics server
type metricsServer struct {
	srv      *http.Server
	listener net.Listener
	done     chan error
}

// Start creates and starts a new metrics server
// Returns nil if ListenAddr is empty
func Start(opt *Options) (*metricsServer, error) {
	if opt == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	listenAddr := fmt.Sprintf(":%d", opt.MetricsPort)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("GET %s", opt.MetricsPath), promhttp.Handler().ServeHTTP)

	s := &metricsServer{
		srv: &http.Server{
			Handler:      mux,
			ReadTimeout:  opt.ReadTimeout,
			WriteTimeout: opt.WriteTimeout,
			IdleTimeout:  opt.IdleTimeout,
		},
		listener: listener,
		done:     make(chan error, 1),
	}

	go func() {
		if err := s.srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.done <- err
		}
		close(s.done)
	}()

	return s, nil
}

// Addr returns the listening address
func (s *metricsServer) Addr() string {
	return s.listener.Addr().String()
}

// Wait blocks until the server stops
func (s *metricsServer) Wait() error {
	return <-s.done
}

// Shutdown gracefully stops the server
func (s *metricsServer) Shutdown(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	return s.srv.Shutdown(ctx)
}
