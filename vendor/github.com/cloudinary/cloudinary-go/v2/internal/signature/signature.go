package signature

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
)

// Length represents the length of the signature.
type Length = uint8

// Algo represent the algorithm of the  signature.
type Algo = string

const (
	// Short signature length
	Short Length = 8
	// Long signature length
	Long Length = 32
)

const (
	// SHA1 algorithm.
	SHA1 Algo = "sha1"
	// SHA256 algorithm.
	SHA256 Algo = "sha256"
)

// Sign signs the content with the provided signature.
func Sign(content string, secret string, algo Algo) ([]byte, error) {
	if len(secret) < 1 {
		return nil, errors.New("must supply api_secret")
	}
	var hashFunc hash.Hash
	switch algo {
	case SHA1:
		hashFunc = sha1.New()
	case SHA256:
		hashFunc = sha256.New()
	default:
		return nil, errors.New("unsupported signature algorithm")
	}

	hashFunc.Write([]byte(content + secret))

	return hashFunc.Sum(nil), nil
}

// SignURL returns the URL signature.
func SignURL(content string, secret string, algo Algo, length uint8) string {
	rawSignature, _ := Sign(content, secret, algo)
	signature := base64.RawURLEncoding.EncodeToString(rawSignature)

	return fmt.Sprintf("s--%s--", signature[:length])
}
