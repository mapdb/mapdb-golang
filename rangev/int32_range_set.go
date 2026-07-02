// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package rangev

// Int32RangeSet is a mutable, auto-coalescing set of cut-regions over int32,
// stored as disjoint, non-empty, pairwise non-connected Int32Ranges (the
// normal form). It auto-coalesces on Add: two ranges merge iff they are
// IsConnected — broader than mere overlap, because an abutment (a cut-touch,
// e.g. [1,3) & [3,5)) is also connected. Every coalescing / split / complement
// / ordering decision reduces to the side-aware cut comparisons of Int32Range;
// there is no (value, inclusive) boolean reasoning and no ±1 endpoint
// arithmetic (the INT_MIN/INT_MAX overflow trap).
//
// Cut-region, not integer-value-set: because Phase 0 has no DiscreteDomain, a
// RangeSet models cut-regions, not the int32 values they happen to contain.
// Add(Open(1, 2)) over int32 produces a NON-EMPTY set whose single stored
// range (1, 2) is cut-non-empty even though Contains is false for every int32.
// So {} and {(1, 2)} are distinct RangeSets. Every set-level predicate
// (IsEmpty, canonicality, Complement, Intersects, Span) is defined on the
// stored cut-regions; only the point queries (Contains / RangeContaining) ask
// about an actual int32.
//
// The backing is a flat slice kept in the normal form (non-empty, pairwise
// non-connected, ascending by lower cut). The order is unobservable beyond
// AsRanges; a tree keyed by lower cut would give identical results.
type Int32RangeSet struct {
	// Normal form: non-empty, pairwise non-connected, ascending by lower cut.
	ranges []Int32Range
}

// NewInt32RangeSet returns an empty range set.
func NewInt32RangeSet() *Int32RangeSet {
	return &Int32RangeSet{}
}

// Add unions range in, coalescing ALL connected stored ranges. A cut-empty
// range (e.g. ClosedOpen(5, 5)) is a no-op, decided by Int32Range.IsEmpty
// (cut-empty), never by discrete cardinality — Add(Open(1, 2)) over int32
// STORES the range. The merged range keeps the outer cuts of every connected
// member (the cut min/max, no ±1 math).
func (s *Int32RangeSet) Add(r Int32Range) {
	if r.IsEmpty() {
		return
	}
	// Merge r with every connected stored range, spanning all of them.
	// Connectivity (overlap OR abutment) is the coalescing predicate.
	merged := r
	out := make([]Int32Range, 0, len(s.ranges)+1)
	for _, existing := range s.ranges {
		if existing.IsConnected(merged) {
			merged = existing.Span(merged)
		} else {
			out = append(out, existing)
		}
	}
	// Insert merged at its ascending-by-lower-cut position.
	pos := len(out)
	for i := range out {
		if out[i].lower.cmp(merged.lower) > 0 {
			pos = i
			break
		}
	}
	out = append(out, Int32Range{})
	copy(out[pos+1:], out[pos:])
	out[pos] = merged
	s.ranges = out
}

// AddAll Adds each range; the final normal form is order-independent.
func (s *Int32RangeSet) AddAll(ranges ...Int32Range) {
	for _, r := range ranges {
		s.Add(r)
	}
}

// Remove subtracts range, splitting any stored range straddling either
// boundary. A cut-empty range is a no-op. The split is pure cut arithmetic —
// the boundary cuts flip (Remove(ClosedOpen(4, 7)) from [1,9] leaves [1,4) and
// [7,9]), never ±1. Abutment alone (cut-empty intersection) does not split.
func (s *Int32RangeSet) Remove(r Int32Range) {
	if r.IsEmpty() {
		return
	}
	out := make([]Int32Range, 0, len(s.ranges)+1)
	for _, existing := range s.ranges {
		i, ok := existing.Intersection(r)
		if ok && !i.IsEmpty() {
			// Left fragment: existing below the removed range's lower cut.
			if existing.lower.cmp(r.lower) < 0 {
				out = append(out, Int32Range{lower: existing.lower, upper: r.lower})
			}
			// Right fragment: existing above the removed range's upper cut.
			if r.upper.cmp(existing.upper) < 0 {
				out = append(out, Int32Range{lower: r.upper, upper: existing.upper})
			}
		} else {
			out = append(out, existing)
		}
	}
	s.ranges = out
}

// Contains reports whether value falls in some stored range. This is the only
// integer-point predicate — (1, 2) correctly contains no int32.
func (s *Int32RangeSet) Contains(value int32) bool {
	for _, r := range s.ranges {
		if r.Contains(value) {
			return true
		}
	}
	return false
}

