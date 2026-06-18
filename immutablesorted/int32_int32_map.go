// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package immutablesorted provides compact immutable sorted map / set types
// (sorted-table-map). Keys (and, for a map, the matching values) are packed
// into contiguous parallel slices and queried by binary search -- the on-heap
// analogue of MapDB 3's SortedTableMap. We port the observable behaviour and
// the packed-array + binary-search mechanism, not the off-heap Volume / byte
// offset machinery.
//
// This is DISTINCT from the frozen-copy Immutable* wrappers (which seal a live
// structure's per-entry layout against mutation) and from interval.Int32 (a
// virtual arithmetic progression with no stored elements): an
// ImmutableInt32Int32SortedMap stores arbitrary sorted key/value pairs in a
// flat array.
//
// Layout is a single flat ascending array pair (a set has one array). MapDB 3
// paged the arrays; paging is a legal but UNOBSERVABLE implementation choice
// (lookup, iteration, range, size results are identical regardless). A flat
// array is the simplest representation that is trivially paging-invariant.
//
// Construction is the only way in. FromSorted takes a strictly-ascending
// snapshot. It PANICS (the family's trap posture, like rangev's out-of-order
// constructor and interval's bad step) unless every adjacent input pair
// satisfies keys[i-1] < keys[i] strictly: out-of-order input panics, a
// duplicate key panics (no last-wins / dedup), and (map) a keys/values length
// mismatch panics. Empty and single-element input are valid. Construction
// COPIES the input, so the built collection is a snapshot independent of the
// caller's source slices.
//
// Immutable: the types expose no Put/Add/Remove/Clear/Set -- the methods simply
// do not exist. Absence is the Go comma-ok idiom: Get/FirstKey/FloorKey/... all
// return (value, ok). Range queries consume rangev.Int32Range; membership is
// exactly rangev.Int32Range.Contains and the in-range slice is bracketed by two
// binary searches over the range's CUT semantics, never v±1 arithmetic (which
// would overflow at INT_MIN/INT_MAX).
//
// v1 ships the int32 surface (the cross-language validation universe); the
// wider matrix widens later exactly as interval and rangev did. These types are
// hand-written (a one-off int32 specialisation like rangev), not codegen.
package immutablesorted

import "github.com/mapdb/mapdb-golang/rangev"

// assertStrictlyAscendingInt32 verifies xs is strictly ascending; it panics
// otherwise. Empty and single-element slices vacuously pass. Comparison is the
// signed int32 natural order.
func assertStrictlyAscendingInt32(xs []int32) {
	for i := 1; i < len(xs); i++ {
		if xs[i-1] >= xs[i] {
			panic("immutablesorted: input must be strictly ascending (no duplicate or out-of-order keys)")
		}
	}
}

// searchInt32 returns the index of key if present (found == true), else the
// lower-bound insertion index (found == false). The midpoint is overflow-safe
// (lo + (hi-lo)/2) and every comparison is signed, so it is correct at
// INT_MIN/INT_MAX.
func searchInt32(sorted []int32, key int32) (int, bool) {
	lo, hi := 0, len(sorted)
	for lo < hi {
		mid := lo + (hi-lo)/2
		switch {
		case sorted[mid] < key:
			lo = mid + 1
		case sorted[mid] > key:
			hi = mid
		default:
			return mid, true
		}
	}
	return lo, false
}

// ImmutableInt32Int32SortedMap is a compact immutable sorted map backed by
// packed parallel int32 arrays (keys[i] -> values[i]), queried by binary
// search. Built once from strictly-ascending input via FromSortedInt32Int32;
// thereafter immutable.
type ImmutableInt32Int32SortedMap struct {
	keys   []int32
	values []int32
}

