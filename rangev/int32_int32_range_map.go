// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package rangev

// Int32Int32RangeMap is a mutable piecewise mapping from disjoint non-empty
// Int32Ranges to int32 values.
//
// Like Int32RangeSet, a RangeMap is always maximally merged — but per value:
// Put is last-writer-wins (it clips/splits every overlapping prior entry) and
// then coalesces the inserted entry with connected neighbours holding an equal
// value. A different value is a barrier and is never absorbed or crossed. The
// normal form therefore carries a global invariant: no two connected entries
// hold an equal value.
//
// Divergence from Guava: TreeRangeMap.put does not coalesce; coalescing lives
// in a separate putCoalescing. We fold it into Put and do not expose
// PutCoalescing. Guava's split is a compatibility retrofit (RangeMap is
// @since 14.0, putCoalescing @since 22.0, by which point put's behaviour was
// observable through asMapOfRanges() and could not be changed); we have no such
// constraint. See spec/features/range-set-map.md §Coalescing.
//
// Every clip / split / merge / ordering decision reduces to the side-aware cut
// comparisons of Int32Range; there is no ±1 endpoint arithmetic (the
// INT_MIN/INT_MAX overflow trap).
//
// The backing is a flat slice of entries kept in normal form: entry ranges
// non-empty, pairwise disjoint, each value mapped by at most one point,
// ascending by lower cut. The order is unobservable beyond AsMapOfRanges; a
// tree keyed by lower cut would give identical results.
type Int32Int32RangeMap struct {
	// Normal form: non-empty, pairwise disjoint, ascending by lower cut.
	entries []Int32Int32Entry
}

// Int32Int32Entry is one (Range, value) pair of an Int32Int32RangeMap.
type Int32Int32Entry struct {
	Range Int32Range
	Value int32
}

// NewInt32Int32RangeMap returns an empty range map.
func NewInt32Int32RangeMap() *Int32Int32RangeMap {
	return &Int32Int32RangeMap{}
}

// Put assigns value to every point of range, last-writer-wins over any prior
// overlap. Existing entries are clipped to the parts outside range (a
// straddling entry splits into two, both keeping the old value); the new
// (range, value) is then coalesced with any connected neighbour holding an
// equal value and inserted. A different value is a barrier. A cut-empty range
// is a no-op, decided before any clipping.
func (m *Int32Int32RangeMap) Put(r Int32Range, value int32) {
	if r.IsEmpty() {
		return
	}
	m.clipOut(r)

	// Coalesce outward from the insertion position. Because the normal form is
	// maintained by every Put, AT MOST ONE entry per side is absorbable: if the
	// neighbour is absorbed, the entry beyond it was already either disconnected
	// from it or differently-valued, and stays so against the grown range. Each
	// loop therefore runs at most once. They are loops rather than ifs so a
	// normal form violated by a bug elsewhere degrades into a correct (if
	// slower) result instead of a malformed map.
	pos := m.insertionPoint(r)
	merged := r

	lo := pos
	for lo > 0 {
		e := m.entries[lo-1]
		if e.Value != value || !e.Range.IsConnected(merged) {
			break
		}
		merged = e.Range.Span(merged)
		lo--
	}

	hi := pos
	for hi < len(m.entries) {
		e := m.entries[hi]
		if e.Value != value || !e.Range.IsConnected(merged) {
			break
		}
		merged = e.Range.Span(merged)
		hi++
	}

	// Replace entries[lo:hi] with the single merged entry.
	rest := append([]Int32Int32Entry{{Range: merged, Value: value}}, m.entries[hi:]...)
	m.entries = append(m.entries[:lo], rest...)
}

// Get returns the value mapped at value and true, or (zero, false) if
// uncovered.
func (m *Int32Int32RangeMap) Get(value int32) (int32, bool) {
	for _, e := range m.entries {
		if e.Range.Contains(value) {
			return e.Value, true
		}
	}
	return 0, false
}

