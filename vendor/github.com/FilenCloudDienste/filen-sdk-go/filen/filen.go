// Package filen provides an SDK interface to interact with the Filen cloud storage service.
// It handles authentication, encryption/decryption, and all API interactions.
package filen

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/client"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/crypto"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
)

// Filen provides the SDK interface for interacting with Filen cloud storage.
// It must be initialized using New or NewWithAPIKey.
type Filen struct {
	// Client is the underlying API client used for all service communication
	Client *client.Client

	// AuthVersion indicates which authentication scheme is being used (1, 2, or 3)
	AuthVersion               crypto.AuthVersion
	FileEncryptionVersion     crypto.FileEncryptionVersion
	MetadataEncryptionVersion crypto.MetadataEncryptionVersion

	// Email is the user's email address
	Email string

	// MasterKeys contains the crypto master keys for the current user. When the user changes
	// their password, a new master key is appended. For decryption, all master keys are tried
	// until one works; for encryption, always use the latest master key. (AuthVersion 2)
	MasterKeys crypto.MasterKeys

	// DEK is the Data Encryption Key used for file encryption (AuthVersion 3)
	DEK crypto.EncryptionKey

	// PrivateKey is the user's RSA private key for asymmetric cryptography operations
	PrivateKey rsa.PrivateKey

	// PublicKey is the user's RSA public key for asymmetric cryptography operations
	PublicKey rsa.PublicKey

	// HMACKey is derived from the private key and used for creating file name hashes
	HMACKey crypto.HMACKey

	// BaseFolder is the root directory of the user's cloud storage
	BaseFolder types.RootDirectory

	// lock provides synchronized access to backend resources
	lock BackendLock
}

// New creates a new Filen instance and initializes it with the given email and password.
// It handles login, authentication, and preparation of encryption keys.
// The appropriate authentication version is automatically determined from the server.
func New(ctx context.Context, email, password, twoFactorCode string) (*Filen, error) {
	unauthorizedClient := client.New(ctx)

	// fetch salt for password derivation
	authInfo, err := unauthorizedClient.PostV3AuthInfo(ctx, email)
	if err != nil {
		return nil, err
	}

	var filen *Filen

	switch authInfo.AuthVersion {
	case 1:
		filen, err = newV1(ctx, email, password, twoFactorCode, *authInfo, unauthorizedClient)
	case 2:
		filen, err = newV2(ctx, email, password, twoFactorCode, *authInfo, unauthorizedClient)
	case 3:
		filen, err = newV3(ctx, email, password, twoFactorCode, *authInfo, unauthorizedClient)
	default:
		panic("unimplemented")
	}
	if err != nil {
		return nil, err
	}

	return filen, nil
}

// NewWithAPIKey creates a new Filen instance using a pre-existing API key.
// This is useful for scenarios where the login step has already been performed
// and the API key is stored securely.
func NewWithAPIKey(ctx context.Context, email, password, apiKey string) (*Filen, error) {
	c := client.NewWithAPIKey(ctx, apiKey)

	authInfo, err := c.PostV3AuthInfo(ctx, email)
	if err != nil {
		return nil, err
	}

	switch authInfo.AuthVersion {
	case 1:
		return newV1WithAPIKey(ctx, email, password, *authInfo, c)
	case 2:
		return newV2WithAPIKey(ctx, email, password, *authInfo, c)
	case 3:
		return newV3WithAPIKey(ctx, email, password, *authInfo, c)
	default:
		panic("unimplemented")
	}
}

// getKeyPair retrieves and decrypts the user's RSA key pair from the server.
// The private key is stored encrypted and is decrypted using the provided meta crypter.
func getKeyPair(ctx context.Context, metaCrypter crypto.MetaCrypter, c *client.Client) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	response, err := c.GetV3UserKeyPairInfo(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get keypair info: %w", err)
	}
	privateKeyStr, err := metaCrypter.DecryptMeta(response.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	privateKey, publicKey, err := crypto.RSAKeyPairFromStrings(privateKeyStr, response.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse rsa keys: %w", err)
	}
	return privateKey, publicKey, nil
}

// getMasterKeys retrieves and processes the user's master encryption keys from the server.
// Master keys are used for file and metadata encryption/decryption.
func getMasterKeys(ctx context.Context, masterKey crypto.MasterKey, c *client.Client) (crypto.MasterKeys, error) {
	encryptedMasterKey := masterKey.EncryptMeta(string(masterKey.Bytes[:]))
	mkResponse, err := c.PostV3UserMasterKeys(ctx, encryptedMasterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get master keys: %w", err)
	}
	masterKeysStr, err := masterKey.DecryptMeta(mkResponse.Keys)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master keys meta: %w", err)
	}

	masterKeys, err := crypto.NewMasterKeys(masterKey, masterKeysStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse master keys: %w", err)
	}

	return masterKeys, nil
}

