# Metrics Overlay

This overlay enables Prometheus metrics collection for the CSI Rclone driver.

## What This Overlay Adds

- **Metrics endpoint** on the CSI node DaemonSet (port 5572)
- **Service** to expose metrics (`csi-rclone-node-metrics`)
- **ServiceMonitor** for Prometheus Operator to auto-discover and scrape metrics

## Prerequisites

This overlay requires the Prometheus Operator to be installed in your cluster, which provides the `ServiceMonitor` CRD.

### Install kube-prometheus-stack

Using Helm:
```bash
# Add Prometheus community Helm repository
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install the kube-prometheus-stack (which includes ServiceMonitor CRDs)
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace
```

This installs:
- Prometheus Operator
- Prometheus server
- Alertmanager
- Grafana
- Various exporters and dashboards

## Deploy CSI Driver with Metrics

```bash
# Deploy using kustomize
kubectl apply -k deploy/overlays/metrics

# Or using skaffold for development
skaffold dev -p metrics
```

## Verify Metrics Collection

### Check ServiceMonitor

```bash
# Verify ServiceMonitor exists with correct labels
kubectl get servicemonitor -n veloxpack csi-rclone-node-metrics -o yaml

# Ensure it has the label: release: kube-prometheus-stack
```

### Check Prometheus Targets

```bash
# Port-forward to Prometheus
kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090:9090

# Open http://localhost:9090/targets
# Look for "csi-rclone-node-metrics" target (should be "UP")
```

### Query Metrics

