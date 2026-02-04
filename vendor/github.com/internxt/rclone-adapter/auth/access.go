package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/internxt/rclone-adapter/config"
	"github.com/internxt/rclone-adapter/crypto"
	"github.com/internxt/rclone-adapter/errors"
	"github.com/tyler-smith/go-bip39"
)

type LoginRequest struct {
	Email string `json:"email"`
}

type LoginResponse struct {
	HasKeys      bool   `json:"hasKeys"`
	SKey         string `json:"sKey"`
	TFA          bool   `json:"tfa"`
	HasKyberKeys bool   `json:"hasKyberKeys"`
	HasEccKeys   bool   `json:"hasEccKeys"`
}

type AccessResponse struct {
	User struct {
		Email               string `json:"email"`
		UserID              string `json:"userId"`
		Mnemonic            string `json:"mnemonic"`
		PrivateKey          string `json:"privateKey"`
		PublicKey           string `json:"publicKey"`
		RevocateKey         string `json:"revocateKey"`
		RootFolderID        string `json:"rootFolderId"`
		Name                string `json:"name"`
		Lastname            string `json:"lastname"`
		UUID                string `json:"uuid"`
		Credit              int    `json:"credit"`
		CreatedAt           string `json:"createdAt"`
		Bucket              string `json:"bucket"`
		RegisterCompleted   bool   `json:"registerCompleted"`
		Teams               bool   `json:"teams"`
		Username            string `json:"username"`
		BridgeUser          string `json:"bridgeUser"`
		SharedWorkspace     bool   `json:"sharedWorkspace"`
		HasReferralsProgram bool   `json:"hasReferralsProgram"`
		BackupsBucket       string `json:"backupsBucket"`
		Avatar              string `json:"avatar"`
		EmailVerified       bool   `json:"emailVerified"`
		LastPasswordChanged string `json:"lastPasswordChangedAt"`
	} `json:"user"`
	Token    string          `json:"token"`
	UserTeam json.RawMessage `json:"userTeam"`
	NewToken string          `json:"newToken"`
}

func RefreshToken(ctx context.Context, cfg *config.Config) (*AccessResponse, error) {
	endpoint := cfg.Endpoints.Drive().Users().Refresh()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute refresh token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(resp, "refresh token")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	var ar AccessResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if ar.NewToken == "" {
		return nil, fmt.Errorf("refresh response missing newToken")
	}

	return &ar, nil
}

func Login(ctx context.Context, cfg *config.Config, email string) (*LoginResponse, error) {
	endpoint := cfg.Endpoints.Drive().Auth().Login()

	reqBody := LoginRequest{Email: email}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(resp, "login")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response: %w", err)
	}

	var lr LoginResponse
	if err := json.Unmarshal(body, &lr); err != nil {
		return nil, fmt.Errorf("failed to parse login response: %w", err)
	}

	return &lr, nil
}

// AccessRequest is the request body for the CLI access endpoint
type AccessRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TFA      string `json:"tfa,omitempty"`
}

// Access authenticates with email and pre-encrypted password hash.
func Access(ctx context.Context, cfg *config.Config, email, encryptedPassword, tfa string) (*AccessResponse, error) {
	endpoint := cfg.Endpoints.Drive().Auth().CLIAccess()

	reqBody := AccessRequest{
		Email:    email,
		Password: encryptedPassword,
		TFA:      tfa,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal access request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create access request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute access request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(resp, "access")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read access response: %w", err)
	}

	var ar AccessResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("failed to parse access response: %w", err)
	}

	return &ar, nil
}

// DoLogin performs the full login flow:
// 1. Calls Login to get sKey (encrypted salt) and TFA status
// 2. Encrypts the password using the salt
// 3. Calls Access to authenticate and get user data
// 4. Decrypts the mnemonic using the password
// Returns the AccessResponse with decrypted mnemonic and the newToken.
func DoLogin(ctx context.Context, cfg *config.Config, email, password, tfa string) (*AccessResponse, error) {
	loginResp, err := Login(ctx, cfg, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get security details: %w", err)
	}

	if loginResp.TFA && tfa == "" {
		return nil, fmt.Errorf("2FA code required")
	}

	encryptedPassword, err := crypto.EncryptPasswordHash(password, loginResp.SKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	accessResp, err := Access(ctx, cfg, email, encryptedPassword, tfa)
	if err != nil {
		return nil, fmt.Errorf("failed to access: %w", err)
	}

	decryptedMnemonic, err := crypto.DecryptTextWithKey(accessResp.User.Mnemonic, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt mnemonic: %w", err)
	}

	if !bip39.IsMnemonicValid(decryptedMnemonic) {
		return nil, fmt.Errorf("invalid mnemonic format")
	}

	accessResp.User.Mnemonic = decryptedMnemonic

	return accessResp, nil
}
