# ISO2 Go F2 review

## Verdict

**BLOCK.** The generated tree-size repair and `AddAtIndex` implementation are sound, and the widening reduction carveout is justified. However, the validation runner still contains four clear G5 violations, including one observably incorrect `sorted_values` implementation, plus a bounded-LRU assertion that needs a production non-touch lookup API to satisfy both G5 and the LRU observation contract.

Validation performed:

- `GOTOOLCHAIN=go1.24.9 go test ./...` passes.
- Staticcheck 2026.1 (`v0.7.0`) reports no findings on `./...`.
- `git diff --check` passes.
- The generated size repair is present in exactly 49 treemap specializations and 7 treeset specializations.

## 1. Full `cmd/validate/main.go` assertion-loop sweep

### Blockers

1. **`cmd/validate/main.go:2249-2255` — TreeMap `sorted_values`**

   The runner ranges over `m.All()` and extracts values itself. It should call the production `m.Values()` method (`treemap/int32_int32_tree_map.go:231`) and sort the returned values for this assertion.

   This is more than a structural G5 violation: the cross-language validation contract defines `sorted_values` as values sorted ascending, while the current implementation emits values in key order. The current scenarios happen to use values whose order tracks their keys, so they produce a false green. This is a **blocker**.

2. **`cmd/validate/main.go:2347-2350` — float32 HashMap `sorted_keys`**

   The runner manually ranges over `m.Keys()` to build a slice. It should call production `m.KeysToSlice()` (`hashmap/float32_int32_hash_map.go:447`) and then apply the runner-side float total-order sort. Sorting for deterministic assertion output is permitted; reconstructing the collection result is not. This is a **G5 blocker**, although output is currently equivalent.

3. **`cmd/validate/main.go:2398-2400` — float32 HashSet `sorted_values` / `to_sorted_array`**

   The runner manually accumulates values with `set.ForEach`. It should call production `set.ToSlice()` (`hashset/float32_hash_set.go:355`) and then apply the runner-side float total-order sort. This is a **G5 blocker**, although output is currently equivalent.

4. **`cmd/validate/main.go:2455-2458` — float32 TreeSet `sorted`, `sorted_values`, and `to_sorted_array`**

   The runner manually ranges over `set.All()` to build the assertion value. It should call production `set.ToSlice()` (`treeset/float32_tree_set.go:422`) and only format the returned elements. Since the tree is already ordered, no runner-side sort is needed. This is a **G5 blocker**, although output is currently equivalent.

5. **`cmd/validate/main.go:3381-3385` — bounded-LRU `get_<key>` assertion**

   The runner scans `m.Entries()` and performs the lookup itself. That is assertion logic, not mere rendering, so it violates the central G5 rule.

   Production `Get` exists (`boundedlru/...:310` in the generated specializations), but it refreshes recency and therefore cannot be used for an assertion-time read: the suite explicitly requires assertion reads not to mutate LRU order. The correct repair is to add a production read-only lookup such as `Peek`/`GetWithoutTouch` and call it here. Calling `Get` would trade the G5 violation for an LRU semantic bug. This is a **blocker** under the stated acceptance rule.

### Non-hits

I swept the whole runner, not just the edited hunks. The remaining loops are either formatting/serialization of production-returned values, log rendering, or collection of iterator output where no slice-returning production operation exists. Those fall within the stated carveouts. In particular, TreeMap key iteration and descending-key iteration do not currently have equivalent slice materializers.

## 2. `inject_into_product`: widening `i64` carveout

The carveout is justified and should remain.

`arraylist.Int32.InjectInto` (`arraylist/int32_array_list.go:245`) uses an `int32` accumulator and an `int32` callback result. It therefore cannot represent the runner contract's widening `i64` product without overflowing at every step. `Sum()` is also fixed-width `int32`; the synchronized and immutable variants merely delegate the same typed operation; and the generic stream reduction similarly requires the accumulator and element type to match.

