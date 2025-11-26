# Kustomize Components

This directory contains reusable Kustomize components for the CSI Rclone driver.

## What are Kustomize Components?

Components are reusable pieces of configuration that can be included in multiple overlays. They allow you to compose functionality without duplication.

## Available Components

### `metrics-basic/`

Default metrics configuration with standard settings.

**What it does:** Configures the CSI driver to expose metrics endpoint

**Configuration:**
- Metrics Address: `:5572`
- Metrics Path: `/metrics`
- Read Timeout: `10s`
- Write Timeout: `10s`
- Idle Timeout: `60s`

**Usage in your overlay:**
```yaml
components:
  - ../../components/metrics-basic
```

---

### `metrics-service/`

Kubernetes Service to expose the metrics endpoint.

**What it does:** Creates a ClusterIP Service for metrics scraping

**Configuration:**
- Service Name: `csi-rclone-node-metrics`
- Port: `5572`
- Selector: `app=csi-rclone-node`

**Usage in your overlay:**
```yaml
components:
  - ../../components/metrics-basic
  - ../../components/metrics-service
```

---

### `metrics-servicemonitor/`

Prometheus Operator ServiceMonitor for automatic discovery.

**What it does:** Creates a ServiceMonitor for Prometheus Operator

**Configuration:**
- Scrape Interval: Configurable (check component file)
- Scrape Timeout: Configurable
- Label: `release: kube-prometheus-stack`

**Usage in your overlay:**
```yaml
components:
  - ../../components/metrics-basic
  - ../../components/metrics-service
  - ../../components/metrics-servicemonitor
```

---

### `metrics-dashboard/`

Grafana dashboard ConfigMap.

**What it does:** Creates dashboard ConfigMap for Grafana sidecar discovery

**Configuration:**
- Dashboard: CSI Driver Rclone Production Metrics
- Label: `grafana_dashboard: "1"`

**Usage in your overlay:**
```yaml
components:
  - ../../components/metrics-dashboard
```

---

### `metrics-custom/`

**Template for custom metrics configuration.**

This is a template directory that you can copy and customize for your specific needs.

---

### `rc-basic/`

Enables the rclone Remote Control (RC) API on every node pod.

**What it does:** Appends the RC CLI flags, exposes container port `5573`, and wires credentials from a Kubernetes secret.

**Prerequisites:**
- Create a secret named `csi-rclone-rc-auth` with `username` and `password` keys:

```bash
kubectl create secret generic csi-rclone-rc-auth \
  --from-literal=username=rc-admin \
  --from-literal=password='change-me' \
  -n veloxpack
```

**Usage in your overlay:**
```yaml
components:
  - ../../components/rc-basic
```

---

### `rc-service/`

Creates a headless `ClusterIP` so in-cluster workloads can call the RC API.

**What it does:** Adds the `csi-rclone-node-rc` Service (port `5573`, selector `app=csi-rclone-node`).

**Usage in your overlay:**
```yaml
components:
  - ../../components/rc-basic     # enables RC on the DaemonSet
  - ../../components/rc-service   # exposes it via a Service
```

## How to Customize Metrics Configuration

### Step 1: Copy the Template

```bash
cp -r deploy/components/metrics-custom deploy/components/my-metrics
```

### Step 2: Modify the Configuration

Edit `deploy/components/my-metrics/metrics-patch.yaml`:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: csi-rclone-node
  namespace: system
spec:
  selector:
    matchLabels:
      app: csi-rclone-node
  template:
    spec:
      containers:
        - name: rclone
          # Complete args array (replaces base args)
          args:
            # Base required args
            - "-v=5"
            - "--nodeid=$(NODE_ID)"
            - "--endpoint=$(CSI_ENDPOINT)"

            # Custom metrics configuration
            - "--metrics-addr=:9090"                    # Custom port
            - "--metrics-path=/custom-metrics"          # Custom path
            - "--metrics-server-read-timeout=30s"       # Increased timeouts
            - "--metrics-server-write-timeout=30s"
            - "--metrics-server-idle-timeout=120s"

          # Port must match --metrics-addr
          ports:
            - name: metrics
              containerPort: 9090
              protocol: TCP
```

**Important:** The `args` array completely replaces the base args, so you must include the base arguments (`-v`, `--nodeid`, `--endpoint`) along with your custom metrics arguments.

### Step 3: Use Your Custom Component

Create your own overlay or modify an existing one:

```yaml
# my-overlay/kustomization.yaml
namespace: veloxpack

resources:
  - ../../base

components:
  - ../../components/my-metrics         # Use your custom metrics config
  - ../../components/metrics-service    # Add Service if needed
  - ../../components/metrics-servicemonitor  # Add ServiceMonitor if needed
