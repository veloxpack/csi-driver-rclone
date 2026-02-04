package buckets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

// NewAES256CTRCipher returns a cipher.Stream that performs AES‑256‑CTR encryption
// with the given 32‑byte key and 16‑byte IV, exactly like Node.js’s
// createCipheriv('aes-256-ctr', key, iv).
func NewAES256CTRCipher(key, iv []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	return cipher.NewCTR(block, iv), nil
}

// EncryptReader wraps the provided src reader in a StreamReader that
// encrypts all data through AES‑256‑CTR (no padding):
//
//	source -> cipher -> …
func EncryptReader(src io.Reader, key, iv []byte) (io.Reader, error) {
	stream, err := NewAES256CTRCipher(key, iv)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption stream: %w", err)
	}
	return cipher.StreamReader{S: stream, R: src}, nil
}

// DecryptReader wraps the provided src reader in a StreamReader that
// decrypts data encrypted with AES‑256‑CTR (no padding):
//
//	encryptedSrc -> source -> …
func DecryptReader(src io.Reader, key, iv []byte) (io.Reader, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher for decryption: %w", err)
	}
	stream := cipher.NewCTR(block, iv)
	return cipher.StreamReader{S: stream, R: src}, nil
}

// GetFileDeterministicKey returns SHA512(key||data)
func GetFileDeterministicKey(key, data []byte) []byte {
	h := sha512.New()
	h.Write(key)
	h.Write(data)
	return h.Sum(nil)
}

// GenerateFileBucketKey derives a bucket-level key from mnemonic and bucketID
func GenerateFileBucketKey(mnemonic, bucketID string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	seed := bip39.NewSeed(mnemonic, "")
	bucketBytes, err := hex.DecodeString(bucketID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bucket ID: %w", err)
	}
	return GetFileDeterministicKey(seed, bucketBytes), nil
}

// GenerateBucketKey generates a 64-character hexadecimal bucket key from a mnemonic and bucket ID.
func GenerateBucketKey(mnem string, bucketID []byte) (string, error) {
	if !bip39.IsMnemonicValid(mnem) {
		return "", fmt.Errorf("invalid mnemonic")
	}
	seed := bip39.NewSeed(mnem, "")
	deterministicKey, err := GetDeterministicKey(seed, bucketID)
	if err != nil {
		return "", fmt.Errorf("failed to get deterministic key: %w", err)
	}
	return hex.EncodeToString(deterministicKey)[:64], nil
}

func GetDeterministicKey(key []byte, data []byte) ([]byte, error) {
	hasher := sha512.New()
	data_bytes, err := hex.DecodeString(hex.EncodeToString(key) + hex.EncodeToString(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode deterministic key data: %w", err)
	}
	hasher.Write(data_bytes)
	return hasher.Sum(nil), nil
}

// GenerateFileKey derives the per-file key and IV from mnemonic, bucketID, and plaintext index
func GenerateFileKey(mnemonic, bucketID, indexHex string) (key, iv []byte, err error) {
	bucketKey, err := GenerateFileBucketKey(mnemonic, bucketID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate bucket key: %w", err)
	}
	indexBytes, err := hex.DecodeString(indexHex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode index: %w", err)
	}
	detKey := GetFileDeterministicKey(bucketKey[:32], indexBytes)
	key = detKey[:32]

	iv = indexBytes[0:16]

	// debug log
	/*
		fmt.Printf(
			"Encrypting file using AES256CTR (key %s, iv %s)...\n",
			hex.EncodeToString(key),
			hex.EncodeToString(iv),
		)
	*/

	return key, iv, nil
}

// Calculates the hash of a file
func CalculateFileHash(reader io.Reader) (string, error) {
	sha256Hasher := sha256.New()

	buf := make([]byte, 4096) // 4KB buffer size
	_, err := io.CopyBuffer(sha256Hasher, reader, buf)
	if err != nil {
		return "", fmt.Errorf("error reading data: %v", err)
	}

	sha256Result := sha256Hasher.Sum(nil)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Result)
	ripemd160Result := ripemd160Hasher.Sum(nil)

	return hex.EncodeToString(ripemd160Result), nil
}

// ComputeFileHash computes RIPEMD-160(SHA-256(data)) from a SHA-256 hash result.
// This is the standard hash algorithm used by all Internxt clients for file integrity.
// Takes the raw SHA-256 hash bytes and returns the hex-encoded RIPEMD-160 hash.
func ComputeFileHash(sha256Sum []byte) string {
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Sum)
	return hex.EncodeToString(ripemd160Hasher.Sum(nil))
}
