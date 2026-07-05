# Changelog

All notable changes to mapdb-golang are documented here.

This module has no version tags yet (effectively pre-1.0 / v0). Per Go module
semantics a pre-v1 breaking change does **not** require a `/v2` import-path
suffix, so the module path is unchanged (`github.com/mapdb/mapdb-golang`). This
release is a coordinated breaking idiom cleanup batched into one bump so
downstream code breaks at most once.

## [Unreleased] — correctness fixes + algo-package catch-up

Correctness-first pass over the merged algorithmic packages and the code
generators. Two **breaking** API changes are called out below; the rest are bug
fixes and additive documentation.

### Fixed — code-generator template bugs (regenerated across all variants)

- **sentinelhashmap**: tombstones left by `Remove` were never counted toward the
  load factor, so `Put`/`Remove` churn filled the table with tombstones until no
  empty slot remained and every probe loop hung forever (~22 cycles on a default
  table). Tombstones now count toward occupancy, and a resize triggered by
  tombstone pressure rehashes at the same capacity (so pure churn no longer grows
  the table without bound) while genuine data growth still doubles.
- **hashmap** (object-keyed and object-valued shapes): `rehashFrom`'s
  backward-shift condition was inverted and the wraparound guard missing, so
  `Remove` silently lost keys and created ghost duplicates. Ported the correct
  primitive×primitive backward-shift logic.
- **treeset / treemap**: the `FromSorted` bulk builder created nodes without the
  subtree-size augmentation, so `Rank`/`Select` returned wrong answers after a
  bulk load. The size is now set bottom-up during the build.

### Fixed — algorithmic packages

- **object.HashMultimap.PutAll**: a zero-value `PutAll(k)` no longer creates a
  phantom key (matches `TreeMultimap`).
- **countmin.NewCountMinOptimal**: rejects derived widths/depths exceeding
  `MaxUint32` (implementation-defined `float→uint32` for tiny epsilon).
- **rangev.PutCoalescing**: coalescing is now transitive and
  direction-independent — an abutting equal-valued chain collapses identically
  regardless of the entries' storage order (previously biased rightward). See the
  companion spec change (`mapdb-collection-spec`,
  `feat/rangemap-coalescing-direction-independent`).
- **boundedlru**: reentering the map from an eviction callback now panics instead
  of silently corrupting the arena.

### Changed — breaking

- **hyperloglog**: `NewHyperLogLogWithPrecision` and `HyperLogLogFromBytes` now
  return `*HyperLogLog` (was `HyperLogLog`). Value copies used to alias the
  register array and the zero value panicked in `Add`. `Registers()` now returns
  a **copy** of the register array (was the internal slice), so callers can no
  longer corrupt the sketch through it.

### Added

- Documentation: README now lists the algorithmic & probabilistic packages
  (`roaring`, `bloom`, `hyperloglog`, `countmin`, `fenwick`, `boundedlru`,
  `rangev`, `immutablesorted`, `multimap`, `pump`, `hash`) plus the `seq` lazy
  layer and `par` parallel layer, and a top-level `LICENSE` file (dual EPL-1.0 /
  EDL-1.0, EDL prominent). The "Composing operations" section now documents the
  current `seq` pipeline (`seq.From`/`seq.Map`/`seq.Sum`, fluent `seq.Seq[T]`)
  instead of the superseded `stream/` package, and points CPU-bound work at
  `par`.
- Tests/CI: `go test -fuzz` targets for the two hand-written byte parsers
  (`roaring.Deserialize`, `HyperLogLogFromBytes`); a `-race` CI lane and a
  gofmt-clean check; the codegen drift gate extracted to
  `scripts/check-codegen.sh` with a `Makefile` mirroring CI.
- Conformance laws (generated): a new `internal/conformance` package expresses
  the collection laws once as generic predicates, and `internal/codegen` stamps
  a `conformance_generated_test.go` per family. Covered so far: law 1
  (`All()` ≡ `ToSlice()`, order-class-aware) across the 8 single-value families;
  the map size-accounting law (`Len()` ≡ `|All()|`) across hashmap /
  sentinelhashmap / treemap, plus treemap key-ascending; and the Segments /
  Segments2 partition law (concat of `Segments(n)` ≡ `All()` as a multiset/map,
  each segment re-runnable) across arraylist, stack, deque, priorityqueue,
  treeset, interval, and treemap.