// getDEK retrieves and decrypts the user's Data Encryption Key (DEK) from the server.
// The DEK is used for file encryption in auth version 3.
func getDEK(ctx context.Context, kek crypto.EncryptionKey, c *client.Client) (*crypto.EncryptionKey, error) {
	encryptedDEK, err := c.GetV3UserDEK(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get DEK: %w", err)
	}
	decryptedDEKStr, err := kek.DecryptMeta(encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}
	dek, err := crypto.MakeEncryptionKeyFromStr(decryptedDEKStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DEK: %w", err)
	}
	return dek, nil
}

// loginV1 performs version 1 authentication with the Filen API.
// It derives the necessary keys from the password, then performs login
// only used in a very limited number of accounts,
// and here for backwards compatibility
func loginV1(ctx context.Context, email, password, twoFactorCode string, info client.V3AuthInfoResponse, uc *client.UnauthorizedClient) (*client.Client, *crypto.MasterKey, error) {
	masterKey, derivedPass, err := crypto.V1DeriveMasterKeyAndDerivedPass(password)
	if err != nil {
		return nil, nil, fmt.Errorf("V1DeriveMasterKeyAndDerivedPass: %w", err)
	}

	response, err := uc.PostV3Login(ctx, email, derivedPass, info.AuthVersion, twoFactorCode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to log in: %w", err)
	}
	c := uc.Authorize(response.APIKey)
	return c, masterKey, nil
}

// loginV2 performs version 2 authentication with the Filen API.
// It derives the necessary keys from the password and salt, then performs login.
func loginV2(ctx context.Context, email, password, twoFactorCode string, info client.V3AuthInfoResponse, uc *client.UnauthorizedClient) (*client.Client, *crypto.MasterKey, error) {
	masterKey, derivedPass, err := crypto.DeriveMKAndAuthFromPassword(password, info.Salt)
	if err != nil {
		return nil, nil, fmt.Errorf("DeriveMKAndAuthFromPassword: %w", err)
	}
	// for simplicity, I'm going to ignore the fact that response here contains the RSAKeypair
	response, err := uc.PostV3Login(ctx, email, derivedPass, info.AuthVersion, twoFactorCode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to log in: %w", err)
	}
	c := uc.Authorize(response.APIKey)
	return c, masterKey, nil
}

// loginV3 performs version 3 authentication with the Filen API.
// It derives the Key Encryption Key (KEK) from the password and salt, then performs login.
func loginV3(ctx context.Context, email, password, twoFactorCode string, info client.V3AuthInfoResponse, uc *client.UnauthorizedClient) (*client.Client, *crypto.EncryptionKey, error) {
	kek, derivedPass, err := crypto.DeriveKEKAndAuthFromPassword(password, info.Salt)
	if err != nil {
		return nil, nil, fmt.Errorf("DeriveKEKAndAuthFromPassword: %w", err)
	}
	// for simplicity, I'm going to ignore the fact that response here contains the RSAKeypair
	response, err := uc.PostV3Login(ctx, email, derivedPass, info.AuthVersion, twoFactorCode)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to log in: %w", err)
	}
	c := uc.Authorize(response.APIKey)

	return c, kek, nil
}

// newV2Authed creates a new Filen instance for auth version 2 with an authenticated client.
// It sets up all required keys and fetches necessary account information.
func newV2Authed(ctx context.Context, email string, info client.V3AuthInfoResponse, c *client.Client, masterKey crypto.MasterKey) (*Filen, error) {
	masterKeys, err := getMasterKeys(ctx, masterKey, c)
	if err != nil {
		return nil, fmt.Errorf("getMasterKeys: %w", err)
	}

	privateKey, publicKey, err := getKeyPair(ctx, &masterKeys, c)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt rsa keys: %w", err)
	}

	baseFolderResponse, err := c.GetV3UserBaseFolder(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get base folder: %w", err)
	}

	return &Filen{
		Client:                    c,
		Email:                     email,
		MasterKeys:                masterKeys,
		PrivateKey:                *privateKey,
		PublicKey:                 *publicKey,
		BaseFolder:                types.NewRootDirectory(baseFolderResponse.UUID),
		AuthVersion:               info.AuthVersion,
		FileEncryptionVersion:     V2AccountFileEncryptionVersion,
		MetadataEncryptionVersion: V2AccountMetadataEncryptionVersion,
		HMACKey:                   crypto.MakeHMACKey(privateKey),
		lock:                      NewBackendLock(),
	}, nil
}

