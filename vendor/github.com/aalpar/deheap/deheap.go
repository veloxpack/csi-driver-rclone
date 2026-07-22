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

// Package deheap provides a doubly-ended heap (min-max heap).
//
// A min-max heap gives O(log n) access to both the smallest AND largest
// element in a collection — something a standard binary heap cannot do.
//
// # How a min-max heap works
//
// Like a binary heap, a min-max heap is a complete binary tree stored in
// a flat slice. The difference is that tree levels alternate between
// "min levels" and "max levels":
//
//	Level 0 (min):               1             ← guaranteed minimum
//	                           /   \
//	Level 1 (max):            9     5          ← maximum is among these
//	                        / | \  / \
//	Level 2 (min):         2  3  4 3  2        ← each ≤ all descendants
//	                      /|
//	Level 3 (max):       6  8                  ← each ≥ all descendants
//
// The invariant: every node on a min level is ≤ all of its descendants,
// and every node on a max level is ≥ all of its descendants.
//
// This means:
//   - The root (index 0) is always the global minimum.
//   - The global maximum is one of the root's children (index 1 or 2).
//   - Both Peek and PeekMax are O(1).
//
// The tree is stored as a flat slice in level-order, exactly like a
// standard binary heap:
//
//	Index:   0  1  2  3  4  5  6  7  8  9
//	Value: [ 1, 9, 5, 2, 3, 4, 3, 2, 6, 8 ]
//	Level:   0  1  1  2  2  2  2  3  3  3
//	Kind:  min max    min            max
//
// # Worked example: Push(0) into [1, 9, 5, 4, 6, 8, 7, 3, 2]
//
// We append 0 at index 9, then bubble up:
//
//	Step 0 — append to end:
//
//	                 1
//	               /   \
//	              9     5
//	            / | \  / \
//	           4  6  8 7  3
//	          /|
//	         2  0  ← new element at index 9 (max level)
//
//	Step 1 — index 9 is on a max level, but 0 < parent 6 (index 4).
//	         0 is smaller than its parent, so it belongs on a min level.
//	         Swap with parent:
//
//	                 1
//	               /   \
//	              9     5
//	            / | \  / \
//	           4  0  8 7  3
//	          /|
//	         2  6
//
//	Step 2 — now at index 4 (max level), but we switched to min-level
//	         bubbling. Compare with grandparent (index 0): 0 < 1.
//	         Swap with grandparent:
//
//	                 0          ← new minimum
//	               /   \
//	              9     5
//	            / | \  / \
//	           4  1  8 7  3
//	          /|
//	         2  6
//
//	Done. The heap property is restored. The new minimum 0 is at the root.
//
// # Worked example: Pop() from [1, 9, 5, 4, 6, 3, 2]
//
//	Step 0 — swap root (index 0) with last element (index 6), then
//	         remove the last element (the old root, value 1):
//
//	         Before:              After swap & remove:
//
//	              1                    2
//	            /   \               /   \
//	           9     5             9     5
//	          /|\   /             /|\
//	         4  6 3 2            4  6  3
//
//	Step 1 — bubble down from index 0 (min level). Find the smallest
//	         among children {9,5} and grandchildren {4,6,3}. Smallest
//	         is 3 at index 5 (a grandchild).
//	         Swap index 0 with index 5:
//
//	              3
//	            /   \
//	           9     5
//	          /|\
//	         4  6  2  ← but wait, 2 < parent 5, so swap with parent
//
//	Step 2 — after grandchild swap, check if the moved element (2)
//	         violates the max-level parent. 2 < 5, so swap 2 and 5:
//
//	              3
//	            /   \
//	           9     2
//	          /|\
//	         4  6  5
//
//	Done. Returned 1 (the old minimum). New minimum is 3.
package deheap

import (
	"container/heap"
	"math/bits"
)

// hparent returns the binary-tree parent of node i.
//
//	     0
//	   /   \
//	  1     2        hparent(5) = (5-1)/2 = 2
//	 / \   / \
//	3   4 5   6
func hparent(i int) int {
	return (i - 1) / 2
}

