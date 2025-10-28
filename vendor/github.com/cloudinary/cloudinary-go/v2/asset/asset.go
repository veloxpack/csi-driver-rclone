package asset

import (
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/config"
	"github.com/cloudinary/cloudinary-go/v2/internal/signature"
	"github.com/cloudinary/cloudinary-go/v2/logger"
	"github.com/cloudinary/cloudinary-go/v2/transformation"
	"hash/crc32"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const shortenAssetType = "iu"

// Asset is the Asset struct.
type Asset struct {
	AssetType      api.AssetType
	DeliveryType   api.DeliveryType
	Transformation transformation.RawTransformation
	Version        int
	PublicID       string
	Suffix         string
	Config         config.Configuration
	AuthToken      AuthToken
	logger         logger.Logger
}

func (a *Asset) setDefaults() {
	a.AssetType = api.Image
	a.DeliveryType = api.Upload
}

// New returns a new Asset instance from the provided configuration.
func New(publicID string, conf *config.Configuration) (*Asset, error) {
	if conf == nil {
		var err error
		conf, err = config.New()
		if err != nil {
			return nil, err
		}
	}

	asset := Asset{PublicID: publicID, Config: *conf}
	asset.setDefaults()
	asset.AuthToken.Config = &conf.AuthToken

	return &asset, nil
}

// Image returns a new image Asset instance from the provided configuration.
func Image(publicID string, conf *config.Configuration) (*Asset, error) {
	return New(publicID, conf)
}

// Video returns a new video Asset instance from the provided configuration.
func Video(publicID string, conf *config.Configuration) (*Asset, error) {
	v, err := New(publicID, conf)
	if err != nil {
		return nil, err
	}
	v.AssetType = api.Video

	return v, nil
}

// File returns a new file Asset instance from the provided configuration.
func File(publicID string, conf *config.Configuration) (*Asset, error) {
	f, err := New(publicID, conf)
	if err != nil {
		return nil, err
	}
	f.AssetType = api.File

	return f, nil
}

// Media returns a new media Asset instance from the provided configuration.
func Media(publicID string, conf *config.Configuration) (*Asset, error) {
	return New(publicID, conf)
}

// String serializes Asset to string.
func (a Asset) String() (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("failed to build URL: %v", r)
			a.logger.Error(msg)
			result = ""
			err = errors.New(msg)
		}
	}()

	assetURL := a.assetURL()
	query := a.query()

	return joinNonEmpty([]interface{}{assetURL, query}, "?"), nil
}

// distribution builds the hostname for the asset distribution.
//
//  1. Customers in shared distribution (e.g. res.cloudinary.com)
//     If CDNSubDomain is true uses res-[1-5].cloudinary.com for both http and https.
//     Setting secureCDNSubDomain to false disables this for https.
//  2. Customers with private cdn
//     If CDNSubDomain is true uses cloudName-res-[1-5].cloudinary.com for http
//     If secureCDNSubDomain is true uses cloudName-res-[1-5].cloudinary.com for https
//     (please contact support if you require this)
//  3. Customers with cname
//     If CDNSubDomain is true uses a[1-5].cname for http.
//     For https, uses the same naming scheme as 1 for shared distribution and as 2 for private distribution.
func distribution(source string, conf config.Configuration) string {
	uc := conf.URL
	useSharedHost := !uc.PrivateCDN
	var hostName string
	if uc.Secure {
		hostName = uc.SecureCName
		if hostName == "" {
			if uc.PrivateCDN {
				hostName = buildHostName(conf.Cloud.CloudName, "", uc.SubDomain, uc.Domain)
			} else {
				hostName = uc.SharedHost
				useSharedHost = true
			}
		}

		secureCDNSubDomain := uc.SecureCDNSubDomain
		if useSharedHost && !secureCDNSubDomain {
			secureCDNSubDomain = uc.CDNSubDomain
		}

		if secureCDNSubDomain {
			hostName = strings.Replace(hostName, uc.SharedHost, buildHostName("", domainShard(source), uc.SubDomain, uc.Domain), 1)
		}
	} else {
		if uc.CName != "" {
			subDomain := ""
			if uc.CDNSubDomain {
				subDomain = "a" + domainShard(source)
			}
			hostName = buildHostName("", "", subDomain, uc.CName)
		} else {
			prefix := ""
			if uc.PrivateCDN {
				prefix = conf.Cloud.CloudName
			}
			suffix := ""
			if uc.CDNSubDomain {
				suffix = domainShard(source)
			}
			hostName = buildHostName(prefix, suffix, uc.SubDomain, uc.Domain)
		}
	}

	distribution := fmt.Sprintf("%s://%s", uc.Protocol(), hostName)

	if useSharedHost {
		distribution += "/" + conf.Cloud.CloudName
	}

	return distribution
}

