# Vendor Patch Management

This directory contains patches that are automatically applied to vendored dependencies during the build process.

## Why Patch Vendored Dependencies?

When using `go mod vendor`, all dependencies are copied into the `vendor/` directory. Sometimes we need to apply temporary fixes or workarounds to these dependencies before they are fixed upstream. However, any manual changes to `vendor/` are lost when running `go mod vendor` again.

This patch system solves this problem by:
1. Storing patches as separate `.patch` files
2. Automatically applying them after vendoring
3. Tracking what changes were made and why

## Current Patches

### go-fuse-v2.patch

**File**: `vendor/github.com/hanwen/go-fuse/v2/fs/bridge.go`

**Issue**: Multiple panics when running multiple pods with ReadWriteMany PVC access:
- **Index out of range** panics in `releaseFileEntry()` at line 948 with indices [1], [2], [3], etc.
- **Index out of range** panic in `GetAttr()` at line 554 accessing `b.files[fh]`
- **Nil pointer dereference** panics in `SetAttr()`, `Read()`, `Write()`, `Flush()`, `Fsync()`, and other file operations
- Panics occur progressively as state corruption accumulates

**Root Cause**: Race condition in go-fuse v2.8.0's file entry tracking where:
1. `entry.nodeIndex` becomes desynchronized from `n.openFiles` slice in concurrent multi-mount scenarios
2. `n.openFiles` contains stale/invalid file handle values that are out of bounds for `b.files` array
3. When one entry is removed, corrupted entries don't get updated, causing accumulating corruption
4. Multiple functions (`inode()`, `GetAttr()`, `releaseFileEntry()`) access `b.files[]` without bounds checking

**Fix**: Comprehensive multi-layered defensive approach across all vulnerable access points:

**In `inode()` function** (line 345-360):
- Add bounds check before accessing `b.files[fh]`
- Return `nil` fileEntry when file handle is out of bounds

**In `GetAttr()` function** (line 552-578):
- Add nil check for fEntry before accessing `.file`
- Add bounds checking in loop over `n.openFiles` before accessing `b.files[fh]`
- Skip corrupted file handle entries gracefully

**In `releaseFileEntry()` function** (line 958-1030):
1. **Early bounds check**: Verify `fh` is valid before accessing `b.files[fh]`
2. **Empty slice guard**: Check if `openFiles` is empty before processing
3. **nodeIndex Validation**: Verify `entry.nodeIndex` is valid AND points to correct file handle
4. **Linear Search Fallback**: When nodeIndex corrupted, search `openFiles` to find actual position
5. **Graceful Degradation**: Return `nil` when state is too corrupted to recover
6. **Safe array access**: Capture values before modification to prevent double-access bugs
7. **Additional bounds checks**: Verify `movedFileHandle` is valid before accessing `b.files`

**In `ReleaseDir()` function** (line 939-956):
- Add nil check matching existing `Release()` behavior
- Gracefully skip cleanup when releaseFileEntry returns nil

**In file operation functions** (`SetAttr`, `Read`, `Write`, `Flush`, `Fsync`, `Fallocate`, `GetLk`, `SetLk`, `SetLkw`):
- Add nil check for `fileEntry` returned from `inode()` before accessing `.file`
- Extract `FileHandle` safely and pass nil-safe handle to operations
- Prevents nil pointer dereference when file handle is corrupted/out-of-bounds

**Impact**: Prevents **all known panic scenarios** in concurrent multi-mount operations, enabling stable operation with 10+ pod replicas accessing same volume simultaneously.

**Status**: Comprehensive workaround for concurrent multi-mount scenarios. Temporary until upstream fix or upgrade to newer go-fuse version.

## Usage

### Automatic Application (Recommended)

Patches are automatically applied when building:

```bash
# Build the project (patches are applied automatically)
make rclone

# Or explicitly apply patches
make apply-patches
```

### Syncing Vendor Directory

When updating dependencies, use `vendor-sync` instead of `go mod vendor`:

```bash
# This runs 'go mod vendor' followed by patch application
make vendor-sync
```

### Manual Application

To manually apply patches:

```bash
# From project root
bash scripts/apply-vendor-patches.sh
```

## Creating New Patches

### Method 1: Using diff (Recommended)