// FromSortedInt32Int32 builds a map from strictly-ascending parallel slices:
// values[i] is the value of keys[i]. The input is COPIED (a snapshot
// independent of the caller's slices).
//
// It PANICS if len(keys) != len(values), if keys are out-of-order, or if any
// key is duplicated. There is no last-wins / dedup and no silent sort. Empty
// and single-element input are valid.
func FromSortedInt32Int32(keys, values []int32) *ImmutableInt32Int32SortedMap {
	if len(keys) != len(values) {
		panic("immutablesorted: FromSorted keys/values length mismatch")
	}
	assertStrictlyAscendingInt32(keys)
	k := make([]int32, len(keys))
	copy(k, keys)
	v := make([]int32, len(values))
	copy(v, values)
	return &ImmutableInt32Int32SortedMap{keys: k, values: v}
}

// Size returns the number of entries.
func (m *ImmutableInt32Int32SortedMap) Size() int { return len(m.keys) }

// IsEmpty reports whether the map has no entries.
func (m *ImmutableInt32Int32SortedMap) IsEmpty() bool { return len(m.keys) == 0 }

// Get returns the value for key and true, or (0, false) if absent.
func (m *ImmutableInt32Int32SortedMap) Get(key int32) (int32, bool) {
	if i, ok := searchInt32(m.keys, key); ok {
		return m.values[i], true
	}
	return 0, false
}

// ContainsKey reports whether key is present.
func (m *ImmutableInt32Int32SortedMap) ContainsKey(key int32) bool {
	_, ok := searchInt32(m.keys, key)
	return ok
}

// FirstKey returns the minimum key and true, or (0, false) if empty.
func (m *ImmutableInt32Int32SortedMap) FirstKey() (int32, bool) {
	if len(m.keys) == 0 {
		return 0, false
	}
	return m.keys[0], true
}

// LastKey returns the maximum key and true, or (0, false) if empty.
func (m *ImmutableInt32Int32SortedMap) LastKey() (int32, bool) {
	if len(m.keys) == 0 {
		return 0, false
	}
	return m.keys[len(m.keys)-1], true
}

// FirstEntry returns the minimum (key, value) entry, or (0, 0, false) if empty.
func (m *ImmutableInt32Int32SortedMap) FirstEntry() (int32, int32, bool) {
	if len(m.keys) == 0 {
		return 0, 0, false
	}
	return m.keys[0], m.values[0], true
}

// LastEntry returns the maximum (key, value) entry, or (0, 0, false) if empty.
func (m *ImmutableInt32Int32SortedMap) LastEntry() (int32, int32, bool) {
	if len(m.keys) == 0 {
		return 0, 0, false
	}
	last := len(m.keys) - 1
	return m.keys[last], m.values[last], true
}

// ── Point navigation (NavigableMap surface, reused verbatim) ─────────────
//
// floor <= k, ceiling >= k, lower < k (strict), higher > k (strict). All
// resolve to a single binary search; the index arithmetic never computes a
// k±1, so it is overflow-safe at the signed extremes.

// floorIndex returns the index of the greatest key <= k, or (-1, false).
func (m *ImmutableInt32Int32SortedMap) floorIndex(k int32) (int, bool) {
	i, ok := searchInt32(m.keys, k)
	if ok {
		return i, true
	}
	if i == 0 {
		return -1, false
	}
	return i - 1, true
}

// lowerIndex returns the index of the greatest key < k (strict), or (-1, false).
func (m *ImmutableInt32Int32SortedMap) lowerIndex(k int32) (int, bool) {
	i, _ := searchInt32(m.keys, k)
	if i == 0 {
		return -1, false
	}
	return i - 1, true
}

// ceilingIndex returns the index of the least key >= k, or (-1, false).
func (m *ImmutableInt32Int32SortedMap) ceilingIndex(k int32) (int, bool) {
	i, _ := searchInt32(m.keys, k)
	if i >= len(m.keys) {
		return -1, false
	}
	return i, true
}