// hlchild returns the left child of node i in the binary tree.
//
//	     0
//	   /   \
//	  1     2        hlchild(1) = (1*2)+1 = 3
//	 / \   / \
//	3   4 5   6
func hlchild(i int) int {
	return (i * 2) + 1
}

// parent returns the grandparent of node i (two levels up).
// Returns -1 if i has no grandparent (i.e., i is on level 0 or 1).
//
// Grandparent links connect nodes on the SAME level type (min↔min
// or max↔max). bubbleup uses these to move an element up through
// its own "sub-heap" without crossing level types.
//
//	Level 0 (min):     0                parent(7) = ((7+1)/4)-1 = 1
//	Level 1 (max):    1  2              parent(3) = ((3+1)/4)-1 = 0
//	Level 2 (min):   3 4 5 6            parent(0) = -1 (no grandparent)
//	Level 3 (max):  7 8
func parent(i int) int {
	return ((i + 1) / 4) - 1
}

// lchild returns the leftmost grandchild of node i (two levels down).
// A node has up to 4 grandchildren at indices lchild(i)..lchild(i)+3.
//
//	Level 0:     0                lchild(0) = ((0+1)*4)-1 = 3
//	Level 1:    1  2              lchild(1) = ((1+1)*4)-1 = 7
//	Level 2:   3 4 5 6
//	Level 3:  7 8 9 ...
func lchild(i int) int {
	return ((i + 1) * 4) - 1
}

// level returns the tree level of index i: floor(log2(i+1)).
// Level 0 is the root, level 1 is its children, etc.
// Computed via bit-length — a single CPU instruction on most architectures.
func level(i int) int {
	return bits.Len(uint(i)+1) - 1
}

// isMinHeap reports whether index i is on a min level (even level).
//
//	Level 0 (min):     0          isMinHeap(0) = true
//	Level 1 (max):    1  2        isMinHeap(1) = false
//	Level 2 (min):   3 4 5 6      isMinHeap(3) = true
func isMinHeap(i int) bool {
	return level(i)%2 == 0
}

// min4 finds the extremum among up to 4 consecutive elements starting at
// index i. When min=true it finds the smallest; when min=false the largest.
// Used to scan the grandchildren of a node during bubbledown — a node can
// have at most 4 grandchildren, stored contiguously at lchild(i)..lchild(i)+3.
//
//	       i
//	     /   \
//	   c0     c1        ← children (scanned by min2)
//	  / \    / \
//	g0  g1  g2  g3      ← grandchildren (scanned by min4)
func min4(h heap.Interface, l int, min bool, i int) int {
	q := i
	i++
	if i >= l {
		return q
	}
	if min == h.Less(i, q) {
		q = i
	}
	i++
	if i >= l {
		return q
	}
	if min == h.Less(i, q) {
		q = i
	}
	i++
	if i >= l {
		return q
	}
	if min == h.Less(i, q) {
		q = i
	}
	return q
}

// min2 finds the extremum among 2 consecutive elements at i and i+1.
// Used to scan the children of a node during bubbledown (a node has at
// most 2 children). Returns i if i+1 is out of bounds.
func min2(h heap.Interface, l int, min bool, i int) int {
	if i+1 < l && min == h.Less(i+1, i) {
		return i + 1
	}
	return i
}

// min3 finds the extremum among up to 3 elements at arbitrary indices
// i, j, k. Indices j or k may be out of bounds (>= l), in which case
// they are skipped. Used in bubbledown to pick the winner among the
// current node (i), its best child (j), and its best grandchild (k).
func min3(h heap.Interface, l int, min bool, i, j, k int) int {
	q := i
	if j < l && h.Less(j, q) == min {
		q = j
	}
	if k < l && h.Less(k, q) == min {
		q = k
	}
	return q
}

