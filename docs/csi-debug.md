## CSI Driver Debug Tips

### Case #1: Volume Create/Delete Failed
> There could be multiple controller pods (only one pod is the leader), if there are no helpful logs, try to get logs from the leader controller pod.

#### Find CSI Driver Controller Pod
```console
$ kubectl get pod -o wide | grep csi-rclone-controller
NAME                                     READY   STATUS    RESTARTS   AGE     IP             NODE
csi-rclone-controller-56bfddd689-dh5tk   5/5     Running   0          35s     10.240.0.19    k8s-agentpool-22533604-0
csi-rclone-controller-56bfddd689-sl4ll   5/5     Running   0          35s     10.240.0.23    k8s-agentpool-22533604-1
```

#### Get Pod Description and Logs
```console
$ kubectl describe pod csi-rclone-controller-56bfddd689-dh5tk > csi-rclone-controller-description.log
$ kubectl logs csi-rclone-controller-56bfddd689-dh5tk -c rclone > csi-rclone-controller.log
```

#### Common Controller Issues
- **Invalid remote configuration**: Check `configData` parameter format
- **Missing required parameters**: Verify `remote` parameter is provided
- **Template variable errors**: Check `remotePath` template syntax
- **Storage class configuration**: Verify StorageClass parameters

### Case #2: Volume Mount/Unmount Failed
> Locate CSI driver pod that does the actual volume mount/unmount on the specific node.

#### Find CSI Driver Node Pod
```console
$ kubectl get pod -o wide | grep csi-rclone-node
NAME                                      READY   STATUS    RESTARTS   AGE     IP             NODE
csi-rclone-node-cvgbs                     3/3     Running   0          7m4s    10.240.0.35    k8s-agentpool-22533604-1
csi-rclone-node-dr4s4                     3/3     Running   0          7m4s    10.240.0.4     k8s-agentpool-22533604-0
```

#### Get Pod Description and Logs
```console
$ kubectl describe po csi-rclone-node-cvgbs > csi-rclone-node-description.log
$ kubectl logs csi-rclone-node-cvgbs -c rclone > csi-rclone-node.log
```

#### Check Rclone Mount Inside Driver
```console
kubectl exec -it csi-rclone-node-cvgbs -c rclone -- mount | grep rclone
kubectl exec -it csi-rclone-node-cvgbs -c rclone -- ps aux | grep rclone
```

#### Common Node Issues
- **Rclone not installed**: Check if rclone binary is available on the node
- **Authentication failures**: Verify credentials in secrets or configData
- **Network connectivity**: Test connection to storage backend
- **Permission errors**: Check file system permissions and mount options
- **Resource constraints**: Verify sufficient memory and disk space

### Case #3: Driver Functionality Issues

#### Check Driver Binary
```console
# Check if the driver binary is working
kubectl exec -it <node-pod> -- /rcloneplugin --help

# Check driver version information (shows when driver starts)
kubectl logs -l app=csi-rclone-node --tail=10 | grep "DRIVER INFORMATION" -A 10

# Check driver logs for rclone integration
kubectl logs -l app=csi-rclone-node --tail=100
```

#### Verify Rclone Library Integration
The driver uses rclone as a Go library directly, so there's no separate rclone binary to install. Check that the driver container includes the rclone functionality:

```console
# Check container image
kubectl describe pod <node-pod> | grep Image

# Verify rclone backends are available
kubectl exec -it <node-pod> -- /rcloneplugin --help | grep -i backend
```

### Case #4: Authentication and Configuration Issues

#### Test Rclone Configuration
```console
# Test remote configuration through the driver
kubectl exec -it <node-pod> -- /rcloneplugin --help

# Check driver logs for configuration parsing
kubectl logs -l app=csi-rclone-node --tail=50 | grep -i config
```

#### Debug Configuration Data
```console
# Check parsed configuration
kubectl exec -it <node-pod> -- cat /tmp/rclone.conf

# Validate INI format
kubectl exec -it <node-pod> -- rclone config show --config /tmp/rclone.conf
```

#### Common Configuration Issues
- **Invalid INI format**: Check `configData` syntax
- **Missing credentials**: Verify all required authentication parameters
- **Wrong endpoint**: Check storage backend URL
- **Region mismatch**: Verify AWS/GCS region settings
- **Permission denied**: Check credential permissions

### Case #5: Network Connectivity Issues

#### Test Storage Backend Connectivity
```console
# Test basic connectivity
kubectl exec -it <node-pod> -- ping storage-backend.com

# Test HTTPS connectivity
kubectl exec -it <node-pod> -- curl -I https://storage-backend.com

# Test with rclone
kubectl exec -it <node-pod> -- rclone lsd remote: --log-level DEBUG
```

#### Common Network Issues
- **DNS resolution**: Check if storage backend hostname resolves
- **Firewall rules**: Verify outbound HTTPS/HTTP access
- **Proxy settings**: Check if cluster uses HTTP proxy
- **SSL/TLS issues**: Verify certificate validity
- **Rate limiting**: Check for API rate limits

### Case #6: Performance Issues

#### Monitor Rclone Performance
```console
# Check rclone stats
kubectl exec -it <node-pod> -- rclone rc core/stats

# Check VFS cache status
kubectl exec -it <node-pod> -- rclone rc vfs/stats

# Monitor system resources
kubectl exec -it <node-pod> -- top
kubectl exec -it <node-pod> -- df -h
```