// higherIndex returns the index of the least key > k (strict), or (-1, false).
func (m *ImmutableInt32Int32SortedMap) higherIndex(k int32) (int, bool) {
	i, ok := searchInt32(m.keys, k)
	if ok {
		i++
	}
	if i >= len(m.keys) {
		return -1, false
	}
	return i, true
}

// FloorKey returns the greatest key <= k and true, or (0, false).
func (m *ImmutableInt32Int32SortedMap) FloorKey(k int32) (int32, bool) {
	if i, ok := m.floorIndex(k); ok {
		return m.keys[i], true
	}
	return 0, false
}

// CeilingKey returns the least key >= k and true, or (0, false).
func (m *ImmutableInt32Int32SortedMap) CeilingKey(k int32) (int32, bool) {
	if i, ok := m.ceilingIndex(k); ok {
		return m.keys[i], true
	}
	return 0, false
}

// LowerKey returns the greatest key < k (strict) and true, or (0, false).
func (m *ImmutableInt32Int32SortedMap) LowerKey(k int32) (int32, bool) {
	if i, ok := m.lowerIndex(k); ok {
		return m.keys[i], true
	}
	return 0, false
}

// HigherKey returns the least key > k (strict) and true, or (0, false).
func (m *ImmutableInt32Int32SortedMap) HigherKey(k int32) (int32, bool) {
	if i, ok := m.higherIndex(k); ok {
		return m.keys[i], true
	}
	return 0, false
}

// FloorEntry returns the greatest key <= k with its value, or (0, 0, false).
func (m *ImmutableInt32Int32SortedMap) FloorEntry(k int32) (int32, int32, bool) {
	if i, ok := m.floorIndex(k); ok {
		return m.keys[i], m.values[i], true
	}
	return 0, 0, false
}

// CeilingEntry returns the least key >= k with its value, or (0, 0, false).
func (m *ImmutableInt32Int32SortedMap) CeilingEntry(k int32) (int32, int32, bool) {
	if i, ok := m.ceilingIndex(k); ok {
		return m.keys[i], m.values[i], true
	}
	return 0, 0, false
}

// LowerEntry returns the greatest key < k (strict) with its value, or (0,0,false).
func (m *ImmutableInt32Int32SortedMap) LowerEntry(k int32) (int32, int32, bool) {
	if i, ok := m.lowerIndex(k); ok {
		return m.keys[i], m.values[i], true
	}
	return 0, 0, false
}

// HigherEntry returns the least key > k (strict) with its value, or (0,0,false).
func (m *ImmutableInt32Int32SortedMap) HigherEntry(k int32) (int32, int32, bool) {
	if i, ok := m.higherIndex(k); ok {
		return m.keys[i], m.values[i], true
	}
	return 0, 0, false
}

// ── Order statistics (rank / select) ─────────────────────────────────────

// Rank returns the number of keys strictly less than key -- the 0-based
// lower-bound index key occupies (if present) or would occupy (if absent), in
// 0..=Size(). Defined for present and absent keys.
func (m *ImmutableInt32Int32SortedMap) Rank(key int32) int {
	i, _ := searchInt32(m.keys, key)
	return i
}

// SelectKey returns the i-th smallest key (0-based) and true, or (0, false) if
// i < 0 or i >= Size(). Round-trips with Rank: SelectKey(Rank(k)) == (k, true)
// for present k.
func (m *ImmutableInt32Int32SortedMap) SelectKey(i int) (int32, bool) {
	if i < 0 || i >= len(m.keys) {
		return 0, false
	}
	return m.keys[i], true
}

// SelectEntry returns the i-th smallest (key, value) entry, or (0, 0, false).
func (m *ImmutableInt32Int32SortedMap) SelectEntry(i int) (int32, int32, bool) {
	if i < 0 || i >= len(m.keys) {
		return 0, 0, false
	}
	return m.keys[i], m.values[i], true
}

// ── Iteration (ascending) ────────────────────────────────────────────────

