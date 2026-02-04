package endpoints

import (
	"net/url"
	"strings"
)

// Config holds the base URL configuration for all API endpoints
type Config struct {
	BaseURL string
}

// Default returns the production endpoints configuration
func Default() *Config {
	return &Config{
		BaseURL: "https://gateway.internxt.com",
	}
}

// NewConfig creates a new endpoints config with a custom base URL
func NewConfig(baseURL string) *Config {
	return &Config{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

// driveURL returns the base drive API URL
func (c *Config) driveURL() string {
	u, _ := url.JoinPath(c.BaseURL, "/drive")
	return u
}

// Drive returns a DriveEndpoints helper for /drive/* endpoints
func (c *Config) Drive() *DriveEndpoints {
	return &DriveEndpoints{base: c.driveURL()}
}

func (c *Config) networkURL() string {
	u, _ := url.JoinPath(c.BaseURL, "/network")
	return u
}

// Network returns a NetworkEndpoints helper for /network/* endpoints
func (c *Config) Network() *NetworkEndpoints {
	return &NetworkEndpoints{base: c.networkURL()}
}

// DriveEndpoints provides endpoints under /drive
type DriveEndpoints struct {
	base string
}

// Auth returns auth-related endpoints
func (d *DriveEndpoints) Auth() *AuthEndpoints {
	base, _ := url.JoinPath(d.base, "/auth")
	return &AuthEndpoints{base: base}
}

// Files returns file-related endpoints
func (d *DriveEndpoints) Files() *FileEndpoints {
	base, _ := url.JoinPath(d.base, "/files")
	return &FileEndpoints{base: base}
}

// Folders returns folder-related endpoints
func (d *DriveEndpoints) Folders() *FolderEndpoints {
	base, _ := url.JoinPath(d.base, "/folders")
	return &FolderEndpoints{base: base}
}

// Users returns user-related endpoints
func (d *DriveEndpoints) Users() *UserEndpoints {
	base, _ := url.JoinPath(d.base, "/users")
	return &UserEndpoints{base: base}
}

// AuthEndpoints : endpoints under /drive/auth
type AuthEndpoints struct {
	base string
}

func (a *AuthEndpoints) Login() string {
	u, _ := url.JoinPath(a.base, "/login")
	return u
}

func (a *AuthEndpoints) CLIAccess() string {
	u, _ := url.JoinPath(a.base, "/cli/login/access")
	return u
}

// FileEndpoints : endpoints under /drive/files
type FileEndpoints struct {
	base string
}

func (f *FileEndpoints) Create() string { return f.base }

func (f *FileEndpoints) Meta(uuid string) string {
	u, _ := url.JoinPath(f.base, uuid, "/meta")
	return u
}

func (f *FileEndpoints) Delete(uuid string) string {
	u, _ := url.JoinPath(f.base, uuid)
	return u
}

func (f *FileEndpoints) Thumbnail() string {
	u, _ := url.JoinPath(f.base, "/thumbnail")
	return u
}

// FolderEndpoints : endpoints under /drive/folders
type FolderEndpoints struct {
	base string
}

func (f *FolderEndpoints) Create() string { return f.base }

func (f *FolderEndpoints) Delete(uuid string) string {
	u, _ := url.JoinPath(f.base, uuid)
	return u
}

func (f *FolderEndpoints) ContentFolders(parentUUID string) string {
	u, _ := url.JoinPath(f.base, "/content", parentUUID, "/folders")
	return u
}

func (f *FolderEndpoints) ContentFiles(parentUUID string) string {
	u, _ := url.JoinPath(f.base, "/content", parentUUID, "/files")
	return u
}

func (f *FolderEndpoints) CheckFilesExistence(parentUUID string) string {
	u, _ := url.JoinPath(f.base, "/content", parentUUID, "/files", "/existence")
	return u
}

// UserEndpoints : endpoints under /users
type UserEndpoints struct {
	base string
}

func (u *UserEndpoints) Usage() string {
	path, _ := url.JoinPath(u.base, "/usage")
	return path
}

func (u *UserEndpoints) Limit() string {
	path, _ := url.JoinPath(u.base, "/limit")
	return path
}

func (u *UserEndpoints) Refresh() string {
	path, _ := url.JoinPath(u.base, "/cli/refresh")
	return path
}

// NetworkEndpoints : endpoints under /buckets and /v2/buckets
type NetworkEndpoints struct {
	base string
}

func (b *NetworkEndpoints) FileInfo(bucketID, fileID string) string {
	u, _ := url.JoinPath(b.base, "/buckets", bucketID, "/files", fileID, "/info")
	return u
}

func (b *NetworkEndpoints) StartUpload(bucketID string) string {
	u, _ := url.JoinPath(b.base, "/v2/buckets", bucketID, "/files/start")
	return u
}

func (b *NetworkEndpoints) FinishUpload(bucketID string) string {
	u, _ := url.JoinPath(b.base, "/v2/buckets", bucketID, "/files/finish")
	return u
}
