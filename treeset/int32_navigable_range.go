// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package treeset

// The NavigableSet Range surface (spec/features/navigable-map.md) consumes the
// rangev.Int32Range value model. rangev is the int32-only v1 specialisation, so
// these methods exist only on the int32 Int32 set; the wider matrix widens later
// exactly as Interval/Range did. They live in this hand-written extension file
// rather than the codegen template so the uniform 7 generated set sources are
// not polluted with one-off int32 logic or a conditional rangev import. The
// key-bound iterator (RangeValues) on the generated type is unchanged; this is
// the additive Range-value surface.
//
// Range membership is EXACTLY rangev.Int32Range.Contains(element): e.g.
// open(1, 2) over int32 matches no element yet is a valid, non-cut-empty range.
// All Range query results are materialized snapshots taken at call time; they
// are read-only and never mutate the set.

import "github.com/mapdb/mapdb-golang/rangev"

// RangeElements returns the elements in the range, ascending. The result is a
// materialized snapshot.
func (s *Int32) RangeElements(r rangev.Int32Range) []int32 {
	out := []int32{}
	for v := range s.All() {
		if r.Contains(v) {
			out = append(out, v)
		}
	}
	return out
}

// DescendingRangeElements returns the elements in the range, descending.
func (s *Int32) DescendingRangeElements(r rangev.Int32Range) []int32 {
	asc := s.RangeElements(r)
	out := make([]int32, len(asc))
	for i, v := range asc {
		out[len(asc)-1-i] = v
	}
	return out
}

// Descending returns all elements, descending. Snapshot at call time.
func (s *Int32) Descending() []int32 {
	asc := s.ToSlice()
	out := make([]int32, len(asc))
	for i, v := range asc {
		out[len(asc)-1-i] = v
	}
	return out
}

// SubSet returns a new independent set of the elements in the range. Mutating
// the snapshot never affects the original and vice versa (it is a materialized
// copy, not a live view). The snapshot keeps the same ascending int32 order as
// the source.
func (s *Int32) SubSet(r rangev.Int32Range) *Int32 {
	out := NewInt32()
	for v := range s.All() {
		if r.Contains(v) {
			out.Add(v)
		}
	}
	return out
}

// RemoveRange removes every element in the range and returns the count removed.
// A range that matches nothing is a no-op returning 0.
func (s *Int32) RemoveRange(r rangev.Int32Range) int {
	victims := s.RangeElements(r)
	for _, v := range victims {
		s.Remove(v)
	}
	return len(victims)
}
