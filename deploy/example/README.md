# Rclone CSI Driver Examples

This directory contains examples demonstrating how to use the Rclone CSI driver with various cloud storage backends.

## Quick Start

1. **Deploy the CSI driver**:
   ```bash
   kubectl apply -k ../
   ```

2. **Choose a storage backend example** and follow the specific instructions.

3. **Create a PVC and test pod**:
   ```bash
   kubectl apply -f rclone-pv-example.yaml
   ```

## Examples Overview

### Basic Examples
- [`rclone-pv-example.yaml`](rclone-pv-example.yaml) - Complete PersistentVolume example with inline rclone configuration
- [`rclone-secret.yaml`](rclone-secret.yaml) - Secret-based configuration example
- [`template-variable-examples.yaml`](template-variable-examples.yaml) - Dynamic path substitution examples

### Storage Backend Examples
- [`minio-deploy.yaml`](minio-deploy.yaml) - MinIO S3-compatible storage setup
- [`nginx-dynamic-path.yaml`](nginx-dynamic-path.yaml) - Dynamic path configuration

### Storage Class Examples
- [`storageclass-s3.yaml`](storageclass-s3.yaml) - Amazon S3 storage class
- [`storageclass-gcs.yaml`](storageclass-gcs.yaml) - Google Cloud Storage storage class
- [`storageclass-azure.yaml`](storageclass-azure.yaml) - Azure Blob Storage storage class
- [`storageclass-minio.yaml`](storageclass-minio.yaml) - MinIO storage class

### Secret Examples
- [`secret-s3.yaml`](secret-s3.yaml) - S3 credentials secret
- [`secret-gcs.yaml`](secret-gcs.yaml) - GCS service account secret
- [`secret-azure.yaml`](secret-azure.yaml) - Azure credentials secret
- [`secret-dropbox.yaml`](secret-dropbox.yaml) - Dropbox token secret
- [`secret-rc-auth.yaml`](secret-rc-auth.yaml) - RC API authentication secret (for Remote Control API)

## Configuration Methods

### Method 1: Secret-based Configuration (Recommended)
Store sensitive credentials in Kubernetes secrets and reference them in StorageClass:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rclone-secret
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
---
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

### Method 2: Inline Configuration
Include configuration directly in StorageClass parameters:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rclone-csi
provisioner: rclone.csi.veloxpack.io
parameters:
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

### Method 3: PersistentVolume Configuration
Configure directly in PersistentVolume volumeAttributes:

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-rclone
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteMany
  csi:
    driver: rclone.csi.veloxpack.io
    volumeHandle: unique-volume-id
    volumeAttributes:
      remote: "s3"
      remotePath: "my-bucket/folder"
      configData: |
        [s3]
        type = s3
        provider = AWS
        access_key_id = YOUR_ACCESS_KEY_ID
        secret_access_key = YOUR_SECRET_ACCESS_KEY
        region = us-east-1
```

## Dynamic Path Substitution

The driver supports template variables in the `remotePath` parameter:

| Variable | Description | Example |
|----------|-------------|---------|
| `${pvc.metadata.name}` | PVC name | `my-pvc-12345` |
| `${pvc.metadata.namespace}` | PVC namespace | `default` |
| `${pv.metadata.name}` | PV name | `pv-rclone-abc123` |

**Example:**
```yaml
parameters:
  remote: "s3"
  remotePath: "my-bucket/${pvc.metadata.namespace}/${pvc.metadata.name}"
