// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package object provides generic (object-typed) collection interfaces and
// implementations for the mapdb-golang port.  These mirror the primitive
// collection interfaces in the collection/ package but use Go generics
// so they work with any element type.
//
// The interface hierarchy follows Go idioms: small, composable interfaces
// (like io.Reader / io.Writer) that can be embedded into larger contracts.
package object

import "iter"

// ── Composable sub-interfaces ─────────────────────────────────────────

// Sized exposes the element count of a collection.
// Use x.Len() == 0 to test for emptiness.
type Sized interface {
	Len() int
}

// Iterable provides element-by-element traversal.
type Iterable[T any] interface {
	All() iter.Seq[T]
	ForEach(f func(T))
}

// Searchable supports membership and predicate queries.
//
// T is unconstrained (any): membership semantics belong to the implementation,
// not the interface. A hash-backed type compares with ==; a tree-backed type
// uses its Comparator; a strategy-backed type uses its equality strategy. This is
// why TreeSet/TreeMultimap/strategy types — the ones with the richest APIs — can
// satisfy the hierarchy at all (11 §4).
type Searchable[T any] interface {
	Contains(value T) bool
	AnySatisfy(predicate func(T) bool) bool
	AllSatisfy(predicate func(T) bool) bool
	NoneSatisfy(predicate func(T) bool) bool
}

// Convertible supports bulk conversion to a slice.
type Convertible[T any] interface {
	ToSlice() []T
}

// ── Composed collection interfaces ────────────────────────────────────

// Collection is the full read-only interface for any collection.
type Collection[T any] interface {
	Sized
	Iterable[T]
	Searchable[T]
	Convertible[T]
}

// MutableCollection adds Clear to Collection.
type MutableCollection[T any] interface {
	Collection[T]
	Clear()
}

// ── Category interfaces ───────────────────────────────────────────────

// List is the read-only interface for ordered collections with positional access.
type List[T any] interface {
	Collection[T]
	Get(index int) T
	IndexOf(value T) int
}

// MutableList extends List with mutation.
type MutableList[T any] interface {
	List[T]
	MutableCollection[T]
	Add(value T)
	Set(index int, value T) T
}

// Set is the read-only interface for collections with unique elements.
type Set[T any] interface {
	Collection[T]
}

// MutableSet extends Set with insertion and removal.
type MutableSet[T any] interface {
	Set[T]
	MutableCollection[T]
	Add(value T) bool
	Remove(value T) bool
}

// Bag is the read-only interface for multisets (elements with occurrence counts).
type Bag[T any] interface {
	Collection[T]
	OccurrencesOf(value T) int
	SizeDistinct() int
}

// MutableBag extends Bag with insertion.
type MutableBag[T any] interface {
	Bag[T]
	MutableCollection[T]
	Add(value T)
}

// Stack is the read-only interface for LIFO stacks.
type Stack[T any] interface {
	Collection[T]
	Peek() (T, bool)
}

// MutableStack extends Stack with push/pop.
type MutableStack[T any] interface {
	Stack[T]
	MutableCollection[T]
	Push(value T)
	Pop() (T, bool)
}

// ── Map interfaces ────────────────────────────────────────────────────

// MapIterable is the read-only interface for any key-value map.
//
// K is unconstrained (any): key identity belongs to the implementation — a
// hash-backed map compares keys with == (and constrains K comparable itself), a
// tree-backed map uses its Comparator. Relaxing K here is what lets TreeMap/
// TreeMultimap/HashMultimap satisfy the map hierarchy (11 §4).
type MapIterable[K any, V any] interface {
	Get(key K) (V, bool)
	ContainsKey(key K) bool
	Len() int
	All() iter.Seq2[K, V]
	Keys() iter.Seq[K]
	Values() iter.Seq[V]
	ForEach(f func(K, V))
	AnySatisfy(predicate func(K, V) bool) bool
	AllSatisfy(predicate func(K, V) bool) bool
	NoneSatisfy(predicate func(K, V) bool) bool
}

// MutableMap extends MapIterable with mutation.
type MutableMap[K any, V any] interface {
	MapIterable[K, V]
	Put(key K, value V) (V, bool)
	Remove(key K) (V, bool)
	Clear()
}

// BiMap is a bidirectional map where both keys and values are unique.
type BiMap[K, V comparable] interface {
	MapIterable[K, V]
	GetInverse(value V) (K, bool)
	ContainsValue(value V) bool
}

// MutableBiMap extends BiMap with mutation.
type MutableBiMap[K, V comparable] interface {
	BiMap[K, V]
	// Put adds a key-value pair. If the value already exists under a different
	// key, that old key is removed (bijection invariant). Returns the old value
	// for the key if it existed.
	Put(key K, value V) (V, bool)
	Remove(key K) (V, bool)
	Clear()
}
