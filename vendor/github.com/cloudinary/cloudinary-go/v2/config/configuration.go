// Package config defines the Cloudinary configuration.
package config

import (
	"errors"
	"net/url"
	"os"

	"github.com/creasty/defaults"
	"github.com/gorilla/schema"
)

// Configuration is the main configuration struct.
type Configuration struct {
	Cloud     Cloud
	API       API
	URL       URL
	AuthToken AuthToken
}

var decoder = schema.NewDecoder()

// New returns a new Configuration instance from the environment variable
func New() (*Configuration, error) {
	return NewFromURL(os.Getenv("CLOUDINARY_URL"))
}

// NewFromURL returns a new Configuration instance from a cloudinary url.
func NewFromURL(cldURLStr string) (*Configuration, error) {
	if cldURLStr == "" {
		return nil, errors.New("must provide CLOUDINARY_URL")
	}

	cldURL, err := url.Parse(cldURLStr)
	if err != nil {
		return nil, err
	}

	pass, _ := cldURL.User.Password()
	params := cldURL.Query()
	conf, err := NewFromQueryParams(cldURL.Host, cldURL.User.Username(), pass, params)
	if err != nil {
		return nil, err
	}

	return conf, err
}

// NewFromParams returns a new Configuration instance from the provided parameters.
func NewFromParams(cloud string, key string, secret string) (*Configuration, error) {
	return NewFromQueryParams(cloud, key, secret, map[string][]string{})
}

// NewFromOAuthToken returns a new Configuration instance from the provided cloud name and OAuth token.
func NewFromOAuthToken(cloud string, oAuthToken string) (*Configuration, error) {
	return NewFromQueryParams(cloud, "", "", map[string][]string{"oauth_token": {oAuthToken}})
}

// NewFromQueryParams returns a new Configuration instance from the provided url query parameters.
func NewFromQueryParams(cloud string, key string, secret string, params map[string][]string) (*Configuration, error) {
	cloudConf := Cloud{
		CloudName: cloud,
		APIKey:    key,
		APISecret: secret,
	}

	conf := &Configuration{
		Cloud:     cloudConf,
		API:       API{},
		URL:       URL{},
		AuthToken: AuthToken{},
	}

	if err := defaults.Set(conf); err != nil {
		return nil, err
	}

	// import configuration keys from parameters

	decoder.IgnoreUnknownKeys(true)

	err := decoder.Decode(&conf.Cloud, params)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(&conf.API, params)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(&conf.URL, params)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(&conf.AuthToken, params)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
