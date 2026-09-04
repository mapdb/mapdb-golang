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

// Int16Sized exposes the element count of a collection.
type Int16Sized interface {
	// Len returns the number of elements. Use x.Len() == 0 to test for
	// emptiness.
	Len() int
}

// Int16Iterable provides element-by-element traversal.
type Int16Iterable interface {
	// All returns an iter.Seq that yields all elements.
	All() iter.Seq[int16]

	// ForEach calls the given function for each element.
	ForEach(f func(int16))
}

// Int16Searchable supports membership and predicate queries.
type Int16Searchable interface {
	// Contains returns true if the collection contains the given value.
	Contains(value int16) bool

	// AnySatisfy returns true if any element satisfies the predicate.
	AnySatisfy(predicate func(int16) bool) bool

	// AllSatisfy returns true if all elements satisfy the predicate.
	AllSatisfy(predicate func(int16) bool) bool

	// NoneSatisfy returns true if no element satisfies the predicate.
	NoneSatisfy(predicate func(int16) bool) bool
}

// Int16Convertible supports bulk conversion to a slice.
type Int16Convertible interface {
	// ToSlice returns all elements as a slice.
	ToSlice() []int16
}

// ── Composed collection interfaces ────────────────────────────────────

// Int16Collection is the full read-only interface for any collection of
// int16 values.  It composes the smaller sub-interfaces above, so a
// caller that only needs iteration can accept Int16Iterable while code
// that needs everything can accept Int16Collection.
//
// Satisfied by: arraylist.Int16, hashset.Int16, bag.HashInt16,
// bag.TreeInt16, stack.Int16, treeset.Int16, and their immutable
// variants.
type Int16Collection interface {
	Int16Sized
	Int16Iterable
	Int16Searchable
	Int16Convertible
	fmt.Stringer
}

// Int16MutableCollection extends Int16Collection with mutation operations.
// Satisfied by: arraylist.Int16, hashset.Int16, bag.HashInt16, stack.Int16.
type Int16MutableCollection interface {
	Int16Collection

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

// Int16List is the read-only interface for ordered lists with positional
// access. Satisfied by: arraylist.Int16, arraylist.ImmutableInt16.
type Int16List interface {
	Int16Collection

	// Get returns the element at the given index. It panics on out-of-range index.
	Get(index int) int16

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value int16) int
}

// Int16MutableList extends Int16List + Int16MutableCollection.
// Satisfied by: arraylist.Int16.
type Int16MutableList interface {
	Int16List
	Int16MutableCollection

	// Add appends a value to the end of the list.
	Add(value int16)

	// Set sets the value at the given index, returning the previous value.
	// It panics on out-of-range index.
	Set(index int, value int16) int16
}

// Int16Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: hashset.Int16, treeset.Int16, hashset.ImmutableInt16.
type Int16Set interface {
	Int16Collection
}

// Int16MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: hashset.Int16, treeset.Int16.
type Int16MutableSet interface {
	Int16Set
	Int16MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value int16) bool
}

// Int16Bag read-only multiset interface with occurrence counts.
// Satisfied by: bag.HashInt16, bag.TreeInt16, bag.ImmutableHashInt16.
type Int16Bag interface {
	Int16Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value int16) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// Int16MutableBag adds insertion.
// Satisfied by: bag.HashInt16, bag.TreeInt16.
type Int16MutableBag interface {
	Int16Bag
	Int16MutableCollection

	// Add adds one occurrence of value.
	Add(value int16)
}

// Int16Stack read-only LIFO stack. Peek returns the top element and false
// if the stack is empty (no error is returned).
// Satisfied by: stack.Int16, stack.ImmutableInt16.
type Int16Stack interface {
	Int16Collection

	// Peek returns the top element without removing it. The bool is false if empty.
	Peek() (int16, bool)
}

// Int16MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: stack.Int16.
type Int16MutableStack interface {
	Int16Stack
	Int16MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value int16)

	// Pop removes and returns the top element. The bool is false if empty.
	Pop() (int16, bool)
}
