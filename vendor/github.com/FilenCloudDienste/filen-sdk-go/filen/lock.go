package filen

import (
	"context"
	"errors"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/client"
	"github.com/FilenCloudDienste/filen-sdk-go/filen/types"
	"github.com/google/uuid"
	"sync"
	"time"
)

// BackendLock is a lock to prevent time of check to time of use bugs
// specifically situations where a file is deleted/moved/renamed
// while a longer action (like a rclone DirMove) is running
// which could cause unexpected states.
//
// It provides a reference counting mechanism so multiple operations can acquire
// the same lock, and the actual backend lock is only released when all operations
// have completed.
type BackendLock struct {
	mu           types.CtxMutex // Mutex for coordinating lock access with context support
	count        int            // Reference count for the lock
	lockUUID     string         // UUID of the current lock on the server
	cancelTicker chan struct{}  // Channel to stop the refresh ticker

	muPoisoned sync.RWMutex // Mutex for the poisoned flag
	poisoned   bool         // Indicates if the lock is poisoned (refresh failed)
}

// Lock configuration constants
const (
	resourceName       = "drive-write"           // Name of the resource to lock on the server
	maxLockAttempts    = 100                     // Maximum number of attempts to acquire the lock
	retryLockSleepTime = 1000 * time.Millisecond // Time to sleep between lock acquisition attempts
	refreshInterval    = 20 * time.Second        // How often to refresh the lock
)

// Common lock errors
var (
	// LockPoisoned is returned when a lock refresh has failed, indicating the lock
	// is no longer valid on the server side.
	LockPoisoned = errors.New("lock refresh failed")
)

// NewBackendLock returns a new BackendLock instance.
// The lock is initially unlocked and ready to use.
func NewBackendLock() BackendLock {
	return BackendLock{
		mu:           types.NewCtxMutex(),
		cancelTicker: make(chan struct{}, 1),
	}
}

// acquireBackendLock attempts to acquire a lock on the backend server.
// It will retry up to maxLockAttempts times with a delay between attempts.
// Once acquired, it starts a background goroutine to refresh the lock periodically.
func (api *Filen) acquireBackendLock(ctx context.Context) error {
	req := client.V3UserLockRequest{
		LockUUID: uuid.NewString(),
		Type:     "acquire",
		Resource: resourceName,
	}
	for i := 0; i < maxLockAttempts; i++ {
		resp, err := api.Client.PostV3UserLock(ctx, req)
		if err != nil {
			return err
		}
		if resp.Acquired {
			break
		}
		time.Sleep(retryLockSleepTime)
	}
	api.lock.lockUUID = req.LockUUID

	// new lock acquired, reset the poison flag
	api.lock.muPoisoned.Lock()
	api.lock.poisoned = false
	api.lock.muPoisoned.Unlock()

	ticker := time.NewTicker(refreshInterval)
	go api.refreshLockHandler(ticker)
	return nil
}

// refreshLockHandler is a background goroutine that periodically refreshes the lock
// on the backend server to prevent it from expiring. If a refresh fails, the lock
// is marked as poisoned.
func (api *Filen) refreshLockHandler(ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			resp, err := api.Client.PostV3UserLock(context.Background(), client.V3UserLockRequest{
				LockUUID: api.lock.lockUUID,
				Type:     "refresh",
				Resource: resourceName,
			})
			if err == nil && resp.Refreshed {
				continue
			}
			api.lock.muPoisoned.Lock()
			api.lock.poisoned = true
			api.lock.muPoisoned.Unlock()
			return
		case <-api.lock.cancelTicker:
			return
		}
	}
}

// releaseBackendLock releases the lock on the backend server.
// This is called when the reference count reaches zero, indicating
// all operations using the lock have completed.
func (api *Filen) releaseBackendLock() {
	api.lock.cancelTicker <- struct{}{}
	// we use context.Background here because this function should always be executed
	// even if the original context is cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := api.Client.PostV3UserLock(ctx, client.V3UserLockRequest{
		LockUUID: api.lock.lockUUID,
		Type:     "release",
		Resource: resourceName,
	})
	api.lock.lockUUID = ""
	if err != nil || !resp.Released {
		api.lock.muPoisoned.Lock()
		api.lock.poisoned = true
		api.lock.muPoisoned.Unlock()
	}
}

// Lock acquires the lock for an operation. If this is the first operation to
// acquire the lock, it will attempt to acquire the actual backend lock.
// Subsequent calls will increment the reference count.
//
// It will return an error if:
// - The context is cancelled
// - The lock is poisoned (a previous refresh failed)
// - The first call fails to acquire the backend lock
//
// This function is context-aware and can be cancelled via the provided context.
func (api *Filen) Lock(ctx context.Context) error {
	err := api.lock.mu.Lock(ctx)
	if err != nil {
		return err
	}
	defer api.lock.mu.Unlock()

	if api.lock.count == 0 {
		err = api.acquireBackendLock(ctx)
		if err != nil {
			return err
		}
	} else {
		api.lock.muPoisoned.RLock()
		if api.lock.poisoned {
			api.lock.muPoisoned.RUnlock()
			return LockPoisoned
		}
		api.lock.muPoisoned.RUnlock()
	}

	api.lock.count++
	return nil
}

// Unlock decrements the lock reference count.
// If the count reaches zero, it releases the backend lock.
//
// This function always completes, even if other operations were cancelled,
// to ensure proper cleanup of resources.
func (api *Filen) Unlock() {
	// we use BlockUntilLock here because this function should always be executed
	api.lock.mu.BlockUntilLock()
	defer api.lock.mu.Unlock()
	api.lock.count--
	if api.lock.count == 0 {
		api.releaseBackendLock()
	}
}
