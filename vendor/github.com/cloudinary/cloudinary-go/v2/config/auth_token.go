package config

// AuthToken defines the configuration for delivering token-based authenticated media assets.
//
// https://cloudinary.com/documentation/control_access_to_media#delivering_token_based_authenticated_media_assets
type AuthToken struct {
	Key        string `schema:"key"`
	IP         string `schema:"ip"`
	ACL        string `schema:"acl"`
	StartTime  int64  `schema:"start_time"`
	Expiration int64  `schema:"expiration"`
	Duration   int64  `schema:"duration"`
}
