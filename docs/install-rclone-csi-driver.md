# Install Rclone CSI Driver

This document explains how to install the Rclone CSI driver on a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.20 or later
- CSI node driver registrar
- kubectl configured to communicate with your cluster
- FUSE support on nodes (for mounting)
- **No rclone installation required** - the driver uses rclone as a Go library directly

## Installation Methods

### Method 1: Using kubectl (Recommended)

#### 1. Deploy the CSI Driver
```bash
kubectl apply -k deploy/
```

This will install:
- CSI Controller (StatefulSet)
- CSI Node Driver (DaemonSet)
- RBAC permissions
- CSIDriver CRD
- Example StorageClass

#### 2. Verify Installation
```bash
# Check controller pods
kubectl get pods -l app=csi-rclone-controller

# Check node pods
kubectl get pods -l app=csi-rclone-node

# Check CSIDriver
kubectl get csidriver rclone.csi.veloxpack.io
```

### Method 2: Using Helm Charts

#### 1. Add Helm Repository
```bash
helm repo add csi-rclone https://veloxpack.github.io/csi-driver-rclone
helm repo update
```

#### 2. Install the Driver
```bash
helm install csi-rclone csi-rclone/csi-driver-rclone --namespace kube-system
```

#### 3. Verify Installation
```bash
helm list -n kube-system
kubectl get pods -l app=csi-rclone-controller
```

### Method 3: Manual Installation

#### 1. Create Namespace
```bash
kubectl apply -f deploy/namespace-csi-rclone.yaml
```

#### 2. Create RBAC
```bash
kubectl apply -f deploy/rbac-csi-rclone.yaml
```

#### 3. Deploy Controller
```bash
kubectl apply -f deploy/csi-rclone-controller.yaml
```

#### 4. Deploy Node Driver
```bash
kubectl apply -f deploy/csi-rclone-node.yaml
```

#### 5. Create CSIDriver
```bash
kubectl apply -f deploy/csi-rclone-driverinfo.yaml
```

## Architecture Benefits

The Rclone CSI driver uses rclone as a Go library directly, which provides several advantages:

- **No External Dependencies**: No need to install rclone binary on nodes
- **No Process Overhead**: Direct library integration means no subprocess spawning
- **Better Resource Management**: Single process with integrated rclone functionality
- **Simplified Deployment**: No additional setup steps required
- **Enhanced Security**: No external process execution reduces attack surface

## Configuration

### 1. Create Storage Class

Create a StorageClass for your rclone backend:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rclone-csi
provisioner: rclone.csi.veloxpack.io
parameters:
  remote: "s3"
  remotePath: "my-bucket"
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true
```

### 2. Create Secret (Optional)

If you want to use secrets for configuration:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rclone-secret
  namespace: default
type: Opaque
stringData:
  remote: "s3"
  remotePath: "my-bucket"
  configData: |
    [s3]
    type = s3
    provider = AWS
    access_key_id = YOUR_ACCESS_KEY_ID
    secret_access_key = YOUR_SECRET_ACCESS_KEY
    region = us-east-1
```

Then reference it in your StorageClass:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rclone-csi
provisioner: rclone.csi.veloxpack.io
parameters:
  remote: "s3"
  remotePath: "my-bucket"
csi.storage.k8s.io/node-publish-secret-name: "rclone-secret"
csi.storage.k8s.io/node-publish-secret-namespace: "default"
```

## Testing the Installation

### 1. Create a Test PVC

```yaml
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
```

### 2. Create a Test Pod

```yaml
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

### 3. Verify Mount

```bash
# Check if PVC is bound
kubectl get pvc test-pvc

# Check if pod is running
kubectl get pod test-pod

# Check mount inside pod
kubectl exec test-pod -- mount | grep rclone
kubectl exec test-pod -- ls -la /data
```

## Uninstallation

### Method 1: Using kubectl

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

### Method 2: Using Helm

```bash
helm uninstall csi-rclone -n kube-system
```

## Troubleshooting

### Check Driver Status

```bash
# Check controller logs
kubectl logs -l app=csi-rclone-controller

# Check node logs
kubectl logs -l app=csi-rclone-node

# Check CSIDriver status
kubectl get csidriver rclone.csi.veloxpack.io -o yaml
```

### Verify Driver Functionality

```bash
# Check if the driver is working correctly
kubectl exec -l app=csi-rclone-node -- /rcloneplugin --help

# Check driver version information (shows when driver starts)
kubectl logs -l app=csi-rclone-node --tail=10 | grep "DRIVER INFORMATION" -A 10
```

### Common Issues

1. **Driver won't start**: Check RBAC permissions and container image
2. **Volume mount fails**: Verify rclone configuration and network connectivity
3. **Permission denied**: Check file system permissions and mount options
4. **Authentication failed**: Verify credentials in secrets or configData

For detailed troubleshooting, see the [debug guide](csi-debug.md).

## Configuration Examples

### Amazon S3

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: s3-csi
provisioner: rclone.csi.veloxpack.io
parameters:
  remote: "s3"
  remotePath: "my-s3-bucket"
  configData: |
    [s3]
    type = s3
    provider = AWS
    access_key_id = YOUR_ACCESS_KEY_ID
    secret_access_key = YOUR_SECRET_ACCESS_KEY
    region = us-east-1
```

### Google Cloud Storage

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gcs-csi
provisioner: rclone.csi.veloxpack.io
parameters:
  remote: "gcs"
  remotePath: "my-gcs-bucket"
  configData: |
    [gcs]
    type = google cloud storage
    project_number = 12345678
    service_account_file = /path/to/service-account.json
```

### MinIO

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: minio-csi
provisioner: rclone.csi.veloxpack.io
parameters:
  remote: "minio"
  remotePath: "my-minio-bucket"
  configData: |
    [minio]
    type = s3
    provider = Minio
    endpoint = http://minio.minio:9000
    access_key_id = minioadmin
    secret_access_key = minioadmin
```

## Security Considerations

1. **Use Secrets**: Store sensitive credentials in Kubernetes secrets
2. **RBAC**: Ensure proper RBAC permissions are configured
3. **Network Policies**: Consider using network policies to restrict access
4. **Image Security**: Use trusted container images
5. **Credential Rotation**: Regularly rotate storage backend credentials

### Resource Limits

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Log Levels

Set log level for debugging:

```yaml
args:
  - "--v=5"  # Verbose logging
  - "--logtostderr=true"
```

## Support

- [Documentation](README.md)
- [Driver Parameters](driver-parameters.md)
- [Development Guide](csi-dev.md)
- [Debug Guide](csi-debug.md)
- [Issue Tracker](https://github.com/veloxpack/csi-driver-rclone/issues)
