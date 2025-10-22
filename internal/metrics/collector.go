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

package metrics

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rclone/rclone/fs/accounting"
	"k8s.io/klog/v2"
)

var (
	// csiCollector is the global CSI metrics collector instance.
	csiCollector *csiRcloneCollector
)

// csiRcloneCollector implements the Prometheus Collector interface
// for CSI-specific metrics.
type csiRcloneCollector struct {
	ctx        context.Context
	nodeID     string
	driverName string
	endpoint   string
	nodeInfo   *prometheus.Desc
}

// NewCollector creates and returns a new CSI metrics collector instance.
func NewCollector(ctx context.Context, nodeID, driverName, endpoint string) *csiRcloneCollector {
	namespace := "csi_driver_rclone_"

	return &csiRcloneCollector{
		ctx:        ctx,
		nodeID:     nodeID,
		driverName: driverName,
		endpoint:   endpoint,
		nodeInfo: prometheus.NewDesc(
			namespace+"driver_info",
			"Information about the CSI driver",
			[]string{"node_id", "driver_name", "endpoint"},
			nil,
		),
	}
}

// InitCollector initializes and registers both the rclone and CSI Prometheus collectors.
// It ensures that collectors are only initialized once.
func InitCollector(ctx context.Context, nodeID, driverName, endpoint string) error {
	if csiCollector != nil {
		klog.V(4).Info("CSI collector already initialized; skipping re-initialization")
		return nil
	}

	// Register rclone collector
	rcloneCollector := accounting.NewRcloneCollector(ctx)
	if err := prometheus.Register(rcloneCollector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			klog.V(4).Info("rclone Prometheus collector already registered")
		} else {
			klog.Warningf("failed to register rclone Prometheus collector: %v", err)
		}
	}

	// Create and register CSI collector
	csiCollector = NewCollector(ctx, nodeID, driverName, endpoint)
	if err := prometheus.Register(csiCollector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			klog.V(4).Info("CSI Prometheus collector already registered")
		} else {
			return fmt.Errorf("failed to register CSI Prometheus collector: %w", err)
		}
	}

	klog.Infof("CSI Prometheus collector initialized for node: %s", nodeID)
	return nil
}

// Collector returns the global CSI collector instance, or nil if not initialized.
func Collector() *csiRcloneCollector {
	return csiCollector
}

// Describe implements prometheus.Collector.
func (c *csiRcloneCollector) Describe(ch chan<- *prometheus.Desc) {
	// No static metrics to describe.
}

// Collect implements prometheus.Collector.
func (c *csiRcloneCollector) Collect(ch chan<- prometheus.Metric) {
	if c == nil {
		klog.Warning("CSI collector is nil; skipping collection")
		return
	}

	// Create the node info metric with labels
	ch <- prometheus.MustNewConstMetric(
		c.nodeInfo,
		prometheus.GaugeValue,
		1,
		c.nodeID,
		c.driverName,
		c.endpoint,
	)

	klog.V(5).Infof("Collected CSI metrics for node: %s", c.nodeID)
}