// Keys returns the keys in ascending order (a fresh snapshot copy).
func (m *ImmutableInt32Int32SortedMap) Keys() []int32 {
	out := make([]int32, len(m.keys))
	copy(out, m.keys)
	return out
}

// Values returns the values in ascending-KEY order (paired with Keys), NOT
// sorted by value (a fresh snapshot copy).
func (m *ImmutableInt32Int32SortedMap) Values() []int32 {
	out := make([]int32, len(m.values))
	copy(out, m.values)
	return out
}

// Entries returns the (key, value) entries in ascending key order.
func (m *ImmutableInt32Int32SortedMap) Entries() []Int32Int32Entry {
	out := make([]Int32Int32Entry, len(m.keys))
	for i := range m.keys {
		out[i] = Int32Int32Entry{Key: m.keys[i], Value: m.values[i]}
	}
	return out
}

// Int32Int32Entry is a (key, value) pair returned by Entries / *Entries.
type Int32Int32Entry struct {
	Key   int32
	Value int32
}

// ── Iteration (descending) — required, not optional ──────────────────────

// DescendingKeys returns all keys, descending.
func (m *ImmutableInt32Int32SortedMap) DescendingKeys() []int32 {
	n := len(m.keys)
	out := make([]int32, n)
	for i, k := range m.keys {
		out[n-1-i] = k
	}
	return out
}

// DescendingEntries returns all (key, value) entries, descending.
func (m *ImmutableInt32Int32SortedMap) DescendingEntries() []Int32Int32Entry {
	n := len(m.keys)
	out := make([]Int32Int32Entry, n)
	for i := range m.keys {
		out[n-1-i] = Int32Int32Entry{Key: m.keys[i], Value: m.values[i]}
	}
	return out
}

// ── Range queries (consume rangev.Int32Range; membership == Contains) ─────
//
// The in-range entries form a contiguous slice of the packed array (the range
// is convex), bracketed by two binary searches via rangev.Int32Range.Bracket.
// The brackets come from the range's CUT semantics, never from v±1 arithmetic,
// so open/closed bounds at INT_MIN/INT_MAX do not overflow. Open(1,2) over
// int32 yields an empty slice (membership is Contains, never inferred
// cut-emptiness). Every result is a fresh snapshot copy.

// RangeKeys returns the keys whose key is in range, ascending.
func (m *ImmutableInt32Int32SortedMap) RangeKeys(r rangev.Int32Range) []int32 {
	lo, hi := r.Bracket(m.keys)
	out := make([]int32, hi-lo)
	copy(out, m.keys[lo:hi])
	return out
}

// RangeEntries returns the (key, value) entries whose key is in range, ascending.
func (m *ImmutableInt32Int32SortedMap) RangeEntries(r rangev.Int32Range) []Int32Int32Entry {
	lo, hi := r.Bracket(m.keys)
	out := make([]Int32Int32Entry, 0, hi-lo)
	for i := lo; i < hi; i++ {
		out = append(out, Int32Int32Entry{Key: m.keys[i], Value: m.values[i]})
	}
	return out
}

// DescendingRangeKeys returns the keys whose key is in range, descending.
func (m *ImmutableInt32Int32SortedMap) DescendingRangeKeys(r rangev.Int32Range) []int32 {
	lo, hi := r.Bracket(m.keys)
	n := hi - lo
	out := make([]int32, n)
	for i := 0; i < n; i++ {
		out[n-1-i] = m.keys[lo+i]
	}
	return out
}

// DescendingRangeEntries returns the (key, value) entries whose key is in
// range, descending.
func (m *ImmutableInt32Int32SortedMap) DescendingRangeEntries(r rangev.Int32Range) []Int32Int32Entry {
	lo, hi := r.Bracket(m.keys)
	n := hi - lo
	out := make([]Int32Int32Entry, n)
	for i := 0; i < n; i++ {
		out[n-1-i] = Int32Int32Entry{Key: m.keys[lo+i], Value: m.values[lo+i]}
	}
	return out
}
