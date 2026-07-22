//
// Copyright 2019-2026 Aaron H. Alpar
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files
// (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge,
// publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
// CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//

package deheap

import "cmp"

// Deheap is a type-safe doubly-ended heap for cmp.Ordered types.
//
// Unlike the v1 interface-based API, this implementation operates
// directly on the underlying []T slice with native < comparisons,
// avoiding interface dispatch, adapter allocation, and boxing overhead.
//
// Usage:
//
//	h := deheap.From(5, 3, 8, 1, 9)
//	h.Peek()    // 1  — O(1) minimum
//	h.PeekMax() // 9  — O(1) maximum
//	h.Pop()     // 1  — O(log n) remove minimum
//	h.PopMax()  // 9  — O(log n) remove maximum
type Deheap[T cmp.Ordered] struct {
	items []T
}

// New returns an empty Deheap.
func New[T cmp.Ordered]() *Deheap[T] {
	return &Deheap[T]{}
}

// From constructs a Deheap from the given elements and initializes
// the heap ordering using Floyd's bottom-up heap construction.
//
// If the input already satisfies the heap property, From returns
// after a linear scan with no modifications.
func From[T cmp.Ordered](items ...T) *Deheap[T] {
	q := &Deheap[T]{items: make([]T, len(items))}
	copy(q.items, items)
	l := len(q.items)
	if !orderedValid(q.items, l) {
		for i := (l - 1) / 2; i >= 0; i-- {
			orderedBubbledown(q.items, l, isMinHeap(i), i)
		}
	}
	return q
}

// Push adds an element to the heap.
// Time complexity is O(log n), where n = h.Len().
func (p *Deheap[T]) Push(o T) {
	p.items = append(p.items, o)
	orderedBubbleup(p.items, isMinHeap(len(p.items)-1), len(p.items)-1)
}

// Pop removes and returns the smallest element from the heap.
// Returns the zero value of T if the heap is empty.
// Time complexity is O(log n), where n = h.Len().
func (p *Deheap[T]) Pop() T {
	if len(p.items) == 0 {
		var zero T
		return zero
	}
	l := len(p.items) - 1
	p.items[0], p.items[l] = p.items[l], p.items[0]
	v := p.items[l]
	p.items = p.items[:l]
	orderedBubbledown(p.items, l, true, 0)
	return v
}

// PopMax removes and returns the largest element from the heap.
// Returns the zero value of T if the heap is empty.
// Time complexity is O(log n), where n = h.Len().
func (p *Deheap[T]) PopMax() T {
	if len(p.items) == 0 {
		var zero T
		return zero
	}
	l := len(p.items)
	j := 0
	if l > 1 {
		j = orderedMin2(p.items, l, false, 1)
	}
	l--
	p.items[j], p.items[l] = p.items[l], p.items[j]
	v := p.items[l]
	p.items = p.items[:l]
	orderedBubbledown(p.items, l, false, j)
	return v
}

// Remove removes and returns the element at index i.
// It panics if i is out of bounds.
// Time complexity is O(log n), where n = h.Len().
func (p *Deheap[T]) Remove(i int) T {
	l := len(p.items) - 1
	p.items[i], p.items[l] = p.items[l], p.items[i]
	v := p.items[l]
	p.items = p.items[:l]
	if l != i {
		q, r := orderedBubbledown(p.items, l, isMinHeap(i), i)
		orderedBubbleup(p.items, isMinHeap(q), q)
		orderedBubbleup(p.items, isMinHeap(r), r)
	}
	return v
}

// Fix re-establishes the heap ordering after the element at index i
// has changed its value. Equivalent to, but cheaper than, Remove(i)
// followed by Push of the new value.
//
// The index i must be in the range [0, p.Len()).
// It panics if i is out of bounds.
//
// The complexity is O(log n) where n = p.Len().
func (p *Deheap[T]) Fix(i int) {
	l := len(p.items)
	min := isMinHeap(i)
	pos := i
	for {
		j := orderedMin2(p.items, l, min, hlchild(pos))
		if j >= l {
			break
		}
		k := orderedMin4(p.items, l, min, lchild(pos))
		v := orderedMin3(p.items, l, min, pos, j, k)
		if v == pos || v >= l {
			break
		}
		p.items[v], p.items[pos] = p.items[pos], p.items[v]
		if v == j {
			pos = v
			break
		}
		hp := hparent(v)
		if orderedLess(p.items, min, hp, v) {
			p.items[hp], p.items[v] = p.items[v], p.items[hp]
			orderedBubbleup(p.items, isMinHeap(hp), hp)
		}
		pos = v
	}
	orderedBubbleup(p.items, isMinHeap(pos), pos)
	if pos != i {
		orderedBubbleup(p.items, isMinHeap(i), i)
	}
}

// Len returns the number of elements in the heap.
func (p *Deheap[T]) Len() int {
	return len(p.items)
}

