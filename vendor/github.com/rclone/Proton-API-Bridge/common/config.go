package common

import (
	"log"
	"net/http"
	"os"
	"runtime"
)

// Logger is the minimum interface the bridge needs to emit log lines.
// Its method set matches resty.Logger and go-proton-api's WithLogger
// option, so a single value can be forwarded straight through to the
// underlying API client without conversion.
type Logger interface {
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

// stdlibLogger preserves the bridge's historical "print to stdlib log"
// behaviour for callers that don't supply a Logger via Config.
type stdlibLogger struct{}

func (stdlibLogger) Errorf(format string, v ...interface{}) { log.Printf("ERROR: "+format, v...) }
func (stdlibLogger) Warnf(format string, v ...interface{})  { log.Printf("WARN: "+format, v...) }
func (stdlibLogger) Debugf(format string, v ...interface{}) { log.Printf("DEBUG: "+format, v...) }

type Config struct {
	/* Constants */
	AppVersion string
	UserAgent  string

	/* HTTP */
	// Transport, if non-nil, is used as the HTTP RoundTripper for
	// all API requests. This lets callers inject their own
	// transport.
	Transport http.RoundTripper

	/* Logging */
	// Logger, if non-nil, receives all log output produced by the bridge
	// and is also forwarded to go-proton-api via proton.WithLogger so
	// that HTTP-layer warnings/errors emitted by resty go through the
	// same sink. When nil, the bridge falls back to printing through the
	// stdlib log package.
	Logger Logger

	/* Login */
	FirstLoginCredential *FirstLoginCredentialData
	ReusableCredential   *ReusableCredentialData
	UseReusableLogin     bool
	CredentialCacheFile  string // If CredentialCacheFile is empty, no credential will be logged

	/* Setting */
	DestructiveIntegrationTest     bool // CAUTION: the integration test requires a clean proton drive
	EmptyTrashAfterIntegrationTest bool // CAUTION: the integration test will clean up all the data in the trash
	ReplaceExistingDraft           bool // for the file upload replace or keep it as-is option
	EnableCaching                  bool // link node caching
	ConcurrentBlockUploadCount     int
	ConcurrentFileCryptoCount      int

	/* Drive */
	DataFolderName string
}

// GetLogger returns the configured Logger or a stdlib-backed default
// when Config.Logger is nil. Callers should use this rather than reading
// Config.Logger directly so they never need to nil-check.
func (c *Config) GetLogger() Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return stdlibLogger{}
}

type FirstLoginCredentialData struct {
	Username        string
	Password        string
	MailboxPassword string
	TwoFA           string
}

type ReusableCredentialData struct {
	UID           string
	AccessToken   string
	RefreshToken  string
	SaltedKeyPass string // []byte <-> base64
}

func NewConfigWithDefaultValues() *Config {
	return &Config{
		AppVersion: "",
		UserAgent:  "",

		FirstLoginCredential: &FirstLoginCredentialData{
			Username:        "",
			Password:        "",
			MailboxPassword: "",
			TwoFA:           "",
		},
		ReusableCredential: &ReusableCredentialData{
			UID:           "",
			AccessToken:   "",
			RefreshToken:  "",
			SaltedKeyPass: "", // []byte <-> base64
		},
		UseReusableLogin:    false,
		CredentialCacheFile: "",

		DestructiveIntegrationTest:     false,
		EmptyTrashAfterIntegrationTest: false,
		ReplaceExistingDraft:           false,
		EnableCaching:                  true,
		ConcurrentBlockUploadCount:     20, // let's be a nice citizen and not stress out proton engineers :)
		ConcurrentFileCryptoCount:      runtime.GOMAXPROCS(0),

		DataFolderName: "data",
	}
}

func NewConfigForIntegrationTests() *Config {
	appVersion := os.Getenv("PROTON_API_BRIDGE_APP_VERSION")
	userAgent := os.Getenv("PROTON_API_BRIDGE_USER_AGENT")

	username := os.Getenv("PROTON_API_BRIDGE_TEST_USERNAME")
	password := os.Getenv("PROTON_API_BRIDGE_TEST_PASSWORD")
	twoFA := os.Getenv("PROTON_API_BRIDGE_TEST_TWOFA")

	useReusableLoginStr := os.Getenv("PROTON_API_BRIDGE_TEST_USE_REUSABLE_LOGIN")
	useReusableLogin := useReusableLoginStr == "1"

	uid := os.Getenv("PROTON_API_BRIDGE_TEST_UID")
	accessToken := os.Getenv("PROTON_API_BRIDGE_TEST_ACCESS_TOKEN")
	refreshToken := os.Getenv("PROTON_API_BRIDGE_TEST_REFRESH_TOKEN")
	saltedKeyPass := os.Getenv("PROTON_API_BRIDGE_TEST_SALTEDKEYPASS")

	return &Config{
		AppVersion: appVersion,
		UserAgent:  userAgent,

		FirstLoginCredential: &FirstLoginCredentialData{
			Username:        username,
			Password:        password,
			MailboxPassword: "",
			TwoFA:           twoFA,
		},
		ReusableCredential: &ReusableCredentialData{
			UID:           uid,
			AccessToken:   accessToken,
			RefreshToken:  refreshToken,
			SaltedKeyPass: saltedKeyPass, // []byte <-> base64
		},
		UseReusableLogin:    useReusableLogin,
		CredentialCacheFile: ".credential",

		DestructiveIntegrationTest:     true,
		EmptyTrashAfterIntegrationTest: true,
		ReplaceExistingDraft:           false,
		EnableCaching:                  true,
		ConcurrentBlockUploadCount:     20,
		ConcurrentFileCryptoCount:      runtime.GOMAXPROCS(0),

		DataFolderName: "data",
	}
}
