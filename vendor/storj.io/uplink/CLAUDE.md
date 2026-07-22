# CLAUDE.md

This file provides guidance for Claude Code when working with the storj/uplink repository.

## Project Overview

Libuplink is the official Go library for the Storj V3 decentralized storage network. It provides end-to-end encrypted access to Storj distributed cloud storage with S3-compatible semantics.

**Module**: `storj.io/uplink`
**Go Version**: 1.24.0
**License**: MIT

## Build & Test Commands

```bash
# Run tests
go test ./...

# Run vet
go vet ./...

# Update dependencies to latest
make bump-dependencies
```

## Project Structure

- Root `*.go` files - Public API (access.go, project.go, bucket.go, object.go, upload.go, download.go, etc.)
- `private/` - Internal implementation (storage streams, erasure coding client, metaclient, transport)
- `edge/` - Edge credentials API
- `testsuite/` - Integration tests (has separate go.mod)
- `examples/walkthrough/` - Usage example

## Key Concepts

- **Access Grants**: API keys + encryption info + satellite address; can be restricted for multitenancy
- **Projects**: Entry point for bucket/object operations, opened with Access Grant
- **Buckets**: Collections of encrypted objects (bucket names not encrypted)
- **Objects**: End-to-end encrypted data with metadata, supporting streaming and random-access reads

## Code Patterns

### Iterator Pattern
List operations return iterators with `Next()`, `Item()`, `Err()` methods:
```go
it := project.ListBuckets(ctx, nil)
for it.Next() {
    bucket := it.Item()
}
if err := it.Err(); err != nil { ... }
```

### Resource Cleanup
Always defer Close() on resources:
```go
project, err := uplink.OpenProject(ctx, access)
if err != nil { ... }
defer project.Close()
```

### Error Handling
Custom error types in `common.go`. Use `github.com/zeebo/errs` for error wrapping.

## Development Guidelines

- All files require Storj copyright header
- Tests required for all changes (especially bug fixes)
- Backward compatibility is high priority
- Use semantic versioning
- Code review via Gerrit (review.dev.storj.tools)

## Key Dependencies

- `storj.io/common` - Common Storj utilities
- `storj.io/drpc` - Distributed RPC
- `storj.io/infectious` - Erasure coding
- `github.com/zeebo/errs` - Error handling