// bubbledown restores the heap property downward from index i.
// Called after a removal places a replacement element at position i.
//
// At each step it finds the best (smallest on min levels, largest on
// max levels) among three candidates: the node itself, its best child,
// and its best grandchild:
//
//	       i           ← current node
//	     /   \
//	   c0     c1       ← children (best picked by min2)
//	  / \    / \
//	g0  g1  g2  g3     ← grandchildren (best picked by min4)
//
// If a grandchild wins, the element moves down two levels and may
// also need a fixup swap with the child in between (which is on the
// opposite level type). This continues until the element is in place.
//
// Returns (q, r): the final positions of the element that was sifted
// and any secondary swap partner. Both are passed to bubbleup by Remove
// to handle the case where the replacement came from a distant part of
// the tree.
func bubbledown(h heap.Interface, l int, min bool, i int) (q int, r int) {
	q = i
	r = i
	for {
		// Best child of i (at most 2 children).
		j := min2(h, l, min, hlchild(i))
		if j >= l {
			break
		}
		// Best grandchild of i (at most 4 grandchildren).
		k := min4(h, l, min, lchild(i))
		// Pick the overall winner among {i, best child, best grandchild}.
		v := min3(h, l, min, i, j, k)
		if v == i || v >= l {
			break // i is already the best — done.
		}
		q = v
		h.Swap(v, i)
		if v == j {
			break // Winner was a child — one swap suffices.
		}
		// Winner was a grandchild. The grandchild's parent (on the
		// opposite level type) may now violate the heap property.
		// If so, swap to fix.
		p := hparent(v)
		if h.Less(p, v) == min {
			h.Swap(p, v)
			r = p
		}
		i = v
	}
	return q, r
}

// bubbleup restores the heap property upward from index i.
// Called after an insertion appends a new element at the end of the slice.
//
// The algorithm has two phases:
//
//  1. Grandparent chain (same level type): if the new element is better
//     than its grandparent, swap and repeat. This moves the element up
//     through nodes on the same level type (min→min or max→max).
//
//  2. Parent fixup (opposite level type): if phase 1 didn't move the
//     element, check the binary-tree parent. If the element is on a min
//     level but is LARGER than its max-level parent (or vice versa),
//     swap with the parent and then continue phase 1 from there on the
//     opposite level type.
//
//     0  (min)         Inserting a new min at index 9:
//     /   \              - Compare with grandparent (index 1, max): skip
//     1       2  (max)    - Compare with parent (index 4, max): swap if needed
//     / | \   / \           - Then compare up grandparent chain on max levels
//     3   4   5 6   (min)
//     /|
//     7  8  [9] ← new          Grandparent links: 9→1→ (root has no grandparent)
//     (max)                     Parent link: 9→4
func bubbleup(h heap.Interface, min bool, i int) (q bool) {
	if i < 0 {
		return false
	}
	j := parent(i)
	for j >= 0 && min == h.Less(i, j) {
		q = true
		h.Swap(i, j)
		i = j
		j = parent(i)
	}
	min = !min
	j = hparent(i)
	for j >= 0 && min == h.Less(i, j) {
		q = true
		h.Swap(i, j)
		i = j
		j = parent(i)
	}
	return q
}

// Pop removes and returns the smallest element from the heap.
// Returns nil if the heap is empty.
//
// The root (index 0) is always the minimum. To remove it, swap it with
// the last element, pop the last element off the slice, and bubbledown
// from the root to restore the heap property.
//
// Time complexity is O(log n), where n = h.Len().
func Pop(h heap.Interface) interface{} {
	if h.Len() == 0 {
		return nil
	}
	l := h.Len() - 1
	h.Swap(0, l)
	q := h.Pop()
	bubbledown(h, l, true, 0)
	return q
}

// PopMax removes and returns the largest element from the heap.
// Returns nil if the heap is empty.
//
// The maximum lives at index 1 or 2 (the root's children, both on
// max level 1). We find which child is larger, swap it with the last
// element, pop the last off the slice, and bubbledown from the vacated
// child position using max-level ordering.
//
//	          min: 1
//	            /   \
//	max: [9]     5       ← max is at index 1; swap it out
//
// Time complexity is O(log n), where n = h.Len().
func PopMax(h heap.Interface) interface{} {
	if h.Len() == 0 {
		return nil
	}
	l := h.Len()
	j := 0
	if l > 1 {
		j = min2(h, l, false, 1)
	}
	l = l - 1
	h.Swap(j, l)
	q := h.Pop()
	bubbledown(h, l, false, j)
	return q
}

