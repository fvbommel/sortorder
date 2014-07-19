package rope

import (
	"fmt"
	"sync"
)

// Debug enables debug output.
var Debug = true

var (
	// A cache of Fibonacci numbers.
	// Initialized to some initial ones to get us started.
	// Note: duplicate 1 is omitted.
	fibCache = []int64{
		1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610, 987,
		1597, 2584, 4181, 6765, 10946, 17711, 28657, 46368, 75025, 121393,
		196418, 317811, 514229, 832040,
	}

	// A lock to hold while accessing fibCache.
	fibLock sync.RWMutex
)

// reverseFib returns the index of the smallest element >= N in a sorted array.
// If the array is empty, 0 is returned. If the array's largest element is <= N,
// len(arr) is returned.
// If multiple array values are equal, any one of them may be returned.
func binSearch(N int64, arr []int64) int {
	// Assuming "virtual" elements -1 and len(arr) which are respectively
	// 1 smaller or larger than the first or last element actually in the array:
	// invariant: arr[start] < N < arr[end]
	start, end := -1, len(arr)
	for end-start > 1 {
		mid := (start + end) / 2
		elt := arr[mid]
		switch {
		case elt < N:
			start = mid
		case N < elt:
			end = mid
		default: // Exactly equal
			return mid
		}
	}
	// start + 1 == end, so the invariant gives us: arr[end-1] < N < arr[end]
	return end
}

// reverseFib returns the index of the smallest Fibonacci number >= N.
func reverseFib(N int64) int {
	return binSearch(N, getFibCache(N))
}

// getFibCache gets a reference to the current fibCache, extended to at least
// cover N. This reference is safe to use without locking since the only
// modification to fibCache is appending, which doesn't affect previous slices.
func getFibCache(N int64) []int64 {
	fibLock.RLock()
	defer fibLock.RUnlock()

	if fibCache[len(fibCache)] < N {
		// Calculate some more numbers
		extendFibs(N)
	}
	return fibCache
}

// extendFibs extends fibCache until the largest number is >= N.
// Precondition: holding a Read-lock on fibLock.
func extendFibs(N int64) {
	// Get a write lock, and make sure we reaquire a read lock on exit.
	fibLock.RUnlock()
	fibLock.Lock()
	defer func() {
		fibLock.Unlock()
		fibLock.RLock()
	}()

	// Note that if we lose the race for a write lock, the loop does not run.
	a, b := fibCache[len(fibCache)-2], fibCache[len(fibCache)-1]
	for b < N {
		a, b = b, a+b
		fibCache = append(fibCache, b)
	}
}

// A rope is balanced if its length <= fib(depth).
func (r Rope) isBalanced() bool {
	if r.node == nil || r.node == emptyNode {
		return true
	}

	len := r.Len()
	maxDepth := reverseFib(len) - 2
	return int(r.node.depth()) <= maxDepth
}

// Rebalance rebalances a rope.
func (r Rope) Rebalance() Rope {
	len := r.Len()
	fibs := getFibCache(len)

	// The concatenation of non-empty nodes in the scratch array, in order of
	// decreasing index, is equivalent to the concatenation of leaves walked
	// so far.
	type B struct {
		n   node
		len int64
	}
	scratch := make([]B, 1+binSearch(len, fibs))
	r.node.walkLeaves(func(l leaf) {
		nLen := l.length()
		bucket := binSearch(nLen, fibs)

		n := node(l)
		for i := bucket; i >= 0; i-- {
			b := scratch[bucket]
			if b.n != nil {
				n = conc(n, b.n, nLen, b.len)
				nLen += b.len
				scratch[bucket].n = nil // clear bucket
			}
		}
		if Debug {
			if b := binSearch(nLen, fibs); b != bucket {
				panic(fmt.Errorf("Bucket grew from %d to %d", bucket, b))
			}
		}
		scratch[bucket] = B{n, nLen}
	})
	nw := B{}
	for _, b := range scratch {
		if b.n != nil {
			if nw.n == nil {
				nw = b
			} else {
				nw.n = conc(b.n, nw.n, b.len, nw.len)
				nw.len += b.len
			}
		}
	}
	if nw.n == r.node {
		return r
	}
	return Rope{nw.n}
}
