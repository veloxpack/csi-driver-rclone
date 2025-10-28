package lib

import (
	"sync"
)

type AtomicValue[T comparable] struct {
	v T
	sync.RWMutex
}

func (av *AtomicValue[T]) Load() T {
	av.RLock()
	defer av.RUnlock()
	return av.v
}

func (av *AtomicValue[T]) Store(val T) {
	av.Lock()
	defer av.Unlock()
	av.v = val
}

func (av *AtomicValue[T]) CompareAndUpdate(cmpValue T, updateFunc func() T) {
	av.Lock()
	defer av.Unlock()
	if av.v == cmpValue {
		av.v = updateFunc()
	}
}
