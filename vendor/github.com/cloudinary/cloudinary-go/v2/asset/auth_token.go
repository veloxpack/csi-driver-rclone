package asset

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/cloudinary/cloudinary-go/v2/config"
	"regexp"
	"strconv"
	"time"
)

const unsafeAuthTokenChars = " \"#%&'/:;<=>?@\\[\\]^`{\\|}~\\"

const authTokenName = "__cld_token__"
const authTokenSeparator = "~"
const authTokenInnerSeparator = "="

// AuthToken is the Authentication Token struct.
type AuthToken struct {
	Config *config.AuthToken
}

func (a AuthToken) isEnabled() bool {
	return a.Config.Key != ""
}

// Generate generates the authentication token.
func (a AuthToken) Generate(path string) string {
	if !a.isEnabled() {
		return ""
	}

	start, expiration := a.handleLifetime()

	if path == "" && a.Config.ACL == "" {
		panic("AuthToken must contain either ACL or URL property")
	}

	var tokenParts []interface{}

	if a.Config.IP != "" {
		tokenParts = append(tokenParts, "ip="+a.Config.IP)
	}
	if start != 0 {
		tokenParts = append(tokenParts, "st="+strconv.FormatInt(start, 10))
	}
	if expiration != 0 {
		tokenParts = append(tokenParts, "exp="+strconv.FormatInt(expiration, 10))
	}
	if a.Config.ACL != "" {
		tokenParts = append(tokenParts, "acl="+escapeToLower(a.Config.ACL))
	}

	toSign := tokenParts

	if path != "" && a.Config.ACL == "" {
		toSign = append(tokenParts, "url="+escapeToLower(path))
	}

	auth := a.digest(joinNonEmpty(toSign, authTokenSeparator))

	tokenParts = append(tokenParts, "hmac="+auth)

	return joinNonEmpty([]interface{}{authTokenName, joinNonEmpty(tokenParts, authTokenSeparator)}, authTokenInnerSeparator)
}

func (a AuthToken) handleLifetime() (int64, int64) {
	expiration := a.Config.Expiration

	if expiration == 0 {
		if a.Config.Duration != 0 {
			start := a.Config.StartTime
			if start == 0 {
				start = time.Now().Unix()
			}
			expiration = start + a.Config.Duration
		} else {
			panic("must provide Expiration or Duration")
		}
	}

	return a.Config.StartTime, expiration
}

func escapeToLower(str string) string {
	re := regexp.MustCompile("([" + regexp.QuoteMeta(unsafeAuthTokenChars) + "])")

	return re.ReplaceAllStringFunc(str, func(s string) string {
		return "%" + hex.EncodeToString([]byte(s))
	})
}

func (a AuthToken) digest(message string) string {
	key, err := hex.DecodeString(a.Config.Key) // H* +`?
	if err != nil {
		panic(err.Error())
	}
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))

	return hex.EncodeToString(h.Sum(nil))
}
