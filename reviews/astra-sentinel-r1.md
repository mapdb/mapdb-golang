# Astra sentinel-map review — round 1

Verdict: **FIX**

Reviewed the uncommitted generator change, generated integer/char and float
variants, and `sentinelhashmap/tombstone_test.go` on 2026-09-25 against the
problem record in `../todo/astra25/06-go-sentinel-fix.md`.

## Blocking finding

**P1 — Same-capacity cleanup becomes an O(capacity) operation on every churn
cycle near the live load limit.**

References: `internal/codegen/sentinelhashmap.go:322`,
`internal/codegen/sentinelhashmap.go:744`, and the allocation/reinsertion at
`internal/codegen/sentinelhashmap.go:764`. This is emitted into all 49 maps.

The physical cleanup threshold equals the live growth threshold. At capacity
C with 3C/4−1 live regular keys, remove one key and put it back. The next put
sees 3C/4−2 live keys and one tombstone. Growth is false, cleanup is true, so
it allocates two new arrays and reinserts the entire remaining map. The map
returns to its initial live count; every repetition rebuilds again. Even
reinserting the exact same key cannot reuse its tombstone because cleanup
runs before probing. A routine stable-size workload therefore changes from
constant expected work per cycle to linear work plus two table allocations
per cycle.

Independent bounded reproduction used capacity 4096, keys 2 through 3072
(3071 live entries), and 100 repetitions of `Remove(2); Put(2, 2)`:

| Code | Rebuilds in 100 cycles | Allocations per cycle |
| --- | ---: | ---: |
| Current working tree | 100 | 2 |
| HEAD generated map via Go overlay | 0 | 0 |

Backing-array identity detected rebuilds; `testing.AllocsPerRun` measured
allocations. No timing threshold was involved. The standalone reproduction
and original-source overlay are in `/tmp/astra-sentinel-review-qf2kim4g`.

Required repair: ensure cleanup leaves capacity-proportional headroom before
another cleanup can be required under stable-size churn. For example, retain
the 3/4 live growth rule but use a higher physical occupancy threshold below
full capacity (such as 7/8), or another policy with equivalent amortization.
Merely moving cleanup after probing addresses same-key reuse but needs care
to avoid the same repeated rebuild for replacement keys landing in new slots.
Preserve the bounded probes, sentinel exclusions, and tombstone bookkeeping.
Regenerate all variants.

Required regression: exercise a table at the maximum regular live count
before growth, repeatedly delete/reinsert, and assert that backing arrays
are not replaced every cycle. Include churn to different keys as well as
same-key reuse, and check content, length, capacity, and a bounded number of
rebuilds across enough cycles. Prefer deterministic structural checks over
elapsed-time assertions. The existing sparse churn tests do not catch this
case. The original no-empty-slot termination tests must continue to pass.

## Other review results

No additional blocking correctness issue was found. Get and remove stop
after one complete probe; put checks for an existing key beyond tombstones
before reusing one. Tombstone increment, reuse decrement, clear reset, and
rebuild reset are consistent. Rebuild preserves dedicated 0/1 sentinel
storage, and float variants also preserve negative zero separately and use
bit equality for NaN keys. Value-type differences do not alter the new
occupancy logic.

The existing tests meaningfully cover the original churn hang, unsuccessful
full-table get/remove, sentinel survival, NaN/signed-zero behavior, and
updates beyond tombstones. The full physical table test's put triggers
cleanup before probing; it does not directly exercise the defensive full
probe insertion fallback. That fallback is unreachable with correctly
maintained current occupancy invariants, so this is not a blocker.

Independent validation: `go test -timeout 120s ./sentinelhashmap` passed
(Go reported cached results); the standalone current-versus-HEAD reproduction
above was executed. Production and test sources were not edited by this review.
