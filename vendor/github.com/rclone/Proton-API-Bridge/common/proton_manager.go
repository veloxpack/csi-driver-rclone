package common

import (
	"github.com/rclone/go-proton-api"
)

func getProtonManager(config *Config) *proton.Manager {
	/* Notes on API calls: if the app version is not specified, the api calls will be rejected. */
	options := []proton.Option{
		proton.WithAppVersion(config.AppVersion),
		proton.WithUserAgent(config.UserAgent),
	}
	if config.Transport != nil {
		options = append(options, proton.WithTransport(config.Transport))
	}
	if config.Logger != nil {
		options = append(options, proton.WithLogger(config.Logger))
	}
	m := proton.New(options...)

	return m
}
