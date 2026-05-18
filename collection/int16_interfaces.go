
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
	// Size returns the number of elements.
	Size() int

	// IsEmpty returns true if the collection contains no elements.
	IsEmpty() bool
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
// Satisfied by: Int16ArrayList, Int16HashSet, Int16HashBag,
// Int16ArrayStack, Int16TreeSet, and their immutable variants.
type Int16Collection interface {
	Int16Sized
	Int16Iterable
	Int16Searchable
	Int16Convertible
	fmt.Stringer
}

// Int16MutableCollection extends Int16Collection with mutation operations.
// Satisfied by: Int16ArrayList, Int16HashSet, Int16HashBag, Int16ArrayStack.
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
// access. Satisfied by: Int16ArrayList, ImmutableInt16ArrayList.
type Int16List interface {
	Int16Collection

	// Get returns the element at the given index, or an error if out of bounds.
	Get(index int) (int16, error)

	// IndexOf returns the index of the first occurrence of value, or -1 if absent.
	IndexOf(value int16) int
}

// Int16MutableList extends Int16List + Int16MutableCollection.
// Satisfied by: Int16ArrayList.
type Int16MutableList interface {
	Int16List
	Int16MutableCollection

	// Add appends a value to the end of the list.
	Add(value int16)

	// Set replaces the element at the given index. Returns the previous value or an error.
	Set(index int, value int16) (int16, error)
}

// Int16Set marker interface for set-like collections (uniqueness implied).
// Satisfied by: Int16HashSet, Int16TreeSet, ImmutableInt16HashSet.
type Int16Set interface {
	Int16Collection
}

// Int16MutableSet adds insertion. Add returns true if the value was newly inserted.
// Satisfied by: Int16HashSet, Int16TreeSet.
type Int16MutableSet interface {
	Int16Set
	Int16MutableCollection

	// Add inserts a value. Returns true if the value was not already present.
	Add(value int16) bool
}

// Int16Bag read-only multiset interface with occurrence counts.
// Satisfied by: Int16HashBag, Int16TreeBag, ImmutableInt16HashBag.
type Int16Bag interface {
	Int16Collection

	// OccurrencesOf returns the number of times value occurs in the bag.
	OccurrencesOf(value int16) int

	// SizeDistinct returns the number of *distinct* values (ignoring multiplicity).
	SizeDistinct() int
}

// Int16MutableBag adds insertion.
// Satisfied by: Int16HashBag, Int16TreeBag.
type Int16MutableBag interface {
	Int16Bag
	Int16MutableCollection

	// Add adds one occurrence of value.
	Add(value int16)
}

// Int16Stack read-only LIFO stack. Peek returns the top element or an error if empty.
// Satisfied by: Int16ArrayStack, ImmutableInt16ArrayStack.
type Int16Stack interface {
	Int16Collection

	// Peek returns the top of the stack without removing it.
	Peek() (int16, error)
}

// Int16MutableStack mutable LIFO stack with Push and Pop.
// Satisfied by: Int16ArrayStack.
type Int16MutableStack interface {
	Int16Stack
	Int16MutableCollection

	// Push pushes a value onto the top of the stack.
	Push(value int16)

	// Pop pops and returns the top of the stack, or an error if empty.
	Pop() (int16, error)
}
