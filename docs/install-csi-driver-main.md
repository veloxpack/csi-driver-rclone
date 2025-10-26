# Install Rclone CSI Driver from Main Branch

This document explains how to install the Rclone CSI driver from the main branch for development and testing purposes.

## Prerequisites

- Kubernetes 1.20 or later
- rclone installed on all nodes
- CSI node driver registrar
- kubectl configured to communicate with your cluster
- Docker or container runtime for building images

## Building from Source

### 1. Clone the Repository

```bash
git clone https://github.com/veloxpack/csi-driver-rclone.git
cd csi-driver-rclone
```

### 2. Build the Driver

```bash
# Build the binary
make build

# Build the container image
make container

# Push to your registry (optional)
make push
```

### 3. Set Environment Variables

```bash
export REGISTRY=your-registry.com
export IMAGE_VERSION=latest
export DRIVER_NAME=rclone.csi.veloxpack.io
```

## Installation Methods

### Method 1: Using kubectl with Custom Image

#### 1. Update Image References

Edit the deployment files to use your custom image:

```bash
# Update controller deployment
sed -i "s|image: .*|image: ${REGISTRY}/csi-rclone:${IMAGE_VERSION}|g" deploy/csi-rclone-controller.yaml

# Update node deployment
sed -i "s|image: .*|image: ${REGISTRY}/csi-rclone:${IMAGE_VERSION}|g" deploy/csi-rclone-node.yaml
```

#### 2. Deploy the Driver

```bash
kubectl apply -k deploy/overlays/default
```

### Method 2: Using Kustomize with Custom Image

#### 1. Create kustomization.yaml

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- namespace-csi-rclone.yaml
- rbac-csi-rclone.yaml
- csi-rclone-controller.yaml
- csi-rclone-node.yaml
- csi-rclone-driverinfo.yaml

images:
- name: csi-rclone
  newName: your-registry.com/csi-rclone
  newTag: latest
```

#### 2. Deploy with Kustomize

```bash
kubectl apply -k .
```

### Method 3: Using Helm with Custom Image

#### 1. Create values.yaml

```yaml
image:
  repository: your-registry.com/csi-rclone
  tag: latest
  pullPolicy: Always

controller:
  replicas: 2
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"

node:
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "512Mi"
      cpu: "500m"
```

#### 2. Install with Helm

```bash
helm install csi-rclone ./charts/csi-driver-rclone -f values.yaml -n kube-system
```

## Development Setup

### 1. Local Development

For local development, you can run the driver outside of Kubernetes:

```bash
# Build the binary
make build

# Run the driver locally
./bin/rcloneplugin --endpoint unix:///tmp/csi.sock --nodeid CSINode -v=5
```

### 2. Testing with csc Tool

Install the CSI testing tool:

```bash
# Install csc
go install github.com/rexray/gocsi/csc@latest
```

Test the driver:

```bash
# Set environment variables
export cap="1,mount,"
export volname="test-$(date +%s)"
export volsize="2147483648"
export endpoint="unix:///tmp/csi.sock"
export target_path="/tmp/targetpath"
export params="remote=s3,remotePath=test-bucket,configData=[s3]\ntype = s3\nprovider = Minio\nendpoint = http://localhost:9000\naccess_key_id = minioadmin\nsecret_access_key = minioadmin"

# Test operations
csc identity plugin-info --endpoint "$endpoint"
csc controller new --endpoint "$endpoint" --cap "$cap" "$volname" --req-bytes "$volsize" --params "$params"
csc node publish --endpoint "$endpoint" --cap "$cap" --vol-context "$params" --target-path "$target_path" "$volumeid"
csc node unpublish --endpoint "$endpoint" --target-path "$target_path" "$volumeid"
csc controller del --endpoint "$endpoint" "$volumeid"
```

## Configuration for Development

### 1. Enable Debug Logging

```yaml
# In controller deployment
args:
  - "--v=5"
  - "--logtostderr=true"
  - "--stderrthreshold=INFO"

# In node deployment
args:
  - "--v=5"
  - "--logtostderr=true"
  - "--stderrthreshold=INFO"
```

### 2. Set Resource Limits

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### 3. Configure Rclone

Create a test configuration:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rclone-secret
  namespace: default
type: Opaque
stringData:
  remote: "s3"
  remotePath: "test-bucket"
  configData: |
    [s3]
    type = s3
    provider = Minio
    endpoint = http://minio.minio:9000
    access_key_id = minioadmin
    secret_access_key = minioadmin
```

