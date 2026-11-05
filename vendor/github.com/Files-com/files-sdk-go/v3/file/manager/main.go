package manager

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/Files-com/files-sdk-go/v3/lib"
)

var (
	// ConcurrentFiles is the default number of files a transfer job may process at once.
	ConcurrentFiles = 50
	// ConcurrentFileParts is the default job-wide cap for concurrent upload/download part work.
	ConcurrentFileParts = 50
	// ConcurrentDirectoryList is the default number of local directory listing operations that may run at once.
	ConcurrentDirectoryList = 100
	// AdaptiveUploadV2ConcurrentFiles is the V2 upload job-level file concurrency cap.
	// Adaptive upload V2 still controls HTTP part concurrency dynamically; this value prevents
	// large multi-file jobs from being serialized while the per-destination controller learns.
	// The default is documented in docs/adaptive-upload-v2-file-concurrency-cap.md.
	AdaptiveUploadV2ConcurrentFiles = 128
	// AdaptiveUploadV2ConcurrentFileParts is the V2 upload HTTP part concurrency cap.
	// The adaptive manager treats this as a maximum, not a fixed target.
	AdaptiveUploadV2ConcurrentFileParts = 1024
	// AdaptiveDownloadV2ConcurrentFiles is the V2 download job-level file concurrency cap.
	AdaptiveDownloadV2ConcurrentFiles = 128
	// AdaptiveDownloadV2ConcurrentFileParts is the V2 download HTTP range concurrency cap.
	// The adaptive manager treats this as a maximum, not a fixed target.
	AdaptiveDownloadV2ConcurrentFileParts = 1024
	// AdaptiveFileAdmissionFloor is the minimum number of concurrent files
	// adaptive transfers keep admitted however low the learned part
	// concurrency falls, so indexing, finishing, and growth probing still
	// overlap.
	AdaptiveFileAdmissionFloor = 4
)

var (
	sharedDefaultOnce    sync.Once
	sharedDefaultManager *Manager
)

var (
	adaptiveUploadV2ConcurrentFilePartsOverride   atomic.Int64
	adaptiveDownloadV2ConcurrentFilePartsOverride atomic.Int64
)

// SetAdaptiveUploadV2ConcurrentFileParts sets the process-wide maximum for
// adaptive upload V2 HTTP part concurrency. Values less than 1 reset the SDK
// to its target-specific defaults.
func SetAdaptiveUploadV2ConcurrentFileParts(maxConcurrentFileParts int) {
	adaptiveUploadV2ConcurrentFilePartsOverride.Store(adaptiveConcurrentFilePartsOverride(maxConcurrentFileParts))
}

// EffectiveAdaptiveUploadV2ConcurrentFileParts returns defaultMax capped by the
// process-wide adaptive upload V2 override, when one is configured.
func EffectiveAdaptiveUploadV2ConcurrentFileParts(defaultMax int) int {
	return effectiveAdaptiveConcurrentFileParts(defaultMax, adaptiveUploadV2ConcurrentFilePartsOverride.Load())
}

// SetAdaptiveDownloadV2ConcurrentFileParts sets the process-wide maximum for
// adaptive download V2 HTTP range concurrency. Values less than 1 reset the SDK
// to its default.
func SetAdaptiveDownloadV2ConcurrentFileParts(maxConcurrentFileParts int) {
	adaptiveDownloadV2ConcurrentFilePartsOverride.Store(adaptiveConcurrentFilePartsOverride(maxConcurrentFileParts))
}

// EffectiveAdaptiveDownloadV2ConcurrentFileParts returns the adaptive download
// V2 max with the process-wide override applied, when one is configured.
func EffectiveAdaptiveDownloadV2ConcurrentFileParts(defaultMax ...int) int {
	maxConcurrentFileParts := AdaptiveDownloadV2ConcurrentFileParts
	if len(defaultMax) > 0 {
		maxConcurrentFileParts = defaultMax[0]
	}
	return effectiveAdaptiveConcurrentFileParts(
		maxConcurrentFileParts,
		adaptiveDownloadV2ConcurrentFilePartsOverride.Load(),
	)
}

