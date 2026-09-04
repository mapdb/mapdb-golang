package collection

import (
	"fmt"
	"iter"
)

// ── Composable sub-interfaces ─────────────────────────────────────────
//
// Following Go's io.Reader / io.Writer / io.ReadWriter pattern, the
// primitive-collection API is built from small, single-concern
// interfaces that can be composed as needed.

// Int64Sized exposes the element count of a collection.
type Int64Sized interface {
	// Len returns the number of elements. Use x.Len() == 0 to test for
	// emptiness.
	Len() int
}

// Int64Iterable provides element-by-element traversal.
type Int64Iterable interface {
	// All returns an iter.Seq that yields all elements.
	All() iter.Seq[int64]

	// ForEach calls the given function for each element.
	ForEach(f func(int64))
}

// Int64Searchable supports membership and predicate queries.
type Int64Searchable interface {
	// Contains returns true if the collection contains the given value.
	Contains(value int64) bool

	// AnySatisfy returns true if any element satisfies the predicate.
	AnySatisfy(predicate func(int64) bool) bool

	// AllSatisfy returns true if all elements satisfy the predicate.
	AllSatisfy(predicate func(int64) bool) bool

	// NoneSatisfy returns true if no element satisfies the predicate.
	NoneSatisfy(predicate func(int64) bool) bool
}

// Int64Convertible supports bulk conversion to a slice.
type Int64Convertible interface {
	// ToSlice returns all elements as a slice.
	ToSlice() []int64
}

// ── Composed collection interfaces ────────────────────────────────────

// Int64Collection is the full read-only interface for any collection of
// int64 values.  It composes the smaller sub-interfaces above, so a
// caller that only needs iteration can accept Int64Iterable while code
// that needs everything can accept Int64Collection.
//
// Satisfied by: arraylist.Int64, hashset.Int64, bag.HashInt64,
// bag.TreeInt64, stack.Int64, treeset.Int64, and their immutable
// variants.
type Int64Collection interface {
	Int64Sized
	Int64Iterable
	Int64Searchable
	Int64Convertible
	fmt.Stringer
}

// Int64MutableCollection extends Int64Collection with mutation operations.
// Satisfied by: arraylist.Int64, hashset.Int64, bag.HashInt64, stack.Int64.
type Int64MutableCollection interface {
	Int64Collection

	// Clear removes all elements.
	Clear()
}

// ── Category interfaces ────────────────────────────────────────────────
//
// These distinguish *what kind of collection* is required without naming
// a concrete type. They mirror Java's IntList / IntSet / IntBag / IntStack
// hierarchy and the matching trait/comptime layers in Rust and Zig.
//
// Note: each category uses its own Add() method shape to match Go's
// type-specific signatures (lists Add() void; sets Add() bool; bags
// Add() int returning new occurrence count; stacks use Push instead).

// Int64List is the read-only interface for ordered lists with positional
// access. Satisfied by: arraylist.Int64, arraylist.ImmutableInt64.
type Int64List interface {
	Int64Collection

	// Get returns the element at the given index. It panics on out-of-range index.
	Get(index int) int64

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value int64) int
}

// Int64MutableList extends Int64List + Int64MutableCollection.
// Satisfied by: arraylist.Int64.
type Int64MutableList interface {
	Int64List
	Int64MutableCollection

	// Add appends a value to the end of the list.
	Add(value int64)

	// Set sets the value at the given index, returning the previous value.
	// It panics on out-of-range index.
	Set(index int, value int64) int64
}

// Int64Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: hashset.Int64, treeset.Int64, hashset.ImmutableInt64.
type Int64Set interface {
	Int64Collection
}

// Int64MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: hashset.Int64, treeset.Int64.
type Int64MutableSet interface {
	Int64Set
	Int64MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value int64) bool
}

// Int64Bag read-only multiset interface with occurrence counts.
// Satisfied by: bag.HashInt64, bag.TreeInt64, bag.ImmutableHashInt64.
type Int64Bag interface {
	Int64Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value int64) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// Int64MutableBag adds insertion.
// Satisfied by: bag.HashInt64, bag.TreeInt64.
type Int64MutableBag interface {
	Int64Bag
	Int64MutableCollection

	// Add adds one occurrence of value.
	Add(value int64)
}

// Int64Stack read-only LIFO stack. Peek returns the top element and false
// if the stack is empty (no error is returned).
// Satisfied by: stack.Int64, stack.ImmutableInt64.
type Int64Stack interface {
	Int64Collection

	// Peek returns the top element without removing it. The bool is false if empty.
	Peek() (int64, bool)
}

// Int64MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: stack.Int64.
type Int64MutableStack interface {
	Int64Stack
	Int64MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value int64)

	// Pop removes and returns the top element. The bool is false if empty.
	Pop() (int64, bool)
}