## [0.2.0] — Breaking idiom cleanup (v2)

This release applies the deferred source-breaking Go-idiom changes that were
intentionally batched out of the additive v1 pass. **All of the following are
breaking.** Migrate by mechanical find-and-replace using the tables below.

### 1. `Size()`/`IsEmpty()` removed — use `Len()`

`Size()` and `IsEmpty()` are gone from every collection type. `Len()` (added
additively in v1) is now the single length accessor, matching Go convention
(`sort.Interface`, `container/list`, `bytes.Buffer`).

| Old | New |
|-----|-----|
| `c.Size()` | `c.Len()` |
| `c.IsEmpty()` | `c.Len() == 0` |
| `!c.IsEmpty()` | `c.Len() != 0` |

`SizeDistinct()` (bag / multimap distinct-key count) is a different method and
is unchanged. `bitset.BitSet.IsEmpty()` is **kept** — for a bit set "no bits
set" (`Cardinality() == 0`) is a domain predicate distinct from length, so it
is not the removed Java-ism.

### 2. Package rename `priority_queue` → `priorityqueue`

Underscores violate Go package naming. The import path and package name change:

| Old | New |
|-----|-----|
| `import ".../priority_queue"` | `import ".../priorityqueue"` |
| `priority_queue.NewInt32PriorityQueue()` | `priorityqueue.NewInt32()` |

### 3. De-stuttered exported type names

Exported types no longer repeat their package name. The package-name word is
dropped from every exported type and constructor.

| Package | Old type | New type | Old constructor | New constructor |
|---------|----------|----------|-----------------|-----------------|
| `arraylist` | `Int32ArrayList` | `Int32` | `NewInt32ArrayList` / `Int32ArrayListOf` | `NewInt32` / `Int32Of` |
| `arraylist` | `ImmutableInt32ArrayList` | `ImmutableInt32` | `NewImmutableInt32ArrayList` | `NewImmutableInt32` |
| `arraylist` | `SynchronizedInt32ArrayList` | `SynchronizedInt32` | `NewSynchronizedInt32ArrayList` | `NewSynchronizedInt32` |
| `hashset` | `Int32HashSet` | `Int32` | `NewInt32HashSet` / `Int32HashSetOf` | `NewInt32` / `Int32Of` |
| `hashmap` | `Int32Int64HashMap` | `Int32Int64` | `NewInt32Int64HashMap` | `NewInt32Int64` |
| `treemap` | `Int32Int64TreeMap` | `Int32Int64` | `NewInt32Int64TreeMap` | `NewInt32Int64` |
| `treeset` | `Int32TreeSet` | `Int32` | `NewInt32TreeSet` / `Int32TreeSetOf` | `NewInt32` / `Int32Of` |
| `sentinelhashmap` | `Int32Int32SentinelHashMap` | `Int32Int32` | `NewInt32Int32SentinelHashMap` / `…WithCapacity` | `NewInt32Int32` / `NewInt32Int32WithCapacity` |
| `stack` | `Int32ArrayStack` | `Int32` | `NewInt32ArrayStack` / `Int32ArrayStackOf` | `NewInt32` / `Int32Of` |
| `deque` | `Int32ArrayDeque` | `Int32` | `NewInt32ArrayDeque` / `Int32ArrayDequeOf` | `NewInt32` / `Int32Of` |
| `priorityqueue` | `Int32PriorityQueue` | `Int32` | `NewInt32PriorityQueue` | `NewInt32` |
| `interval` | `Int32Interval` | `Int32` | `NewInt32Interval` / `Int32IntervalFromTo` | `NewInt32` / `Int32FromTo` |

**Bag** keeps the hash-vs-tree discriminator, moved in front of the primitive
(drops `Bag`):

