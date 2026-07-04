// Package pump holds the shared types and sentinel errors for the data-pump
// (bulk-import) feature: the duplicate policy, the idiomatic error model, the
// red-black build coloring math, and the hash-presize formula. The per-primitive
// bulk-load constructors and Sink builders across the collection families
// reference these, so the contract is defined in exactly one place. The package
// depends on no collection package, so every family can import it without an
// import cycle.
package pump

import (
	"errors"
	"math/bits"
)

// DuplicatePolicy controls how a bulk-load handles a duplicate key in the
// source. There is deliberately no Overwrite ("last value wins"): in a
// single-pass, insert-only build last-wins is a read-modify-write and
// contradicts the pump's reason to exist. Callers wanting last-wins should use
// the ordinary Put loop instead.
//
// The policy does not apply to bag value multiplicity: bags count duplicate
// runs regardless of the policy.
type DuplicatePolicy int

const (
	// ErrorOnDuplicate fails the build on the first duplicate key.
	ErrorOnDuplicate DuplicatePolicy = iota
	// IgnoreDuplicates keeps the first occurrence of a key and skips the rest.
	IgnoreDuplicates
)

// ErrNotSorted is returned by an ordered bulk-load (FromSorted / a tree or
// multimap Sink) when the source is not in ascending order according to the
// collection's own comparator (the IEEE-754 total order for float keys). Use
// errors.Is to test for it.
var ErrNotSorted = errors.New("mapdb: bulk-load input not in ascending key order")

// ErrDuplicateKey is returned by a bulk-load with ErrorOnDuplicate when the
// source contains a duplicate key. Use errors.Is to test for it.
var ErrDuplicateKey = errors.New("mapdb: duplicate key in bulk-load input")

// ErrDuplicateValue is returned by a BiMap bulk-load when the source contains a
// duplicate value (a BiMap requires a bijection, so a repeated value breaks the
// inverse mapping). Use errors.Is to test for it.
var ErrDuplicateValue = errors.New("mapdb: duplicate value in bulk-load input")

// ErrTooManyElements is returned by a BulkLoadExact constructor when the source
// yields more than the promised n elements; honouring it would force a rehash
// and break the zero-rehash guarantee.
var ErrTooManyElements = errors.New("mapdb: bulk-load source exceeds the exact size")

// HashCapacityFor returns the open-addressing table capacity that fits n
// elements at the 0.75 load factor with zero mid-load rehash, using the strict
// growth rule cap*3 >= 4*n + 1. It is the single source of truth for hash
// presizing across the generated families.
//
//	required = floor(4*n/3) + 1     // == ceil((4n+1)/3); overflow-safe
//	cap      = nextPow2(required)   // n == 0 -> 0, the empty-table sentinel
//
// required is computed in unsigned-64 arithmetic via a division-decompose
// (floor(4n/3) = 4*(n/3) + floor(4*(n%3)/3)), so the intermediate 4q never
// overflows even for n near MaxInt. n so large that no power-of-two capacity
// fits in a positive int (required would exceed the top representable power of
// two) is a programmer error and panics rather than silently wrapping.
func HashCapacityFor(n int) int {
	if n <= 0 {
		return 0
	}
	// required = floor(4n/3) + 1, computed in uint64 so 4q cannot overflow.
	q := uint64(n) / 3
	r := uint64(n) % 3
	required := q*4 + (r*4)/3 + 1
	// nextPow2 must stay within a positive int. The largest power of two that
	// fits is 1<<(bits.UintSize-2) (e.g. 1<<62 on 64-bit). If required exceeds
	// it, no valid capacity exists.
	maxPow2 := uint64(1) << (bits.UintSize - 2)
	if required > maxPow2 {
		panic("mapdb: HashCapacityFor: n too large for any hash-table capacity")
	}
	return nextPow2Int(int(required))
}

// RedBlackRedLevel returns the tree depth (0 == root) at which nodes must be
// coloured red when building a perfectly balanced binary search tree from n
// sorted elements via the classic JDK buildFromSorted construction. Every node
// above this level is black; the nodes on this single deepest, possibly
// incomplete level are red. This yields a valid red-black tree (root black, no
// two consecutive reds, uniform black-height), so subsequent insert/remove keep
// the invariant. n must be >= 0; n == 0 returns 0 (no nodes to colour).
//
// This is the per-tree-type coloring math factored out of the generated
// recursive builders so the only matrix-multiplied code is the trivial
// type-specific recursion.
func RedBlackRedLevel(n int) int {
	level := 0
	for m := n - 1; m >= 0; m = m/2 - 1 {
		level++
	}
	return level
}

// nextPow2Int rounds n up to the next power of two. n must be > 0 (n <= 1
// returns 1 — a different floor from internal/bits.NextPowerOfTwo, which floors
// at the 16-slot default hash capacity). Width-defined shifts, so the >>32 step
// is a no-op on 32-bit and required on 64-bit.
func nextPow2Int(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}
