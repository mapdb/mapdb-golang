# v2 breaking idiom changes — DONE

The breaking Go-idiom cleanup batched out of the additive v1 pass has now been
applied as a single coordinated bump (see `CHANGELOG.md` for the full
consumer migration tables). This module has no version tags yet, so per Go
semantics the pre-1.0 breaking change does **not** require a `/v2` module-path
suffix — the module path is unchanged.

All items below are **DONE**:

- **DONE — Dropped `Size()` / `IsEmpty()`.** `Len()` is now the single length
  accessor; emptiness is `Len() == 0`. `SizeDistinct()` (bag/multimap) is a
  different method and stays. `bitset.BitSet.IsEmpty()` is kept (it means "no
  bits set", a domain predicate distinct from length — not the removed Java-ism).

- **DONE — Package rename `priority_queue` → `priorityqueue`.** The directory,
  package clause, `//go:generate` directive, the codegen dispatcher key, and the
  shared `genCmpFloat` package argument all moved. Nothing outside the package
  imported it, so the blast radius was contained.

- **DONE — De-stuttered exported type names.** The package-name word is dropped
  from every exported type and constructor: `arraylist.Int32ArrayList` →
  `arraylist.Int32`, `sentinelhashmap.NewFloat32Int32SentinelHashMapWithCapacity`
  → `sentinelhashmap.NewFloat32Int32WithCapacity`, etc. Bag keeps the hash/tree
  discriminator in front of the primitive (`bag.HashInt32` / `bag.TreeInt32`);
  multimap keeps the list/set discriminator after the names
  (`multimap.Int32Int32List` / `multimap.Int32Int32Set`). `tuple.*Pair` and the
  generic `object.*` collections do not stutter and are unchanged.

- **DONE — Unified error conventions.** Index misuse (`Get`/`Set`/`RemoveAtIndex`
  by index, stack `PeekAt`, interval `Get`) now **panics** like a native slice
  and returns the value directly. Empty-collection accessors
  (`Pop`/`Peek`/`RemoveFirst`/`RemoveLast`/`PeekFirst`/`PeekLast` on
  stack/deque/priorityqueue) are now comma-ok `(T, bool)`. Map/set lookups were
  already comma-ok and are unchanged. The immutable stack `Pop` is now
  `(*Immutable…, T, bool)`.

- **DONE — Renamed the fluent `With`/`Without` mutators.** On mutable and
  synchronized types these mutate the receiver and return it, so they are now
  `AddReturning` / `RemoveReturning` / `AddAllReturning` / `RemoveAllReturning`,
  making the mutation explicit. (Immutable types never had these.)

- **DONE — `sort.Slice` → `slices.SortFunc`.** Internal ordering uses the modern
  stdlib API: `cmp.Compare` for integer element types, the IEEE-754 total-order
  comparator for floats.

- **DONE — Uniform zero-value usability for the open-addressed tables.** The
  `*HashMap` / `*HashSet` / sentinel tables are now usable from their zero value:
  reads guard an empty table and the first write lazy-initializes via resize,
  matching the map-backed families (bag, multimap, the map-backed `object/`
  generics). **The `Synchronized*` wrappers remain construct-only** — a
  synchronized wrapper holds a pointer `delegate` whose READ methods would also
  nil-deref on a zero value; making it fully zero-value usable would require the
  wrapper to learn its delegate's constructor, which is out of scope. Use the
  `NewSynchronized*` constructors.

- **DONE — Trimmed the producer-side `collection/` interfaces.** The 49
  per-key/value `*_map_interfaces.go` files were unused by any consumer
  (production, tests, validation runner) and were removed. The 7 per-primitive
  capability interface files (`Sized`→`Len`, `Iterable`, `Searchable`,
  `Convertible`, and the composed `Collection`/`List`/`Set`/`Bag`/`Stack`) are
  kept, updated to the v2 signatures, and still statically verified by
  `collection/category_verify_test.go`.
