// Package crypto provides the cryptographic functions required within the SDK.
//
// There are two kinds of decrypted data:
//   - Metadata means any small string data, typically file metadata, but also e.g. directory names.
//   - Data means file content.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/dromara/dongle/md2"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/pbkdf2"
	"slices"
	"strings"
)

type AuthVersion int
type FileEncryptionVersion int
type MetadataEncryptionVersion int

type MetaCrypter interface {
	EncryptMeta(metadata string) EncryptedString
	DecryptMeta(encrypted EncryptedString) (string, error)
}

// EncryptedString denotes that a string is encrypted and can't be used meaningfully before being decrypted.
type EncryptedString string

// NewEncryptedStringV2 creates a new EncryptedString with the v2 format
func NewEncryptedStringV2(encrypted []byte, nonce [12]byte) EncryptedString {
	return EncryptedString("002" + string(nonce[:]) + base64.StdEncoding.EncodeToString(encrypted))
}

// NewEncryptedStringV3 creates a new EncryptedString with the v3 format
func NewEncryptedStringV3(encrypted []byte, nonce [12]byte) EncryptedString {
	return EncryptedString("003" + hex.EncodeToString(nonce[:]) + base64.StdEncoding.EncodeToString(encrypted))
}

// MasterKeys is a slice of MasterKey, this is used by the V1 and V2 encryption schemes
type MasterKeys []MasterKey

// NewMasterKeys creates a new MasterKeys slice
func NewMasterKeys(encryptionKey MasterKey, stringKeys string) (MasterKeys, error) {
	keys := make([]MasterKey, 0)
	for _, key := range strings.Split(stringKeys, "|") {
		mk, err := NewMasterKey([]byte(key))
		if err != nil {
			return nil, fmt.Errorf("NewMasterKey: %w", err)
		}
		if encryptionKey.DerivedBytes == mk.DerivedBytes {
			continue
		}
		keys = append(keys, *mk)
	}
	keys = slices.Insert(keys, 0, encryptionKey)
	return keys, nil
}

func getCipherForKey(key [32]byte) (cipher.AEAD, error) {
	c, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("getCipherForKey: %v", err)
	}
	derivedGcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, fmt.Errorf("getCipherForKey: %v", err)
	}
	return derivedGcm, nil
}

// DecryptMeta should be avoided, and Filen.DecryptMeta should be used instead,
// but this is necessary for RSA Keypair decryption
func (ms *MasterKeys) DecryptMeta(encrypted EncryptedString) (string, error) {
	if encrypted[0:8] == "U2FsdGVk" {
		return ms.DecryptMetaV1(encrypted)
	}
	if encrypted[0:3] == "002" {
		return ms.DecryptMetaV2(encrypted)
	}
	return "", fmt.Errorf("unknown metadata format")
}

// MasterKey is a key used to encrypt and decrypt metadata
// in the v1 and v2 encryption schemes
type MasterKey struct {
	Bytes        []byte
	DerivedBytes [32]byte
	cipher       cipher.AEAD
}

// NewMasterKey creates a new MasterKey from a byte slice
func NewMasterKey(key []byte) (*MasterKey, error) {
	derivedKey := pbkdf2.Key(key, key, 1, 32, sha512.New)
	derivedBytes := [32]byte{}
	copy(derivedBytes[:], derivedKey[:32])
	c, err := getCipherForKey(derivedBytes)
	if err != nil {
		return nil, fmt.Errorf("NewMasterKey: %v", err)
	}
	return &MasterKey{
		Bytes:        key,
		DerivedBytes: derivedBytes,
		cipher:       c,
	}, nil
}

// EncryptMeta should be avoided, and Filen.EncryptMeta should be used instead
func (m *MasterKey) EncryptMeta(metadata string) EncryptedString {
	nonce := [12]byte([]byte(GenerateRandomString(12)))
	encrypted := m.cipher.Seal(nil, nonce[:], []byte(metadata), nil)
	return NewEncryptedStringV2(encrypted, nonce)
}

func (m *MasterKey) decryptMetaV1(metadata EncryptedString) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(metadata))
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	salt := decoded[8:16]
	cipherText := decoded[16:]

	keyBytes, ivBytes := deriveKeyAndIV(m.Bytes[:], salt, 32, 16)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)

	plaintext := make([]byte, len(cipherText))
	mode.CryptBlocks(plaintext, cipherText)

	paddingLen := int(plaintext[len(plaintext)-1])
	if paddingLen > aes.BlockSize || paddingLen <= 0 {
		return "", fmt.Errorf("invalid padding size")
	}

	return string(plaintext[:len(plaintext)-paddingLen]), nil
}

