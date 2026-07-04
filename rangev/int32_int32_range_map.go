// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package rangev

// Int32Int32RangeMap is a mutable piecewise mapping from disjoint non-empty
// Int32Ranges to int32 values.
//
// Unlike Int32RangeSet, a RangeMap does NOT coalesce across different values.
// Put is last-writer-wins: it clips/splits every overlapping prior entry and
// inserts the new (range, value), but leaves adjacent equal-valued entries
// distinct. PutCoalescing is the variant that merges connected neighbours
// holding an equal value.
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
// (range, value) is then inserted. A cut-empty range is a no-op. Put does NOT
// coalesce — an adjacent equal value stays a distinct entry.
func (m *Int32Int32RangeMap) Put(r Int32Range, value int32) {
	if r.IsEmpty() {
		return
	}
	m.clipOut(r)
	m.insertEntry(r, value)
}

// PutCoalescing is like Put, then merges the inserted entry with any connected
// (overlapping or abutting) neighbour whose value equals value, producing one
// entry spanning the union. Neighbours with a different value are left
// untouched (clipped by the Put step as usual).
func (m *Int32Int32RangeMap) PutCoalescing(r Int32Range, value int32) {
	if r.IsEmpty() {
		return
	}
	m.clipOut(r)
	// Span over every connected entry with an EQUAL value, dropping them. This
	// repeats to a fixpoint rather than taking a single ascending pass: growing
	// `merged` (in EITHER direction) can newly connect it to an entry that was
	// already visited earlier in the pass, so one pass is direction-biased
	// (a chain of abutting equal-valued entries coalesces rightward but not
	// leftward). Iterating until no entry is absorbed makes the result
	// direction-independent and leaves no two connected equal-valued entries.
	merged := r
	for {
		grew := false
		out := m.entries[:0]
		for _, e := range m.entries {
			if e.Value == value && e.Range.IsConnected(merged) {
				merged = e.Range.Span(merged)
				grew = true
			} else {
				out = append(out, e)
			}
		}
		m.entries = out
		if !grew {
			break
		}
	}
	m.insertEntry(merged, value)
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

// insertEntry inserts (r, value) at its ascending-by-lower-cut position.
// Callers must have already cleared the overlap (via clipOut); r is disjoint
// from every remaining entry.
func (m *Int32Int32RangeMap) insertEntry(r Int32Range, value int32) {
	pos := len(m.entries)
	for i := range m.entries {
		if m.entries[i].Range.lower.cmp(r.lower) > 0 {
			pos = i
			break
		}
	}
	m.entries = append(m.entries, Int32Int32Entry{})
	copy(m.entries[pos+1:], m.entries[pos:])
	m.entries[pos] = Int32Int32Entry{Range: r, Value: value}
}
