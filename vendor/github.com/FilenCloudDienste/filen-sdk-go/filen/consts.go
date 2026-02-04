package filen

import "github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"

const (
	// ChunkSize is the maximum size of a chunk in bytes
	// as defined by Filen when uploading files (1 MiB)
	ChunkSize = 1024 * 1024
	// MaxSmallCallers is the maximum number of concurrent goroutines
	// running at the same time making smaller requests, this is mostly arbitrary
	// and mainly used to limit the number of API calls during a large DirMove from rclone
	MaxSmallCallers = 64
	// MaxDownloadThreadsPerFile is the number of chunks to keep in memory at once.
	// This controls memory usage during downloads.
	V2AccountFileEncryptionVersion     = crypto.FileEncryptionVersion(2)
	V2AccountMetadataEncryptionVersion = crypto.MetadataEncryptionVersion(2)
)
