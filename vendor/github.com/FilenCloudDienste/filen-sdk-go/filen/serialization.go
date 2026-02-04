package filen

import (
	"context"
	"crypto/x509"
	"encoding/gob"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/client"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
	"io"
)

// serializableFilen is an internal structure used to serialize and deserialize
// the Filen SDK state. It contains only the essential data needed to reconstruct
// a fully functional Filen object, focusing on cryptographic keys and identifiers.
type serializableFilen struct {
	APIKey                    string             // API key for authentication
	AuthVersion               crypto.AuthVersion // Authentication version (2 or 3)
	FileEncryptionVersion     crypto.FileEncryptionVersion
	MetadataEncryptionVersion crypto.MetadataEncryptionVersion
	Email                     string   // User's email address
	MasterKeys                [][]byte // Master encryption keys
	DEK                       [32]byte // Data Encryption Key (for auth v3)
	KEK                       [32]byte // Key Encryption Key (for auth v3)
	PrivateKey                []byte   // RSA private key in PKCS1 format
	HMACKey                   [32]byte // Key used for HMAC operations
	BaseFolderUUID            string   // UUID of user's root directory
}

// serialize converts a Filen instance to a serializable format.
// It extracts all the necessary cryptographic keys and identifiers
// needed to later reconstruct the Filen object.
func (api *Filen) serialize() *serializableFilen {
	masterKeys := make([][]byte, len(api.MasterKeys))
	for i, masterKey := range api.MasterKeys {
		masterKeys[i] = masterKey.Bytes
	}
	return &serializableFilen{
		APIKey:                    api.Client.APIKey,
		AuthVersion:               api.AuthVersion,
		FileEncryptionVersion:     api.FileEncryptionVersion,
		MetadataEncryptionVersion: api.MetadataEncryptionVersion,
		Email:                     api.Email,
		MasterKeys:                masterKeys,
		DEK:                       api.DEK.Bytes,
		PrivateKey:                x509.MarshalPKCS1PrivateKey(&api.PrivateKey),
		HMACKey:                   api.HMACKey,
		BaseFolderUUID:            api.BaseFolder.GetUUID(),
	}
}

// deserialize reconstructs a Filen object from its serialized form.
// It recreates all the cryptographic keys and initializes a new
// API client with the stored API key.
func (s *serializableFilen) deserialize() (*Filen, error) {
	masterKeys := make([]crypto.MasterKey, len(s.MasterKeys))
	for i, masterKey := range s.MasterKeys {
		mk, err := crypto.NewMasterKey(masterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse master key: %w", err)
		}
		masterKeys[i] = *mk
	}
	var (
		dek crypto.EncryptionKey
	)
	if s.AuthVersion >= 3 {
		dekPtr, err := crypto.MakeEncryptionKeyFromBytes(s.DEK)
		if err != nil {
			return nil, fmt.Errorf("failed to parse DEK: %w", err)
		}
		dek = *dekPtr
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(s.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &Filen{
		Client:      client.NewWithAPIKey(context.Background(), s.APIKey),
		AuthVersion: s.AuthVersion,
		Email:       s.Email,
		MasterKeys:  masterKeys,
		DEK:         dek,
		PrivateKey:  *privateKey,
		PublicKey:   privateKey.PublicKey,
		HMACKey:     s.HMACKey,
		BaseFolder:  types.NewRootDirectory(s.BaseFolderUUID),
	}, nil
}

// SerializeTo serializes the Filen object to the provided writer.
// This allows saving the current state of the SDK, including all encryption keys
// and authentication information, for later restoration without going through
// the login process again.
//
// The serialized data should be treated as sensitive, as it contains encryption keys
// that could be used to access the user's files if compromised.
func (api *Filen) SerializeTo(w io.Writer) error {
	s := api.serialize()
	encoder := gob.NewEncoder(w)
	return encoder.Encode(s)
}

// DeserializeFrom reconstructs a Filen object from previously serialized data.
// It reads the serialized state from the provided reader and instantiates a
// fully functional Filen SDK instance with all the necessary encryption keys
// and authentication details.
//
// This allows resuming a session without going through the login and key
// derivation process again.
func DeserializeFrom(r io.Reader) (*Filen, error) {
	var s serializableFilen
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&s); err != nil {
		return nil, err
	}
	return s.deserialize()
}

// TSConfig holds the necessary information to initialize a Filen object
// from the TypeScript SDK. This provides interoperability between the Go
// and TypeScript implementations of the Filen SDK.
type TSConfig struct {
	Email          string   // User's email address
	MasterKeys     []string // Master keys as hex strings
	APIKey         string   // API key for authentication
	PublicKey      string   // RSA public key
	PrivateKey     string   // RSA private key
	AuthVersion    int      // Authentication version (2 or 3)
	BaseFolderUUID string   // UUID of user's root directory
}

// NewFromTSConfig creates a new Filen object from a TypeScript SDK configuration.
// This function serves as a bridge between the TypeScript and Go SDKs, allowing
// seamless integration in applications that use both languages.
//
// It handles the differences in key formats and authentication versions between
// the TypeScript and Go implementations.
func NewFromTSConfig(tsconfig TSConfig) (*Filen, error) {
	switch tsconfig.AuthVersion {
	case 1, 2:
		masterKeys := make([]crypto.MasterKey, len(tsconfig.MasterKeys))
		for i, masterKey := range tsconfig.MasterKeys {
			masterKey, err := crypto.NewMasterKey([]byte(masterKey))
			if err != nil {
				panic(err)
			}
			masterKeys[i] = *masterKey
		}
		privateKey, publicKey, err := crypto.RSAKeyPairFromTSConfig(tsconfig.PrivateKey, tsconfig.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rsa keys: %w", err)
		}
		return &Filen{
			Client:                    client.NewWithAPIKey(context.Background(), tsconfig.APIKey),
			AuthVersion:               crypto.AuthVersion(tsconfig.AuthVersion),
			FileEncryptionVersion:     V2AccountFileEncryptionVersion,
			MetadataEncryptionVersion: V2AccountMetadataEncryptionVersion,
			Email:                     tsconfig.Email,
			MasterKeys:                masterKeys,
			PrivateKey:                *privateKey,
			PublicKey:                 *publicKey,
			HMACKey:                   crypto.MakeHMACKey(privateKey),
			BaseFolder:                types.NewRootDirectory(tsconfig.BaseFolderUUID),
		}, nil
	case 3:
		dek, err := crypto.MakeEncryptionKeyFromStr(tsconfig.MasterKeys[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse DEK: %w", err)
		}
		private, public, err := crypto.RSAKeyPairFromTSConfig(tsconfig.PrivateKey, tsconfig.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rsa keys: %w", err)
		}
		return &Filen{
			Client:                    client.NewWithAPIKey(context.Background(), tsconfig.APIKey),
			AuthVersion:               crypto.AuthVersion(tsconfig.AuthVersion),
			FileEncryptionVersion:     V2AccountFileEncryptionVersion,
			MetadataEncryptionVersion: V2AccountMetadataEncryptionVersion,
			Email:                     tsconfig.Email,
			MasterKeys:                make(crypto.MasterKeys, 0),
			DEK:                       *dek,
			PrivateKey:                *private,
			PublicKey:                 *public,
			HMACKey:                   crypto.MakeHMACKey(private),
			BaseFolder:                types.NewRootDirectory(tsconfig.BaseFolderUUID),
		}, nil
	default:
		return nil, fmt.Errorf("invalid auth version: %d", tsconfig.AuthVersion)
	}
}