1. Make a backup of the original file:
```bash
cp vendor/path/to/file.go vendor/path/to/file.go.orig
```

2. Edit the vendored file with your changes:
```bash
vim vendor/path/to/file.go
```

3. Generate the patch:
```bash
diff -u vendor/path/to/file.go.orig vendor/path/to/file.go > patches/descriptive-name.patch
```

4. Clean up the backup:
```bash
rm vendor/path/to/file.go.orig
```

### Method 2: Using git diff

If you're tracking vendor changes in git temporarily:

```bash
# Make your changes to the vendored file
vim vendor/path/to/file.go

# Generate the patch from unstaged changes
git diff vendor/path/to/file.go > patches/descriptive-name.patch
```

### Patch Naming Convention

Use descriptive names that include:
- Package name
- Version (if relevant)
- What the patch does

Examples:
- `go-fuse-v2.patch`
- `rclone-sftp-connection-pool-fix.patch`
- `kubernetes-client-timeout-increase.patch`

## Patch File Format

Patches must be in unified diff format (`diff -u`) and use `-p0` format (paths relative to project root):

```diff
--- vendor/package/file.go.orig	2025-10-28 00:00:00.000000000 +0000
+++ vendor/package/file.go	2025-10-28 00:00:00.000000000 +0000
@@ -10,7 +10,9 @@
 func example() {
-    // old code
+    // new code
+    // additional line
 }
```

## Troubleshooting

### Patch Fails to Apply

If a patch fails to apply:

1. Check if the vendored file has changed (dependency update):
```bash
go mod vendor
# Try applying the patch manually to see errors
patch -p0 --dry-run < patches/your-patch.patch
```

2. Regenerate the patch:
   - Update dependencies: `go mod vendor`
   - Recreate the patch following the steps above
   - Test the new patch: `make apply-patches`

### Patch Already Applied

The script automatically detects if a patch is already applied and skips it:
```
⊙ Patch example.patch already applied (skipping)
```

This is normal and not an error.

### Conflicts with Multiple Patches

If multiple patches modify the same file:
1. Apply patches in order (alphabetically by filename)
2. Consider combining related patches into one
3. Test thoroughly after applying all patches

## Best Practices

1. **Document Why**: Always include comments in the patch explaining:
   - What issue it fixes
   - Link to upstream issue/PR if available
   - When it can be removed (upstream fix version)

2. **Keep Patches Small**: Each patch should address one specific issue

3. **Test After Applying**: Run tests after applying patches:
```bash
make apply-patches
make unit-test
```

4. **Review Before Committing**: Always review patch contents before committing to repository

5. **Plan for Removal**: Patches should be temporary. Track upstream fixes and remove patches when possible.

## Workflow Examples

### Updating Dependencies

```bash
# Update go.mod
go get -u github.com/some/package

# Sync vendor with patches
make vendor-sync

# If patches fail, regenerate them
# (follow "Creating New Patches" steps above)

# Test
make unit-test

# Commit
git add go.mod go.sum vendor/ patches/
git commit -m "Update dependencies and regenerate patches"
```

### Adding a New Patch

```bash
# Start with clean vendor
make vendor-sync

# Make changes to vendored file and create patch
# (follow "Creating New Patches" steps above)

# Test the patch
make apply-patches
make unit-test

# Commit
git add patches/new-patch.patch
git commit -m "Add patch for [specific issue]"
```

## CI/CD Integration

The patch system is integrated into the build process:
- `make rclone` automatically applies patches before building
- `make container` builds the container with patched dependencies
- CI pipelines should use standard make targets (patches apply automatically)

## Maintenance

### When to Remove a Patch

Remove a patch when:
1. Upstream has fixed the issue
2. We've upgraded to a version with the fix
3. The workaround is no longer needed

To remove a patch:
```bash
# Remove the patch file
rm patches/patch-name.patch

# Sync vendor to clean state
make vendor-sync

# Verify build still works
make rclone
make unit-test
```

### Periodic Review

Regularly review patches (e.g., quarterly):
1. Check if upstream issues are fixed
2. Test removing patches one by one
3. Update patch comments with current status
4. Consider upgrading dependencies to avoid patches

## Support

For questions or issues with the patch system:
1. Check this README
2. Review existing patches for examples
3. Open an issue in the project repository

