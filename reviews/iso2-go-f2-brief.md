# Review brief: mapdb-golang iso2 findings G1-F2 / G1-F6 / G3

You are reviewing an uncommitted change in the git repo `/home/play2/mapdb/mapdb-golang`
(branch `main`, base commit `c4d4fa4`). Read the repo directly; the key diff is also at
`/tmp/iso2-go-f2.diff` (generator + runner + CI + test files; the 60-odd purely
mechanical regenerated `treemap/*.go`, `treeset/*.go`, `arraylist/*.go` files are excluded
from that diff file but ARE in the working tree — inspect them with `git diff` in the repo
if you want).

## Background

There is a 298+ scenario cross-language conformance suite at
`/home/play2/mapdb/mapdb-collection-spec/cross-language-validation/`
(`scenarios/`, `validate.sh`, `check-runners.sh` + `runners.json` guard, `README.md`).

The suite's central rule (README §G5): **a runner MUST obtain every assertion value by
calling the production method that the assertion names. A runner-local loop over the
collection's contents is a bug even when the comparator used is production code.**
The Go runner is `/home/play2/mapdb/mapdb-golang/cmd/validate/main.go`.

Three findings were assigned (from `/home/play2/mapdb/todo/iso2/PROGRESS.md`):

- **G1-F2 (Go)**: `evalSetAssertion` computed union / intersect / difference /
  symmetric-difference over a runner-local `map[int32]struct{}` instead of calling
  production `hashset.Int32.Union/Intersect/Difference/SymmetricDifference`.
- **G1-F6 (Go)**: the list `add_at` operation snapshotted the list, `Clear()`ed it, and
  `AddAll`ed a rebuilt slice, instead of calling a production insert-at-index method.
- **G3**: `staticcheck` (which catches U1000 dead code — a self-recursive dead renderer
  helper in this runner once hid a real bug) was not a required CI step.

## What I changed

### 1. G1-F2 — set algebra now comes from production methods
`cmd/validate/main.go` `evalSetAssertion`: all eight set-algebra assertions
(`union_sorted`, `intersect_sorted`, `difference_sorted`, `symmetric_difference_sorted`
and the four `*_size` twins) now call `set.Union(other)` / `.Intersect(other)` /
`.Difference(other)` / `.SymmetricDifference(other)` — all four exist as production
methods in `hashset/int32_hash_set.go` (lines ~296-350), so no composition was needed for
symmetric difference. `*_size` reads `Len()` off the production result. `setToSorted` now
uses the production `ToSlice()` and sorts the RESULT (a hash set has no order, so sorting
for rendering is allowed by the rule).

### 2. G1-F2 sweep — the rest of the runner
I swept the whole runner for the same pattern and rewrote `evalListAssertion` to use
production `arraylist.Int32` methods: `Len`, `Contains`, `Get`, `Select`, `Reject`,
`Detect`, `Count`, `AnySatisfy`, `AllSatisfy`, `NoneSatisfy`, `Sort`, `ToSlice`
(previously all of these were runner-local `for _, v := range values` loops).
`evalBagAssertion` `sorted_distinct` now uses production `AllDistinct()` (was a local
`map[int32]struct{}`) and `to_sorted_array` uses production `ToSlice()` (was a local
re-flatten via `ForEachWithOccurrences`). The f32 ArrayList runner now uses `Len()` /
`ToSlice()` rather than a pre-snapshotted local slice.

**Deliberately left runner-local (please challenge this):** `inject_into_product`
in `evalListAssertion`. The assertion wants a WIDENING i64 product; production
`InjectInto` is typed `InjectInto(initial int32, f func(int32, int32) int32) int32`, so
its accumulator is i32 and would wrap. `Sum()` is the only widening reduction and is
fixed to addition. Documented in a comment at the case.

### 3. G1-F6 — production insert-at-index
`arraylist.Int32` genuinely had no insert-at (only `Add`, `AddAll`, `Set`,
`RemoveAtIndex`). The frozen Java reference has it
(`eclipse-collections-code-generator/.../mutablePrimitiveList.stg`: `void addAtIndex(int
index, <type> element)`), so I added `AddAtIndex(index int, value T)` to the arraylist
codegen template (`internal/codegen/arraylist.go`) plus a `Synchronized<T>ArrayList`
delegating wrapper, regenerated all 7 primitive lists + 7 synchronized wrappers, and added
a unit test `TestInt32_AddAtIndex` in `arraylist/int32_array_list_test.go`
(front / middle / `index == Len()` append boundary / empty list / both out-of-range
panics). The runner's `add_at` case now calls `l.AddAtIndex(idx, v)`.

### 4. G3 — staticcheck as a required CI step
`.github/workflows/ci.yml` gained a `Staticcheck` step between `Test` and the codegen
drift gate, pinned: `go install honnef.co/go/tools/cmd/staticcheck@v0.7.0` then
`"$(go env GOPATH)/bin/staticcheck" ./...`. There is no Makefile in this repo, so CI was
the only place to add it. `go.mod` says `go 1.24` and CI uses `setup-go` with
`go-version-file: go.mod`.

staticcheck found 9 issues; all fixed:
- `cmd/validate/main.go` `snapshotList` unused (U1000) after the rewrite — deleted.
- `internal/codegen/tuple.go` SA9009 "ineffectual compiler directive due to extraneous
  space: `// go:generate.`" — it was prose at line start; reworded so `go:generate` is
  mid-line.
