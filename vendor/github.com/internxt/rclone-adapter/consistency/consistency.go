// Package consistency provides a gate to handle eventual consistency
// when creating folders. After a folder is created on the server, it
// may not be immediately visible to other API endpoints. TrackFolder
// records the creation time, and AwaitFolder blocks only for the
// remaining window before the folder is expected to be consistent.
// Entries self-evict via time.AfterFunc, keeping memory bounded.
// This aims to prevent this issue: https://inxt.atlassian.net/browse/PB-1446
package consistency

import (
	"context"
	"sync"
	"time"
)

var recentFolders sync.Map

const window = 500 * time.Millisecond

// TrackFolder records that a folder was just created. The entry
// self-deletes after the consistency window elapses.
func TrackFolder(uuid string) {
	recentFolders.Store(uuid, time.Now())
	time.AfterFunc(window, func() {
		recentFolders.Delete(uuid)
	})
}

// AwaitFolder blocks until the consistency window has elapsed for a
// recently created folder. Returns immediately for unknown or already
// consistent folders.
func AwaitFolder(ctx context.Context, folderUUID string) error {
	v, ok := recentFolders.Load(folderUUID)
	if !ok {
		return nil
	}

	remaining := window - time.Since(v.(time.Time))
	if remaining <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(remaining):
		return nil
	}
}
