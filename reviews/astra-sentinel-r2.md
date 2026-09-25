# Astra sentinel-map review — round 2

Verdict: **SHIP**

Reviewed the complete final uncommitted fix and the round-1 repair on
2026-09-25. No blocking correctness or performance findings remain.

## Round-1 finding resolved

`internal/codegen/sentinelhashmap.go:729` now uses a 7/8 physical occupancy
threshold, while `internal/codegen/sentinelhashmap.go:712` retains the 3/4
live growth threshold. A cleanup therefore leaves capacity-proportional
space before another cleanup can be necessary. Each deletion creates at
most one tombstone, and reuse reduces the count, so stable-size churn cannot
cause the previous rebuild on every operation.

The new `TestInt32Int32NearLimitChurnAmortizesRehash` in
`sentinelhashmap/tombstone_test.go` checks 3071 regular entries in a 4096-slot
table, plus sentinel entries. It covers both 1000 same-key delete/reinsert
cycles and 1024 replacements with different keys, bounds rebuild counts
using backing-array identity, and verifies capacity, length, and retained
contents. These assertions directly cover the round-1 regression without
relying on timing.

Independently rerunning the original round-1 standalone reproduction on the
repaired tree produced **zero rebuilds in 100 cycles and zero allocations per
cycle**, compared with 100 rebuilds and two allocations per cycle in round 1.

## Final correctness assessment

The generator and all 49 generated variants consistently track deletion,
tombstone reuse, clear, and rebuild. Occupancy excludes the dedicated 0/1
entries and, for float keys, negative zero. Rebuild retains these sentinel
values and preserves NaN keys through bit equality. Probe bounds ensure get
and remove terminate even without an empty physical slot; put preserves the
existing-key search beyond tombstones and retains its bounded-probe fallback.
The higher cleanup threshold still stays below full physical occupancy.

The original churn, forced full-table, update-past-tombstone, sentinel, and
float-special-key regression tests continue to pass. No production or test
source was changed by this review.

Independent checks:

- `go test -count=1 -timeout 120s ./sentinelhashmap` passed without cached results.
- Round-1 standalone allocation/rebuild reproduction passed with the results above.
- Checked all 49 generated maps for the final threshold, three bounded probe
  loops, deletion increment, both reuse decrements, and clear/rebuild resets.

The parent reports full build, vet, repository tests, staticcheck, and
regeneration checks passing; this review's independent checks are listed
separately above. Occasional O(capacity) cleanup and replacement-array
allocation remain inherent to the selected policy; the repeated per-cycle
cliff identified in round 1 is resolved.
