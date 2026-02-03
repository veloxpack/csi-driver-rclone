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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	metricsserver "github.com/veloxpack/csi-driver-rclone/internal/metrics"
	rcserver "github.com/veloxpack/csi-driver-rclone/internal/rc"
	"github.com/veloxpack/csi-driver-rclone/pkg/rclone"
	"k8s.io/klog/v2"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID     = flag.String("nodeid", "", "node id")
	driverName = flag.String("drivername", rclone.DefaultDriverName, "name of the driver")
	remount    = flag.Bool("remount", false, "remount existing volume mount points on startup")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")

	metricsOpts := metricsserver.NewOptions()
	rcOpts := rcserver.NewOptions()

	// Metrics Options
	flag.StringVar(&metricsOpts.MetricsAddr, "metrics-addr",
		metricsOpts.MetricsAddr, "Metrics server listening address")
	flag.StringVar(&metricsOpts.MetricsPath, "metrics-path",
		metricsOpts.MetricsPath, "HTTP path where metrics are exposed")
	flag.DurationVar(&metricsOpts.ReadTimeout, "metrics-server-read-timeout",
		metricsOpts.ReadTimeout, "Metrics server read timeout")
	flag.DurationVar(&metricsOpts.WriteTimeout, "metrics-server-write-timeout",
		metricsOpts.WriteTimeout, "Metrics server write timeout")
	flag.DurationVar(&metricsOpts.IdleTimeout, "metrics-server-idle-timeout",
		metricsOpts.IdleTimeout, "Metrics server idle timeout")
	// RC Options
	flag.BoolVar(&rcOpts.Enabled, "rc",
		rcOpts.Enabled, "Enable rclone Remote Control (RC) API")
	flag.StringVar(&rcOpts.Address, "rc-addr",
		rcOpts.Address, "RC server listening address")
	flag.BoolVar(&rcOpts.NoAuth, "rc-no-auth",
		rcOpts.NoAuth, "Disable authentication for RC (insecure)")
	flag.StringVar(&rcOpts.Username, "rc-user",
		rcOpts.Username, "RC basic auth username")
	flag.StringVar(&rcOpts.Password, "rc-pass",
		rcOpts.Password, "RC basic auth password")

	flag.Parse()

	if *nodeID == "" {
		klog.Warning("nodeid is empty")
	}

	ctx := context.Background()

	// Track servers for shutdown (using interface{} since they have different Shutdown signatures)
	var metricsSrv interface {
		Addr() string
		Shutdown(context.Context) error
	}
	var rcSrv rcserver.Server

	// Start metrics server if enabled
	if metricsOpts.MetricsAddr != "" {
		// Start metrics server
		srv, err := metricsserver.Start(metricsOpts)
		if err != nil {
			klog.Fatalf("Failed to start metrics server: %v", err)
		}
		if srv != nil {
			metricsSrv = srv
			klog.Infof("Metrics server listening on http://%s%s", srv.Addr(), metricsOpts.MetricsPath)
		}
	}

	// Start RC server if enabled
	if rcOpts.Enabled {
		srv, err := rcserver.Start(ctx, rcOpts)
		if err != nil {
			klog.Fatalf("Failed to start RC server: %v", err)
		}
		if srv != nil {
			rcSrv = srv
			klog.Infof("RC server listening on %s", rcOpts.Address)
		}
	}

	driverOptions := rclone.DriverOptions{
		NodeID:     *nodeID,
		DriverName: *driverName,
		Endpoint:   *endpoint,
		Remount:    *remount,
	}

	driver := rclone.NewDriver(&driverOptions)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGTERM, // Kubernetes sends this for graceful shutdown
		syscall.SIGINT,  // Ctrl+C for local testing
		syscall.SIGUSR1, // Custom: dump mount info
		syscall.SIGUSR2, // Custom: force cache sync
	)

	// Start driver in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		driver.Run(false)
	}()

	// Wait for signal
	sig := <-sigChan
	klog.Infof("Received signal: %v", sig)

	// Handle different signals
	switch sig {
	case syscall.SIGTERM, syscall.SIGINT:
		// Graceful shutdown
		klog.Infof("Starting graceful shutdown...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()

		// Shutdown metrics server
		if metricsSrv != nil {
			klog.V(2).Info("Shutting down metrics server...")
			metricsShutdownCtx, metricsCancel := context.WithTimeout(ctx, 5*time.Second)
			defer metricsCancel()
			if err := metricsSrv.Shutdown(metricsShutdownCtx); err != nil {
				klog.Errorf("Error shutting down metrics server: %v", err)
			}
		}

		// Shutdown RC server
		if rcSrv != nil {
			klog.V(2).Info("Shutting down RC server...")
			if err := rcSrv.Shutdown(); err != nil {
				klog.Errorf("Error shutting down RC server: %v", err)
			}
		}

		// Perform driver shutdown
		if err := driver.Shutdown(shutdownCtx); err != nil {
			klog.Errorf("Error during driver shutdown: %v", err)
			os.Exit(1)
		}

		klog.Info("Graceful shutdown completed")
		os.Exit(0)

	case syscall.SIGUSR1:
		// Dump mount information (non-terminating)
		klog.Info("=== MOUNT STATUS DUMP (SIGUSR1) ===")
		driver.DumpMountInfo()
		// Continue running after dump
		wg.Wait()

	case syscall.SIGUSR2:
		// Force cache sync (non-terminating)
		klog.Info("=== FORCING CACHE SYNC (SIGUSR2) ===")
		if err := driver.ForceCacheSync(ctx); err != nil {
			klog.Errorf("Cache sync failed: %v", err)
		} else {
			klog.Info("Cache sync completed successfully")
		}
		// Continue running after sync
		wg.Wait()
	}
}