- 6× SA4006 in `object/multimap_test.go` ("this value of X is never used" on
  `got = append(got, 100)` etc.) — these are INTENTIONAL: the tests prove the multimap's
  `Get`/`GetCopy`/`ForEachKeyMultiValues` are defensive, i.e. a caller's append must not
  reach internal storage. Suppressed with `//lint:ignore SA4006 <reason>`.
- 1× SA4000 in `hashmap/hash_utils_test.go` (`hashComparable(k) != hashComparable(k)`) —
  intentional determinism test; suppressed with `//lint:ignore SA4000 <reason>`.

### 5. A REAL PRODUCTION BUG found by the gate
Running `validate.sh` turned up 1 red: `17-bulk-load/treemap_i32_from_sorted`,
`rank_21: expected=4 got=2`. Root cause: the bottom-up bulk builders
`build<Node>` in the treemap and treeset codegen templates never set `node.size`
— the per-node subtree-size augmentation that `Rank`/`SelectKey`/`Select` read.
A bulk-built tree never passes through the insert/rotation paths that maintain it, so
order statistics were silently wrong on any `*FromSorted` / `Sink` built tree.
Fix: call `<Node>FixSize(node)` after both children are built, in
`internal/codegen/treemap.go` and `internal/codegen/treeset.go`; regenerated (this is why
~60 `treemap/*.go` + `treeset/*.go` files changed by 5 lines each). Regression tests
added: `TestInt32Int32_RankSelectAfterFromSorted` in
`treemap/int32_int32_rank_select_test.go` and `TestInt32_RankSelectAfterFromSorted` in
`treeset/int32_rank_select_test.go` (small tree + a 257-element tree, plus the existing
`assertSizeInvariant` / `assertSetSizeInvariant` whole-tree invariant checks).

## Gate results (all run locally, after the change)

- `go build ./...` — OK
- `go vet ./...` — OK
- `go test ./...` — all packages ok
- `gofmt -l .` — empty
- codegen drift gate (the inline delete-then-regenerate from `.github/workflows/ci.yml`;
  there is no `scripts/check-codegen.sh` in this repo) — clean, no drift, no strays
- `staticcheck ./...` — clean (run as `GOTOOLCHAIN=go1.24.9 staticcheck ./...`; the host
  has Go 1.27 whose export-data version staticcheck 2026.1/v0.7.0 cannot decode, hence the
  toolchain pin locally; CI is on 1.24 via `go-version-file: go.mod`)
- `./check-runners.sh --root /home/play2/mapdb` — PASS go: 26 checked (all 5 langs PASS)
- `./validate.sh --skip-java --skip-rust --skip-ts --skip-zig` —
  **before: 306 scenarios, 305 pass / 1 fail. after: 306 pass / 0 fail.**

## Questions — answer each explicitly

1. **Is any assertion in `cmd/validate/main.go` still computed by a runner-local loop
   when a production method exists?** Sweep the WHOLE file, not just the sections I named
   (there are runners for hashmap, treemap, treeset, multimap, deque, stack, priority
   queue, bitset, bloom, roaring, interval, range, rangeset, bounded-LRU, hash pipeline,
   f32 variants, …). Name file:line for each hit and the production method that should
   have been called. Note: reading `Len()` off a production result, or sorting a
   production result for rendering when the collection is unordered, is explicitly
   allowed.
2. **Is my `inject_into_product` carve-out correct**, or is there a production reduction
   on `arraylist.Int32` (or a reachable production helper) that gives a widening i64
   product? If there is, say which.
3. **Is staticcheck actually enforced by CI now?** Read
   `.github/workflows/ci.yml` as a whole. Will the step fail the job on a finding? Is the
   pin right (`honnef.co/go/tools/cmd/staticcheck@v0.7.0` — does that tag exist and is it
   compatible with the Go version `setup-go` will pick from `go.mod`'s `go 1.24`)? Would
   `go install` inside the module directory perturb `go.mod`/`go.sum`?
4. **Are my 7 `//lint:ignore` suppressions legitimate**, or is any of them masking a real
   defect? Read the surrounding test bodies.
5. **Is the treemap/treeset subtree-size fix correct and complete?** Specifically:
   (a) is `<Node>FixSize(node)` after both children the right placement, (b) did I miss
   any other bulk/bottom-up builder in the repo that also skips the augmentation
   (check `object/`, `immutablesorted/`, `sentinelhashmap/`, `multimap/`, `interval/`,
   `rangev/`, and any `*Sink` types), (c) is `Len()`/`m.size` consistent with the node
   augmentation on a bulk-built tree, (d) is the added test coverage adequate?
6. **Is the `AddAtIndex` contract right?** Compare against the frozen Java reference
   (`/home/play2/mapdb/mapdb-java/eclipse-collections-code-generator/src/main/resources/impl/list/mutable/primitiveArrayList.stg`
   `addAtIndex`) — bounds behaviour (`index == size` must append; `index > size` and
   `index < 0` must reject), and against the other ports' lists. Also: should the
   `Immutable<T>ArrayList` or any interface in `collection/` have gained anything?
   Is my implementation (`append` a zero, `copy` shift right, assign) correct including
   the aliasing/reallocation case?
7. **Any production bug misclassified?** i.e. did I paper over a genuine production defect
   by changing the runner, rather than fixing the collection? Check especially the
   `evalListAssertion` and `evalSetAssertion` rewrites: the new code must not be more
   permissive than the old.

Be concrete: file:line, and for each finding say whether it BLOCKS the commit or is a
nit. Disagreement with evidence is welcome.

Write your answer to `/tmp/iso2-go-f2-review.md`.
