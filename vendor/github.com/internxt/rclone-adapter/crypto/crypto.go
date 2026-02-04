package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	AppCryptoSecret = "6KYQBP847D4ATSFA"
	saltedPrefix = "Salted__"
)

// getKeyAndIvFrom derives AES key and IV from a secret and salt using 3 rounds of MD5.
func getKeyAndIvFrom(secret string, salt []byte) (key, iv []byte) {
	const transformRounds = 3
	password := append([]byte(secret), salt...)
	md5Hashes := make([][]byte, transformRounds)
	digest := password

	for i := 0; i < transformRounds; i++ {
		h := md5.New()
		h.Write(digest)
		md5Hashes[i] = h.Sum(nil)
		digest = append(md5Hashes[i], password...)
	}

	key = append(md5Hashes[0], md5Hashes[1]...)
	iv = md5Hashes[2]
	return key, iv
}

// DecryptTextWithKey decrypts an AES-256-CBC encrypted text using a secret.
func DecryptTextWithKey(encryptedHex, secret string) (string, error) {
	cipherText, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %w", err)
	}

	if len(cipherText) < 16 {
		return "", fmt.Errorf("ciphertext too short")
	}
	if string(cipherText[:8]) != saltedPrefix {
		return "", fmt.Errorf("invalid OpenSSL format: missing Salted__ prefix")
	}

	salt := cipherText[8:16]
	encryptedContent := cipherText[16:]

	key, iv := getKeyAndIvFrom(secret, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(encryptedContent)%aes.BlockSize != 0 {
		return "", fmt.Errorf("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(encryptedContent))
	mode.CryptBlocks(plainText, encryptedContent)

	plainText, err = pkcs7Unpad(plainText)
	if err != nil {
		return "", fmt.Errorf("failed to unpad: %w", err)
	}

	return string(plainText), nil
}

// EncryptTextWithKey encrypts a plain text using AES-256-CBC with a secret.
func EncryptTextWithKey(plainText, secret string) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	key, iv := getKeyAndIvFrom(secret, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	paddedPlainText := pkcs7Pad([]byte(plainText), aes.BlockSize)

	cipherText := make([]byte, len(paddedPlainText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, paddedPlainText)

	result := append([]byte(saltedPrefix), salt...)
	result = append(result, cipherText...)

	return hex.EncodeToString(result), nil
}

// DecryptText decrypts using the default AppCryptoSecret
func DecryptText(encryptedHex string) (string, error) {
	return DecryptTextWithKey(encryptedHex, AppCryptoSecret)
}

// EncryptText encrypts using the default AppCryptoSecret
func EncryptText(plainText string) (string, error) {
	return EncryptTextWithKey(plainText, AppCryptoSecret)
}

// PassToHash generates a password hash using PBKDF2 with SHA1.
func PassToHash(password, saltHex string) (string, error) {
	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt hex: %w", err)
	}

	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha1.New)
	return hex.EncodeToString(hash), nil
}

// EncryptPasswordHash encrypts a password for the login flow
func EncryptPasswordHash(password, encryptedSalt string) (string, error) {
	salt, err := DecryptText(encryptedSalt)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt salt: %w", err)
	}

	hash, err := PassToHash(password, salt)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	encryptedHash, err := EncryptText(hash)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt hash: %w", err)
	}

	return encryptedHash, nil
}

// pkcs7Pad adds PKCS7 padding to data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// pkcs7Unpad removes PKCS7 padding from data
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}
