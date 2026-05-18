
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

// Float32Sized exposes the element count of a collection.
type Float32Sized interface {
	// Size returns the number of elements.
	Size() int

	// IsEmpty returns true if the collection contains no elements.
	IsEmpty() bool
}

// Float32Iterable provides element-by-element traversal.
type Float32Iterable interface {
	// All returns an iter.Seq that yields all elements.
	All() iter.Seq[float32]

	// ForEach calls the given function for each element.
	ForEach(f func(float32))
}

// Float32Searchable supports membership and predicate queries.
type Float32Searchable interface {
	// Contains returns true if the collection contains the given value.
	Contains(value float32) bool

	// AnySatisfy returns true if any element satisfies the predicate.
	AnySatisfy(predicate func(float32) bool) bool

	// AllSatisfy returns true if all elements satisfy the predicate.
	AllSatisfy(predicate func(float32) bool) bool

	// NoneSatisfy returns true if no element satisfies the predicate.
	NoneSatisfy(predicate func(float32) bool) bool
}

// Float32Convertible supports bulk conversion to a slice.
type Float32Convertible interface {
	// ToSlice returns all elements as a slice.
	ToSlice() []float32
}

// ── Composed collection interfaces ────────────────────────────────────

// Float32Collection is the full read-only interface for any collection of
// float32 values.  It composes the smaller sub-interfaces above, so a
// caller that only needs iteration can accept Float32Iterable while code
// that needs everything can accept Float32Collection.
//
// Satisfied by: Float32ArrayList, Float32HashSet, Float32HashBag,
// Float32ArrayStack, Float32TreeSet, and their immutable variants.
type Float32Collection interface {
	Float32Sized
	Float32Iterable
	Float32Searchable
	Float32Convertible
	fmt.Stringer
}

// Float32MutableCollection extends Float32Collection with mutation operations.
// Satisfied by: Float32ArrayList, Float32HashSet, Float32HashBag, Float32ArrayStack.
type Float32MutableCollection interface {
	Float32Collection

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

// Float32List is the read-only interface for ordered lists with positional
// access. Satisfied by: Float32ArrayList, ImmutableFloat32ArrayList.
type Float32List interface {
	Float32Collection

	// Get returns the element at the given index, or an error if out of bounds.
	Get(index int) (float32, error)

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value float32) int
}

// Float32MutableList extends Float32List + Float32MutableCollection.
// Satisfied by: Float32ArrayList.
type Float32MutableList interface {
	Float32List
	Float32MutableCollection

	// Add appends a value to the end of the list.
	Add(value float32)

	// Set replaces the element at the given index. Returns the previous value or an error.
	Set(index int, value float32) (float32, error)
}

// Float32Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: Float32HashSet, Float32TreeSet, ImmutableFloat32HashSet.
type Float32Set interface {
	Float32Collection
}

// Float32MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: Float32HashSet, Float32TreeSet.
type Float32MutableSet interface {
	Float32Set
	Float32MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value float32) bool
}

// Float32Bag read-only multiset interface with occurrence counts.
// Satisfied by: Float32HashBag, Float32TreeBag, ImmutableFloat32HashBag.
type Float32Bag interface {
	Float32Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value float32) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// Float32MutableBag adds insertion.
// Satisfied by: Float32HashBag, Float32TreeBag.
type Float32MutableBag interface {
	Float32Bag
	Float32MutableCollection

	// Add adds one occurrence of value.
	Add(value float32)
}

// Float32Stack read-only LIFO stack. Peek returns the top element or an error if empty.
// Satisfied by: Float32ArrayStack, ImmutableFloat32ArrayStack.
type Float32Stack interface {
	Float32Collection

	// Peek returns the top of the stack without removing it.
	Peek() (float32, error)
}

// Float32MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: Float32ArrayStack.
type Float32MutableStack interface {
	Float32Stack
	Float32MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value float32)

	// Pop pops and returns the top of the stack, or an error if empty.
	Pop() (float32, error)
}
