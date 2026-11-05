// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

package storj

// ObjectChecksumAlgorithm identifies the algorithm used to compute an object's checksum.
type ObjectChecksumAlgorithm int

const (
	// ObjectChecksumAlgorithmNone indicates that no algorithm was used to compute an object's checksum.
	// This is used when an object was not uploaded with a checksum.
	ObjectChecksumAlgorithmNone = ObjectChecksumAlgorithm(0)

	// ObjectChecksumAlgorithmCRC32 indicates that the CRC32 algorithm was used to compute an object's checksum.
	ObjectChecksumAlgorithmCRC32 = ObjectChecksumAlgorithm(1)

	// ObjectChecksumAlgorithmCRC32C indicates that the CRC32C algorithm was used to compute an object's checksum.
	ObjectChecksumAlgorithmCRC32C = ObjectChecksumAlgorithm(2)

	// ObjectChecksumAlgorithmCRC64NVME indicates that the CRC64NVME algorithm was used to compute an object's checksum.
	ObjectChecksumAlgorithmCRC64NVME = ObjectChecksumAlgorithm(3)

	// ObjectChecksumAlgorithmSHA1 indicates that the SHA-1 algorithm was used to compute an object's checksum.
	ObjectChecksumAlgorithmSHA1 = ObjectChecksumAlgorithm(4)

	// ObjectChecksumAlgorithmSHA256 indicates that the SHA-256 algorithm was used to compute an object's checksum.
	ObjectChecksumAlgorithmSHA256 = ObjectChecksumAlgorithm(5)
)
