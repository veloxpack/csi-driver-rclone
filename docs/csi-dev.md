# Rclone CSI Driver Development Guide

## How to Build This Project

### Prerequisites
- Go 1.21 or later
- Docker (for building container images)
- Kubernetes cluster (for testing)
- rclone installed on all nodes

### Clone Repository
```console
$ mkdir -p $GOPATH/src/github.com/veloxpack/
$ git clone https://github.com/veloxpack/csi-driver-rclone $GOPATH/src/github.com/veloxpack/csi-driver-rclone
$ cd $GOPATH/src/github.com/veloxpack/csi-driver-rclone
```

### Build CSI Driver
```console
$ make build
```

### Build Container Image
```console
$ make container
```

### Push Container Image
```console
$ make push
```

## How to Test CSI Driver in Local Environment

### Install csc Tool
Install the CSI testing tool according to [gocsi documentation](https://github.com/rexray/gocsi/tree/master/csc):
```console
$ mkdir -p $GOPATH/src/github.com
$ cd $GOPATH/src/github.com
$ git clone https://github.com/rexray/gocsi.git
$ cd rexray/gocsi/csc
$ make build
```

### Start CSI Driver Locally
```console
$ cd $GOPATH/src/github.com/veloxpack/csi-driver-rclone
$ ./bin/rcloneplugin --endpoint unix:///tmp/csi.sock --nodeid CSINode -v=5 &
```

### Test CSI Operations

#### 0. Set Environment Variables
```console
$ cap="1,mount,"
$ volname="test-$(date +%s)"
$ volsize="2147483648"
$ endpoint="unix:///tmp/csi.sock"
$ target_path="/tmp/targetpath"
$ params="remote=s3,remotePath=test-bucket,configData=[s3]\ntype = s3\nprovider = Minio\nendpoint = http://localhost:9000\naccess_key_id = minioadmin\nsecret_access_key = minioadmin"
```

#### 1. Get Plugin Info
```console
$ csc identity plugin-info --endpoint "$endpoint"
"rclone.csi.veloxpack.io"    "v1.0.0"
```

#### 2. Create a New Rclone Volume
```console
$ value="$(csc controller new --endpoint "$endpoint" --cap "$cap" "$volname" --req-bytes "$volsize" --params "$params")"
$ sleep 15
$ volumeid="$(echo "$value" | awk '{print $1}' | sed 's/"//g')"
$ echo "Got volume id: $volumeid"
```

#### 3. Publish a Rclone Volume
```console
$ csc node publish --endpoint "$endpoint" --cap "$cap" --vol-context "$params" --target-path "$target_path" "$volumeid"
```

#### 4. Unpublish a Rclone Volume
```console
$ csc node unpublish --endpoint "$endpoint" --target-path "$target_path" "$volumeid"
```

#### 5. Validate Volume Capabilities
```console
$ csc controller validate-volume-capabilities --endpoint "$endpoint" --cap "$cap" "$volumeid"
```

#### 6. Delete the Rclone Volume
```console
$ csc controller del --endpoint "$endpoint" "$volumeid" --timeout 10m
```

#### 7. Get NodeID
```console
$ csc node get-info --endpoint "$endpoint"
CSINode
```

## How to Test CSI Driver in a Kubernetes Cluster

### Set Environment Variables
```console
export REGISTRY=<your-docker-registry>
export IMAGE_VERSION=latest
```

### Build and Push Container Image
```console
# Run `docker login` first
# Build docker image
make container
# Push the docker image
make push
```

### Deploy to Kubernetes Cluster
Make sure `kubectl get nodes` works on your development machine.

### Run E2E Tests
```console
# Install Rclone CSI Driver on the Kubernetes cluster
make e2e-bootstrap

# Run the E2E test
make e2e-test
```

## Development Workflow

### Code Style
- Follow Go standard formatting: `gofmt -s -w .`
- Run linter: `golangci-lint run --config .golangci.yml ./...`
- Ensure all tests pass: `go test ./...`

### Testing
```console
# Run unit tests
go test ./pkg/rclone/...

# Run integration tests
go test ./test/...

# Run with coverage
go test -cover ./pkg/rclone/...
```

### Building for Different Architectures
```console
# Build for Linux AMD64
GOOS=linux GOARCH=amd64 make build

# Build for Linux ARM64
GOOS=linux GOARCH=arm64 make build
```

## Debugging

### Enable Debug Logging
Set the log level in the driver configuration:
```yaml
args:
  - "--v=5"  # Verbose logging
```

### Common Debug Commands
```console
# Check driver logs
kubectl logs -l app=csi-rclone-controller
kubectl logs -l app=csi-rclone-node

# Check rclone installation on nodes
kubectl exec -it <node-pod> -- which rclone
kubectl exec -it <node-pod> -- rclone version

# Test rclone configuration
kubectl exec -it <node-pod> -- rclone lsd remote:path
```

### Local Development with MinIO
For local testing, you can use MinIO as an S3-compatible backend:

```console
# Start MinIO locally
docker run -p 9000:9000 -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"

# Test with rclone
rclone config create test s3 provider=Minio endpoint=http://localhost:9000 access_key_id=minioadmin secret_access_key=minioadmin
rclone lsd test:
```

## Architecture Overview

This driver is based on the [csi-driver-nfs](https://github.com/kubernetes-csi/csi-driver-nfs) reference implementation and draws inspiration from the original [csi-rclone](https://github.com/wunderio/csi-rclone) implementation by WunderIO.

The CSI driver consists of three main components:

### 1. Identity Server
- Provides driver information and capabilities
- Handles health checks
- Implements: `GetPluginInfo`, `GetPluginCapabilities`, `Probe`

### 2. Controller Server
- Manages volume lifecycle (create/delete)
- Validates volume parameters
- Implements: `CreateVolume`, `DeleteVolume`, `ValidateVolumeCapabilities`

### 3. Node Server
- Handles volume mounting/unmounting on nodes
- Manages rclone mount operations
- Implements: `NodePublishVolume`, `NodeUnpublishVolume`

### Key Design Decisions

1. **No Staging**: Rclone volumes don't require staging, so `NodeStageVolume`/`NodeUnstageVolume` are not implemented.

2. **Direct Rclone Integration**: Uses rclone's Go library directly instead of shelling out to the rclone binary.

3. **Remote Creation**: Creates temporary rclone remotes for each mount to avoid conflicts.

4. **VFS Caching**: Leverages rclone's VFS (Virtual File System) for improved performance with cloud storage.

5. **Template Variable Support**: Supports dynamic path substitution using PVC/PV metadata.

## Contributing

### Before Submitting Code
1. Ensure code passes linting: `golangci-lint run ./...`
2. Update documentation if needed
3. Add tests for new functionality

### Code Review Process
1. Create a feature branch
2. Make your changes
3. Run tests and linting
4. Submit a pull request
5. Address review feedback

### Release Process
1. Update version in `pkg/rclone/version.go`
2. Update CHANGELOG.md
3. Create a release tag
4. Build and push container images
5. Update documentation

## Troubleshooting

### Common Issues

#### Driver Won't Start
- Check that rclone is installed on all nodes
- Verify Kubernetes RBAC permissions
- Check driver logs for configuration errors

#### Volume Mount Failures
- Verify rclone configuration is correct
- Check network connectivity to storage backend
- Ensure credentials have proper permissions
- Check node resources (memory, disk space)

#### Performance Issues
- Adjust VFS cache settings
- Consider using `--vfs-cache-mode=full` for better performance
- Monitor disk usage for cache files
- Check network bandwidth to storage backend

### Getting Help
- Check the [troubleshooting guide](csi-debug.md)
- Review [driver parameters](driver-parameters.md)
- Open an issue on GitHub
- Check rclone documentation for backend-specific issues