// Peek returns the smallest element without removing it.
// It panics if the heap is empty.
func (p *Deheap[T]) Peek() T {
	return p.items[0]
}

// PeekMax returns the largest element without removing it.
// It panics if the heap is empty.
//
// In a min-max heap the maximum is always one of the root's children
// (index 1 or 2), since they sit on the first max level. With only
// one element the root is both min and max; with two elements the
// sole child at index 1 is the max.
//
//	    1          ← min (root)
//	  /   \
//	[9]    5       ← max is the larger child
func (p *Deheap[T]) PeekMax() T {
	if len(p.items) <= 1 {
		return p.items[0]
	}
	if len(p.items) == 2 {
		return p.items[1]
	}
	if p.items[1] > p.items[2] {
		return p.items[1]
	}
	return p.items[2]
}

// Verify reports whether the heap satisfies the min-max heap property.
//
// Time complexity is O(n), where n = p.Len().
func (p *Deheap[T]) Verify() bool {
	return orderedValid(p.items, len(p.items))
}

// ---------------------------------------------------------------------------
// Generic algorithm functions
//
// These mirror the v1 functions in deheap.go (bubbleup, bubbledown,
// min2, min3, min4) but operate directly on []T with native <
// comparisons instead of going through heap.Interface. This eliminates
// interface dispatch, adapter allocation, and interface{} boxing.
//
// The navigation helpers (hparent, hlchild, parent, lchild, level,
// isMinHeap) are shared — they are pure index arithmetic with no
// type dependency.
// ---------------------------------------------------------------------------

// orderedValid reports whether items satisfies the min-max heap property.
// See valid in deheap.go for the algorithm description.
func orderedValid[T cmp.Ordered](items []T, l int) bool {
	for i := 1; i < l; i++ {
		hp := hparent(i)
		if isMinHeap(i) {
			if items[hp] < items[i] {
				return false
			}
		} else {
			if items[i] < items[hp] {
				return false
			}
		}
		if i >= 3 {
			gp := hparent(hp)
			if isMinHeap(i) {
				if items[i] < items[gp] {
					return false
				}
			} else {
				if items[gp] < items[i] {
					return false
				}
			}
		}
	}
	return true
}

// orderedLess compares two elements, respecting the min flag.
// When min=true, returns whether items[a] < items[b].
// When min=false, returns whether items[a] > items[b].
// This mirrors the v1 pattern: min == h.Less(a, b).
func orderedLess[T cmp.Ordered](items []T, min bool, a, b int) bool {
	if min {
		return items[a] < items[b]
	}
	return items[b] < items[a]
}

// orderedMin2 finds the extremum among 2 consecutive elements at i and i+1.
func orderedMin2[T cmp.Ordered](items []T, l int, min bool, i int) int {
	if i+1 < l && orderedLess(items, min, i+1, i) {
		return i + 1
	}
	return i
}

// orderedMin3 finds the extremum among up to 3 elements at indices i, j, k.
func orderedMin3[T cmp.Ordered](items []T, l int, min bool, i, j, k int) int {
	q := i
	if j < l && orderedLess(items, min, j, q) {
		q = j
	}
	if k < l && orderedLess(items, min, k, q) {
		q = k
	}
	return q
}

// orderedMin4 finds the extremum among up to 4 consecutive elements
// starting at index i.
func orderedMin4[T cmp.Ordered](items []T, l int, min bool, i int) int {
	q := i
	i++
	if i >= l {
		return q
	}
	if orderedLess(items, min, i, q) {
		q = i
	}
	i++
	if i >= l {
		return q
	}
	if orderedLess(items, min, i, q) {
		q = i
	}
	i++
	if i >= l {
		return q
	}
	if orderedLess(items, min, i, q) {
		q = i
	}
	return q
}

// orderedBubbledown restores the heap property downward from index i.
func orderedBubbledown[T cmp.Ordered](items []T, l int, min bool, i int) (q int, r int) {
	q = i
	r = i
	for {
		j := orderedMin2(items, l, min, hlchild(i))
		if j >= l {
			break
		}
		k := orderedMin4(items, l, min, lchild(i))
		v := orderedMin3(items, l, min, i, j, k)
		if v == i || v >= l {
			break
		}
		q = v
		items[v], items[i] = items[i], items[v]
		if v == j {
			break
		}
		p := hparent(v)
		if orderedLess(items, min, p, v) {
			items[p], items[v] = items[v], items[p]
			r = p
		}
		i = v
	}
	return q, r
}

// orderedBubbleup restores the heap property upward from index i.
func orderedBubbleup[T cmp.Ordered](items []T, min bool, i int) {
	if i < 0 {
		return
	}
	j := parent(i)
	for j >= 0 && orderedLess(items, min, i, j) {
		items[i], items[j] = items[j], items[i]
		i = j
		j = parent(i)
	}
	min = !min
	j = hparent(i)
	for j >= 0 && orderedLess(items, min, i, j) {
		items[i], items[j] = items[j], items[i]
		i = j
		j = parent(i)
	}
}