## Testing the Installation

### 1. Create Test Resources

```yaml
# StorageClass
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rclone-csi
provisioner: rclone.csi.veloxpack.io
parameters:
  remote: "s3"
  remotePath: "test-bucket"
csi.storage.k8s.io/node-publish-secret-name: "rclone-secret"
csi.storage.k8s.io/node-publish-secret-namespace: "default"
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true

---
# PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: rclone-csi

---
# Test Pod
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx
    volumeMounts:
    - name: test-volume
      mountPath: /data
  volumes:
  - name: test-volume
    persistentVolumeClaim:
      claimName: test-pvc
```

### 2. Verify Installation

```bash
# Check driver pods
kubectl get pods -l app=csi-rclone-controller
kubectl get pods -l app=csi-rclone-node

# Check CSIDriver
kubectl get csidriver rclone.csi.veloxpack.io

# Check PVC status
kubectl get pvc test-pvc

# Check pod status
kubectl get pod test-pod

# Test mount
kubectl exec test-pod -- mount | grep rclone
kubectl exec test-pod -- ls -la /data
```

## Development Workflow

### 1. Making Changes

```bash
# Make your changes to the code
vim pkg/rclone/nodeserver.go

# Run tests
make test

# Run linter
make lint

# Build and test
make build
make container
```

### 2. Testing Changes

```bash
# Deploy updated driver
kubectl apply -k deploy/overlays/default

# Check logs
kubectl logs -l app=csi-rclone-controller -f
kubectl logs -l app=csi-rclone-node -f
```

### 3. Debugging

```bash
# Enable debug logging
kubectl patch deployment csi-rclone-controller -p '{"spec":{"template":{"spec":{"containers":[{"name":"rclone","args":["--v=5","--logtostderr=true"]}]}}}}'

# Check driver logs
kubectl logs -l app=csi-rclone-controller --tail=100
kubectl logs -l app=csi-rclone-node --tail=100
```

## Continuous Integration

### 1. GitHub Actions

Create `.github/workflows/ci.yaml`:

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    - name: Build
      run: make build
    - name: Test
      run: make test
    - name: Lint
      run: make lint
```

### 2. Build and Push Images

```yaml
name: Build and Push

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v1
    - name: Login to Registry
      uses: docker/login-action@v1
      with:
        registry: your-registry.com
        username: ${{ secrets.REGISTRY_USERNAME }}
        password: ${{ secrets.REGISTRY_PASSWORD }}
    - name: Build and push
      uses: docker/build-push-action@v2
      with:
        context: .
        push: true
        tags: your-registry.com/csi-rclone:latest
```

## Troubleshooting

### Common Issues

1. **Build failures**: Check Go version and dependencies
2. **Image push failures**: Verify registry credentials
3. **Driver won't start**: Check rclone installation and configuration
4. **Volume mount fails**: Verify rclone configuration and network connectivity

### Debug Commands

```bash
# Check build logs
make build 2>&1 | tee build.log

# Check container logs
docker logs csi-rclone-container

# Check Kubernetes logs
kubectl logs -l app=csi-rclone-controller --tail=100
kubectl logs -l app=csi-rclone-node --tail=100

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

## Uninstallation

### Remove Development Installation

```bash
# Delete all resources
kubectl delete -k deploy/

# Or delete individually
kubectl delete -f deploy/csi-rclone-driverinfo.yaml
kubectl delete -f deploy/csi-rclone-node.yaml
kubectl delete -f deploy/csi-rclone-controller.yaml
kubectl delete -f deploy/rbac-csi-rclone.yaml
kubectl delete -f deploy/namespace-csi-rclone.yaml
```

### Clean Up Local Development

```bash
# Stop local driver
pkill -f rcloneplugin

# Clean up test files
rm -rf /tmp/targetpath
rm -f /tmp/csi.sock
```

## Next Steps

1. **Read the documentation**: [README.md](../README.md)
2. **Understand driver parameters**: [driver-parameters.md](driver-parameters.md)
3. **Learn about debugging**: [csi-debug.md](csi-debug.md)
4. **Contribute**: Follow the [development guide](csi-dev.md)

## Support

- [Documentation](../README.md)
- [Issue Tracker](https://github.com/veloxpack/csi-driver-rclone/issues)
- [Discussions](https://github.com/veloxpack/csi-driver-rclone/discussions)
