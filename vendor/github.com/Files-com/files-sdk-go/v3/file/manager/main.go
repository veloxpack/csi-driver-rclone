package manager

import (
	"fmt"
	"math"
	"net/http"

	"github.com/Files-com/files-sdk-go/v3/lib"
)

var (
	ConcurrentFiles         = 50
	ConcurrentFileParts     = 50
	ConcurrentDirectoryList = 100
)

type Manager struct {
	FilesManager            ConcurrencyManager
	FilePartsManager        ConcurrencyManager
	DirectoryListingManager ConcurrencyManager
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
