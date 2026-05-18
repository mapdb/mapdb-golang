package collection

//
// Compile-time verification that concrete generated types satisfy the
// category interfaces. Each blank assignment fails at compile time if
// the interface is not satisfied. Int32 is a representative — the same
// hierarchy is generated for every primitive.

import (
	"github.com/mapdb/mapdb-golang/arraylist"
	"github.com/mapdb/mapdb-golang/bag"
	"github.com/mapdb/mapdb-golang/hashset"
	"github.com/mapdb/mapdb-golang/interval"
	"github.com/mapdb/mapdb-golang/stack"
	"github.com/mapdb/mapdb-golang/treeset"
)

// ── Sub-interface verification ────────────────────────────────────────
// Every concrete type satisfies the small composable interfaces.

// Int32Sized
var _ Int32Sized = (*arraylist.Int32ArrayList)(nil)
var _ Int32Sized = (*arraylist.ImmutableInt32ArrayList)(nil)
var _ Int32Sized = (*hashset.Int32HashSet)(nil)
var _ Int32Sized = (*hashset.ImmutableInt32HashSet)(nil)
var _ Int32Sized = (*bag.Int32HashBag)(nil)
var _ Int32Sized = (*bag.Int32TreeBag)(nil)
var _ Int32Sized = (*bag.ImmutableInt32HashBag)(nil)
var _ Int32Sized = (*stack.Int32ArrayStack)(nil)
var _ Int32Sized = (*stack.ImmutableInt32ArrayStack)(nil)
var _ Int32Sized = (*treeset.Int32TreeSet)(nil)
var _ Int32Sized = (*interval.Int32Interval)(nil)

// Int32Iterable
var _ Int32Iterable = (*arraylist.Int32ArrayList)(nil)
var _ Int32Iterable = (*arraylist.ImmutableInt32ArrayList)(nil)
var _ Int32Iterable = (*hashset.Int32HashSet)(nil)
var _ Int32Iterable = (*hashset.ImmutableInt32HashSet)(nil)
var _ Int32Iterable = (*bag.Int32HashBag)(nil)
var _ Int32Iterable = (*bag.Int32TreeBag)(nil)
var _ Int32Iterable = (*bag.ImmutableInt32HashBag)(nil)
var _ Int32Iterable = (*stack.Int32ArrayStack)(nil)
var _ Int32Iterable = (*stack.ImmutableInt32ArrayStack)(nil)
var _ Int32Iterable = (*treeset.Int32TreeSet)(nil)
var _ Int32Iterable = (*interval.Int32Interval)(nil)

// Int32Searchable
var _ Int32Searchable = (*arraylist.Int32ArrayList)(nil)
var _ Int32Searchable = (*arraylist.ImmutableInt32ArrayList)(nil)
var _ Int32Searchable = (*hashset.Int32HashSet)(nil)
var _ Int32Searchable = (*hashset.ImmutableInt32HashSet)(nil)
var _ Int32Searchable = (*bag.Int32HashBag)(nil)
var _ Int32Searchable = (*bag.Int32TreeBag)(nil)
var _ Int32Searchable = (*bag.ImmutableInt32HashBag)(nil)
var _ Int32Searchable = (*stack.Int32ArrayStack)(nil)
var _ Int32Searchable = (*stack.ImmutableInt32ArrayStack)(nil)
var _ Int32Searchable = (*treeset.Int32TreeSet)(nil)
var _ Int32Searchable = (*interval.Int32Interval)(nil)

// Int32Convertible
var _ Int32Convertible = (*arraylist.Int32ArrayList)(nil)
var _ Int32Convertible = (*arraylist.ImmutableInt32ArrayList)(nil)
var _ Int32Convertible = (*hashset.Int32HashSet)(nil)
var _ Int32Convertible = (*hashset.ImmutableInt32HashSet)(nil)
var _ Int32Convertible = (*bag.Int32HashBag)(nil)
var _ Int32Convertible = (*bag.Int32TreeBag)(nil)
var _ Int32Convertible = (*bag.ImmutableInt32HashBag)(nil)
var _ Int32Convertible = (*stack.Int32ArrayStack)(nil)
var _ Int32Convertible = (*stack.ImmutableInt32ArrayStack)(nil)
var _ Int32Convertible = (*treeset.Int32TreeSet)(nil)
var _ Int32Convertible = (*interval.Int32Interval)(nil)

// ── Composed Int32Collection verification ─────────────────────────────
var _ Int32Collection = (*arraylist.Int32ArrayList)(nil)
var _ Int32Collection = (*arraylist.ImmutableInt32ArrayList)(nil)
var _ Int32Collection = (*hashset.Int32HashSet)(nil)
var _ Int32Collection = (*hashset.ImmutableInt32HashSet)(nil)
var _ Int32Collection = (*bag.Int32HashBag)(nil)
var _ Int32Collection = (*bag.Int32TreeBag)(nil)
var _ Int32Collection = (*bag.ImmutableInt32HashBag)(nil)
var _ Int32Collection = (*stack.Int32ArrayStack)(nil)
var _ Int32Collection = (*stack.ImmutableInt32ArrayStack)(nil)
var _ Int32Collection = (*treeset.Int32TreeSet)(nil)
var _ Int32Collection = (*interval.Int32Interval)(nil)

// ── Category interface verification ───────────────────────────────────

// Int32List / Int32MutableList
var _ Int32List = (*arraylist.Int32ArrayList)(nil)
var _ Int32MutableList = (*arraylist.Int32ArrayList)(nil)
var _ Int32List = (*arraylist.ImmutableInt32ArrayList)(nil)

// Int32Set / Int32MutableSet
var _ Int32Set = (*hashset.Int32HashSet)(nil)
var _ Int32MutableSet = (*hashset.Int32HashSet)(nil)
var _ Int32Set = (*treeset.Int32TreeSet)(nil)
var _ Int32MutableSet = (*treeset.Int32TreeSet)(nil)

// Int32Bag / Int32MutableBag
var _ Int32Bag = (*bag.Int32HashBag)(nil)
var _ Int32MutableBag = (*bag.Int32HashBag)(nil)
var _ Int32Bag = (*bag.Int32TreeBag)(nil)
var _ Int32MutableBag = (*bag.Int32TreeBag)(nil)

// Int32Stack / Int32MutableStack
var _ Int32Stack = (*stack.Int32ArrayStack)(nil)
var _ Int32MutableStack = (*stack.Int32ArrayStack)(nil)

// Immutable variants satisfy their category interfaces.
var _ Int32Set = (*hashset.ImmutableInt32HashSet)(nil)
var _ Int32Bag = (*bag.ImmutableInt32HashBag)(nil)
var _ Int32Stack = (*stack.ImmutableInt32ArrayStack)(nil)
