package crypto

import (
	"crypto/rand"
	"crypto/sha512"
	"math/big"
)

func RunSHA512(b []byte) []byte {
	hasher := sha512.New()
	hasher.Write(b)
	return hasher.Sum(nil)
}

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GenerateRandomString generates a cryptographically secure random string based on a selection of alphanumerical characters.
func GenerateRandomString(length int) string {
	str := ""
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			panic(err)
		}
		str += string(runes[idx.Int64()])
	}
	return str
}

// GenerateRandomBytes generates a cryptographically secure random byte array
func GenerateRandomBytes(length int) []byte {
	b := make([]byte, length)
	// rand.Read fills b with random bytes and never errors according to doc
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