#### Common Performance Issues
- **Slow uploads/downloads**: Check network bandwidth and latency
- **High memory usage**: Adjust VFS cache settings
- **Disk space issues**: Monitor cache directory usage
- **CPU usage**: Check for excessive rclone processes

### Case #7: Volume Lifecycle Issues

#### Check Volume Status
```console
# Check PVC status
kubectl get pvc -o wide

# Check PV status
kubectl get pv -o wide

# Check pod volume mounts
kubectl describe pod <pod-name>
```

#### Debug Volume Operations
```console
# Check volume events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check CSI driver logs
kubectl logs -l app=csi-rclone-controller --tail=100
kubectl logs -l app=csi-rclone-node --tail=100
```

### Case #8: Troubleshooting Connection Failure on Agent Node

#### Manual Mount Test
```console
# Create test directory
mkdir /tmp/test

# Test manual mount
rclone mount remote:path /tmp/test --daemon --log-level DEBUG

# Check mount status
mount | grep rclone
ps aux | grep rclone

# Unmount test
fusermount -u /tmp/test
```

#### Debug Mount Issues
```console
# Check FUSE installation
kubectl exec -it <node-pod> -- which fusermount
kubectl exec -it <node-pod> -- lsmod | grep fuse

# Check mount permissions
kubectl exec -it <node-pod> -- ls -la /tmp/test
```

### Case #9: Secret and RBAC Issues

#### Check Secret Configuration
```console
# Verify secret exists
kubectl get secret rclone-secret -o yaml

# Check secret content
kubectl get secret rclone-secret -o jsonpath='{.data.configData}' | base64 -d

# Test secret access
kubectl exec -it <node-pod> -- ls /var/run/secrets/kubernetes.io/serviceaccount/
```

#### Check RBAC Permissions
```console
# Check service account
kubectl get serviceaccount csi-rclone-controller -n kube-system

# Check cluster role
kubectl get clusterrole csi-rclone-controller

# Check cluster role binding
kubectl get clusterrolebinding csi-rclone-controller
```

### Case #10: Log Analysis

#### Enable Debug Logging
```yaml
# In driver deployment
args:
  - "--v=5"  # Verbose logging
  - "--logtostderr=true"
```

#### Common Log Patterns
- **"failed to create remote config"**: Configuration parsing error
- **"failed to initialize filesystem"**: Backend connection issue
- **"failed to mount"**: FUSE or permission issue
- **"authentication failed"**: Credential problem
- **"network error"**: Connectivity issue

#### Log Collection
```console
# Collect all relevant logs
kubectl logs -l app=csi-rclone-controller > controller.log
kubectl logs -l app=csi-rclone-node > node.log
kubectl get events --sort-by=.metadata.creationTimestamp > events.log
kubectl describe pvc <pvc-name> > pvc-description.log
kubectl describe pv <pv-name> > pv-description.log
```

### Case #11: Storage Backend Specific Issues

#### Amazon S3
```console
# Test S3 connectivity
kubectl exec -it <node-pod> -- rclone lsd s3:my-bucket --s3-provider=AWS --s3-region=us-east-1

# Check S3 permissions
kubectl exec -it <node-pod> -- rclone lsd s3:my-bucket --s3-provider=AWS --s3-region=us-east-1 --dry-run
```

#### Google Cloud Storage
```console
# Test GCS connectivity
kubectl exec -it <node-pod> -- rclone lsd gcs:my-bucket

# Check service account
kubectl exec -it <node-pod> -- cat /path/to/service-account.json
```

#### MinIO
```console
# Test MinIO connectivity
kubectl exec -it <node-pod> -- rclone lsd minio:my-bucket --s3-provider=Minio --s3-endpoint=http://minio:9000
```

### Case #12: Resource and Limits Issues

#### Check Resource Usage
```console
# Check pod resources
kubectl top pod -l app=csi-rclone-controller
kubectl top pod -l app=csi-rclone-node

# Check node resources
kubectl top nodes
kubectl describe node <node-name>
```

#### Adjust Resource Limits
```yaml
# In driver deployment
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Case #13: Cleanup and Recovery

#### Force Unmount Stuck Volumes
```console
# Find stuck mounts
kubectl exec -it <node-pod> -- mount | grep rclone

# Force unmount
kubectl exec -it <node-pod> -- fusermount -u /path/to/mount

# Kill stuck rclone processes
kubectl exec -it <node-pod> -- pkill -f rclone
```

#### Clean Up Resources
```console
# Delete PVC and PV
kubectl delete pvc <pvc-name>
kubectl delete pv <pv-name>

# Restart driver pods
kubectl rollout restart daemonset/csi-rclone-node -n kube-system
kubectl rollout restart deployment/csi-rclone-controller -n kube-system
```

### Getting Additional Help

1. **Check rclone documentation**: [https://rclone.org/](https://rclone.org/)
2. **Review CSI specification**: [https://github.com/container-storage-interface/spec](https://github.com/container-storage-interface/spec)
3. **Open an issue**: [https://github.com/veloxpack/csi-driver-rclone/issues](https://github.com/veloxpack/csi-driver-rclone/issues)
4. **Check Kubernetes CSI documentation**: [https://kubernetes-csi.github.io/docs/](https://kubernetes-csi.github.io/docs/)