| Old | New |
|-----|-----|
| `bag.Int32HashBag` | `bag.HashInt32` |
| `bag.Int32TreeBag` | `bag.TreeInt32` |
| `bag.ImmutableInt32HashBag` | `bag.ImmutableHashInt32` |
| `bag.SynchronizedInt32HashBag` | `bag.SynchronizedHashInt32` |
| `bag.NewInt32HashBag` / `bag.Int32HashBagOf` | `bag.NewHashInt32` / `bag.HashInt32Of` |
| `bag.NewInt32TreeBag` / `bag.Int32TreeBagOf` | `bag.NewTreeInt32` / `bag.TreeInt32Of` |

**Multimap** keeps the list-vs-set discriminator after the key/value names
(drops `Multimap`):

| Old | New |
|-----|-----|
| `multimap.Int32Int32ListMultimap` | `multimap.Int32Int32List` |
| `multimap.Int32Int32SetMultimap` | `multimap.Int32Int32Set` |
| `multimap.NewInt32Int32ListMultimap` | `multimap.NewInt32Int32List` |
| `multimap.NewInt32Int32SetMultimap` | `multimap.NewInt32Int32Set` |

`tuple` (`tuple.Int32Int64Pair`) and the generic `object.*` collections are
**unchanged** — their type names do not stutter against their package name.

### 4. Unified error conventions

| Situation | Old | New |
|-----------|-----|-----|
| Index out of range (`Get`/`Set`/`RemoveAtIndex` by index, stack `PeekAt`, interval `Get`) | `(T, error)` | **panics** like a native slice; returns the value directly (`T`) |
| Empty-collection accessor (stack/deque/priorityqueue `Pop`/`Peek`/`RemoveFirst`/`RemoveLast`/`PeekFirst`/`PeekLast`) | `(T, error)` | comma-ok `(T, bool)` (`false` when empty) |
| Map/set lookup (`Get`/`Put`/`Remove`, `Detect`, `Min`/`Max`/…) | comma-ok `(V, bool)` | unchanged |

Migration examples:

```go
// old
v, err := list.Get(i)
if err != nil { ... }
// new — panics on bad index, like list[i]
v := list.Get(i)

// old
top, err := stk.Pop()
if err != nil { ... }
// new — comma-ok
top, ok := stk.Pop()
if !ok { ... }
```

The immutable stack `Pop` changes from `(*Immutable…, T, error)` to
`(*Immutable…, T, bool)`.

### 5. Fluent `With`/`Without` renamed to `…Returning`

The fluent mutators on **mutable** and **synchronized** types mutate the
receiver and return it — `With` wrongly implied a copy to a Go reader. They are
renamed to make the mutation explicit (immutable types never had these):

| Old | New |
|-----|-----|
| `c.With(v)` | `c.AddReturning(v)` |
| `c.Without(v)` | `c.RemoveReturning(v)` |
| `c.WithAll(vs...)` | `c.AddAllReturning(vs...)` |
| `c.WithoutAll(vs...)` | `c.RemoveAllReturning(vs...)` |

### 6. `sort.Slice` → `slices.SortFunc`

Internal ordering now uses the modern stdlib `slices.SortFunc` (integer keys via
`cmp.Compare`, floats via the IEEE-754 total-order comparator). No public
signature change beyond what is already covered above.

### 7. Zero-value usability

The open-addressed tables (`*HashMap`, `*HashSet`, sentinel maps) are usable
from their zero value — reads guard an empty table and the first write
lazy-initializes via resize — matching the map-backed families. The
`Synchronized*` wrappers remain **construct-only** (they hold a `delegate`
pointer that a zero value leaves `nil`); use their `NewSynchronized*`
constructors.

### 8. Trimmed producer-side `collection/` interfaces

The 49 per-key/value `collection/*_map_interfaces.go` files (Eclipse-style
producer map interfaces) were unused by any consumer (production, tests, or the
validation runner) and have been **removed**. The 7 per-primitive capability
interface files (`collection/<prim>_interfaces.go`: `Sized`→`Len`-only,
`Iterable`, `Searchable`, `Convertible`, and the composed `Collection`/`List`/
`Set`/`Bag`/`Stack` interfaces) are **kept** and updated to the v2 signatures;
`collection/category_verify_test.go` still statically asserts that the
primitive collection types satisfy them.