// buildHostName is a helper method for building hostname of form:
//
//	subDomainPrefix-subDomain-subDomainSuffix.domain
//
// For example:
//
//	cloudName-res-3.cloudinary.com
func buildHostName(subDomainPrefix string, subDomainSuffix string, subDomain string, domain string) string {
	return joinNonEmpty([]interface{}{joinNonEmpty([]interface{}{subDomainPrefix, subDomain, subDomainSuffix}, "-"), domain}, ".")
}

var crc32q = crc32.MakeTable(crc32.IEEE)

// domainShard computes the domain shard from the source.
func domainShard(source string) string {
	return strconv.Itoa(int(crc32.Checksum([]byte(source), crc32q)%5 + 1))
}

var suffixSupportedDeliveryTypes = map[api.AssetType]map[api.DeliveryType]string{
	api.Image: {
		api.Upload:        "images",
		api.Private:       "private_images",
		api.Authenticated: "authenticated_images",
	},
	api.Video: {
		api.Upload: "videos",
	},
	api.File: {
		api.Upload: "files",
	},
}

func (a Asset) assetType() string {
	if a.AssetType == api.Image && a.DeliveryType == api.Upload {
		if a.Config.URL.UseRootPath {
			return ""
		}

		if a.Config.URL.Shorten {
			return shortenAssetType
		}
	}

	if a.Suffix == "" {
		return joinURL([]interface{}{a.AssetType, a.DeliveryType})
	}

	assetType, found := suffixSupportedDeliveryTypes[a.AssetType][a.DeliveryType]
	if !found {
		panic(fmt.Sprintf("URL Suffix is not supported for %v/%v", a.AssetType, a.DeliveryType))
	}

	return assetType
}

// signature returns URL signature.
//
// https://cloudinary.com/documentation/advanced_url_delivery_options#generating_delivery_url_signatures
func (a Asset) signature() string {
	if !a.Config.URL.SignURL || a.AuthToken.isEnabled() {
		return ""
	}

	algo, length := a.getSignatureAlgorithmAndLength()

	toSign := joinURL([]interface{}{a.Transformation, a.PublicID})

	return signature.SignURL(toSign, a.Config.Cloud.APISecret, algo, length)
}

func (a Asset) getSignatureAlgorithmAndLength() (signature.Algo, signature.Length) {
	if a.Config.URL.GetSignatureLength() == signature.Long {
		return signature.SHA256, signature.Long
	}

	return a.Config.Cloud.GetSignatureAlgorithm(), a.Config.URL.GetSignatureLength()
}

// version finalizes the version part (v123) of the asset URL.
func (a Asset) version() string {
	var versionRegexp = regexp.MustCompile(`^v\d+`)
	version := a.Version
	if version == 0 &&
		a.Config.URL.ForceVersion &&
		filepath.Dir(a.PublicID) != "." &&
		!isURL(a.PublicID) &&
		!versionRegexp.MatchString(a.PublicID) {
		version = 1
	}

	if version != 0 {
		return fmt.Sprintf("v%d", version)
	}

	return ""
}

// version finalizes the source part (PublicID + Suffix) of the asset URL.
func (a Asset) source() string {
	source := fileNameWithoutExt(a.PublicID)

	if !isURL(source) {
		var err error
		source, err = url.QueryUnescape(strings.Replace(source, "%20", "+", -1))
		if err != nil {
			panic(err)
		}
	}

	source = smartEscape(source)

	if a.Suffix != "" {
		source += fmt.Sprintf("/%s", a.Suffix)
	}

	if filepath.Ext(a.PublicID) != "" {
		source += filepath.Ext(a.PublicID)
	}

	return source
}

func (a *Asset) path() string {
	return joinURL([]interface{}{a.assetType(), a.signature(), a.Transformation, a.version(), a.source()})
}

func (a *Asset) assetURL() string {
	return joinURL([]interface{}{distribution(a.PublicID, a.Config), a.path()})
}

func (a *Asset) query() string {
	// Currently, analytics is not supported with AuthToken. Just return AuthToken if it is configured.
	if a.Config.URL.SignURL && a.AuthToken.isEnabled() {
		u, err := url.Parse(a.assetURL())
		if err != nil {
			panic(err)
		}

		return a.AuthToken.Generate(u.Path)
	}

	if !a.Config.URL.Analytics {
		return ""
	}

	return fmt.Sprintf("%s=%s", queryString, sdkAnalyticsSignature())
}
func joinNonEmpty(items []interface{}, sep string) string {
	var parts []string
	for _, i := range items {
		s := fmt.Sprintf("%v", i)
		if strings.TrimSpace(s) != "" {
			parts = append(parts, s)
		}
	}

	return strings.Join(parts, sep)
}

func joinURL(items []interface{}) string {
	return joinNonEmpty(items, "/")
}

func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

var urlRegexp = regexp.MustCompile(`^https?://`)

func isURL(candidate string) bool {
	return urlRegexp.MatchString(candidate)
}

func smartEscape(str string) string {
	revert := strings.NewReplacer("%3A", ":", "%2F", "/")

	return revert.Replace(url.QueryEscape(str))
}
