package config

// API defines the configuration for making requests to the Cloudinary API.
type API struct {
	UploadPrefix  string `schema:"upload_prefix" default:"https://api.cloudinary.com"`
	Timeout       int64  `schema:"timeout" default:"60"` // seconds
	UploadTimeout int64  `schema:"upload_timeout"`
	ChunkSize     int64  `schema:"chunk_size" default:"20000000"` // bytes
}
