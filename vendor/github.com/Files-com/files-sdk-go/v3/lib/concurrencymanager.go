package lib

import "context"

type ConcurrencyManager interface {
	// Wait until a slot is available for the new goroutine.
	Wait()

	// Done Mark a goroutine as finished
	Done()

	// WaitAllDone Wait for all goroutines are done
	WaitAllDone()

	// RunningCount Returns the number of goroutines which are running
	RunningCount() int

	// WaitWithContext Acquires a semaphore to allow a new goroutine to run or returns false if the context is done
	WaitWithContext(ctx context.Context) bool

	// WaitForADone Blocks until at least one goroutine has completed.
	WaitForADone() bool
}

type ConcurrencyManagerWithSubWorker interface {
	ConcurrencyManager
	NewSubWorker() ConcurrencyManager
}
