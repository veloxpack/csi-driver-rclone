package lib

import (
	"sync"
)

type KeyedMutex struct {
	m *sync.Map
}

func NewKeyedMutex() KeyedMutex {
	m := sync.Map{}
	return KeyedMutex{&m}
}

func (s KeyedMutex) Unlock(key interface{}) {
	l, exist := s.m.Load(key)
	if !exist {
		panic("KeyedMutex: unlock of unlocked mutex")
	}
	l_ := l.(*sync.Mutex)
	s.m.Delete(key)
	l_.Unlock()
}

func (s KeyedMutex) Lock(key interface{}) {
	m := sync.Mutex{}
	m_, _ := s.m.LoadOrStore(key, &m)
	mm := m_.(*sync.Mutex)
	mm.Lock()
	if mm != &m {
		mm.Unlock()
		s.Lock(key)
		return
	}
	return
}
