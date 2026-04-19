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
type Sized interface {
	Size() int
	IsEmpty() bool
}

// Iterable provides element-by-element traversal.
type Iterable[T any] interface {
	All() iter.Seq[T]
	ForEach(f func(T))
}

// Searchable supports membership and predicate queries.
// T must be comparable so Contains can use ==.
type Searchable[T comparable] interface {
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
type Collection[T comparable] interface {
	Sized
	Iterable[T]
	Searchable[T]
	Convertible[T]
}

// MutableCollection adds Clear to Collection.
type MutableCollection[T comparable] interface {
	Collection[T]
	Clear()
}

// ── Category interfaces ───────────────────────────────────────────────

// List is the read-only interface for ordered collections with positional access.
type List[T comparable] interface {
	Collection[T]
	Get(index int) (T, error)
	IndexOf(value T) int
}

// MutableList extends List with mutation.
type MutableList[T comparable] interface {
	List[T]
	MutableCollection[T]
	Add(value T)
	Set(index int, value T) (T, error)
}

// Set is the read-only interface for collections with unique elements.
type Set[T comparable] interface {
	Collection[T]
}

// MutableSet extends Set with insertion and removal.
type MutableSet[T comparable] interface {
	Set[T]
	MutableCollection[T]
	Add(value T) bool
	Remove(value T) bool
}

// Bag is the read-only interface for multisets (elements with occurrence counts).
type Bag[T comparable] interface {
	Collection[T]
	OccurrencesOf(value T) int
	SizeDistinct() int
}

// MutableBag extends Bag with insertion.
type MutableBag[T comparable] interface {
	Bag[T]
	MutableCollection[T]
	Add(value T)
}

// Stack is the read-only interface for LIFO stacks.
type Stack[T comparable] interface {
	Collection[T]
	Peek() (T, error)
}

// MutableStack extends Stack with push/pop.
type MutableStack[T comparable] interface {
	Stack[T]
	MutableCollection[T]
	Push(value T)
	Pop() (T, error)
}

// ── Map interfaces ────────────────────────────────────────────────────

// MapIterable is the read-only interface for any key-value map.
type MapIterable[K comparable, V any] interface {
	Get(key K) (V, bool)
	ContainsKey(key K) bool
	Size() int
	IsEmpty() bool
	All() iter.Seq2[K, V]
	Keys() iter.Seq[K]
	Values() iter.Seq[V]
	ForEach(f func(K, V))
	AnySatisfy(predicate func(K, V) bool) bool
	AllSatisfy(predicate func(K, V) bool) bool
	NoneSatisfy(predicate func(K, V) bool) bool
}

// MutableMap extends MapIterable with mutation.
type MutableMap[K comparable, V any] interface {
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
