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
// why TreeSet and HashSetWithStrategy — comparator/strategy types with the richest
// APIs — can satisfy the hierarchy at all (11 §4).
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

// SortedSet is a Set whose iteration is ordered (by a Comparator) and which
// supports order-based navigation and order-statistic queries. Satisfied by the
// comparator-backed TreeSet — the relaxation of Set to T-any (11 §4) is what lets
// it model an interface at all. Read-only: navigation, not mutation.
type SortedSet[T any] interface {
	Set[T]
	// Min/Max return the least/greatest element, or false if empty.
	Min() (T, bool)
	Max() (T, bool)
	// Floor/Ceiling return the greatest ≤ / least ≥ value; Lower/Higher are their
	// strict (<, >) counterparts. The bool is false when no such element exists.
	Floor(value T) (T, bool)
	Ceiling(value T) (T, bool)
	Lower(value T) (T, bool)
	Higher(value T) (T, bool)
	// Rank returns the number of elements strictly less than value (its 0-based
	// insertion position); Select returns the i-th smallest (0-based), false if out
	// of range. Together they are the order-statistic pair (spec: rank-select.md).
	Rank(value T) int
	Select(i int) (T, bool)
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
// tree-backed map uses its Comparator, a strategy-backed map uses its strategy.
// Relaxing K here is what lets TreeMap and HashMapWithStrategy satisfy the map
// hierarchy (11 §4). Multimaps (Get returns []V, distinct Put shape) do NOT model
// MapIterable — they get their own Multimap interface in a later slice.
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

// SortedMap is a MapIterable whose keys are ordered (by a Comparator), exposing
// the endpoints and the three range views (mirroring Java's SortedMap:
// firstKey/lastKey + headMap/tailMap/subMap). Range views are lazy iter.Seq2.
type SortedMap[K any, V any] interface {
	MapIterable[K, V]
	// Min/Max return the entry with the least/greatest key, false if empty.
	Min() (K, V, bool)
	Max() (K, V, bool)
	// HeadMap yields entries with key < toKey; TailMap yields key ≥ fromKey; SubMap
	// yields fromKey ≤ key < toKey — all in ascending key order, all half-open.
	HeadMap(toKey K) iter.Seq2[K, V]
	TailMap(fromKey K) iter.Seq2[K, V]
	SubMap(fromKey, toKey K) iter.Seq2[K, V]
}

// NavigableMap extends SortedMap with point navigation, order statistics, and
// descending views (mirroring Java's NavigableMap). Satisfied by TreeMap.
type NavigableMap[K any, V any] interface {
	SortedMap[K, V]
	// Floor/Ceiling return the entry with the greatest key ≤ / least key ≥ key;
	// Lower/Higher are the strict counterparts. The bool is false if none exists.
	Floor(key K) (K, V, bool)
	Ceiling(key K) (K, V, bool)
	Lower(key K) (K, V, bool)
	Higher(key K) (K, V, bool)
	// Rank returns the number of keys strictly less than key; SelectEntry returns
	// the i-th smallest entry (0-based), false if out of range.
	Rank(key K) int
	SelectEntry(i int) (K, V, bool)
	// DescendingMap / DescendingKeys iterate in descending key order.
	DescendingMap() iter.Seq2[K, V]
	DescendingKeys() iter.Seq[K]
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

// ── Multimap interfaces ───────────────────────────────────────────────

// Multimap is the read-only interface for a key→many-values grouping. Satisfied by
// HashMultimap and the ordered TreeMultimap; the K-any relaxation (11 §4) is what
// lets the comparator-backed tree multimap model it. This is the interface a
// grouping terminal (stream/par GroupBy) can return instead of a concrete
// *HashMultimap — the whole point of giving multimaps an interface (11 §4).
//
// Multimap is NOT a MapIterable: its Get returns []V (a key's values), not a
// single (V, bool), so the two do not share a contract.
type Multimap[K any, V any] interface {
	// Get returns the values stored under key, or nil if the key is absent. The
	// current implementations return a fresh copy that is safe to retain and
	// mutate; a single accessor (not also GetCopy) keeps the contract minimal and
	// leaves room for a future view-returning multimap — treat the result as
	// read-only for forward compatibility.
	Get(key K) []V
	// ContainsKey reports whether key has at least one value.
	ContainsKey(key K) bool
	// Len is the total number of key-value pairs; SizeDistinct is the number of
	// distinct keys.
	Len() int
	SizeDistinct() int
	// All yields every key-value pair (a key with n values yields n times); Keys
	// yields each distinct key once; Values yields every value across all keys.
	All() iter.Seq2[K, V]
	Keys() iter.Seq[K]
	Values() iter.Seq[V]
	// ForEach visits every pair; ForEachKey visits each distinct key once;
	// ForEachKeyMultiValues visits each key with its whole (copied) value slice.
	ForEach(f func(K, V))
	ForEachKey(f func(K))
	ForEachKeyMultiValues(f func(K, []V))
}

// MutableMultimap extends Multimap with mutation.
type MutableMultimap[K any, V any] interface {
	Multimap[K, V]
	// Put adds value under key; PutAll adds several values under key at once.
	Put(key K, value V)
	PutAll(key K, values ...V)
	// RemoveKey removes and returns all values previously under key (nil if
	// absent). RemoveMatching removes the values equal to target (per eq) under
	// key and returns how many were removed.
	RemoveKey(key K) []V
	RemoveMatching(key K, target V, eq func(V, V) bool) int
	Clear()
}
