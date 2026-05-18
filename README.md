# mapdb-collections (Go)

High-performance primitive-specialized and generic collections for Go, inspired by [Eclipse Collections](https://eclipse.dev/collections/).

## Why?

Go's standard library provides `map` and `slice` — great general-purpose containers, but they box all values and lack the rich iteration API that Eclipse Collections is known for. This library provides:

- **Primitive-specialized types** (`Int32ArrayList`, `Int32HashSet`, `Int32Int64HashMap`, etc.) with no interface boxing and contiguous memory layout
- **Generic object collections** (`ArrayList[T]`, `HashSet[T]`, `HashMap[K,V]`, etc.) with the full Eclipse Collections API
- Rich functional methods: `Select`, `Reject`, `Detect`, `AnySatisfy`, `AllSatisfy`, `InjectInto`, `Collect`, and more

## Primitive Collections

Mutable types, one per primitive. Primitive counts by type:

| Type | Example | Supported primitives |
|------|---------|----------------------|
| **ArrayList** | `Int32ArrayList` | 7 (int8, int16, int32, int64, uint16, float32, float64) |
| **HashSet** | `Int32HashSet` | 8 (the above + bool) |
| **HashBag** | `Int32HashBag` | 7 |
| **ArrayStack** | `Int32ArrayStack` | 7 |
| **ArrayDeque** | `Int32ArrayDeque` | 7 — ring-buffered, O(1) both ends |
| **HashMap** | `Int32Int64HashMap` | 49 pairs (7×7) |
| **TreeSet** | `Int32TreeSet` | 7 |
| **TreeMap** | `Int32Int64TreeMap` | 49 pairs, NavigableMap-style API |
| **PriorityQueue** | `Int32PriorityQueue` | 7 |
| **BitSet** | `BitSet` | — |
| **Pair** | `Int32Int64Pair` | 49 pairs + Object variants |
| **Interval** | `Int32Interval` | signed int types only (4) |

`uint16` is the Go mapping of Java's `char` — it flows through every template that supports unsigned arithmetic. `bool` is supported on HashSet today; extending it to the other containers is tracked as follow-up work (requires gating out the Sort / Sum / Min / Max / BinarySearch paths in the individual templates).

Every mutable type has an `Immutable` counterpart (`ImmutableInt32ArrayList`, etc.) obtained via `ToImmutable()`. A `Synchronized` wrapper exposes the full mutable surface under an internal `sync.RWMutex` so reads, writes, and callback-based functional methods can be mixed freely from multiple goroutines.

## Object Collections

Generic collections using Go 1.24 type parameters:

| Type | Description |
|------|-------------|
| `ArrayList[T]` | Ordered list backed by `[]T` |
| `HashSet[T]` | Unordered set backed by `map[T]struct{}` |
| `HashMap[K, V]` | Key-value map backed by `map[K]V` |
| `HashBag[T]` | Counting bag backed by `map[T]int` |
| `ArrayStack[T]` | LIFO stack backed by `[]T` |
| `HashBiMap[K, V]` | Bidirectional map with unique keys and values |
| `LinkedHashMap[K, V]` | HashMap with insertion-order iteration |
| `LinkedHashSet[T]` | HashSet with insertion-order iteration |
| `TreeMap[K, V]` | Red-black tree with Comparator; Floor / Ceiling / Higher / Lower / HeadMap / TailMap / SubMap / PollFirstEntry / PollLastEntry / DescendingMap / DescendingKeys |
| `TreeSet[T]` | Comparator-ordered set |
| `HashMultimap[K, V]` | Multiple values per key, unordered keys |
| `TreeMultimap[K, V]` | Multiple values per key, Comparator-ordered keys |

All generic types implement composable interfaces: `Collection[T]`, `MutableList[T]`, `MutableSet[T]`, `MutableBag[T]`, `MutableStack[T]`, `MutableMap[K,V]`, `MutableBiMap[K,V]`.

## Composing operations

Go's type system doesn't allow generic methods on generic types, so
operations that change the element type (`Collect`, `GroupBy`, `Partition`,
`Zip`, `ZipWithIndex`, `Chunk`, `FlatCollect`) can't be methods on `ArrayList[T]` directly. They live in the `stream/` package as free functions.

In practice this means:

```go
// Eclipse Collections Java (fluent):
//   people.select(p -> p.hasCats()).collectInt(Person::getAge).sum()

// mapdb-golang:
catOwners := people.Select(func(p *Person) bool { return p.hasCats() })
ages      := stream.Map(catOwners.All(), func(p *Person) int { return p.Age })
total     := stream.Sum(ages)
```

The `.All()` call produces an `iter.Seq[T]` that `stream.*` consumes; the pipeline is lazy until you materialise it (`stream.ToSlice`, `stream.Sum`, `stream.Reduce`, etc.). `stream.Partition` materialises eagerly and returns two re-runnable seqs — safe over single-shot sources.

## Quick Start

```go
import (
    "github.com/mapdb/mapdb-golang/arraylist"
    "github.com/mapdb/mapdb-golang/object"
)

// Primitive ArrayList
list := arraylist.Int32ArrayListOf(3, 1, 4, 1, 5)
list.Sort()
selected := list.Select(func(v int32) bool { return v > 2 })

// Generic ArrayList
names := object.NewArrayList[string]()
names.Add("Alice"); names.Add("Bob"); names.Add("Charlie")
found, _ := names.Detect(func(s string) bool { return s[0] == 'B' })

// Generic HashBag
bag := object.NewHashBag[string]()
bag.Add("apple")
bag.AddOccurrences("apple", 3)
// bag.OccurrencesOf("apple") == 4

// Multimap
mm := object.NewHashMultimap[string, int]()
mm.PutAll("alice", 1, 2, 3)
mm.Put("bob", 10)

// NavigableMap on TreeMap
tm := object.NewTreeMap[int, string](object.NaturalComparator[int]())
tm.Put(10, "a"); tm.Put(20, "b"); tm.Put(30, "c")
if k, _, ok := tm.Higher(15); ok {
    _ = k // 20 — smallest key strictly > 15
}
for k, v := range tm.SubMap(10, 25) {
    _ = k; _ = v
}
```

## Thread safety

Each mutable primitive type has a `Synchronized*` wrapper that exposes the full mutable surface under a `sync.RWMutex`:

- read-only methods hold an `RLock`
- writes hold the write `Lock`
- functional methods (`Select`, `ForEach`, `AnySatisfy`, …) snapshot under `RLock` and run the callback unlocked — callbacks can re-enter the wrapper without deadlocking
- binary methods (`Equals`, `Union`, `Intersect`, …) acquire the pair of locks in pointer-address order, so `A.op(B)` and `B.op(A)` racing can't deadlock

## Requirements

- Go 1.24+ (uses `hash/maphash.Comparable` in generic hashing strategies)
- Zero external runtime dependencies

## Status

Early development. Licensed EPL-1.0 / EDL-1.0 to match Eclipse Collections. See `LICENSE-EPL-1.0.txt` and `LICENSE-EDL-1.0.txt`. Every file carries a `USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND` notice; please validate against your workload before relying on it in production.
