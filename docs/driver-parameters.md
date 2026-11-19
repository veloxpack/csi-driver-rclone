## Driver Parameters
> This driver requires rclone to be installed on all nodes and supports dynamic provisioning of Persistent Volumes via Persistent Volume Claims by mounting cloud storage backends using rclone.

### Storage Class Usage (Dynamic Provisioning)
> [`StorageClass` example](../deploy/example/storageclass-rclone.yaml)

| Name | Meaning | Example Value | Mandatory | Default Value |
|------|---------|---------------|-----------|---------------|
| `remote` | Rclone remote name (backend type) | `"s3"`, `"gcs"`, `"azureblob"`, `"dropbox"` | Yes | - |
| `remotePath` | Path within the remote storage | `"my-bucket/folder"`, `"/data"` | No | Root of remote |
| `configData` | Inline rclone configuration (INI format) | See examples below | No | - |

**VolumeID Format:**
```
{remote}#{volume-name}
```
> Example: `s3#my-pvc-12345`

### PersistentVolume/PersistentVolumeClaim Usage (Static Provisioning)
> [`PersistentVolume` example](../deploy/example/pv-rclone-csi.yaml)

| Name | Meaning | Example Value | Mandatory | Default Value |
|------|---------|---------------|-----------|---------------|
| `volumeHandle` | Unique identifier for the volume | `"s3#my-volume"` | Yes | - |
| `volumeAttributes.remote` | Rclone remote name | `"s3"`, `"gcs"`, `"azureblob"` | Yes | - |
| `volumeAttributes.remotePath` | Path within remote storage | `"my-bucket/folder"` | No | Root of remote |
| `volumeAttributes.configData` | Inline rclone configuration | See examples below | No | - |

### Configuration Methods

The driver accepts rclone configuration (`remote`, `remotePath`, `configData`, and mount options) from three places: a Kubernetes Secret, StorageClass `parameters`, and PersistentVolume `volumeAttributes`. Keep credentials in a Secret. The other two are stored in plain text in the cluster API.

#### 1. Kubernetes Secrets (Recommended for Sensitive Data)

##### How a Secret is linked to a volume

