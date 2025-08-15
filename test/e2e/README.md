# E2E Tests for CSI Driver Rclone

This directory contains end-to-end (E2E) tests for the csi-driver-rclone project, following the testing patterns established in csi-driver-nfs.

## Structure

```
test/
├── e2e/                        # End-to-end tests
│   ├── driver/                 # Driver interface implementations
│   │   ├── driver.go          # Base driver interfaces
│   │   └── rclone_driver.go   # Rclone-specific driver implementation
│   ├── testsuites/             # Test suite definitions
│   │   ├── testsuites.go      # Helper functions and test utilities
│   │   ├── specs.go           # Test specifications and volume details
│   │   ├── dynamically_provisioned_cmd_volume_tester.go
│   │   ├── dynamically_provisioned_delete_pod_tester.go
│   │   ├── dynamically_provisioned_read_only_volume_tester.go
│   │   ├── dynamically_provisioned_collocated_pod_tester.go
│   │   ├── dynamically_provisioned_pod_with_multiple_pv.go
│   │   ├── dynamically_provisioned_inline_volume.go
│   │   ├── dynamically_provisioned_reclaim_policy_tester.go
│   │   └── dynamically_provisioned_volume_subpath_tester.go
│   ├── dynamic_provisioning_test.go  # Main dynamic provisioning tests
│   └── e2e_suite_test.go      # Test suite setup and teardown
├── external-e2e/               # Kubernetes external storage tests
│   ├── run.sh                 # Test runner script
│   └── testdriver.yaml        # Driver capabilities configuration
└── integration/                # Integration tests
```

## Test Categories

### Dynamic Provisioning Tests

The tests cover various scenarios for dynamically provisioned volumes:

1. **Basic Volume Provisioning**: Create a volume with rclone parameters and verify basic read/write operations
2. **Pod Deletion and Recreation**: Test data persistence across pod deletions
3. **Multiple Volumes per Pod**: Verify a single pod can mount multiple PVCs
4. **Read-only Volumes**: Ensure read-only mounts are enforced
5. **Collocated Pods**: Test multiple pods on the same node accessing different volumes
6. **Volume Subpath**: Test mounting volumes with subpaths
7. **CSI Inline Volumes**: Test ephemeral inline volumes without PVCs
8. **Reclaim Policies**: Test PV retention and deletion with different reclaim policies (Delete/Retain)

### External E2E Tests

Located in `test/external-e2e/`, these tests use Kubernetes' external storage test suite to validate CSI driver compliance. See the [external-e2e documentation](../external-e2e/) for more details.

### Unsupported Features

The following features are **not supported** by rclone CSI driver (and therefore not tested):

- **Volume Snapshots**: Cloud storage backends handle snapshots differently
- **Volume Expansion/Resize**: Cloud storage doesn't support resizing volumes
- **Volume Cloning**: Not applicable for cloud object storage
- **Block Volumes**: Only filesystem volumes are supported (rclone uses FUSE mounts)

## Prerequisites

Before running the E2E tests, you need:

1. A Kubernetes cluster (can be local like kind, minikube, or remote)
2. kubectl configured to access the cluster
3. The CSI driver rclone installed and running in the cluster
4. Rclone configured with test remotes (e.g., a test-remote pointing to S3, GCS, or local storage)

## Configuration

### Environment Variables

- `KUBECONFIG`: Path to kubeconfig file (defaults to `~/.kube/config`)
- `NODE_ID`: Node identifier for the CSI driver
- `RCLONE_CSI_DRIVER`: Override the default driver name (default: `rclone.csi.veloxpack.io`)
- `TEST_WINDOWS`: Set to run Windows-specific test commands

### Storage Class Parameters

Default storage class parameters used in tests:

```yaml
# Basic storage class
parameters:
  remote: "test-remote"
  remotePath: "test"
  csi.storage.k8s.io/provisioner-secret-name: "mount-options"
  csi.storage.k8s.io/provisioner-secret-namespace: "default"

# Dynamic path storage class (multi-tenant isolation)
parameters:
  remote: "test-remote"
  remotePath: "test/${pvc.metadata.namespace}/${pvc.metadata.name}"
  csi.storage.k8s.io/provisioner-secret-name: "mount-options"
  csi.storage.k8s.io/provisioner-secret-namespace: "default"
```

**Template Variables**: The driver supports dynamic path substitution in `remotePath`:
- `${pvc.metadata.name}` - Name of the PersistentVolumeClaim
- `${pvc.metadata.namespace}` - Namespace of the PersistentVolumeClaim
- `${pv.metadata.name}` - Name of the PersistentVolume

## Running the Tests

### Prerequisites Setup

1. **Install the CSI driver**:
   ```bash
   kubectl apply -f deploy/kubernetes/
   ```

