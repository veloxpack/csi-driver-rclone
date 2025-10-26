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
package metricsserver

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
	MetricsPath  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MetricsAddr  string
}

// Validate checks if the options are valid
func (o *Options) Validate() error {
	if o == nil {
		return fmt.Errorf("options cannot be nil")
	}

	if o.MetricsAddr == "" {
		return fmt.Errorf("metrics address cannot be empty")
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

	listener, err := net.Listen("tcp", opt.MetricsAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", opt.MetricsAddr, err)
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