The driver doesn't look for a Secret by a fixed name. The StorageClass (dynamic provisioning) or the PersistentVolume (static provisioning) references it with the standard CSI node-publish secret fields. When a pod mounts the volume, the kubelet reads that Secret and passes its keys to the driver's `NodePublishVolume` call, where they are merged with the other parameters (see [Parameter Priority](#parameter-priority)).

```
StorageClass ──(node-publish-secret-name/namespace)──▶ Secret
     ▲                                                    │
     │ storageClassName                                   │ keys merged at mount time
     │                                                    ▼
    PVC ──────────────▶ PV (created by the provisioner) ──▶ rclone mount
```

- The Secret can have **any name** and live in **any namespace**. `rclone-secret` / `default` in the examples are only placeholders.
- Any parameter can go in the Secret, including `remote` and `remotePath`. Then the StorageClass only needs the Secret reference.
- The CSI node plugin needs no extra RBAC. The kubelet fetches the Secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-s3-creds
  namespace: my-namespace
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

##### Option A: One Secret per StorageClass

Every PVC that uses this StorageClass mounts with the same credentials. For several backends, create one StorageClass per backend.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rclone-s3
provisioner: rclone.csi.veloxpack.io
parameters:
  csi.storage.k8s.io/node-publish-secret-name: "my-s3-creds"
  csi.storage.k8s.io/node-publish-secret-namespace: "my-namespace"
```

> Full example: [`rclone-secret-storageclass.yaml`](../deploy/example/rclone-secret-storageclass.yaml)

##### Option B: Each PVC selects its own Secret (multiple endpoints, one StorageClass)

The CSI external-provisioner resolves template variables in the secret reference when the volume is provisioned. This lets each PVC name its own Secret with an annotation, and keeps the Secret in the PVC's namespace:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rclone-csi-dynamic
provisioner: rclone.csi.veloxpack.io
parameters:
  csi.storage.k8s.io/node-publish-secret-name: ${pvc.annotations['rclone.csi.veloxpack.io/secret-name']}
  csi.storage.k8s.io/node-publish-secret-namespace: ${pvc.namespace}
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: team-a-data
  namespace: team-a
  annotations:
    rclone.csi.veloxpack.io/secret-name: team-a-s3   # Secret in namespace team-a
spec:
  storageClassName: rclone-csi-dynamic
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
```

Supported templates for `node-publish-secret-name`: `${pv.name}`, `${pvc.namespace}`, `${pvc.name}`, `${pvc.annotations['<key>']}`. For `node-publish-secret-namespace`: `${pv.name}`, `${pvc.namespace}`. See the [external-provisioner documentation](https://kubernetes-csi.github.io/docs/secrets-and-credentials-storage-class.html) for details.

The secret reference is resolved once, when the PV is created. Changing the annotation on an existing PVC has no effect. Changing the *contents* of the referenced Secret takes effect on the next mount.

> Full example with two independent backends: [`multi-tenant.yaml`](../deploy/example/multi-tenant.yaml)

##### Static provisioning: `nodePublishSecretRef`

For a PersistentVolume you create yourself, reference the Secret directly and keep only non-sensitive values in `volumeAttributes`:

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
    nodePublishSecretRef:
      name: my-s3-creds
      namespace: my-namespace
```

#### 2. Inline Configuration in StorageClass

> **Not recommended for credentials.** Any user who can read StorageClasses can read StorageClass parameters, and they are copied into each PV's `volumeAttributes`. Use this for non-sensitive settings, or for backends that need no credentials, and put secrets in a Secret (see above).

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
```

#### 3. PersistentVolume volumeAttributes

> **Not recommended for credentials.** Like StorageClass parameters, `volumeAttributes` are stored in plain text on the PV object. Prefer `nodePublishSecretRef` (see above) outside of testing.

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
        provider = Minio
        endpoint = https://play.min.io
        access_key_id = Q3AM3UQ867SPQQA43P2F
        secret_access_key = zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG
```

### Parameter Priority
Configuration parameters are merged in the following order (later values override earlier ones):
1. **Secrets** (lowest priority)
2. **StorageClass parameters** (medium priority)
3. **PersistentVolume volumeAttributes** (highest priority)

> With dynamic provisioning, StorageClass parameters are copied into the PV's `volumeAttributes`, so a key set in both the StorageClass and the Secret always takes the StorageClass value. If each PVC's Secret should control `remote`/`remotePath` (Option B above), leave those keys out of the StorageClass.

### Dynamic Path Substitution
The `remotePath` parameter supports template variables that are replaced with PVC/PV metadata:

| Template Variable | Replaced With | Example |
|------------------|---------------|---------|
| `${pvc.metadata.name}` | PVC name | `my-pvc-12345` |
| `${pvc.metadata.namespace}` | PVC namespace | `default` |
| `${pv.metadata.name}` | PV name | `pv-rclone-abc123` |

**Example:**
```yaml
parameters:
  remote: "s3"
  remotePath: "my-bucket/${pvc.metadata.namespace}/${pvc.metadata.name}"
```

### Supported Rclone Backends
The driver supports all rclone backends. Common examples:

#### Amazon S3
```yaml
configData: |
  [s3]
  type = s3
  provider = AWS
  access_key_id = YOUR_ACCESS_KEY_ID
  secret_access_key = YOUR_SECRET_ACCESS_KEY
  region = us-east-1
```

#### Google Cloud Storage
```yaml
configData: |
  [gcs]
  type = google cloud storage
  project_number = 12345678
  service_account_file = /path/to/service-account.json
```

#### Azure Blob Storage
```yaml
configData: |
  [azureblob]
  type = azureblob
  account = your_storage_account
  key = your_storage_key
  endpoint = https://your_storage_account.blob.core.windows.net/
```

#### MinIO (S3-Compatible)
```yaml
configData: |
  [minio]
  type = s3
  provider = Minio
  endpoint = http://minio.minio:9000
  access_key_id = minioadmin
  secret_access_key = minioadmin
```

#### Dropbox
```yaml
configData: |
  [dropbox]
  type = dropbox
  token = {"access_token":"your_token","token_type":"bearer","expiry":"2024-01-01T00:00:00Z"}
```

#### SFTP
```yaml
configData: |
  [sftp]
  type = sftp
  host = sftp.example.com
  user = username
  pass = obscured_password
```

### Mount Options
The driver supports rclone mount flags through two methods:

1. **PersistentVolume mountOptions** - Standard Kubernetes mount options
2. **Direct rcloneplugin flags** - Many rclone mount flags are directly supported as command-line options

For a complete list of supported flags and their descriptions, see the [official rclone mount documentation](https://rclone.org/commands/rclone_mount/).

#### Method 1: PersistentVolume mountOptions
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-rclone-example
spec:
  mountOptions:
    - allow-non-empty=true
    - debug-fuse=true
    - vfs-cache-mode=writes
    - vfs-cache-max-size=10G
    - dir-cache-time=30s
  csi:
    driver: rclone.csi.veloxpack.io
    volumeHandle: rclone-example-volume
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

#### Method 2: Direct rcloneplugin Flags
The rcloneplugin binary directly supports many rclone mount flags as command-line options. These can be configured in the CSI driver deployment.

**Common Debug Options:**
- `--debug-fuse` - Debug FUSE internals
- `--allow-non-empty` - Allow mounting over non-empty directory
- `--allow-other` - Allow access to other users

**Common Performance Options:**
- `--vfs-cache-mode=writes` - Cache mode (off|minimal|writes|full)

**Complete Flag Reference:**
For the complete list of all supported flags and their descriptions, see the [official rclone mount documentation](https://rclone.org/commands/rclone_mount/).

### Troubleshooting Parameters

#### Common Issues
1. **Invalid remote configuration**: Check that `configData` is valid INI format
2. **Authentication failures**: Verify credentials in `configData`
3. **Network connectivity**: Ensure nodes can reach the storage backend
4. **Permission errors**: Check that credentials have proper access rights

#### Debug Mode
Enable debug logging using either method. For all available logging options, see the [rclone mount documentation](https://rclone.org/commands/rclone_mount/):

**Method 1: PersistentVolume mountOptions**
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-rclone-debug
spec:
  mountOptions:
    - debug-fuse=true
    - vv
    - log-level=DEBUG
  csi:
    driver: rclone.csi.veloxpack.io
    volumeHandle: debug-volume
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

**Method 2: Direct rcloneplugin flags**
Configure in the CSI driver deployment (DaemonSet/Deployment):
```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: csi-rclone-node
spec:
  template:
    spec:
      containers:
      - name: rcloneplugin
        image: csi-rclone:latest
        args:
          - --debug-fuse=true
          - -vv
          - --log-level=DEBUG
```

#### Validation
Test your rclone configuration locally before using it in Kubernetes:
```bash
# Test remote configuration
rclone lsd remote:path

# Test mount
rclone mount remote:path /tmp/mount --daemon
```

### Remote Control (RC) API
The CSI node plugin can expose rclone's [Remote Control API](https://rclone.org/rc/) for operational tasks (e.g., triggering cache refreshes). The API is disabled by default and must be explicitly enabled in the Helm chart:

```yaml
node:
  rc:
    enabled: true
    addr: ":5573"
    service:
      enabled: true
      annotations:
        prometheus.io/scrape: "false"
    basicAuth:
      existingSecret: rclone-rc-auth
```

**Security Notes**
- Always keep `rc.noAuth` at `false` unless you understand the risks.
- Use a Kubernetes secret (`node.rc.basicAuth.existingSecret`) to provide the `username` and `password` keys consumed by the DaemonSet.
- Expose the RC service only on internal networks; the default service template creates a headless `ClusterIP`.

Once enabled, any in-cluster workload can reach the RC service on the published port. For example:

```bash
kubectl run rc-shell --image=ghcr.io/veloxpack/csi-driver-rclone \
  --rm -it -- bash -lc 'curl -u user:pass http://csi-rclone-node-rc.veloxpack.svc:5573/core/stats'
```