// GetEntry returns the (range, value) entry covering value and true, or
// (zero, zero, false) if uncovered.
func (m *Int32Int32RangeMap) GetEntry(value int32) (Int32Range, int32, bool) {
	for _, e := range m.entries {
		if e.Range.Contains(value) {
			return e.Range, e.Value, true
		}
	}
	return Int32Range{}, 0, false
}

// Remove unmaps range, splitting any entry straddling either boundary (both
// fragments keep the old value). A cut-empty range is a no-op.
func (m *Int32Int32RangeMap) Remove(r Int32Range) {
	if r.IsEmpty() {
		return
	}
	m.clipOut(r)
}

// Span returns the minimum range enclosing all entry ranges and true, or
// (zero, false) on an empty map.
func (m *Int32Int32RangeMap) Span() (Int32Range, bool) {
	if len(m.entries) == 0 {
		return Int32Range{}, false
	}
	return Int32Range{
		lower: m.entries[0].Range.lower,
		upper: m.entries[len(m.entries)-1].Range.upper,
	}, true
}

// SubRangeMap returns a NEW independent Int32Int32RangeMap restricted to view
// (each entry range clipped to view, values preserved).
func (m *Int32Int32RangeMap) SubRangeMap(view Int32Range) *Int32Int32RangeMap {
	out := make([]Int32Int32Entry, 0, len(m.entries))
	for _, e := range m.entries {
		if i, ok := e.Range.Intersection(view); ok && !i.IsEmpty() {
			out = append(out, Int32Int32Entry{Range: i, Value: e.Value})
		}
	}
	return &Int32Int32RangeMap{entries: out}
}

// AsMapOfRanges returns the canonical disjoint (range, value) entries,
// ascending by lower cut, as a fresh slice (mutating it does not affect the
// map).
func (m *Int32Int32RangeMap) AsMapOfRanges() []Int32Int32Entry {
	out := make([]Int32Int32Entry, len(m.entries))
	copy(out, m.entries)
	return out
}

// IsEmpty reports whether the map has no entries.
func (m *Int32Int32RangeMap) IsEmpty() bool {
	return len(m.entries) == 0
}

// Clear removes all entries.
func (m *Int32Int32RangeMap) Clear() {
	m.entries = m.entries[:0]
}

// ---- internals -----------------------------------------------------------

// clipOut clips every entry to the parts outside r (the Remove /
// overlap-resolution split). A straddling entry becomes two fragments; an entry
// fully inside r is dropped. Pure cut arithmetic — the boundary cuts flip,
// never ±1. Abutment alone (cut-empty intersection) leaves an entry untouched.
func (m *Int32Int32RangeMap) clipOut(r Int32Range) {
	out := make([]Int32Int32Entry, 0, len(m.entries)+1)
	for _, e := range m.entries {
		i, ok := e.Range.Intersection(r)
		if ok && !i.IsEmpty() {
			// Left fragment below the removed range's lower cut.
			if e.Range.lower.cmp(r.lower) < 0 {
				out = append(out, Int32Int32Entry{
					Range: Int32Range{lower: e.Range.lower, upper: r.lower},
					Value: e.Value,
				})
			}
			// Right fragment above the removed range's upper cut.
			if r.upper.cmp(e.Range.upper) < 0 {
				out = append(out, Int32Int32Entry{
					Range: Int32Range{lower: r.upper, upper: e.Range.upper},
					Value: e.Value,
				})
			}
		} else {
			out = append(out, e)
		}
	}
	m.entries = out
}

// insertionPoint returns the ascending-by-lower-cut index at which r belongs:
// the first index whose lower cut is above r's. Callers must have already
// cleared the overlap (via clipOut), so r is disjoint from every remaining entry
// and every entry below the returned index lies strictly to its left.
func (m *Int32Int32RangeMap) insertionPoint(r Int32Range) int {
	for i := range m.entries {
		if m.entries[i].Range.lower.cmp(r.lower) > 0 {
			return i
		}
	}
	return len(m.entries)
}