Using Prometheus UI (http://localhost:9090):
```promql
# Total bytes transferred
rclone_bytes_transferred_total

# Files transferred
rclone_files_transferred_total

# Transfer errors
rclone_errors_total

# HTTP status codes
rclone_http_status_code
```

Or using curl:
```bash
curl -s 'http://localhost:9090/api/v1/query?query=rclone_bytes_transferred_total' | jq .
```

## Access Grafana

```bash
# Port-forward to Grafana
kubectl port-forward -n monitoring svc/kube-prometheus-stack-grafana 3000:80

# Open http://localhost:3000
# Default credentials: admin / prom-operator
```

### Import Production Dashboard

This repository includes a production-grade Grafana dashboard (`grafana-dashboard.json`) designed specifically for the CSI Rclone driver.

**Dashboard Features:**
- **Overview & Health Status**: Real-time health checks, active mounts, error rates
- **VFS Cache Performance**: File handles, disk cache usage, metadata cache, upload queues
- **Transfer Statistics**: Transfer speed, data volumes, file operations, server-side operations
- **Mount Health Details**: Detailed mount information with health timeline
- **System Resources**: CPU, memory, and goroutine monitoring

**Import the Dashboard:**

Method 1 - Via UI:
1. Open Grafana (http://localhost:3000)
2. Navigate to **Dashboards** → **Import**
3. Click **Upload JSON file**
4. Select `deploy/overlays/metrics/grafana-dashboard.json`
5. Select your Prometheus datasource
6. Click **Import**

Method 2 - Via ConfigMap (GitOps):
```bash
kubectl create configmap csi-rclone-dashboard \
  --from-file=grafana-dashboard.json=deploy/overlays/metrics/grafana-dashboard.json \
  -n monitoring \
  --dry-run=client -o yaml | kubectl apply -f -

# Add label for Grafana sidecar to discover it
kubectl label configmap csi-rclone-dashboard \
  grafana_dashboard=1 \
  -n monitoring
```

Method 3 - Direct API:
```bash
# Get Grafana admin password
kubectl get secret -n monitoring kube-prometheus-stack-grafana \
  -o jsonpath="{.data.admin-password}" | base64 --decode

# Import via API
curl -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -u admin:PASSWORD \
  -d @deploy/overlays/metrics/grafana-dashboard.json
```

**Dashboard Template Variables:**
- `datasource`: Prometheus datasource (auto-populated)
- `job`: Service job name (default: `csi-rclone-node-metrics`)
- `namespace`: Kubernetes namespace (default: `veloxpack`)
- `volume`: Filter by volume ID (multi-select, default: All)

## Available Metrics

The CSI driver exposes the following Prometheus metrics:

### CSI Driver Specific Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `csi_driver_info` | Gauge | `node_id`, `driver_name`, `endpoint`, `rclone_version`, `driver_version` | Information about the CSI driver |
| `csi_driver_vfs_file_handles_in_use` | Gauge | `volume_id`, `remote_name` | Number of file handles currently in use |
| `csi_driver_vfs_metadata_cache_dirs_total` | Gauge | `volume_id`, `remote_name` | Number of directories in metadata cache |
| `csi_driver_vfs_metadata_cache_files_total` | Gauge | `volume_id`, `remote_name` | Number of files in metadata cache |
| `csi_driver_vfs_disk_cache_bytes_used` | Gauge | `volume_id`, `remote_name` | Bytes used by VFS disk cache |
| `csi_driver_vfs_disk_cache_files_total` | Gauge | `volume_id`, `remote_name` | Number of files in VFS disk cache |
| `csi_driver_vfs_disk_cache_errored_files_total` | Counter | `volume_id`, `remote_name` | Number of files with errors in disk cache |
| `csi_driver_vfs_uploads_in_progress_total` | Gauge | `volume_id`, `remote_name` | Number of uploads currently in progress |
| `csi_driver_vfs_uploads_queued_total` | Gauge | `volume_id`, `remote_name` | Number of uploads queued for processing |
| `csi_driver_vfs_disk_cache_out_of_space` | Gauge | `volume_id`, `remote_name` | Disk cache out of space indicator (1=yes, 0=no) |
| `csi_driver_mount_healthy` | Gauge | `volume_id`, `pod_id`, `target_path`, `remote_name`, `mount_type`, `device_name`, `volume_name`, `read_only`, `mount_duration_seconds` | Mount health status (1=healthy, 0=unhealthy) |
| `csi_driver_remote_transfer_speed_bytes_per_second` | Gauge | - | Current transfer speed in bytes per second |
| `csi_driver_remote_transfer_eta_seconds` | Gauge | - | Estimated time to completion in seconds |
| `csi_driver_remote_checks_total` | Counter | - | Total number of file checks completed |
| `csi_driver_remote_deletes_total` | Counter | - | Total number of files deleted |
| `csi_driver_remote_server_side_copies_total` | Counter | - | Total number of server-side copies |
| `csi_driver_remote_server_side_moves_total` | Counter | - | Total number of server-side moves |
| `csi_driver_remote_transferring_files` | Gauge | - | Number of files currently being transferred |
| `csi_driver_remote_checking_files` | Gauge | - | Number of files currently being checked |

### Rclone Built-in Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `rclone_bytes_transferred_total` | Counter | Total bytes transferred by rclone |
| `rclone_checked_files_total` | Counter | Total files checked |
| `rclone_files_transferred_total` | Counter | Total files transferred |
| `rclone_files_deleted_total` | Counter | Total files deleted |
| `rclone_files_renamed_total` | Counter | Total files renamed |
| `rclone_dirs_deleted_total` | Counter | Total directories deleted |
| `rclone_entries_listed_total` | Counter | Total directory entries listed |
| `rclone_errors_total` | Counter | Total transfer errors |
| `rclone_fatal_error` | Gauge | Fatal error flag (1 if error occurred) |
| `rclone_http_status_code` | Gauge | HTTP response status codes |
| `rclone_speed` | Gauge | Current transfer speed |

### Common Labels

All metrics include standard Kubernetes labels:
- `job`: Service job name (e.g., `csi-rclone-node-metrics`)
- `namespace`: Kubernetes namespace
- `pod`: Name of the CSI node pod
- `instance`: Pod IP and port
- `service`: Service name
- `endpoint`: Metrics endpoint name

## Configuration

Metrics are configured via environment variables in the DaemonSet:

```yaml
env:
  - name: METRICS_ADDR
    value: ":5572"
  - name: METRICS_PATH
    value: "/metrics"
  - name: METRICS_READ_TIMEOUT
    value: "10s"
  - name: METRICS_WRITE_TIMEOUT
    value: "10s"
  - name: METRICS_IDLE_TIMEOUT
    value: "60s"
```

## Troubleshooting

### ServiceMonitor Not Discovered

Check if the ServiceMonitor has the required label for your Prometheus instance:

```bash
# Check what Prometheus is looking for
kubectl get prometheus -A -o jsonpath='{range .items[*]}{.metadata.namespace}{"\t"}{.spec.serviceMonitorSelector}{"\n"}{end}'

# Ensure your ServiceMonitor has matching labels
kubectl get servicemonitor -n veloxpack csi-rclone-node-metrics -o jsonpath='{.metadata.labels}'
```

### Metrics Endpoint Not Responding

```bash
# Check if the metrics port is exposed
kubectl get pods -n veloxpack -l app=csi-rclone-node -o jsonpath='{.items[0].spec.containers[?(@.name=="rclone")].ports}'

# Test the metrics endpoint directly
kubectl port-forward -n veloxpack pod/<csi-rclone-node-pod> 5572:5572
curl http://localhost:5572/metrics
```

### No Metrics Data in Prometheus

1. Verify the CSI driver is performing operations (mount/unmount volumes)
2. Check scrape interval in ServiceMonitor (default: 30s)
3. Check Prometheus logs: `kubectl logs -n monitoring -l app.kubernetes.io/name=prometheus`
