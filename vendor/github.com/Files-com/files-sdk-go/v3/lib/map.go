package lib

import "sync"

type Map[T any] struct {
	sync.Map
}

func (m *Map[T]) Load(key any) (value T, ok bool) {
	var v interface{}
	v, ok = m.Map.Load(key)
	if ok {
		return v.(T), ok
	}
	return
}

func (m *Map[T]) Store(key any, value T) {
	m.Map.Store(key, value)
}

func (m *Map[T]) Delete(key any) {
	m.LoadAndDelete(key)
}