// DecryptMetaV2 should be avoided, and Filen.DecryptMeta should be used instead
func (m *MasterKey) DecryptMetaV2(metadata EncryptedString) (string, error) {
	nonce := metadata[3:15]
	decoded, err := base64.StdEncoding.DecodeString(string(metadata[15:]))
	if err != nil {
		return "", fmt.Errorf("DecryptMetadataV2: %v", err)
	}
	decoded, err = m.cipher.Open(decoded[:0], []byte(nonce), decoded, nil)
	if err != nil {
		return "", fmt.Errorf("DecryptMetadataV2: %v", err)
	}
	return string(decoded), nil
}

// DecryptMeta should be avoided, and Filen.DecryptMeta should be used instead
func (m *MasterKey) DecryptMeta(metadata EncryptedString) (string, error) {
	if metadata[0:8] == "U2FsdGVk" {
		return m.decryptMetaV1(metadata)
	}
	switch metadata[0:3] {
	case "002":
		return m.DecryptMetaV2(metadata)
	default:
		return "", fmt.Errorf("unknown metadata format")
	}
}

// AllKeysFailedError denotes that no key passed to [DecryptMetadataAllKeys] worked.
type AllKeysFailedError struct {
	Errors []error // errors thrown in the process
}

func (e *AllKeysFailedError) Error() string {
	return fmt.Sprintf("all keys failed: %v", e.Errors)
}

func (ms *MasterKeys) decryptMeta(metadata EncryptedString, decryptFunc func(m *MasterKey, encryptedString EncryptedString) (string, error)) (string, error) {
	errs := make([]error, 0)
	for _, masterKey := range *ms {
		var decrypted string
		decrypted, err := decryptFunc(&masterKey, metadata)
		if err == nil {
			return decrypted, nil
		}
		errs = append(errs, err)
	}
	return "", &AllKeysFailedError{Errors: errs}
}

// DecryptMetaV1 should be avoided, and Filen.DecryptMeta should be used instead
func (ms *MasterKeys) DecryptMetaV1(metadata EncryptedString) (string, error) {
	return ms.decryptMeta(metadata, (*MasterKey).decryptMetaV1)
}

// DecryptMetaV2 should be avoided, and Filen.DecryptMeta should be used instead
func (ms *MasterKeys) DecryptMetaV2(metadata EncryptedString) (string, error) {
	return ms.decryptMeta(metadata, (*MasterKey).DecryptMetaV2)
}

// EncryptMeta should be avoided, and Filen.EncryptMeta should be used instead
func (ms *MasterKeys) EncryptMeta(metadata string) EncryptedString {
	// potential null dereference which makes me uncomfortable
	// this function should only ever be called on non-empty MasterKeys
	// which should be safe since in v2 there must be at least 1 master key,
	// and in v3 we won't be using this function
	return (*ms)[0].EncryptMeta(metadata)
}

// DerivedPassword is derived from the user password, and used to authenticate the user to the backend
type DerivedPassword string

// DeriveMKAndAuthFromPassword returns a MasterKey and a DerivedPassword
func DeriveMKAndAuthFromPassword(password string, salt string) (*MasterKey, DerivedPassword, error) {
	// makes a 128 byte string
	derived := hex.EncodeToString(pbkdf2.Key([]byte(password), []byte(salt), 200000, 64, sha512.New))
	var (
		rawMasterKey [64]byte
	)
	copy(rawMasterKey[:], derived[:64])

	hasher := sha512.New()
	hasher.Write([]byte(derived[64:])) // write password
	derivedPass := DerivedPassword(hex.EncodeToString(hasher.Sum(nil)))

	masterKey, err := NewMasterKey(rawMasterKey[:])
	if err != nil {
		return nil, "", fmt.Errorf("NewMasterKey: %v\n", err)
	}
	return masterKey, derivedPass, nil
}

// v3

// EncryptionKey is used to encrypt and decrypt data
// these keys are used as the v3 KEK, DEK and v2/v3 file Keys
type EncryptionKey struct {
	Bytes  [32]byte
	Cipher cipher.AEAD
}

// MakeNewFileKey returns a new encryption key
func MakeNewFileKey(v FileEncryptionVersion) (*EncryptionKey, error) {
	switch v {
	case 1:
		panic("unsupported version")
	case 2:
		encryptionKeyStr := GenerateRandomString(32)
		encryptionKey, err := MakeEncryptionKeyFromBytes([32]byte([]byte(encryptionKeyStr)))
		if err != nil {
			return nil, fmt.Errorf("NewKeyEncryptionKey auth version 2: %w", err)
		}
		return encryptionKey, nil
	case 3:
		encryptionKey, err := NewEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("NewKeyEncryptionKey auth version 3: %w", err)
		}
		return encryptionKey, nil
	default:
		panic("unsupported version")
	}
}

