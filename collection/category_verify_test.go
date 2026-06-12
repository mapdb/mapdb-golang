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
var _ Int32Sized = (*arraylist.Int32)(nil)
var _ Int32Sized = (*arraylist.ImmutableInt32)(nil)
var _ Int32Sized = (*hashset.Int32)(nil)
var _ Int32Sized = (*hashset.ImmutableInt32)(nil)
var _ Int32Sized = (*bag.HashInt32)(nil)
var _ Int32Sized = (*bag.TreeInt32)(nil)
var _ Int32Sized = (*bag.ImmutableHashInt32)(nil)
var _ Int32Sized = (*stack.Int32)(nil)
var _ Int32Sized = (*stack.ImmutableInt32)(nil)
var _ Int32Sized = (*treeset.Int32)(nil)
var _ Int32Sized = (*interval.Int32)(nil)

// Int32Iterable
var _ Int32Iterable = (*arraylist.Int32)(nil)
var _ Int32Iterable = (*arraylist.ImmutableInt32)(nil)
var _ Int32Iterable = (*hashset.Int32)(nil)
var _ Int32Iterable = (*hashset.ImmutableInt32)(nil)
var _ Int32Iterable = (*bag.HashInt32)(nil)
var _ Int32Iterable = (*bag.TreeInt32)(nil)
var _ Int32Iterable = (*bag.ImmutableHashInt32)(nil)
var _ Int32Iterable = (*stack.Int32)(nil)
var _ Int32Iterable = (*stack.ImmutableInt32)(nil)
var _ Int32Iterable = (*treeset.Int32)(nil)
var _ Int32Iterable = (*interval.Int32)(nil)

// Int32Searchable
var _ Int32Searchable = (*arraylist.Int32)(nil)
var _ Int32Searchable = (*arraylist.ImmutableInt32)(nil)
var _ Int32Searchable = (*hashset.Int32)(nil)
var _ Int32Searchable = (*hashset.ImmutableInt32)(nil)
var _ Int32Searchable = (*bag.HashInt32)(nil)
var _ Int32Searchable = (*bag.TreeInt32)(nil)
var _ Int32Searchable = (*bag.ImmutableHashInt32)(nil)
var _ Int32Searchable = (*stack.Int32)(nil)
var _ Int32Searchable = (*stack.ImmutableInt32)(nil)
var _ Int32Searchable = (*treeset.Int32)(nil)
var _ Int32Searchable = (*interval.Int32)(nil)

// Int32Convertible
var _ Int32Convertible = (*arraylist.Int32)(nil)
var _ Int32Convertible = (*arraylist.ImmutableInt32)(nil)
var _ Int32Convertible = (*hashset.Int32)(nil)
var _ Int32Convertible = (*hashset.ImmutableInt32)(nil)
var _ Int32Convertible = (*bag.HashInt32)(nil)
var _ Int32Convertible = (*bag.TreeInt32)(nil)
var _ Int32Convertible = (*bag.ImmutableHashInt32)(nil)
var _ Int32Convertible = (*stack.Int32)(nil)
var _ Int32Convertible = (*stack.ImmutableInt32)(nil)
var _ Int32Convertible = (*treeset.Int32)(nil)
var _ Int32Convertible = (*interval.Int32)(nil)

// ── Composed Int32Collection verification ─────────────────────────────
var _ Int32Collection = (*arraylist.Int32)(nil)
var _ Int32Collection = (*arraylist.ImmutableInt32)(nil)
var _ Int32Collection = (*hashset.Int32)(nil)
var _ Int32Collection = (*hashset.ImmutableInt32)(nil)
var _ Int32Collection = (*bag.HashInt32)(nil)
var _ Int32Collection = (*bag.TreeInt32)(nil)
var _ Int32Collection = (*bag.ImmutableHashInt32)(nil)
var _ Int32Collection = (*stack.Int32)(nil)
var _ Int32Collection = (*stack.ImmutableInt32)(nil)
var _ Int32Collection = (*treeset.Int32)(nil)
var _ Int32Collection = (*interval.Int32)(nil)

// ── Category interface verification ───────────────────────────────────

// Int32List / Int32MutableList
var _ Int32List = (*arraylist.Int32)(nil)
var _ Int32MutableList = (*arraylist.Int32)(nil)
var _ Int32List = (*arraylist.ImmutableInt32)(nil)

// Int32Set / Int32MutableSet
var _ Int32Set = (*hashset.Int32)(nil)
var _ Int32MutableSet = (*hashset.Int32)(nil)
var _ Int32Set = (*treeset.Int32)(nil)
var _ Int32MutableSet = (*treeset.Int32)(nil)

// Int32Bag / Int32MutableBag
var _ Int32Bag = (*bag.HashInt32)(nil)
var _ Int32MutableBag = (*bag.HashInt32)(nil)
var _ Int32Bag = (*bag.TreeInt32)(nil)
var _ Int32MutableBag = (*bag.TreeInt32)(nil)

// Int32Stack / Int32MutableStack
var _ Int32Stack = (*stack.Int32)(nil)
var _ Int32MutableStack = (*stack.Int32)(nil)

// Immutable variants satisfy their category interfaces.
var _ Int32Set = (*hashset.ImmutableInt32)(nil)
var _ Int32Bag = (*bag.ImmutableHashInt32)(nil)
var _ Int32Stack = (*stack.ImmutableInt32)(nil)
