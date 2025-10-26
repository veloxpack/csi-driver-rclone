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
	"time"

	metricsserver "github.com/veloxpack/csi-driver-rclone/internal/metrics"
	"github.com/veloxpack/csi-driver-rclone/pkg/rclone"
	"k8s.io/klog/v2"
)

var (
	endpoint   = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeID     = flag.String("nodeid", "", "node id")
	driverName = flag.String("drivername", rclone.DefaultDriverName, "name of the driver")
)

func main() {
	klog.InitFlags(nil)
	_ = flag.Set("logtostderr", "true")

	metricsOpts := metricsserver.NewOptions()

	// Metrics Options
	flag.StringVar(&metricsOpts.MetricsAddr, "metrics-addr", metricsOpts.MetricsAddr, "Metrics server listening address")
	flag.StringVar(&metricsOpts.MetricsPath, "metrics-path", metricsOpts.MetricsPath, "HTTP path where metrics are exposed")
	flag.DurationVar(&metricsOpts.ReadTimeout, "metrics-server-read-timeout", metricsOpts.ReadTimeout, "Metrics server read timeout")
	flag.DurationVar(&metricsOpts.WriteTimeout, "metrics-server-write-timeout", metricsOpts.WriteTimeout, "Metrics server write timeout")
	flag.DurationVar(&metricsOpts.IdleTimeout, "metrics-serve-idle-timeout", metricsOpts.IdleTimeout, "Metrics server idle timeout")

	flag.Parse()

	if *nodeID == "" {
		klog.Warning("nodeid is empty")
	}

	// Start metrics server if enabled
	if metricsOpts.MetricsAddr != "" {
		ctx := context.Background()

		// Start metrics server
		metricsSrv, err := metricsserver.Start(metricsOpts)
		if err != nil {
			klog.Fatalf("Failed to start metrics server: %v", err)
		}
		if metricsSrv != nil {
			klog.Infof("Metrics server listening on http://%s%s", metricsSrv.Addr(), metricsOpts.MetricsPath)
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
					klog.Errorf("Error shutting down metrics server: %v", err)
				}
			}()
		}
	}

	driverOptions := rclone.DriverOptions{
		NodeID:     *nodeID,
		DriverName: *driverName,
		Endpoint:   *endpoint,
	}

	driver := rclone.NewDriver(&driverOptions)
	driver.Run(false)
	os.Exit(0)
}
