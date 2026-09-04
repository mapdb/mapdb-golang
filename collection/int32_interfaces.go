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

// Int32Sized exposes the element count of a collection.
type Int32Sized interface {
	// Len returns the number of elements. Use x.Len() == 0 to test for
	// emptiness.
	Len() int
}

// Int32Iterable provides element-by-element traversal.
type Int32Iterable interface {
	// All returns an iter.Seq that yields all elements.
	All() iter.Seq[int32]

	// ForEach calls the given function for each element.
	ForEach(f func(int32))
}

// Int32Searchable supports membership and predicate queries.
type Int32Searchable interface {
	// Contains returns true if the collection contains the given value.
	Contains(value int32) bool

	// AnySatisfy returns true if any element satisfies the predicate.
	AnySatisfy(predicate func(int32) bool) bool

	// AllSatisfy returns true if all elements satisfy the predicate.
	AllSatisfy(predicate func(int32) bool) bool

	// NoneSatisfy returns true if no element satisfies the predicate.
	NoneSatisfy(predicate func(int32) bool) bool
}

// Int32Convertible supports bulk conversion to a slice.
type Int32Convertible interface {
	// ToSlice returns all elements as a slice.
	ToSlice() []int32
}

// ── Composed collection interfaces ────────────────────────────────────

// Int32Collection is the full read-only interface for any collection of
// int32 values.  It composes the smaller sub-interfaces above, so a
// caller that only needs iteration can accept Int32Iterable while code
// that needs everything can accept Int32Collection.
//
// Satisfied by: arraylist.Int32, hashset.Int32, bag.HashInt32,
// bag.TreeInt32, stack.Int32, treeset.Int32, and their immutable
// variants.
type Int32Collection interface {
	Int32Sized
	Int32Iterable
	Int32Searchable
	Int32Convertible
	fmt.Stringer
}

// Int32MutableCollection extends Int32Collection with mutation operations.
// Satisfied by: arraylist.Int32, hashset.Int32, bag.HashInt32, stack.Int32.
type Int32MutableCollection interface {
	Int32Collection

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

// Int32List is the read-only interface for ordered lists with positional
// access. Satisfied by: arraylist.Int32, arraylist.ImmutableInt32.
type Int32List interface {
	Int32Collection

	// Get returns the element at the given index. It panics on out-of-range index.
	Get(index int) int32

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value int32) int
}

// Int32MutableList extends Int32List + Int32MutableCollection.
// Satisfied by: arraylist.Int32.
type Int32MutableList interface {
	Int32List
	Int32MutableCollection

	// Add appends a value to the end of the list.
	Add(value int32)

	// Set sets the value at the given index, returning the previous value.
	// It panics on out-of-range index.
	Set(index int, value int32) int32
}

// Int32Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: hashset.Int32, treeset.Int32, hashset.ImmutableInt32.
type Int32Set interface {
	Int32Collection
}

// Int32MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: hashset.Int32, treeset.Int32.
type Int32MutableSet interface {
	Int32Set
	Int32MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value int32) bool
}

// Int32Bag read-only multiset interface with occurrence counts.
// Satisfied by: bag.HashInt32, bag.TreeInt32, bag.ImmutableHashInt32.
type Int32Bag interface {
	Int32Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value int32) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// Int32MutableBag adds insertion.
// Satisfied by: bag.HashInt32, bag.TreeInt32.
type Int32MutableBag interface {
	Int32Bag
	Int32MutableCollection

	// Add adds one occurrence of value.
	Add(value int32)
}

// Int32Stack read-only LIFO stack. Peek returns the top element and false
// if the stack is empty (no error is returned).
// Satisfied by: stack.Int32, stack.ImmutableInt32.
type Int32Stack interface {
	Int32Collection

	// Peek returns the top element without removing it. The bool is false if empty.
	Peek() (int32, bool)
}

// Int32MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: stack.Int32.
type Int32MutableStack interface {
	Int32Stack
	Int32MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value int32)

	// Pop removes and returns the top element. The bool is false if empty.
	Pop() (int32, bool)
}
