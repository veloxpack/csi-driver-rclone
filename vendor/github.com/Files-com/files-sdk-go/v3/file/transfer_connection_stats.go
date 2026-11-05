package file

import (
	"sync"

	"github.com/Files-com/files-sdk-go/v3/file/manager"
	"github.com/Files-com/files-sdk-go/v3/lib"
)

// TransferStats describes current transfer worker usage.
type TransferStats = manager.TransferStats

// DirectionalTransferStats groups transfer worker usage by direction.
type DirectionalTransferStats struct {
	Upload   TransferStats
	Download TransferStats
}

type adaptiveManagerRegistry[K comparable] struct {
	mu       sync.Mutex
	managers map[K]*lib.AdaptiveConcurrencyManager
}

// AdaptiveTransferStats returns transfer worker usage for the process-wide
// adaptive upload and download managers.
func AdaptiveTransferStats() DirectionalTransferStats {
	return DirectionalTransferStats{
		Upload:   uploadV2SharedAdaptiveManagers.transferStats(),
		Download: downloadV2SharedAdaptiveManagers.transferStats(),
	}
}

func (r *adaptiveManagerRegistry[K]) managerFor(key K, create func() *lib.AdaptiveConcurrencyManager) *lib.AdaptiveConcurrencyManager {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.managers == nil {
		r.managers = make(map[K]*lib.AdaptiveConcurrencyManager)
	}
	if manager := r.managers[key]; manager != nil {
		return manager
	}
	manager := create()
	r.managers[key] = manager
	return manager
}

func (r *adaptiveManagerRegistry[K]) transferStats() TransferStats {
	r.mu.Lock()
	defer r.mu.Unlock()
	var stats TransferStats
	for _, adaptiveManager := range r.managers {
		if adaptiveManager == nil {
			continue
		}
		active := max(adaptiveManager.RunningCount(), 0)
		if active == 0 {
			continue
		}
		stats.Active += active
		stats.Max += max(adaptiveManager.Max(), 0)
	}
	return stats
}