// EncryptMeta should be avoided, and Filen.EncryptMeta should be used instead
func (key *EncryptionKey) EncryptMeta(metadata string) EncryptedString {
	nonce := [12]byte(GenerateRandomBytes(12))
	encrypted := key.Cipher.Seal(nil, nonce[:], []byte(metadata), nil)
	return NewEncryptedStringV3(encrypted, nonce)
}

func (key *EncryptionKey) ToMasterKey() (*MasterKey, error) {
	return NewMasterKey(key.Bytes[:])
}

// DecryptMeta should be avoided, and Filen.DecryptMeta should be used instead
func (key *EncryptionKey) DecryptMeta(metadata EncryptedString) (string, error) {
	if metadata[0:3] != "003" {
		return "", fmt.Errorf("unsupported metadata %s format (allowed: 003)", metadata[0:3])
	}
	nonce, err := hex.DecodeString(string(metadata[3:27]))
	if err != nil {
		return "", fmt.Errorf("decoding nonce: %v", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(metadata[27:]))
	if err != nil {
		return "", fmt.Errorf("decoding metadata: %v", err)
	}
	decrypted, err := key.Cipher.Open(nil, nonce[:], decoded, nil)
	if err != nil {
		return "", fmt.Errorf("decrypting: %v", err)
	}
	return string(decrypted), nil
}

// MakeEncryptionKeyFromBytes returns a new encryption key
// from a 32 byte array
func MakeEncryptionKeyFromBytes(key [32]byte) (*EncryptionKey, error) {
	c, err := getCipherForKey(key)
	if err != nil {
		return nil, fmt.Errorf("MakeEncryptionKeyFromBytes: %v", err)
	}
	return &EncryptionKey{
		Bytes:  key,
		Cipher: c,
	}, nil
}

// MakeEncryptionKeyFromStr returns a new encryption key
// from a 64 char hex encoded string
func MakeEncryptionKeyFromStr(key string) (*EncryptionKey, error) {
	decoded, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("decoding DEK: %w", err)
	}
	return MakeEncryptionKeyFromBytes([32]byte(decoded))
}

// NewEncryptionKey generates a new encryption key using a random 32 byte array
func NewEncryptionKey() (*EncryptionKey, error) {
	return MakeEncryptionKeyFromBytes([32]byte(GenerateRandomBytes(32)))
}

// ToString returns a 64 char hex encoded string representation
// of the encryption key
func (key *EncryptionKey) ToString() string {
	return hex.EncodeToString(key.Bytes[:])
}

// DeriveKEKAndAuthFromPassword returns a KEK and a DerivedPassword
// derived from the user password
func DeriveKEKAndAuthFromPassword(password string, salt string) (*EncryptionKey, DerivedPassword, error) {
	bytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, "", fmt.Errorf("decoding salt: %v", err)
	}
	derived := hex.EncodeToString(argon2.IDKey([]byte(password), bytes, 3, 65536, 4, 64))

	kek, err := MakeEncryptionKeyFromStr(derived[:len(derived)/2])
	if err != nil {
		return nil, "", fmt.Errorf("MakeEncryptionKeyFromBytes: %v", err)
	}
	return kek, DerivedPassword(derived[len(derived)/2:]), nil
}

// MakeEncryptionKeyFromUnknownStr returns a new encryption key
// from either a 32 character string or a 64 character hex encoded string
func MakeEncryptionKeyFromUnknownStr(key string) (*EncryptionKey, error) {
	switch len(key) {
	case 32: // v1 & v2
		return MakeEncryptionKeyFromBytes([32]byte([]byte(key)))
	case 64: // v3
		return MakeEncryptionKeyFromStr(key)
	default:
		return nil, fmt.Errorf("key length wrong")
	}
}

func (key *EncryptionKey) encrypt(nonce []byte, data []byte) []byte {
	return key.Cipher.Seal(data[:0], nonce, data, nil)
}

// EncryptData encrypts file data using the encryption key
// generates a nonce and prepends it to the data
func (key *EncryptionKey) EncryptData(data []byte) []byte {
	nonce := GenerateRandomBytes(12)
	data = key.encrypt(nonce[:], data)
	return append(nonce, data...)
}

func (key *EncryptionKey) decrypt(nonce []byte, data []byte) error {
	data, err := key.Cipher.Open(data[:0], nonce, data, nil)
	if err != nil {
		return fmt.Errorf("open: %v", err)
	}
	return nil
}

