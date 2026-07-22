# deheap

[![Go Reference](https://pkg.go.dev/badge/github.com/aalpar/deheap.svg)](https://pkg.go.dev/github.com/aalpar/deheap)
[![Go Report Card](https://goreportcard.com/badge/github.com/aalpar/deheap)](https://goreportcard.com/report/github.com/aalpar/deheap)
[![codecov](https://codecov.io/gh/aalpar/deheap/graph/badge.svg)](https://codecov.io/gh/aalpar/deheap)

A doubly-ended heap (min-max heap) for Go. Provides O(log n) access to both
the minimum and maximum elements of a collection through a single data
structure, with zero external dependencies.

```
go get github.com/aalpar/deheap
```

## Why a doubly-ended heap?

A standard binary heap gives you efficient access to one extremum — the
smallest or the largest element — but not both. Many practical problems need
both ends simultaneously:

**Scheduling and resource allocation.** Operating system schedulers and job
queues routinely need the highest-priority task (to run next) and the
lowest-priority task (to evict or age). A doubly-ended priority queue avoids
maintaining two separate heaps and the bookkeeping to keep them synchronized.

**Bounded-size caches and buffers.** When a priority queue has a capacity
limit, insertions must discard the least valuable element. With a min-max
heap, both the insertion (against the max) and the eviction (from the min, or
vice versa) are logarithmic — no linear scan required.

**Median maintenance and order statistics.** Streaming median algorithms
typically partition data into a max-heap of the lower half and a min-heap of
the upper half. A single min-max heap can serve double duty, simplifying the
implementation.

**Network packet scheduling.** Rate-controlled and deadline-aware packet
schedulers (e.g., in QoS systems) need to dequeue by earliest deadline and
drop by lowest priority, both efficiently.

**Search algorithms.** Algorithms like SMA\* (Simplified Memory-Bounded A\*)
maintain an open set where the node with the lowest f-cost is expanded next
and the node with the highest f-cost is pruned when memory is exhausted.

## API

`deheap` provides two API surfaces.

### Generic API (Go 1.21+)

For `cmp.Ordered` types — `int`, `float64`, `string`, and friends — use the
type-safe generic API directly:

```go
import "github.com/aalpar/deheap"

// Construct from existing elements.
h := deheap.From(5, 3, 8, 1, 9)

// Or build incrementally.
h := deheap.New[int]()
h.Push(5)
h.Push(3)

// O(1) access to both extrema.
fmt.Println(h.Peek())    // smallest
fmt.Println(h.PeekMax()) // largest

// O(log n) removal from either end.
min := h.Pop()
max := h.PopMax()

// Remove by index.
val := h.Remove(2)

// Update an element and restore heap order.
h.Push(10)
h.Push(20)
// ... mutate the element at index 0 ...
h.Fix(0)

// Check if the heap is valid.
fmt.Println(h.Verify()) // true
```

### Interface API

For custom types, implement `heap.Interface` and use the package-level
functions. This is the original v1 API and remains stable.

```go
import "github.com/aalpar/deheap"

type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) { *h = append(*h, x.(int)) }

func (h *IntHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[:n-1]
    return x
}

h := &IntHeap{2, 1, 5, 6}
deheap.Init(h)
deheap.Push(h, 3)
min := deheap.Pop(h)
max := deheap.PopMax(h)

// Update an element and restore heap order.
(*h)[0] = 42
deheap.Fix(h, 0)

// Check if the heap is valid.
fmt.Println(deheap.Verify(h)) // true
```

## Implementation

The underlying data structure is a **min-max heap** stored in a flat slice
with no pointers and no additional bookkeeping. Nodes on even levels
(0, 2, 4, ...) satisfy the min-heap property with respect to their
descendants, and nodes on odd levels (1, 3, 5, ...) satisfy the max-heap
property. The root is always the minimum; the maximum is one of its two
children.

Level parity is determined by bit-length of the 1-based index, computed via
`math/bits.Len` — a single CPU instruction on most architectures. Insertions
bubble up through grandparent links; deletions bubble down through
grandchild links, with a secondary swap against the binary-tree parent when
the element crosses a level boundary.

The generic API implements the same algorithms directly on `[]T` with
native `<` comparisons, eliminating interface dispatch, adapter allocation,
and boxing overhead.

### Complexity

| Operation | Time     | Space |
|-----------|----------|-------|
| `Push`    | O(log n) | O(1) amortized |
| `Pop`     | O(log n) | O(1) |
| `PopMax`  | O(log n) | O(1) |
| `Remove`  | O(log n) | O(1) |
| `Fix`     | O(log n) | O(1) |
| `Peek`    | O(1)     | O(1) |
| `PeekMax` | O(1)     | O(1) |
| `Init`    | O(n)     | O(1) |
| `Verify`  | O(n)     | O(1) |

Storage is a single contiguous slice — one element per slot, no child
pointers, no color bits, no auxiliary arrays. Memory overhead beyond the
elements themselves is the slice header (24 bytes on 64-bit systems).

### Benchmarks

Measured on Apple M4 Max (arm64), Go 1.23:

**Generic API** (`Deheap[T]` — direct `<` comparisons, no interface dispatch):

| Operation   | ns/op  | B/op | allocs/op |
|-------------|--------|------|-----------|
| Push        | 12     | 45   | 0         |
| Pop         | 225    | 0    | 0         |
| PopMax      | 225    | 0    | 0         |

**Interface API** (v1 — `heap.Interface`):

| Operation   | ns/op  | B/op | allocs/op |
|-------------|--------|------|-----------|
| Push        | 21     | 54   | 0         |
| Pop         | 288    | 7    | 0         |
| PopMax      | 282    | 7    | 0         |

**Standard library** (`container/heap`) for comparison:

| Operation   | ns/op  | B/op | allocs/op |
|-------------|--------|------|-----------|
| Push        | 22     | 55   | 0         |
| Pop         | 208    | 7    | 0         |

The generic API is faster than `container/heap` on Push. Pop is ~8% slower
due to examining up to four grandchildren per node (vs two children in a
binary heap), but provides O(log n) access to both extrema. The v1 interface
API pays the same `heap.Interface` dispatch cost as `container/heap`. All
operations are zero-allocation in steady state.

#### Benchmark descriptions

The benchmarks compare deheap's two API surfaces against each other and
against `container/heap`. Baseline measurements isolate the cost of slice
operations from heap logic. All benchmarks use a heap of 10,000 `int`
elements to ensure the working set fits in L1 cache while exercising the
full tree depth.

Sample run (Apple M4 Max, Go 1.23):

```
$ go test -bench . -benchmem
goos: darwin
goarch: arm64
pkg: github.com/aalpar/deheap
cpu: Apple M4 Max
BenchmarkMin4-16              	261428533	         4.784 ns/op	       0 B/op	       0 allocs/op
BenchmarkBaselinePush-16      	1000000000	         6.411 ns/op	      42 B/op	       0 allocs/op
BenchmarkPush-16              	44383689	        23.16 ns/op	      50 B/op	       0 allocs/op
BenchmarkPop-16               	 5666707	       326.3 ns/op	       7 B/op	       0 allocs/op
BenchmarkPopMax-16            	 5644575	       319.7 ns/op	       7 B/op	       0 allocs/op
BenchmarkPushPop-16           	 5098502	       297.2 ns/op	      65 B/op	       1 allocs/op
BenchmarkHeapPushPop-16       	 6203408	       256.5 ns/op	      56 B/op	       1 allocs/op
BenchmarkHeapPop-16           	 8475319	       234.7 ns/op	       7 B/op	       0 allocs/op
BenchmarkHeapPush-16          	46837188	        21.57 ns/op	      48 B/op	       0 allocs/op
BenchmarkOrderedPush-16       	100000000	        12.38 ns/op	      45 B/op	       0 allocs/op
BenchmarkOrderedPop-16        	 8680756	       249.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkOrderedPopMax-16     	10103798	       237.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkOrderedPushPop-16    	 8926204	       233.2 ns/op	      44 B/op	       0 allocs/op
PASS
```

| Benchmark | What it measures |
|-----------|-----------------|
| `OrderedPush` | Generic `Deheap[int].Push`: append + bubble-up with direct `<` comparisons. |
| `OrderedPop` | Generic `Deheap[int].Pop`: remove minimum and bubble-down with direct `<`. |
| `OrderedPopMax` | Generic `Deheap[int].PopMax`: remove maximum and bubble-down with direct `<`. |
| `OrderedPushPop` | Generic push-then-drain throughput. |
| `Min4` | Cost of `min4`, the internal function that finds the extremum among up to 4 grandchildren during bubble-down. |
| `BaselinePush` | Raw slice append with no heap ordering — establishes the floor cost of memory allocation and copying. |
| `Push` | `deheap.Push` (v1): append + bubble-up via `heap.Interface`. |
| `Pop` | `deheap.Pop` (v1): remove the minimum element and bubble-down via `heap.Interface`. |
| `PopMax` | `deheap.PopMax` (v1): remove the maximum element and bubble-down via `heap.Interface`. |
| `PushPop` | v1 push-then-drain throughput. |
| `HeapPushPop` | Same push-then-drain pattern using `container/heap` for direct comparison. |
| `HeapPop` | `container/heap.Pop` in isolation, comparable to `Pop` above. |
| `HeapPush` | `container/heap.Push` in isolation, comparable to `Push` above. |

## Testing

The test suite includes 54 test functions covering internal helpers,
algorithmic correctness, edge cases (empty, single-element, two-element
heaps), and large-scale randomized validation. Six native Go fuzz targets
(`testing.F`) exercise both API surfaces under arbitrary input. Tests are run
against Go 1.21, 1.22, and 1.23 in CI.

```bash
go test ./...          # run all tests
go test -bench . ./... # run benchmarks
go test -fuzz .        # run fuzz tests
```

## Requirements

- Go 1.21 or later (generic API)
- Go 1.13 or later (interface API only)
- Zero external dependencies

## References

1. M.D. Atkinson, J.-R. Sack, N. Santoro, and T. Strothotte. "Min-Max
   Heaps and Generalized Priority Queues." *Communications of the ACM*,
   29(10):996–1000, October 1986.
   https://doi.org/10.1145/6617.6621

2. J. van Leeuwen and D. Wood. "Interval Heaps." *The Computer Journal*,
   36(3):209–216, 1993.

3. P. Brass. *Advanced Data Structures*. Cambridge University Press, 2008.

## License

MIT — see [LICENSE](LICENSE).