```

### Step 4: Update the Service (if needed)

If you changed the port, create your own service component or patch the existing one:

**Option A: Create a custom service component**

```bash
# Copy and modify the service component
cp -r deploy/components/metrics-service deploy/components/my-metrics-service

# Edit the port in deploy/components/my-metrics-service/service.yaml
# Change port: 5572 to port: 9090
```

**Option B: Patch the service in your overlay**

```yaml
# my-overlay/kustomization.yaml
namespace: veloxpack

resources:
  - ../../base

components:
  - ../../components/my-metrics          # Custom metrics config
  - ../../components/metrics-service     # Standard service

# Patch the service port
patches:
  - patch: |-
      apiVersion: v1
      kind: Service
      metadata:
        name: csi-rclone-node-metrics
      spec:
        ports:
          - name: metrics
            port: 9090
            targetPort: metrics
    target:
      kind: Service
      name: csi-rclone-node-metrics
```

## Common Customization Scenarios

### Scenario 1: Change Metrics Port

**Why:** Port 5572 conflicts with another service.

**Solution:**
1. Copy `metrics-custom` to `metrics-port-8080`
2. Change `--metrics-addr=:8080` in args
3. Change `containerPort` to `8080`
4. Update Service component or create custom service component

### Scenario 2: Increase Timeouts for Slow Networks

**Why:** Prometheus scrapes are timing out.

**Solution:**
1. Copy `metrics-custom` to `metrics-slow-network`
2. Increase timeout args: `--metrics-server-read-timeout=30s`, `--metrics-server-write-timeout=30s`, `--metrics-server-idle-timeout=180s`

### Scenario 3: Use a Different Metrics Path

**Why:** Corporate policy requires metrics at `/observability/metrics`.

**Solution:**
1. Copy `metrics-custom` to `metrics-corp-policy`
2. Change `--metrics-path=/observability/metrics` in args
3. Update ServiceMonitor path if using Prometheus Operator

### Scenario 4: Adjust Scrape Interval

**Why:** You need faster or slower metric updates for your data transfer workload.

**Solution:**
1. Copy `metrics-servicemonitor` to `metrics-servicemonitor-custom`
2. Change the `interval` field in `servicemonitor.yaml`
3. Recommended intervals:
   - **5s**: Development/debugging
   - **15s**: Production data transfer (recommended)
   - **30s**: Resource-constrained environments

## Testing Your Custom Component

Test that your component generates valid Kubernetes manifests:

```bash
# Dry-run to see the output
kubectl kustomize my-overlay/

# Validate against your cluster
kubectl kustomize my-overlay/ | kubectl apply --dry-run=client -f -
```

## Best Practices

1. **Naming Convention:** Use descriptive names like `metrics-port-9090` or `metrics-production`
2. **Documentation:** Add comments in your patch explaining why you made changes
3. **Version Control:** Keep your custom components in version control
4. **Testing:** Always test with `kubectl kustomize` before applying
5. **Validation:** Use `kubectl apply --dry-run` to validate

## Component vs. Overlay

**Use a Component when:**
- The configuration is reusable across multiple overlays
- You want to compose functionality
- The change is orthogonal to the base configuration

**Use an Overlay when:**
- The configuration is environment-specific
- You're combining multiple components and resources
- You need to set namespace or other top-level settings

## Example: Complete Custom Deployment

```bash
# 1. Copy and customize the metrics component
cp -r deploy/components/metrics-custom deploy/components/metrics-prod

# 2. Edit deploy/components/metrics-prod/metrics-patch.yaml
# (increase timeouts for production)

# 3. Optionally customize ServiceMonitor scrape interval
cp -r deploy/components/metrics-servicemonitor deploy/components/metrics-servicemonitor-prod
# Edit deploy/components/metrics-servicemonitor-prod/servicemonitor.yaml
# Change interval to 15s for production

# 4. Create your production overlay
mkdir -p deploy/overlays/production

# 5. Create deploy/overlays/production/kustomization.yaml
cat <<EOF > deploy/overlays/production/kustomization.yaml
namespace: prod-csi

resources:
  - ../../base

components:
  - ../../components/metrics-prod                    # Custom metrics config
  - ../../components/metrics-service                 # Service
  - ../../components/metrics-servicemonitor-prod     # Custom ServiceMonitor
  - ../../components/metrics-dashboard               # Grafana dashboard

# Add production-specific labels to dashboard
patches:
  - target:
      kind: ConfigMap
      labelSelector: grafana_dashboard=1
    patch: |-
      - op: add
        path: /metadata/labels/environment
        value: production
EOF

# 6. Deploy
kubectl apply -k deploy/overlays/production
```

## Need Help?

See the [overlays README](../overlays/README.md) for more examples of how overlays and components work together.