2. **Create necessary secrets** for rclone configuration:
   ```bash
   # Example: Create secret with rclone config for S3
   kubectl create secret generic mount-options \
     --from-literal=remote="test-remote" \
     --from-literal=remotePath="test" \
     --from-literal=configData="$(cat <<EOF
   [test-remote]
   type = s3
   provider = AWS
   access_key_id = YOUR_ACCESS_KEY
   secret_access_key = YOUR_SECRET_KEY
   region = us-east-1
   EOF
   )" \
     --namespace=default

   # Or for local testing with MinIO
   kubectl create secret generic mount-options \
     --from-literal=remote="test-remote" \
     --from-literal=remotePath="test" \
     --from-literal=configData="$(cat <<EOF
   [test-remote]
   type = s3
   provider = Minio
   endpoint = http://minio.default.svc.cluster.local:9000
   access_key_id = minioadmin
   secret_access_key = minioadmin
   EOF
   )" \
     --namespace=default
   ```

### Run Tests

```bash
# From the project root
cd test/e2e

# Run all E2E tests
go test -v -timeout=30m

# Run specific test
go test -v -timeout=30m -ginkgo.focus="should create a volume on demand"

# Run with custom kubeconfig
KUBECONFIG=/path/to/kubeconfig go test -v -timeout=30m
```

### Using Make Targets

If Make targets are configured in the project:

```bash
# Run E2E tests
make e2e-test

# Bootstrap E2E environment
make e2e-bootstrap

# Cleanup E2E environment
make e2e-teardown
```

### Running External E2E Tests

External E2E tests validate the driver against Kubernetes' official storage test suite:

```bash
# From the project root
cd test/external-e2e

# Run external e2e tests
./run.sh

# Or set up and run manually
export KUBECONFIG=~/.kube/config
chmod +x run.sh
./run.sh
```

The external tests will:
1. Download Kubernetes test binaries
2. Deploy the CSI driver
3. Run external storage tests
4. Collect driver logs
5. Clean up resources

## Writing New Tests

### Test Structure

Each test follows this general pattern:

```go
ginkgo.It("should test some functionality", func(ctx ginkgo.SpecContext) {
    pods := []testsuites.PodDetails{
        {
            Cmd: "test command",
            Volumes: []testsuites.VolumeDetails{
                {
                    ClaimSize: "10Gi",
                    VolumeMount: testsuites.VolumeMountDetails{
                        NameGenerate:      "test-volume-",
                        MountPathGenerate: "/mnt/test-",
                    },
                },
            },
        },
    }
    test := testsuites.DynamicallyProvisionedCmdVolumeTest{
        CSIDriver:              testDriver,
        Pods:                   pods,
        StorageClassParameters: defaultStorageClassParameters,
    }
    test.Run(ctx, cs, ns)
})
```

### Adding New Test Suites

To add a new test suite:

1. Create a new file in `testsuites/` directory
2. Define a test struct with required fields
3. Implement the `Run` method with test logic
4. Add test cases in `dynamic_provisioning_test.go`

Example:

```go
type MyNewTest struct {
    CSIDriver              driver.DynamicPVTestDriver
    Pods                   []PodDetails
    StorageClassParameters map[string]string
}

func (t *MyNewTest) Run(ctx context.Context, client clientset.Interface, namespace *v1.Namespace) {
    // Test implementation
}
```

## Test Framework

The tests use:
- **Ginkgo v2**: BDD-style testing framework
- **Gomega**: Matcher library for assertions
- **Kubernetes E2E Framework**: Provides utilities for Kubernetes testing

## Debugging

### View Test Logs

```bash
# Run with verbose output
go test -v -ginkgo.v

# Show detailed pod logs for controller
kubectl logs -n kube-system -l app=csi-rclone-controller --all-containers=true

# Show detailed pod logs for node daemonset
kubectl logs -n kube-system -l app=csi-rclone-node --all-containers=true

# Check CSI driver status
kubectl get csidrivers
kubectl get pods -n kube-system | grep rclone

# Describe CSI driver
kubectl describe csidriver rclone.csi.veloxpack.io
```

### Common Issues

1. **PVC stuck in Pending**: Check if CSI driver pods are running and healthy
2. **Mount failures**: Verify rclone configuration and remote accessibility
3. **Permission errors**: Ensure proper RBAC permissions for the CSI driver

## CI/CD Integration

To integrate E2E tests into CI/CD:

```yaml
# Example GitHub Actions workflow
- name: Run E2E Tests
  run: |
    kind create cluster
    kubectl apply -f deploy/kubernetes/
    cd test/e2e
    go test -v -timeout=30m
```

## Contributing

When adding new tests:
- Follow existing test patterns
- Ensure tests are idempotent
- Clean up resources properly (use defer for cleanup)
- Add appropriate test descriptions using `ginkgo.By()`
- Document any special setup requirements

## References

- [Kubernetes CSI Developer Documentation](https://kubernetes-csi.github.io/docs/)
- [Ginkgo Testing Framework](https://onsi.github.io/ginkgo/)
- [Gomega Matcher Library](https://onsi.github.io/gomega/)
- [csi-driver-nfs Tests](https://github.com/kubernetes-csi/csi-driver-nfs/tree/master/test) - Reference implementation
- [Rclone Documentation](https://rclone.org/docs/)
- [CSI Driver Rclone Examples](../../deploy/example/) - Example configurations and templates