func adaptiveConcurrentFilePartsOverride(maxConcurrentFileParts int) int64 {
	if maxConcurrentFileParts < 1 {
		return 0
	}
	return int64(maxConcurrentFileParts)
}

func effectiveAdaptiveConcurrentFileParts(defaultMax int, override int64) int {
	if defaultMax < 1 {
		defaultMax = 1
	}
	if override < 1 {
		return defaultMax
	}
	if overrideMax := int(override); overrideMax < defaultMax {
		return overrideMax
	}
	return defaultMax
}

type Manager struct {
	FilesManager            ConcurrencyManager
	FilePartsManager        ConcurrencyManager
	DirectoryListingManager ConcurrencyManager
}

// TransferStats describes current transfer worker usage.
type TransferStats struct {
	Active int
	Max    int
}

type ConcurrencyManager struct {
	*lib.ConstrainedWorkGroup
	DownloadFilesAsSingleStream bool
}

func (ConcurrencyManager) New(maxGoRoutines int, downloadFilesAsSingleStream ...bool) ConcurrencyManager {
	if len(downloadFilesAsSingleStream) == 0 {
		downloadFilesAsSingleStream = append(downloadFilesAsSingleStream, false)
	}

	return ConcurrencyManager{ConstrainedWorkGroup: lib.NewConstrainedWorkGroup(maxGoRoutines), DownloadFilesAsSingleStream: downloadFilesAsSingleStream[0]}
}

func (c ConcurrencyManager) Max() int {
	return c.ConstrainedWorkGroup.Max()
}

func New(files, fileParts, directoryListing int) *Manager {
	return &Manager{
		FilesManager:            ConcurrencyManager{}.New(files),
		FilePartsManager:        ConcurrencyManager{}.New(fileParts),
		DirectoryListingManager: ConcurrencyManager{}.New(directoryListing),
	}
}

func Default() *Manager {
	return New(ConcurrentFiles, ConcurrentFileParts, ConcurrentDirectoryList)
}

func SharedDefault() *Manager {
	sharedDefaultOnce.Do(func() {
		sharedDefaultManager = Default()
	})
	return sharedDefaultManager
}

func Sync() *Manager {
	return New(1, 1, 1)
}

func Build(maxConcurrentConnections, maxConcurrentDirectoryLists int, downloadFilesAsSingleStream ...bool) *Manager {
	maxConcurrentConnections = int(math.Max(float64(maxConcurrentConnections), 1))
	return &Manager{
		FilesManager:            ConcurrencyManager{}.New(maxConcurrentConnections),
		FilePartsManager:        ConcurrencyManager{}.New(maxConcurrentConnections, downloadFilesAsSingleStream...),
		DirectoryListingManager: ConcurrencyManager{}.New(int(math.Max(float64(maxConcurrentDirectoryLists), 1))),
	}
}

func (m *Manager) CreateMatchingClient(client *http.Client) *http.Client {
	if fmt.Sprintf("%T", client.Transport) == "*recorder.Recorder" {
		// Can't modify VCR in Test mode.
		return client
	}

	switch t := client.Transport.(type) {
	case *lib.Transport:
		t.MaxConnsPerHost = m.FilePartsManager.Max()
		return client
	default:
		defaultTransport := lib.DefaultPooledTransport()
		defaultTransport.MaxConnsPerHost = m.FilePartsManager.Max()
		client.Transport = defaultTransport
		return client
	}
}

// TransferStats returns the transfer worker usage for this manager.
func (m *Manager) TransferStats() TransferStats {
	if m == nil || m.FilePartsManager.ConstrainedWorkGroup == nil {
		return TransferStats{}
	}
	return TransferStats{
		Active: max(m.FilePartsManager.RunningCount(), 0),
		Max:    max(m.FilePartsManager.Max(), 0),
	}
}
