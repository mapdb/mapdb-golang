// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package treemap

// The NavigableMap Range surface (spec/features/navigable-map.md) consumes the
// rangev.Int32Range value model. rangev is the int32-only v1 specialisation, so
// these methods exist only on the int32-keyed Int32Int32 map; the wider matrix
// widens later exactly as Interval/Range did. They live in this hand-written
// extension file rather than the codegen template so the uniform 49 generated
// map sources are not polluted with one-off int32 logic or a conditional rangev
// import. The key-bound iterators (RangeKeys/SubMap/HeadMap/TailMap) on the
// generated type are unchanged; this is the additive Range-value surface.
//
// Range membership is EXACTLY rangev.Int32Range.Contains(key): e.g. open(1, 2)
// over int32 matches no key yet is a valid, non-cut-empty range — emptiness is
// never inferred from the cuts. All Range query results are materialized
// snapshots taken at call time; they are read-only and never mutate the map.

import "github.com/mapdb/mapdb-golang/rangev"

// Int32Int32Entry is a key-value pair returned by the entry-form Range queries.
type Int32Int32Entry struct {
	Key   int32
	Value int32
}

// RangeKeysIn returns the keys whose key is in the range, ascending. The result
// is a materialized snapshot.
func (m *Int32Int32) RangeKeysIn(r rangev.Int32Range) []int32 {
	out := []int32{}
	for k := range m.Keys() {
		if r.Contains(k) {
			out = append(out, k)
		}
	}
	return out
}

// RangeEntriesIn returns the entries whose key is in the range, ascending. The
// result is a materialized snapshot.
func (m *Int32Int32) RangeEntriesIn(r rangev.Int32Range) []Int32Int32Entry {
	out := []Int32Int32Entry{}
	for k, v := range m.All() {
		if r.Contains(k) {
			out = append(out, Int32Int32Entry{Key: k, Value: v})
		}
	}
	return out
}

// DescendingRangeKeys returns the keys whose key is in the range, descending.
func (m *Int32Int32) DescendingRangeKeys(r rangev.Int32Range) []int32 {
	out := []int32{}
	for k := range m.DescendingKeys() {
		if r.Contains(k) {
			out = append(out, k)
		}
	}
	return out
}

// DescendingRangeEntries returns the entries whose key is in the range,
// descending.
func (m *Int32Int32) DescendingRangeEntries(r rangev.Int32Range) []Int32Int32Entry {
	out := []Int32Int32Entry{}
	for k, v := range m.DescendingMap() {
		if r.Contains(k) {
			out = append(out, Int32Int32Entry{Key: k, Value: v})
		}
	}
	return out
}

// DescendingEntries returns all entries, descending. Snapshot at call time.
func (m *Int32Int32) DescendingEntries() []Int32Int32Entry {
	out := []Int32Int32Entry{}
	for k, v := range m.DescendingMap() {
		out = append(out, Int32Int32Entry{Key: k, Value: v})
	}
	return out
}

// SubMapRange returns a new independent map of the entries whose key is in the
// range. Mutating the snapshot never affects the original and vice versa (it is
// a materialized copy, not a live view). The snapshot keeps the same ascending
// int32 key order as the source.
func (m *Int32Int32) SubMapRange(r rangev.Int32Range) *Int32Int32 {
	out := NewInt32Int32()
	for k, v := range m.All() {
		if r.Contains(k) {
			out.Put(k, v)
		}
	}
	return out
}

// RemoveRange removes every entry whose key is in the range and returns the
// count removed. A range that matches nothing is a no-op returning 0.
func (m *Int32Int32) RemoveRange(r rangev.Int32Range) int {
	victims := m.RangeKeysIn(r)
	for _, k := range victims {
		m.Remove(k)
	}
	return len(victims)
}
