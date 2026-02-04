package thumbnails

// Config holds thumbnail generation settings
type Config struct {
	MaxWidth  int
	MaxHeight int
	Quality   int
	Format    string
}

func DefaultConfig() *Config {
	return &Config{
		MaxWidth:  300,
		MaxHeight: 300,
		Quality:   100,
		Format:    "png",
	}
}

// CreateThumbnailRequest matches the API payload structure for POST /drive/files/thumbnail
type CreateThumbnailRequest struct {
	FileUUID       string `json:"fileUuid"`
	MaxWidth       int    `json:"maxWidth"`
	MaxHeight      int    `json:"maxHeight"`
	Type           string `json:"type"`
	Size           int64  `json:"size"`
	BucketID       string `json:"bucketId"`
	BucketFile     string `json:"bucketFile"`
	EncryptVersion string `json:"encryptVersion"`
}
