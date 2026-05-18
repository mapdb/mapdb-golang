
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

// Float64Sized exposes the element count of a collection.
type Float64Sized interface {
	// Size returns the number of elements.
	Size() int

	// IsEmpty returns true if the collection contains no elements.
	IsEmpty() bool
}

// Float64Iterable provides element-by-element traversal.
type Float64Iterable interface {
	// All returns an iter.Seq that yields all elements.
	All() iter.Seq[float64]

	// ForEach calls the given function for each element.
	ForEach(f func(float64))
}

// Float64Searchable supports membership and predicate queries.
type Float64Searchable interface {
	// Contains returns true if the collection contains the given value.
	Contains(value float64) bool

	// AnySatisfy returns true if any element satisfies the predicate.
	AnySatisfy(predicate func(float64) bool) bool

	// AllSatisfy returns true if all elements satisfy the predicate.
	AllSatisfy(predicate func(float64) bool) bool

	// NoneSatisfy returns true if no element satisfies the predicate.
	NoneSatisfy(predicate func(float64) bool) bool
}

// Float64Convertible supports bulk conversion to a slice.
type Float64Convertible interface {
	// ToSlice returns all elements as a slice.
	ToSlice() []float64
}

// ── Composed collection interfaces ────────────────────────────────────

// Float64Collection is the full read-only interface for any collection of
// float64 values.  It composes the smaller sub-interfaces above, so a
// caller that only needs iteration can accept Float64Iterable while code
// that needs everything can accept Float64Collection.
//
// Satisfied by: Float64ArrayList, Float64HashSet, Float64HashBag,
// Float64ArrayStack, Float64TreeSet, and their immutable variants.
type Float64Collection interface {
	Float64Sized
	Float64Iterable
	Float64Searchable
	Float64Convertible
	fmt.Stringer
}

// Float64MutableCollection extends Float64Collection with mutation operations.
// Satisfied by: Float64ArrayList, Float64HashSet, Float64HashBag, Float64ArrayStack.
type Float64MutableCollection interface {
	Float64Collection

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

// Float64List is the read-only interface for ordered lists with positional
// access. Satisfied by: Float64ArrayList, ImmutableFloat64ArrayList.
type Float64List interface {
	Float64Collection

	// Get returns the element at the given index, or an error if out of bounds.
	Get(index int) (float64, error)

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value float64) int
}

// Float64MutableList extends Float64List + Float64MutableCollection.
// Satisfied by: Float64ArrayList.
type Float64MutableList interface {
	Float64List
	Float64MutableCollection

	// Add appends a value to the end of the list.
	Add(value float64)

	// Set replaces the element at the given index. Returns the previous value or an error.
	Set(index int, value float64) (float64, error)
}

// Float64Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: Float64HashSet, Float64TreeSet, ImmutableFloat64HashSet.
type Float64Set interface {
	Float64Collection
}

// Float64MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: Float64HashSet, Float64TreeSet.
type Float64MutableSet interface {
	Float64Set
	Float64MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value float64) bool
}

// Float64Bag read-only multiset interface with occurrence counts.
// Satisfied by: Float64HashBag, Float64TreeBag, ImmutableFloat64HashBag.
type Float64Bag interface {
	Float64Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value float64) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// Float64MutableBag adds insertion.
// Satisfied by: Float64HashBag, Float64TreeBag.
type Float64MutableBag interface {
	Float64Bag
	Float64MutableCollection

	// Add adds one occurrence of value.
	Add(value float64)
}

// Float64Stack read-only LIFO stack. Peek returns the top element or an error if empty.
// Satisfied by: Float64ArrayStack, ImmutableFloat64ArrayStack.
type Float64Stack interface {
	Float64Collection

	// Peek returns the top of the stack without removing it.
	Peek() (float64, error)
}

// Float64MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: Float64ArrayStack.
type Float64MutableStack interface {
	Float64Stack
	Float64MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value float64)

	// Pop pops and returns the top of the stack, or an error if empty.
	Pop() (float64, error)
}
