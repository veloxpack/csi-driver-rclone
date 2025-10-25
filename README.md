
<h1 align="center">
  <a href="https://www.veloxpack.io/docs/csi-driver-rclone">
    <img src=".github/banner.png" alt="Rclone UI" width="100%">
  </a>
  <br>
  <a href="https://www.veloxpack.io/docs/csi-driver-rclone">
    Rclone CSI Driver for Kubernetes
  </a>
</h1>

![build status](https://github.com/veloxpack/csi-driver-rclone/actions/workflows/test.yaml/badge.svg)
[![Trivy vulnerability scanner](https://github.com/veloxpack/csi-driver-rclone/actions/workflows/trivy.yaml/badge.svg?branch=main)](https://github.com/veloxpack/csi-driver-rclone/actions/workflows/trivy.yaml)

### Overview

This is a repository for [Rclone](https://rclone.org/) [CSI](https://kubernetes-csi.github.io/docs/) driver, csi plugin name: `rclone.csi.veloxpack.io`. This driver enables Kubernetes pods to mount cloud storage backends as persistent volumes using rclone, supporting 50+ storage providers including S3, Google Cloud Storage, Azure Blob, Dropbox, and many more.

### Container Images & Kubernetes Compatibility:
|driver version  | supported k8s version | status |
|----------------|-----------------------|--------|
|main branch     | 1.20+                 | GA     |
|v0.1.0          | 1.20+                 | GA     |

### Install driver on a Kubernetes cluster

#### Option 1: Install via Helm (Recommended)

Install directly from the OCI registry:

```bash
# Install with default configuration
helm install csi-rclone oci://registry-1.docker.io/veloxpack/csi-driver-rclone-charts

# Install in a specific namespace
helm install csi-rclone oci://registry-1.docker.io/veloxpack/csi-driver-rclone-charts \
  --namespace veloxpack --create-namespace
```

Verify the installation:

```bash
# Check release status
helm list -n veloxpack

# Verify pods are running
kubectl get pods -n veloxpack -l app.kubernetes.io/name=csi-driver-rclone
```

#### Option 2: Install via kubectl
Follow the [manual installation guide](./docs/install-rclone-csi-driver.md)

### Driver parameters
Please refer to [`rclone.csi.veloxpack.io` driver parameters](./docs/driver-parameters.md)

### Examples
 - [Basic usage](./deploy/example/README.md)
 - [S3 Storage](./deploy/example/storageclass-s3.yaml)
 - [Google Cloud Storage](./deploy/example/storageclass-gcs.yaml)
 - [Azure Blob Storage](./deploy/example/storageclass-azure.yaml)
 - [MinIO](./deploy/example/storageclass-minio.yaml)
 - [Dropbox](./deploy/example/secret-dropbox.yaml)
 - [SFTP](./deploy/example/secret-sftp.yaml)

### Troubleshooting
 - [CSI driver troubleshooting guide](./docs/csi-debug.md)

## Kubernetes Development
Please refer to [development guide](./docs/csi-dev.md)

## Features

- **50+ Storage Providers**: Supports Amazon S3, Google Cloud Storage, Azure Blob, Dropbox, SFTP, and many more
- **No External Dependencies**: Uses rclone as a Go library directly - no rclone binary installation required
- **No Process Overhead**: Direct library integration means no subprocess spawning or external process management
- **Dynamic Volume Provisioning**: Create persistent volumes via StorageClass
- **Secret-based Configuration**: Secure credential management using Kubernetes secrets
- **Inline Configuration**: Direct configuration in StorageClass parameters
- **Template Variable Support**: Dynamic path substitution using PVC/PV metadata
- **VFS Caching**: High-performance caching with configurable options
- **No Staging Required**: Direct mount without volume staging
- **Flexible Backend Support**: Choose between minimal or full backend support for smaller images

## Requirements

- Kubernetes 1.20 or later
- CSI node driver registrar
- FUSE support on nodes (for mounting)
- **No rclone installation required** - the driver uses rclone as a Go library directly

## Quick Start

### 1. Deploy the CSI Driver

```bash
kubectl apply -k deploy/
```

This will install:
- CSI Controller (StatefulSet)
- CSI Node Driver (DaemonSet)
- RBAC permissions
- CSIDriver CRD
- Example StorageClass

### 2. Configure Storage Backend

Create a secret with your storage backend configuration:

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

### 3. Create StorageClass

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
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true
```

### 4. Create PVC and Pod

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-rclone
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
  storageClassName: rclone-csi
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx-rclone
spec:
  containers:
  - name: nginx
    image: nginx
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: pvc-rclone
```

## Configuration Methods

### Method 1: Kubernetes Secrets (Recommended)
Store sensitive credentials in Kubernetes secrets and reference them in StorageClass.

### Method 2: Inline Configuration
Include configuration directly in StorageClass parameters.

### Method 3: PersistentVolume Configuration
Configure directly in PersistentVolume volumeAttributes.

**Priority**: volumeAttributes > StorageClass parameters > Secrets

## Supported Storage Backends

The driver supports all rclone backends, including:

- **Amazon S3** and S3-compatible storage (MinIO, DigitalOcean Spaces, etc.)
- **Google Cloud Storage**
- **Azure Blob Storage**
- **Dropbox**
- **SFTP/SSH**
- **Google Drive**
- **OneDrive**
- **Box**
- **Backblaze B2**
- **WebDAV**
- **FTP**
- **And 50+ more backends**

See [examples](./deploy/example/README.md) for configuration examples.

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

### VFS Cache Options
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-rclone-performance
spec:
  mountOptions:
    - vfs-cache-mode=writes
    - vfs-cache-max-size=10G
    - dir-cache-time=30s
  csi:
    driver: rclone.csi.veloxpack.io
    volumeHandle: performance-volume
    volumeAttributes:
      remote: "s3"
      remotePath: "my-bucket"
      configData: |
        [s3]
        type = s3
        provider = AWS
        access_key_id = YOUR_ACCESS_KEY_ID
        secret_access_key = YOUR_SECRET_ACCESS_KEY
```

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
kubectl get pods -n kube-system -l app=csi-rclone-controller

# Check node pods
kubectl get pods -n kube-system -l app=csi-rclone-node

# Check logs
kubectl logs -n kube-system -l app=csi-rclone-controller
kubectl logs -n kube-system -l app=csi-rclone-node
```

### Verify Driver Functionality
```bash
# Check if the driver is working correctly
kubectl exec -n kube-system -l app=csi-rclone-node -- /rcloneplugin --help

# Check driver version information (shows when driver starts)
kubectl logs -n kube-system -l app=csi-rclone-node --tail=10 | grep "DRIVER INFORMATION" -A 10
```

### Common Issues
1. **Authentication failures**: Verify credentials in secrets or configData
2. **Network connectivity**: Ensure nodes can reach the storage backend
3. **Permission errors**: Check that credentials have proper access rights
4. **Configuration format**: Ensure configData is valid INI format
5. **Resource constraints**: Verify sufficient memory and disk space

For detailed troubleshooting, see the [debug guide](./docs/csi-debug.md).

## Building from Source

```bash
# Clone repository
git clone https://github.com/veloxpack/csi-driver-rclone.git
cd csi-driver-rclone

# Build binary
make build

# Build Docker image
make container

# Push to registry
make push
```

### Docker Build Options

The driver supports two backend configurations for different use cases:

#### Full Backend Support (Default)
Includes all 50+ rclone backends for maximum compatibility:

```bash
# Build with all backends (default)
docker build -t csi-rclone:latest .

# Or explicitly specify
docker build --build-arg RCLONE_BACKEND_MODE=all -t csi-rclone:latest .
```

#### Minimal Backend Support
Includes only the most common backends for smaller image size:

```bash
# Build with minimal backends
docker build --build-arg RCLONE_BACKEND_MODE=minimal -t csi-rclone:minimal .
```

**Minimal backends include:**
- Amazon S3 and S3-compatible storage
- Google Cloud Storage
- Azure Blob Storage
- Dropbox
- Google Drive
- OneDrive
- Box
- Backblaze B2
- SFTP
- WebDAV
- FTP
- Local filesystem

**Benefits of minimal build:**
- Smaller Docker image size
- Faster container startup
- Reduced attack surface
- Lower memory footprint

Choose the build that fits your needs - full support for maximum compatibility or minimal for production efficiency.

## Development

### Running Linter
```bash
./bin/golangci-lint run --config .golangci.yml ./...
```

### Testing
```bash
go test ./pkg/rclone/...
```

### Local Development
```bash
# Run driver locally
./bin/rcloneplugin --endpoint unix:///tmp/csi.sock --nodeid CSINode -v=5
```

## Architecture

This driver is based on the [csi-driver-nfs](https://github.com/kubernetes-csi/csi-driver-nfs) reference implementation, following CSI specification best practices. It also draws inspiration from the original [csi-rclone](https://github.com/wunderio/csi-rclone) implementation by WunderIO.

**Components:**
- **Identity Server**: Plugin metadata and health checks
- **Controller Server**: Volume lifecycle management (create/delete)
- **Node Server**: Volume mounting/unmounting on nodes

**Key Design Decisions:**
1. **No Staging**: Rclone volumes don't require staging
2. **Direct Rclone Integration**: Uses rclone's Go library directly
3. **Remote Creation**: Creates temporary remotes for each mount
4. **VFS Caching**: Leverages rclone's VFS for improved performance
5. **Template Variable Support**: Dynamic path substitution using PVC/PV metadata

## Security Considerations

1. **Use Secrets**: Store sensitive credentials in Kubernetes secrets
2. **RBAC**: Ensure proper RBAC permissions are configured
3. **Network Policies**: Consider using network policies to restrict access
4. **Image Security**: Use trusted container images
5. **Credential Rotation**: Regularly rotate storage backend credentials

### Log Levels
Set log level for debugging:
```yaml
args:
  - "--v=5"  # Verbose logging
  - "--logtostderr=true"
```

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack channel](https://kubernetes.slack.com/messages/sig-storage)
- [Mailing list](https://groups.google.com/forum/#!forum/kubernetes-sig-storage)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please ensure:
- All code passes `golangci-lint` checks
- Follow existing code patterns
- Add tests for new functionality
- Update documentation

## Acknowledgments

This project builds upon the excellent work of several open source communities:

- **[WunderIO/csi-rclone](https://github.com/wunderio/csi-rclone)** - The original rclone CSI driver implementation that inspired this project
- **[Kubernetes CSI NFS Driver](https://github.com/kubernetes-csi/csi-driver-nfs)** - Reference implementation and architectural patterns
- **[Rclone](https://rclone.org/)** - The powerful cloud storage sync tool that makes this driver possible
- **[Kubernetes CSI Community](https://github.com/kubernetes-csi)** - For the Container Storage Interface specification and ecosystem

Special thanks to the maintainers and contributors of these projects for their dedication to open source software.

## Support

- [Rclone Documentation](https://rclone.org/)
- [CSI Specification](https://github.com/container-storage-interface/spec)
- [Issue Tracker](https://github.com/veloxpack/csi-driver-rclone/issues)
- [Discussions](https://github.com/veloxpack/csi-driver-rclone/discussions)