// DecryptData decrypts file data using the encryption key
// returns the decrypted data, assumes that the nonce is the first 12 bytes
func (key *EncryptionKey) DecryptData(data []byte) ([]byte, error) {
	nonce := data[:12]
	err := key.decrypt(nonce, data[12:])
	if err != nil {
		return nil, err
	}
	return data[12 : len(data)-key.Cipher.Overhead()], nil
}

func (key *EncryptionKey) ToStringWithVersion(v FileEncryptionVersion) string {
	if v == 3 {
		return hex.EncodeToString(key.Bytes[:])
	}
	return string(key.Bytes[:])
}

// RSAKeyPairFromStrings returns a private and public key pair
// from base64 encoded strings
func RSAKeyPairFromStrings(privKey string, pubKey string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	publicKeyDecoded, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding public key: %v", err)
	}
	privateKeyDecoded, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding private key: %v", err)
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(publicKeyDecoded)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing public key: %v", err)
	}

	publicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		return nil, nil, fmt.Errorf("parsing public key, failed to cast: %v", err)
	}

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privateKeyDecoded)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing private key: %v", err)
	}

	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("parsing private key, failed to cast: %v", err)
	}

	if !publicKey.Equal(&privateKey.PublicKey) {
		return nil, nil, fmt.Errorf("public and private key mismatch")
	}

	return privateKey, publicKey, nil
}

// RSAKeyPairFromTSConfig returns a private and public key pair
// from base64 encoded strings where the private key is encoded with PKCS1 DER
func RSAKeyPairFromTSConfig(privKey string, pubKey string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	publicKeyDecoded, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding public key: %v", err)
	}
	privateKeyDecoded, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding private key: %v", err)
	}
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyDecoded)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing public key: %v", err)
	}

	privateKeyAny, err := x509.ParsePKCS8PrivateKey(privateKeyDecoded)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing private key: %v", err)
	}

	privateKey, ok := privateKeyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("parsing private key, failed to cast: %v", err)
	}

	if !publicKey.Equal(&privateKey.PublicKey) {
		return nil, nil, fmt.Errorf("public and private key mismatch")
	}

	return privateKey, publicKey, nil
}

// HMACKey is a 256 bit key used as a generic hashing key
// any time we want a hash of a string
type HMACKey [32]byte

// MakeHMACKey derives a 256 bit key from a private key
// this is to allow a single key to derivable from both V2 and V3 accounts
func MakeHMACKey(privateKey *rsa.PrivateKey) HMACKey {
	key := HMACKey{}
	derivedKey := hkdf.New(sha256.New, privateKey.D.Bytes(), nil, []byte("hmac-sha256-key"))
	_, err := derivedKey.Read(key[:])
	if err != nil {
		// this should never happen
		// we do not read enough from the hkdf for it to be an issue
		panic("error generating hkdf key: " + err.Error())
	}
	return key
}

// Hash hashes a string using the key
func (h HMACKey) Hash(data []byte) string {
	hasher := hmac.New(sha256.New, h[:])
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// V2Hash hashes a string using the V2 algorithm
// this was used before HMACKey was introduced,
// and is still used in some places for v2 accounts
func V2Hash(data []byte) string {
	outerHasher := sha1.New()
	innerHasher := sha512.New()
	innerHasher.Write(data)
	outerHasher.Write([]byte(hex.EncodeToString(innerHasher.Sum(nil))))
	return hex.EncodeToString(outerHasher.Sum(nil))
}

// PublicEncrypt encrypts data using a public key
func PublicEncrypt(publicKey *rsa.PublicKey, data string) (EncryptedString, error) {
	encrypted, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, publicKey, []byte(data), nil)
	if err != nil {
		return "", err
	}
	return EncryptedString(base64.StdEncoding.EncodeToString(encrypted)), nil
}

// PublicKeyFromString returns a public key from a base64 encoded string
func PublicKeyFromString(pubKey string) (*rsa.PublicKey, error) {
	publicKeyDecoded, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return nil, fmt.Errorf("decoding public key: %v", err)
	}
	publicKeyAny, err := x509.ParsePKIXPublicKey(publicKeyDecoded)
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %v", err)
	}
	publicKey, ok := publicKeyAny.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsing public key, failed to cast: %v", err)
	}
	return publicKey, nil
}

