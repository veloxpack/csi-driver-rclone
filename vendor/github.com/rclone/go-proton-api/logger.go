package proton

import (
	"sync/atomic"

	"github.com/go-resty/resty/v2"
)

// pkgLogger is used by package-level code paths that don't have access
// to a Manager (currently Keys.Unlock and the retry-after middleware).
// It is updated whenever a Manager is built with the WithLogger option;
// the most recently configured logger wins. When nil, the call sites
// fall back to logrus so the library's historical behaviour is
// preserved for callers that don't supply a logger.
var pkgLogger atomic.Pointer[resty.Logger]

func setPkgLogger(l resty.Logger) {
	if l == nil {
		return
	}
	pkgLogger.Store(&l)
}

func getPkgLogger() resty.Logger {
	if l := pkgLogger.Load(); l != nil {
		return *l
	}
	return nil
}