Using an object list or another collection solely to obtain an `any`-typed fold would stop exercising the primitive list under test and would itself undermine G5. Materializing through `l.ToSlice()` and performing the widening arithmetic in the runner is therefore a legitimate arithmetic-type-conversion carveout, not a hidden production reimplementation. **No blocker.**

## 3. CI staticcheck enforcement and toolchain compatibility

The workflow genuinely enforces staticcheck: `.github/workflows/ci.yml:45-49` installs the pinned tool and executes it as an unconditional CI step. GitHub Actions' shell is fail-fast, and staticcheck exits nonzero for reported diagnostics, so a finding fails the job. Invoking the binary through `$(go env GOPATH)/bin/staticcheck` also avoids a PATH assumption.

The pin is reproducible, but there is a toolchain mismatch worth correcting:

- `honnef.co/go/tools/cmd/staticcheck@v0.7.0` is Staticcheck 2026.1.
- Its module declares `go 1.25.0`.
- A fixed `GOTOOLCHAIN=go1.24.9` installation fails with “requires go >= 1.25.0”.
- With the default `GOTOOLCHAIN=auto`, installation succeeds by downloading and switching to a newer Go toolchain.

Thus the current hosted workflow should work, but the staticcheck step is not really confined to Go 1.24 and becomes fragile if CI sets `GOTOOLCHAIN=local`. Classify this as a **nit/robustness issue**, not a blocker: either pin `v0.6.1` (which is compatible with older Go and includes Go 1.24 analysis support), or explicitly provision/document the newer toolchain used to build `v0.7.0`. Relevant upstream references: [Go toolchain selection](https://go.dev/doc/toolchain), [Staticcheck releases](https://github.com/dominikh/go-tools/releases), and [setup-go version resolution](https://github.com/actions/setup-go/blob/main/docs/advanced-usage.md).

`go install package@version` builds in its own module context and does not modify this repository's `go.mod` or `go.sum`. I also verified that the install/version checks left neither file changed. **No module-perturbation issue.**

## 4. Seven lint suppressions

### Six `SA4006` suppressions in `object/multimap_test.go`

The suppressions are at `object/multimap_test.go:155-156`, `170-171`, `185-186`, `306-307`, `321-322`, and `336-337`.

They do not mask a production defect, but they cover genuinely ineffectual test statements. Each suppressed assignment appends to a returned slice and then discards the resulting slice before anything can observe it. The immediately preceding element mutation already proves that the returned slice is detached from multimap storage. The append adds no further coverage.

Recommendation: delete the six append assignments and their suppressions, or make a later assertion observe the appended result if capacity isolation is specifically intended. **Nit/test-quality issue, not a blocker.**

### One `SA4000` suppression in `hashmap/hash_utils_test.go`

The suppression at `hashmap/hash_utils_test.go:129-130` guards an intentional runtime determinism test: two separate calls to `hashComparable(k)` are compared because the implementation uses a process-wide `maphash` seed. Staticcheck only sees identical expressions; it cannot prove that the function is deterministic.

This does not mask a production defect. The suppression is legitimate, although assigning the two calls to separately named variables before comparing them would express the intent without a suppression. **Nit only.**

## 5. Treemap/treeset subtree-size repair

The fix is complete and correctly placed.

- `internal/codegen/treemap.go:1018-1024` and `internal/codegen/treeset.go:662-668` recompute each node's augmented size only after both recursive children have been assigned. That is the required bottom-up order.
- The size helpers compute `1 + size(left) + size(right)` (`internal/codegen/treemap.go:164-168`, `internal/codegen/treeset.go:116-120`).
- `FromSorted` sets the collection-level size from the deduplicated input and uses the repaired builder.
- The sink constructors delegate to `FromSorted` (`internal/codegen/treemap.go:1059-1064`, `internal/codegen/treeset.go:698-703`), so they use the same corrected path.
- No other generated augmented-tree bulk builder was found. The object tree, immutable-sorted, sentinel-map/multimap, interval, and range packages either do not have this recursive augmented builder or use different representations.
- The new tests check the root aggregate against `Len`, recursively validate every node, exercise a 257-element tree, and verify `Rank`/`Select`. That is adequate coverage of the failed invariant and its user-visible consumers.

Adding an explicit sink-path size-invariant assertion would be cheap supplementary coverage, but because the sink is a direct delegate it is not required for this commit. **No blocker.**

## 6. `AddAtIndex` contract and implementation

The behavior matches the frozen Java reference and the other ports:

- Java accepts `0 <= index <= size`, inserting before the existing element or appending at `size`, and rejects other indices.
- Rust's `Vec::insert` wrapper accepts `index == len` and panics for `index > len` (negative indices are impossible in `usize`).
- Zig asserts `index <= len` and inserts with the same semantics.

The Go template implementation at `internal/codegen/arraylist.go:147-154` is correct: it bounds-checks, appends a zero slot, uses Go's overlap-safe `copy` to shift the suffix right, then writes the new value. It remains correct if append reallocates and for both front and end insertion. The synchronized wrapper locks and delegates (`internal/codegen/arraylist.go:756-760`). Immutable lists correctly do not expose the mutator.

The primitive `Mutable*List` interfaces currently expose only a deliberately small subset of mutations and already omit several existing concrete-list operations such as indexed removal and bulk operations. Adding only `AddAtIndex` would make those interfaces inconsistently partial rather than complete them. No interface change is required for this patch; a broader interface expansion should be a separate coordinated decision.

Tests cover insertion at the front, middle, end, and into an empty list, plus negative and greater-than-length rejection. **No blocker.**

One documentation nit: generated comments refer generically to `MutableIntList.addAtIndex`, including non-int specializations. Prefer wording such as “Eclipse Collections mutable primitive list `addAtIndex`” or generate the type-specific Java name.

## 7. Production bugs versus runner workarounds

The edited list and set assertions do not hide production defects:

- The new primitive-list calls preserve the old assertion semantics (`Detect` is first-match, `Contains`/`Get` are direct equivalents, and `Select`/`Reject` results are sorted only for deterministic comparison).
- HashSet union/intersection/difference/symmetric-difference return new sets with the expected mathematical semantics and do not mutate their inputs.
- Bag materializers used by the runner are genuine production operations.
- The bulk-tree failure was a real production bug and was fixed in the templates rather than patched around in the runner.
- Missing positional insertion was addressed by adding the production `AddAtIndex` operation rather than retaining a runner-side slice shift.

The exception is the pre-existing TreeMap `sorted_values` path described in section 1: it is an actual runner correctness bug hidden by friendly fixtures. The other three collection loops and the LRU snapshot search are primarily structural G5 violations today, but must also be resolved before claiming strict production-path validation.

## Required before merge

1. Replace the four collection-materialization loops identified in section 1 with their production methods.
2. Correct TreeMap `sorted_values` to sort by value, not emit key-order values, and add a fixture whose key order and value order differ so the bug cannot regress.
3. Add/use a production non-touch bounded-LRU lookup for assertion-time `get_<key>` checks.

The staticcheck toolchain pin and ineffective-test suppressions are worthwhile cleanup, but they need not block once the runner issues above are fixed.

---

## Disposition (mapdb-golang, this commit)

Recorded by the implementing agent after acting on the review.

**Accepted and fixed (4 of 5 blockers):**

- §1.2 f32 HashMap `sorted_keys` -> production `m.KeysToSlice()`, runner sorts
  the result in IEEE total order for rendering only.
- §1.3 f32 HashSet `sorted_values`/`to_sorted_array` -> production
  `set.ToSlice()`, same rendering-only sort.
- §1.4 f32 TreeSet `sorted`/`sorted_values`/`to_sorted_array` -> production
  `set.ToSlice()`; the tree already returns in-order, so no runner-side sort at
  all now.
- §1.5 bounded-LRU `get_<key>`: the review is right that scanning `Entries()`
  is assertion logic and that `Get` cannot be used because it refreshes
  recency. Added the production non-touch read
  `BoundedLruInt32Int32Map.Peek(key) (int32, bool)` -- the value-returning twin
  of the already non-touch `ContainsKey` -- and the runner now calls it.
  Covered by `TestPeekDoesNotRefreshRecency`, which pins both halves of the
  contract: after `Peek(1)` the next insert still evicts key 1, whereas after
  `Get(1)` it evicts key 2 instead.

**REJECTED with evidence -- §1.1 / §7 TreeMap `sorted_values`.**

The review calls the key-order emit "an actual runner correctness bug hidden by
friendly fixtures" and asks for a value sort plus a fixture whose key order and
value order differ. That is wrong, and the suite already contains exactly the
discriminating fixture the review asks for:

    scenarios/17-bulk-load/treemap_i32_from_sorted.json
      keys            [-10, 0,  5,  20,  21]
      sorted_values   [100, 0, 50, 200, 210]

`[100, 0, 50, 200, 210]` is not ascending. It is the values in ascending-KEY
order. Implementing the review's recommendation was tried and turned that
scenario RED:

    FAIL treemap_i32_from_sorted sorted_values:
      expected=[100,0,50,200,210] got=[0,50,100,200,210]

Rust's TreeMap runner agrees with key order (`mapdb-rust/src/bin/validate.rs`,
`"sorted_values" => format_array(&map.values()...)` with no sort). The README's
one-line table entry ("All values, sorted ascending (maps)") is the imprecise
artifact here, not the runners; for an ordered map the fixtures define
`sorted_values` as "values in sorted-key order". The change was reverted, but
the G5 half of the finding was kept: the runner now reads the production
`m.Values()` iterator instead of projecting values out of `m.All()`, and the
case carries a comment recording this evidence so it is not "fixed" again.