```

## Supported Storage Backends

### Amazon S3
- **Provider**: AWS
- **Configuration**: Access key, secret key, region
- **Example**: [`storageclass-s3.yaml`](storageclass-s3.yaml)

### Google Cloud Storage
- **Provider**: Google Cloud
- **Configuration**: Service account JSON or OAuth
- **Example**: [`storageclass-gcs.yaml`](storageclass-gcs.yaml)

### Azure Blob Storage
- **Provider**: Microsoft Azure
- **Configuration**: Storage account name and key
- **Example**: [`storageclass-azure.yaml`](storageclass-azure.yaml)

### MinIO
- **Provider**: MinIO (S3-compatible)
- **Configuration**: Endpoint, access key, secret key
- **Example**: [`storageclass-minio.yaml`](storageclass-minio.yaml)

### Dropbox
- **Provider**: Dropbox
- **Configuration**: OAuth token
- **Example**: [`secret-dropbox.yaml`](secret-dropbox.yaml)

### SFTP
- **Provider**: SFTP server
- **Configuration**: Host, username, password/key
- **Example**: [`secret-sftp.yaml`](secret-sftp.yaml)

## Remote Control (RC) API

The driver can expose rclone's Remote Control API for programmatic control of mounts. This is useful for:

- **VFS Cache Refresh**: Trigger cache refresh for specific paths
- **Statistics**: Get real-time mount statistics
- **Operations**: Control rclone operations programmatically

### Setup

1. **Create the RC auth secret**:
   ```bash
   kubectl apply -f secret-rc-auth.yaml
   ```

2. **Enable RC API** in your deployment (via Helm or kustomize)

3. **Use the RC API** from within your cluster:
   ```bash
   # Get RC service endpoint
   RC_SERVICE=$(kubectl get svc -n veloxpack csi-rclone-node-rc -o jsonpath='{.metadata.name}')

   # Example: Refresh VFS cache
   curl -X POST http://${RC_SERVICE}:5573/vfs/refresh \
     -u admin:secure-password \
     -H "Content-Type: application/json" \
     -d '{"recursive": true, "dir": "/path/to/mount"}'
   ```

For more information, see the [RC API documentation](https://rclone.org/rc/).

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

## Troubleshooting

### Check Driver Status
```bash
# Check controller pods
kubectl get pods -n veloxpack -l app=csi-rclone-controller

# Check node pods
kubectl get pods -n veloxpack -l app=csi-rclone-node

# Check logs
kubectl logs -n veloxpack -l app=csi-rclone-controller
kubectl logs -n veloxpack -l app=csi-rclone-node
```

### Verify Driver Functionality
```bash
# Check if the driver is working correctly
kubectl exec -n veloxpack -l app=csi-rclone-node -- /rcloneplugin --help

# Check driver version information (shows when driver starts)
kubectl logs -n veloxpack -l app=csi-rclone-node --tail=10 | grep "DRIVER INFORMATION" -A 10
```

### Test Configuration
```bash
# Check driver logs for configuration parsing
kubectl logs -n veloxpack -l app=csi-rclone-node --tail=50 | grep -i config
```

## Security Best Practices

1. **Use Secrets**: Store sensitive credentials in Kubernetes secrets
2. **RBAC**: Ensure proper RBAC permissions are configured
3. **Network Policies**: Consider using network policies to restrict access
4. **Credential Rotation**: Regularly rotate storage backend credentials
5. **Least Privilege**: Use credentials with minimal required permissions

## Common Issues

1. **Authentication failures**: Verify credentials in secrets or configData
2. **Network connectivity**: Ensure nodes can reach the storage backend
3. **Permission errors**: Check that credentials have proper access rights
4. **Configuration format**: Ensure configData is valid INI format
5. **Resource constraints**: Verify sufficient memory and disk space

For detailed troubleshooting, see the [debug guide](../../docs/csi-debug.md).

## Contributing

To add new examples:

1. Create a new YAML file with descriptive name
2. Include comments explaining the configuration
3. Test the example thoroughly
4. Update this README with the new example
5. Submit a pull request

## Acknowledgments

These examples are based on patterns from the [csi-driver-nfs](https://github.com/kubernetes-csi/csi-driver-nfs) project and inspired by the original [csi-rclone](https://github.com/wunderio/csi-rclone) implementation.

## Support

- [Documentation](../../README.md)
- [Driver Parameters](../../docs/driver-parameters.md)
- [Debug Guide](../../docs/csi-debug.md)
- [Issue Tracker](https://github.com/veloxpack/csi-driver-rclone/issues)
