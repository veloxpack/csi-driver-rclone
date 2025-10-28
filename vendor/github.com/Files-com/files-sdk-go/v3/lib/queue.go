package lib

import "sync"

type queue[T any] []T

func (q *queue[T]) Push(x T) {
	*q = append(*q, x)
}

func (q *queue[T]) Pop() T {
	h := *q
	var el T
	l := len(h)
	if l == 0 {
		return el
	}
	el, *q = h[0], h[1:l]
	return el
}

func (q *queue[T]) Clear() {
	*q = queue[T]{}
}

type Queue[T any] struct {
	queue queue[T]
	*sync.RWMutex
}

func (q *Queue[T]) Init(size int) *Queue[T] {
	q.RWMutex = &sync.RWMutex{}
	q.queue = make(queue[T], 0, size)
	return q
}

func (q *Queue[T]) Len() int {
	q.RLock()
	defer q.RUnlock()
	return len(q.queue)
}

func (q *Queue[T]) Push(item T) {
	q.Lock()
	defer q.Unlock()
	q.queue.Push(item)
}

func (q *Queue[T]) Pop() T {
	q.Lock()
	defer q.Unlock()
	return q.queue.Pop()
}

func (q *Queue[T]) Clear() {
	q.Lock()
	defer q.Unlock()
	*q = Queue[T]{}
}
