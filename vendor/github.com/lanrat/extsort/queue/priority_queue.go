// Package queue provides a generic priority queue implementation optimized for external sorting.
// It is a binary heap of values that follows the same algorithm as Go's container/heap,
// without its interface: elements are not boxed, so Push and Pop do not allocate.
// The priority queue supports any type E with a user-provided comparison function.
package queue

import (
	"fmt"
)

// PriorityQueue is a generic priority queue that maintains elements in sorted order
// according to a user-provided comparison function. It provides efficient O(log n)
// insertion and removal of the minimum/maximum element. This implementation is
// specifically optimized for the external sorting use case where elements need
// to be efficiently merged from multiple sorted streams.
type PriorityQueue[E any] struct {
	items       []E // a binary heap: items[i] is not after its children 2i+1 and 2i+2
	compareFunc func(E, E) int
}

// NewPriorityQueue creates a new priority queue with the given comparison function.
// The cmpFunc should return a negative integer if the first argument should have higher priority
// (appear earlier) than the second argument, zero if they are equal, and a positive integer
// if the first should appear later. For ascending order, use cmp.Compare(a, b).
// The queue starts empty and elements can be added with Push().
func NewPriorityQueue[E any](cmpFunc func(E, E) int) *PriorityQueue[E] {
	return &PriorityQueue[E]{compareFunc: cmpFunc}
}

// Len returns the current number of elements in the priority queue.
// This operation is O(1).
func (pq *PriorityQueue[E]) Len() int {
	return len(pq.items)
}

// Push adds a new element to the priority queue, maintaining heap properties.
// The element will be positioned according to the comparison function provided
// during queue creation. This operation is O(log n).
func (pq *PriorityQueue[E]) Push(x E) {
	pq.items = append(pq.items, x)
	pq.up(len(pq.items) - 1)
}

// Pop removes and returns the highest priority element from the queue.
// The returned element is the one that would be returned by Peek().
// This operation is O(log n). Panics if the queue is empty.
func (pq *PriorityQueue[E]) Pop() E {
	n := len(pq.items) - 1
	top := pq.items[0]
	pq.items[0] = pq.items[n]
	var zero E
	pq.items[n] = zero // don't keep the popped element reachable
	pq.items = pq.items[:n]
	pq.down(0)
	return top
}

// Peek returns the highest priority element without removing it from the queue.
// This allows inspection of the next element that would be returned by Pop().
// This operation is O(1). Panics if the queue is empty.
func (pq *PriorityQueue[E]) Peek() E {
	return pq.items[0]
}

// PeekUpdate must be called after modifying the value returned by Peek() in-place.
// This re-establishes the heap property when the priority of the top element changes.
// This is more efficient than Pop() followed by Push() when updating the top element.
// This operation is O(log n).
func (pq *PriorityQueue[E]) PeekUpdate() {
	pq.down(0)
}

// Print outputs the current contents of the priority queue to stdout.
// Note that elements are printed in heap order, not priority order.
// This method is primarily intended for debugging purposes.
func (pq *PriorityQueue[E]) Print() {
	fmt.Print("[")
	for i := range pq.items {
		fmt.Print(pq.items[i], ", ")

	}
	fmt.Println("]")
}

func (pq *PriorityQueue[E]) less(i, j int) bool {
	return pq.compareFunc(pq.items[i], pq.items[j]) < 0
}

// up moves element j towards the root until it is not before its parent.
func (pq *PriorityQueue[E]) up(j int) {
	for j > 0 {
		i := (j - 1) / 2 // parent
		if !pq.less(j, i) {
			break
		}
		pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
		j = i
	}
}

// down moves element i towards the leaves until neither child is before it.
func (pq *PriorityQueue[E]) down(i int) {
	n := len(pq.items)
	for {
		j := 2*i + 1         // left child
		if j >= n || j < 0 { // j < 0 after int overflow
			break
		}
		if j2 := j + 1; j2 < n && pq.less(j2, j) {
			j = j2 // right child
		}
		if !pq.less(j, i) {
			break
		}
		pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
		i = j
	}
}