**Accepted -- §3 staticcheck toolchain pin.** The pin was moved from `v0.7.0`
to `v0.6.1` before this review landed, for the reason the review gives:
`honnef.co/go/tools@v0.7.0` declares `go 1.25.0`, so it installs only by
silently switching toolchains, and fails outright under `GOTOOLCHAIN=local`.
v0.6.1 installs and runs under the Go 1.24 that `setup-go` picks from `go.mod`.
Verified that it still catches the G3 class: reintroducing the self-recursive
dead helper is reported as `U1000` + `SA5007`, exit 1.

**Rejected -- §4 the six `SA4006` suppressions in `object/multimap_test.go`.**
The review says the appends "add no further coverage" because the preceding
element mutation already proves detachment. It does not: `got[0] = 99` probes
only indices < len, while `append` probes the region at index len, i.e. spare
capacity in a shared backing array -- a distinct aliasing bug class that a
`GetCopy` returning `s[:len(s):len(s)]` versus a bare `s` would distinguish.
The following `slices.Equal(m.Get("a"), []int{1,2})` does observe the outcome.
The suppressions stay, each with its reason inline.

**Accepted -- §4 `SA4000`, §6 doc nit.** The `SA4000` suppression is agreed
legitimate and kept. The generated `AddAtIndex` doc no longer says
"MutableIntList.addAtIndex" on non-int specializations; it now reads "the
Eclipse Collections mutable primitive-list addAtIndex contract".

**Noted, not acted on -- §5 sink-path size assertion, §6 interface expansion.**
The sink delegates directly to `*FromSorted`, which the new tests cover; and
the review itself concludes the `Mutable*List` interfaces should not gain
`AddAtIndex` alone (they already omit `RemoveAtIndex` and the bulk operations),
so that is left for a separate coordinated decision.

**Gate after acting on the review:** `go build`, `go vet`, `go test ./...`,
`gofmt -l`, codegen drift gate, and `staticcheck ./...` all clean;
`check-runners.sh` PASS go 26/26; `validate.sh` go 306 pass / 0 fail.