// newV2 handles the complete initialization process for auth version 2.
// It performs login and then completes setup with the authenticated client.
func newV2(ctx context.Context, email, password, twoFactorCode string, info client.V3AuthInfoResponse, uc *client.UnauthorizedClient) (*Filen, error) {
	c, masterKey, err := loginV2(ctx, email, password, twoFactorCode, info, uc)
	if err != nil {
		return nil, fmt.Errorf("loginV2: %w", err)
	}

	return newV2Authed(ctx, email, info, c, *masterKey)
}

// newV2WithAPIKey initializes a Filen instance for auth version 2 using a pre-existing API key.
// It derives the master key from the password but skips the login step.
func newV2WithAPIKey(ctx context.Context, email, password string, info client.V3AuthInfoResponse, c *client.Client) (*Filen, error) {
	masterKey, _, err := crypto.DeriveMKAndAuthFromPassword(password, info.Salt)
	if err != nil {
		return nil, fmt.Errorf("DeriveMKAndAuthFromPassword: %w", err)
	}

	return newV2Authed(ctx, email, info, c, *masterKey)
}

// newV3Authed creates a new Filen instance for auth version 3 with an authenticated client.
// It sets up all required keys and fetches necessary account information.
func newV3Authed(ctx context.Context, email string, info client.V3AuthInfoResponse, c *client.Client, kek crypto.EncryptionKey) (*Filen, error) {
	dek, err := getDEK(ctx, kek, c)
	if err != nil {
		return nil, fmt.Errorf("getDEK: %w", err)
	}

	privateKey, publicKey, err := getKeyPair(ctx, dek, c)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt rsa keys: %w", err)
	}

	baseFolderResponse, err := c.GetV3UserBaseFolder(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get base folder: %w", err)
	}
	return &Filen{
		Client:                    c,
		Email:                     email,
		MasterKeys:                make(crypto.MasterKeys, 0),
		DEK:                       *dek,
		PrivateKey:                *privateKey,
		PublicKey:                 *publicKey,
		BaseFolder:                types.NewRootDirectory(baseFolderResponse.UUID),
		AuthVersion:               info.AuthVersion,
		FileEncryptionVersion:     3,
		MetadataEncryptionVersion: 3,
		HMACKey:                   crypto.MakeHMACKey(privateKey),
		lock:                      NewBackendLock(),
	}, nil
}

// newV3 handles the complete initialization process for auth version 3.
// It performs login and then completes setup with the authenticated client.
func newV3(ctx context.Context, email, password, twoFactorCode string, info client.V3AuthInfoResponse, uc *client.UnauthorizedClient) (*Filen, error) {
	c, kek, err := loginV3(ctx, email, password, twoFactorCode, info, uc)
	if err != nil {
		return nil, fmt.Errorf("loginV3: %w", err)
	}

	return newV3Authed(ctx, email, info, c, *kek)
}

// newV3WithAPIKey initializes a Filen instance for auth version 3 using a pre-existing API key.
// It derives the Key Encryption Key (KEK) from the password but skips the login step.
func newV3WithAPIKey(ctx context.Context, email, password string, info client.V3AuthInfoResponse, c *client.Client) (*Filen, error) {
	kek, _, err := crypto.DeriveKEKAndAuthFromPassword(password, info.Salt)
	if err != nil {
		return nil, fmt.Errorf("DeriveKEKAndAuthFromPassword: %w", err)
	}

	return newV3Authed(ctx, email, info, c, *kek)
}

func newV1(ctx context.Context, email, password, twoFactorCode string, info client.V3AuthInfoResponse, uc *client.UnauthorizedClient) (*Filen, error) {
	c, masterKey, err := loginV1(ctx, email, password, twoFactorCode, info, uc)
	if err != nil {
		return nil, fmt.Errorf("loginV1: %w", err)
	}

	return newV2Authed(ctx, email, info, c, *masterKey)
}

func newV1WithAPIKey(ctx context.Context, email, password string, info client.V3AuthInfoResponse, c *client.Client) (*Filen, error) {
	masterKey, _, err := crypto.V1DeriveMasterKeyAndDerivedPass(password)
	if err != nil {
		return nil, fmt.Errorf("V1DeriveMasterKeyAndDerivedPass: %w", err)
	}

	return newV2Authed(ctx, email, info, c, *masterKey)
}