// RangeContaining returns the stored range containing value and true, or
// (zero, false) if none.
func (s *Int32RangeSet) RangeContaining(value int32) (Int32Range, bool) {
	for _, r := range s.ranges {
		if r.Contains(value) {
			return r, true
		}
	}
	return Int32Range{}, false
}

// Encloses reports whether some single stored range encloses range
// (cut-defined Int32Range.Encloses). A set covering {[1,3), [5,9)} does NOT
// enclose [2,6) — no single stored range does.
func (s *Int32RangeSet) Encloses(r Int32Range) bool {
	for _, existing := range s.ranges {
		if existing.Encloses(r) {
			return true
		}
	}
	return false
}

// EnclosesAll reports whether Encloses holds for every argument.
func (s *Int32RangeSet) EnclosesAll(ranges ...Int32Range) bool {
	for _, r := range ranges {
		if !s.Encloses(r) {
			return false
		}
	}
	return true
}

// Intersects reports whether range has a cut-non-empty intersection with some
// stored range — pure cut algebra. An abutment is NOT an intersection
// (Intersects(ClosedOpen(3, 5)) against [5,9) is false); a cut-empty query
// never intersects; but a discrete-empty-yet-cut-non-empty overlap DOES count
// (Intersects(Open(1, 2)) against stored (1, 2) is true, though no int32 lies
// in it).
func (s *Int32RangeSet) Intersects(r Int32Range) bool {
	for _, existing := range s.ranges {
		if i, ok := existing.Intersection(r); ok && !i.IsEmpty() {
			return true
		}
	}
	return false
}

// Span returns the minimum enclosing range [min lower cut, max upper cut] and
// true, or (zero, false) on an empty set.
func (s *Int32RangeSet) Span() (Int32Range, bool) {
	if len(s.ranges) == 0 {
		return Int32Range{}, false
	}
	return Int32Range{
		lower: s.ranges[0].lower,
		upper: s.ranges[len(s.ranges)-1].upper,
	}, true
}

// Complement returns a NEW independent Int32RangeSet of the cut-region gaps
// between the stored ranges over the full (-inf, +inf) domain.
// Complement(empty) = {All()}; Complement({All()}) = {}; no spurious ±inf gap
// when an end is already unbounded; the boundary side flips (closed<->open at
// the same cut value). Complement(Complement(s)) == s.
func (s *Int32RangeSet) Complement() *Int32RangeSet {
	out := make([]Int32Range, 0, len(s.ranges)+1)
	// Walking cut: the lower cut of the next gap. Starts at -inf.
	cursor := cut{kind: cutBelowAll}
	for _, r := range s.ranges {
		// Gap from cursor up to this range's lower cut, when non-empty.
		if cursor.cmp(r.lower) < 0 {
			out = append(out, Int32Range{lower: cursor, upper: r.lower})
		}
		// Next gap starts just past this range's upper cut.
		cursor = r.upper
	}
	// Trailing gap from the last upper cut to +inf, when non-empty.
	aboveAll := cut{kind: cutAboveAll}
	if cursor.cmp(aboveAll) < 0 {
		out = append(out, Int32Range{lower: cursor, upper: aboveAll})
	}
	return &Int32RangeSet{ranges: out}
}

// SubRangeSet returns a NEW independent Int32RangeSet = this set intersected
// with view (each stored range clipped to view). SubRangeSet(ClosedOpen(3, 6))
// of {[1,5), [8,9]} = {[3,5)}.
func (s *Int32RangeSet) SubRangeSet(view Int32Range) *Int32RangeSet {
	out := make([]Int32Range, 0, len(s.ranges))
	// The stored ranges are ascending and disjoint, so their clipped images
	// stay ascending, disjoint, and non-connected.
	for _, r := range s.ranges {
		if i, ok := r.Intersection(view); ok && !i.IsEmpty() {
			out = append(out, i)
		}
	}
	return &Int32RangeSet{ranges: out}
}

// AsRanges returns the canonical disjoint ranges, ascending by lower cut, as a
// fresh slice (mutating it does not affect the set).
func (s *Int32RangeSet) AsRanges() []Int32Range {
	out := make([]Int32Range, len(s.ranges))
	copy(out, s.ranges)
	return out
}

// IsEmpty reports whether the set has no stored ranges. A cut-region predicate
// — {(1, 2)} is NOT empty even though it contains no int32.
func (s *Int32RangeSet) IsEmpty() bool {
	return len(s.ranges) == 0
}

// Clear removes all ranges.
func (s *Int32RangeSet) Clear() {
	s.ranges = s.ranges[:0]
}