// for backwards compatibility with V1 only
func V1HashPassword(password string) DerivedPassword {
	sha1Hasher := sha1.New()
	sha256Hasher := sha256.New()
	sha384Hasher := sha512.New384()
	sha512Hasher := sha512.New()

	md2Hasher := md2.New()
	md4Hasher := md4.New()
	md5Hasher := md5.New()
	sha512Hasher2 := sha512.New()

	sha1Hasher.Write([]byte(password))
	sha256Hasher.Write([]byte(hex.EncodeToString(sha1Hasher.Sum(nil))))
	sha384Hasher.Write([]byte(hex.EncodeToString(sha256Hasher.Sum(nil))))
	sha512Hasher.Write([]byte(hex.EncodeToString(sha384Hasher.Sum(nil))))
	part1 := hex.EncodeToString(sha512Hasher.Sum(nil))

	md2Hasher.Write([]byte(password))
	md4Hasher.Write([]byte(hex.EncodeToString(md2Hasher.Sum(nil))))
	md5Hasher.Write([]byte(hex.EncodeToString(md4Hasher.Sum(nil))))
	sha512Hasher2.Write([]byte(hex.EncodeToString(md5Hasher.Sum(nil))))
	part2 := hex.EncodeToString(sha512Hasher2.Sum(nil))

	return DerivedPassword(part1 + part2)
}

// for backwards compatibility with V1 only
func V1DeriveMasterKeyAndDerivedPass(password string) (*MasterKey, DerivedPassword, error) {
	pass := V1HashPassword(password)
	masterKeyStr := V2Hash([]byte(password))
	masterKey, err := NewMasterKey([]byte(masterKeyStr))
	if err != nil {
		return nil, "", err
	}
	return masterKey, pass, nil
}

// Simplified EVP_BytesToKey implementation
// this is used to decrypt V1 metadata
func deriveKeyAndIV(key, salt []byte, keyLen, ivLen int) ([]byte, []byte) {
	keyAndIV := make([]byte, keyLen+ivLen)

	data := make([]byte, 0, 16+len(key))
	for offset := 0; offset < keyLen+ivLen; {
		hash := md5.New()
		hash.Write(data)
		hash.Write(key)
		hash.Write(salt)
		digest := hash.Sum(nil)

		copyLen := min(len(digest), keyLen+ivLen-offset)
		copy(keyAndIV[offset:], digest[:copyLen])
		offset += copyLen

		data = digest
	}

	return keyAndIV[:keyLen], keyAndIV[keyLen:]
}

// V1Decrypt decrypts data using the V1 encryption scheme
func V1Decrypt(data, key []byte) ([]byte, error) {
	// Old and deprecated, not in use anymore, just here for backwards compatibility
	firstBytes := data[:16]
	asciiString := string(firstBytes)
	base64String := base64.StdEncoding.EncodeToString(firstBytes)
	utf8String := string(firstBytes)

	needsConvert := true
	isCBC := true

	if strings.HasPrefix(asciiString, "Salted_") ||
		strings.HasPrefix(base64String, "Salted_") ||
		strings.HasPrefix(utf8String, "Salted_") {
		needsConvert = false
	}

	if strings.HasPrefix(asciiString, "Salted_") ||
		strings.HasPrefix(base64String, "Salted_") ||
		strings.HasPrefix(utf8String, "U2FsdGVk") ||
		strings.HasPrefix(asciiString, "U2FsdGVk") ||
		strings.HasPrefix(utf8String, "Salted_") ||
		strings.HasPrefix(base64String, "U2FsdGVk") {
		isCBC = false
	}

	if needsConvert && !isCBC {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, err
		}
		data = decoded
	}

	if !isCBC {
		saltBytes := data[8:16]

		keyBytes, ivBytes := deriveKeyAndIV(key, saltBytes, 32, 16)

		block, err := aes.NewCipher(keyBytes)
		if err != nil {
			return nil, err
		}

		mode := cipher.NewCBCDecrypter(block, ivBytes)
		ciphertext := data[16:]
		plaintext := make([]byte, len(ciphertext))
		mode.CryptBlocks(plaintext, ciphertext)

		// Remove PKCS#7 padding
		padding := int(plaintext[len(plaintext)-1])
		return plaintext[:len(plaintext)-padding], nil
	} else {
		keyBytes := key
		ivBytes := keyBytes[:16]

		block, err := aes.NewCipher(keyBytes)
		if err != nil {
			return nil, err
		}

		mode := cipher.NewCBCDecrypter(block, ivBytes)
		plaintext := make([]byte, len(data))
		mode.CryptBlocks(plaintext, data)

		// Remove PKCS#7 padding
		padding := int(plaintext[len(plaintext)-1])
		return plaintext[:len(plaintext)-padding], nil
	}
}
