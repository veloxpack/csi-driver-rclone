# Deployment Overlays

This directory contains Kustomize overlays for different deployment scenarios.

## Available Overlays

### Default Overlay

- **`default/`** - Basic CSI driver deployment without metrics

```bash
kubectl apply -k deploy/overlays/default
```

### Metrics Overlays (Modular)

Choose the overlay that matches your monitoring setup:

#### 1. **`metrics/`** - Basic Metrics (Minimal)
Enables metrics endpoint on the DaemonSet without creating Service or ServiceMonitor.

**Use when:**
- You want to manually scrape metrics
- You're using a custom metrics collection setup
- You want minimal configuration

```bash
kubectl apply -k deploy/overlays/metrics
```

**What it includes:**
- ✅ Metrics endpoint on port 5572
- ❌ No Kubernetes Service
- ❌ No ServiceMonitor
- ❌ No Grafana Dashboard

---

#### 2. **`metrics-service/`** - Metrics with Service
Adds a Kubernetes Service to expose the metrics endpoint.

**Use when:**
- You want to expose metrics via a Service
- You're using annotation-based Prometheus scraping
- You need service discovery

```bash
kubectl apply -k deploy/overlays/metrics-service
```

**What it includes:**
- ✅ Metrics endpoint on port 5572
- ✅ ClusterIP Service
- ❌ No ServiceMonitor
- ❌ No Grafana Dashboard

---

#### 3. **`metrics-prometheus/`** - Metrics with Prometheus Operator
Complete Prometheus Operator integration with ServiceMonitor.

**Use when:**
- You have Prometheus Operator installed
- You want automatic scraping via ServiceMonitor
- You need full Prometheus integration

**Requirements:**
- Prometheus Operator must be installed
- ServiceMonitor CRD available

```bash
kubectl apply -k deploy/overlays/metrics-prometheus
```

**What it includes:**
- ✅ Metrics endpoint on port 5572
- ✅ ClusterIP Service
- ✅ ServiceMonitor with label `release: kube-prometheus-stack`
- ❌ No Grafana Dashboard

---

#### 4. **`metrics-dashboard/`** - Grafana Dashboard Only
Deploys only the Grafana dashboard ConfigMap.

**Use when:**
- You already have metrics enabled via another method
- You only want to add the dashboard
- You want to deploy dashboard separately

**Requirements:**
- Grafana with sidecar configured to watch ConfigMaps with `grafana_dashboard: "1"` label
- Metrics must be enabled separately (use another overlay or Helm)

```bash
kubectl apply -k deploy/overlays/metrics-dashboard
```

**What it includes:**
- ❌ No metrics configuration
- ❌ No Service
- ❌ No ServiceMonitor
- ✅ Grafana Dashboard ConfigMap (auto-discovered by Grafana sidecar)

---

#### 5. **`metrics-full/`** - Complete Monitoring Stack
Everything included: Service, ServiceMonitor, and Grafana Dashboard.

**Use when:**
- You have Prometheus Operator + Grafana installed
- You want the complete monitoring setup
- You want everything in one command

**Requirements:**
- Prometheus Operator installed (kube-prometheus-stack)
- Grafana with dashboard sidecar enabled

```bash
kubectl apply -k deploy/overlays/metrics-full
```

**What it includes:**
- ✅ Metrics endpoint on port 5572
- ✅ ClusterIP Service
- ✅ ServiceMonitor with label `release: kube-prometheus-stack`
- ✅ Grafana Dashboard ConfigMap (auto-discovered by Grafana sidecar)

---

## Customizing Metrics Configuration

All metrics overlays use the `deploy/components/metrics-basic` component with default settings. To customize:

### Quick Customization Example

```bash
# 1. Copy the template component
cp -r deploy/components/metrics-custom deploy/components/my-metrics

# 2. Edit deploy/components/my-metrics/metrics-patch.yaml
# Change port, timeouts, or other settings

# 3. Use your custom component in an overlay
cat <<EOF > my-overlay/kustomization.yaml
namespace: veloxpack

resources:
  - ../../base

components:
  - ../../components/my-metrics          # Your custom metrics config
  - ../../components/metrics-service     # Add Service if needed
  - ../../components/metrics-servicemonitor  # Add ServiceMonitor if needed
EOF

# 4. Deploy
kubectl apply -k my-overlay/
```

See the [components README](../components/README.md) for detailed customization guide.

## Composing Overlays

You can also compose components to create custom deployments. For example, metrics + dashboard without Prometheus:

Create your own `kustomization.yaml`:

```yaml
namespace: veloxpack

resources:
  - ../../base

components:
  - ../../components/metrics-basic       # Enable metrics endpoint
  - ../../components/metrics-service     # Add Service
  - ../../components/metrics-dashboard   # Add Grafana dashboard
  # Note: Omitting metrics-servicemonitor for non-Prometheus setup
```

## Quick Reference Table

| Overlay | Metrics Endpoint | Service | ServiceMonitor | Dashboard | Requirements |
|---------|-----------------|---------|----------------|-----------|--------------|
| `default` | ❌ | ❌ | ❌ | ❌ | None |
| `metrics` | ✅ | ❌ | ❌ | ❌ | None |
| `metrics-service` | ✅ | ✅ | ❌ | ❌ | None |
| `metrics-prometheus` | ✅ | ✅ | ✅ | ❌ | Prometheus Operator |
| `metrics-dashboard` | ❌ | ❌ | ❌ | ✅ | Grafana sidecar |
| `metrics-full` | ✅ | ✅ | ✅ | ✅ | Prometheus Operator + Grafana |

## Common Scenarios

### Scenario 1: Development with Skaffold
```bash
skaffold dev -p metrics-full
```

### Scenario 2: Production with Prometheus Operator
```bash
kubectl apply -k deploy/overlays/metrics-prometheus
# Add dashboard later if needed
kubectl apply -k deploy/overlays/metrics-dashboard
```

### Scenario 3: Prometheus without Operator (annotation-based)
```bash
kubectl apply -k deploy/overlays/metrics-service
```

### Scenario 4: Custom Metrics Collection
```bash
kubectl apply -k deploy/overlays/metrics
# Then configure your own scraping
```

## Namespace Considerations

- All CSI resources are deployed to the `veloxpack` namespace by default
- The Grafana dashboard ConfigMap is created in the `veloxpack` namespace
  - Grafana sidecar watches **ALL** namespaces by default, so it will be auto-discovered
  - If your Grafana only watches the `monitoring` namespace, manually copy the ConfigMap there
- To change the namespace, edit the `namespace:` field in the kustomization.yaml

## Next Steps

After deploying, see the [metrics README](./metrics/README.md) for:
- Verifying the deployment
- Accessing Grafana dashboards
- Querying metrics
- Troubleshooting
