# Deferred breaking idiom changes (batch for a future v2)

Phase 7 applied the **additive** Go idiom improvements (no source breakage):
`Len()` added as an alias for `Size()`, `math.FloatNbits` replacing needless
`unsafe.Pointer` float-bit reinterpretation, and lazy initialization so
zero-value map-backed collections no longer panic.

The following are genuinely **breaking** and are intentionally batched for a
single coordinated v2 so downstream code breaks at most once:

- **Drop / rename `Size()` → `Len()`.** Today both exist; v2 would remove
  `Size()` (and likely `IsEmpty()` in favor of `Len() == 0`).
- **Package rename `priority_queue` → `priorityqueue`.** Underscores violate Go
  package naming; renaming changes every import path.
- **De-stutter exported type names** — e.g. `arraylist.Int32ArrayList` →
  `arraylist.Int32`, `sentinelhashmap.NewFloat32Int32SentinelHashMapWithCapacity`.
- **Unify error conventions.** Today `(T, error)` (index OOB), comma-ok
  (missing key), and panic (constructor misuse) coexist. v2: panic on index
  misuse (like slices), comma-ok elsewhere.
- **`With`/`Without`/`WithAll` on mutable types** mutate the receiver and return
  it; to a Go reader `With` implies a copy. Rename (e.g. `AddReturning`) or make
  copy semantics uniform.
- **`sort.Slice` → `slices.SortFunc`** where the public ordering surface changes.
- **Zero-value `Synchronized*` wrappers remain construct-only.** Unlike the
  base collections (now zero-value usable via lazy init), a synchronized
  wrapper holds a pointer `delegate` whose READ methods would also nil-deref;
  making it fully zero-value usable is out of scope. Use the `NewSynchronized*`
  constructors. (Considered for v2 only if the wrapper learns its delegate's
  constructor.)
- **Trim the 197 producer-side `collection/` interfaces** — Go interfaces are
  small and consumer-defined; the Eclipse-style hierarchy is rarely imported.