// Remove removes and returns the element at index i from the heap.
//
// The element at i is swapped with the last element, the last element
// is popped off the slice, and then the replacement is sifted into
// place. Because the replacement can come from anywhere in the tree,
// both bubbledown AND bubbleup are needed to restore the invariant.
//
// The complexity is O(log n) where n = h.Len().
func Remove(h heap.Interface, i int) (q interface{}) {
	l := h.Len() - 1
	h.Swap(i, l)
	q = h.Pop()
	if l != i {
		q, r := bubbledown(h, l, isMinHeap(i), i)
		bubbleup(h, isMinHeap(q), q)
		bubbleup(h, isMinHeap(r), r)
	}
	return q
}

// Fix re-establishes the heap ordering after the element at index i
// has changed its value. Equivalent to, but cheaper than, Remove(h, i)
// followed by Push of the new value.
//
// The index i must be in the range [0, h.Len()).
// It panics if i is out of bounds.
//
// The complexity is O(log n) where n = h.Len().
func Fix(h heap.Interface, i int) {
	l := h.Len()
	// Sift down from i. At each cross-level fixup, immediately
	// bubbleup the displaced element so it is not lost when the
	// loop overwrites the position on the next iteration.
	min := isMinHeap(i)
	pos := i
	for {
		j := min2(h, l, min, hlchild(pos))
		if j >= l {
			break
		}
		k := min4(h, l, min, lchild(pos))
		v := min3(h, l, min, pos, j, k)
		if v == pos || v >= l {
			break
		}
		h.Swap(v, pos)
		if v == j {
			pos = v
			break
		}
		p := hparent(v)
		if (min && h.Less(p, v)) || (!min && h.Less(v, p)) {
			h.Swap(p, v)
			bubbleup(h, isMinHeap(p), p)
		}
		pos = v
	}
	// Fix upward from the final sift position (where the modified
	// element landed) and from the original position (which now
	// holds a descendant that may violate ancestor constraints).
	bubbleup(h, isMinHeap(pos), pos)
	if pos != i {
		bubbleup(h, isMinHeap(i), i)
	}
}

// Push adds element o to the heap, maintaining the min-max heap property.
//
// The element is appended to the end of the slice (the next open slot
// in the complete binary tree), then bubbled up to its correct position.
//
// Time complexity is O(log n), where n = h.Len().
func Push(h heap.Interface, o interface{}) {
	h.Push(o)
	l := h.Len()
	i := l - 1
	bubbleup(h, isMinHeap(i), i)
}

// valid reports whether h satisfies the min-max heap property.
//
// It checks each node against its binary parent (adjacent level type)
// and grandparent (same level type). These two local checks suffice:
// transitivity along grandparent chains establishes the global property.
//
// The scan is sequential over indices 1..l-1, making it cache-friendly.
func valid(h heap.Interface, l int) bool {
	for i := 1; i < l; i++ {
		hp := hparent(i)
		if isMinHeap(i) {
			// i on min level, hp on max level: hp must be ≥ i.
			if h.Less(hp, i) {
				return false
			}
		} else {
			// i on max level, hp on min level: i must be ≥ hp.
			if h.Less(i, hp) {
				return false
			}
		}
		if i >= 3 {
			gp := hparent(hp)
			if isMinHeap(i) {
				// Both min: gp must be ≤ i.
				if h.Less(i, gp) {
					return false
				}
			} else {
				// Both max: gp must be ≥ i.
				if h.Less(gp, i) {
					return false
				}
			}
		}
	}
	return true
}

// Verify reports whether h satisfies the min-max heap property.
//
// Time complexity is O(n), where n = h.Len().
func Verify(h heap.Interface) bool {
	return valid(h, h.Len())
}

// Init establishes the min-max heap ordering on an arbitrary slice.
// Call this once on a non-empty slice before calling Pop, PopMax, or Push.
//
// If the data already satisfies the heap property, Init returns after
// a linear scan with no modifications. Otherwise it uses Floyd's
// bottom-up heap construction, processing nodes from the last non-leaf
// down to the root. Most work happens near the leaves where subtrees
// are small and memory accesses are local.
//
// Time complexity is O(n), where n = h.Len().
func Init(h heap.Interface) {
	l := h.Len()
	if valid(h, l) {
		return
	}
	for i := (l - 1) / 2; i >= 0; i-- {
		bubbledown(h, l, isMinHeap(i), i)
	}
}
