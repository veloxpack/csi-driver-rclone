# CSI Rclone Driver Helm Charts

This directory contains Helm charts for the CSI Rclone Driver, which allows you to mount cloud storage as persistent volumes in Kubernetes.

## Charts

- **csi-driver-rclone**: The main chart for deploying the CSI Rclone Driver

## Quick Start

1. Add the Helm repository:
   ```bash
   helm repo add csi-rclone https://veloxpack.github.io/csi-driver-rclone/charts
   helm repo update
   ```

2. Install the driver:
   ```bash
   helm install csi-rclone csi-rclone/csi-driver-rclone
   ```

3. Create a StorageClass:
   ```yaml
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: rclone-s3
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
   reclaimPolicy: Delete
   volumeBindingMode: Immediate
   mountOptions:
     - vfs-cache-mode=writes
     - vfs-cache-max-size=10G
   ```

## Configuration

The chart supports various configuration options through the `values.yaml` file:

- **Image configuration**: Customize container images and tags
- **Resource limits**: Set CPU and memory limits for containers
- **Node selection**: Control where pods are scheduled
- **Tolerations**: Handle node taints
- **Storage classes**: Create multiple storage classes with different configurations

## Examples

See the `examples/` directory for various configuration examples:

- S3 storage class
- Google Cloud Storage storage class
- Azure Blob Storage storage class
- MinIO storage class

## Documentation

For detailed documentation, see:
- [Main README](../README.md)
- [Driver Parameters](../docs/driver-parameters.md)
- [Installation Guide](../docs/install-rclone-csi-driver.md)

## Support

- GitHub Issues: https://github.com/veloxpack/csi-driver-rclone/issues
- Documentation: https://github.com/veloxpack/csi-driver-rclone/blob/main/README.md
