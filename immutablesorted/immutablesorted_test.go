// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Native tests for the compact immutable sorted map / set. These cover the
// obligations the cross-language JSON suite cannot express (panics, snapshot
// independence, iterator key-order pairing, signed-edge brackets, round-trip
// identity) per spec/features/sorted-table-map.md §"Native-only".
package immutablesorted

import (
	"math"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/rangev"
)

const (
	intMin = math.MinInt32
	intMax = math.MaxInt32
)

// expectPanic runs fn and fails the test unless it panics.
func expectPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

// ── Construction traps (native-only; RESERVED expect_panic) ──────────────

func TestMapUnsortedInputTraps(t *testing.T) {
	expectPanic(t, "map unsorted", func() {
		FromSortedInt32Int32([]int32{10, 30, 20}, []int32{1, 3, 2})
	})
}

func TestMapDuplicateKeyTraps(t *testing.T) {
	expectPanic(t, "map duplicate", func() {
		FromSortedInt32Int32([]int32{10, 20, 20, 30}, []int32{1, 2, 99, 3})
	})
}

func TestMapLengthMismatchTraps(t *testing.T) {
	expectPanic(t, "map length mismatch", func() {
		FromSortedInt32Int32([]int32{10, 20, 30}, []int32{1, 2})
	})
}

func TestSetUnsortedInputTraps(t *testing.T) {
	expectPanic(t, "set unsorted", func() {
		FromSortedInt32([]int32{10, 30, 20})
	})
}

func TestSetDuplicateTraps(t *testing.T) {
	expectPanic(t, "set duplicate", func() {
		FromSortedInt32([]int32{10, 20, 20})
	})
}

// ── Empty + single (valid, not a trap) ───────────────────────────────────

func TestEmptyMapIsValidAndAllAbsence(t *testing.T) {
	m := FromSortedInt32Int32([]int32{}, []int32{})
	if m.Size() != 0 || !m.IsEmpty() {
		t.Fatalf("empty map: size=%d isEmpty=%v", m.Size(), m.IsEmpty())
	}
	if _, ok := m.Get(5); ok {
		t.Error("Get on empty must be absent")
	}
	if m.ContainsKey(5) {
		t.Error("ContainsKey on empty must be false")
	}
	for _, probe := range []func() (int32, bool){
		m.FirstKey, m.LastKey,
		func() (int32, bool) { return m.FloorKey(5) },
		func() (int32, bool) { return m.CeilingKey(5) },
		func() (int32, bool) { return m.LowerKey(5) },
		func() (int32, bool) { return m.HigherKey(5) },
		func() (int32, bool) { return m.SelectKey(0) },
	} {
		if _, ok := probe(); ok {
			t.Error("nav/select on empty must be absent")
		}
	}
	if m.Rank(5) != 0 {
		t.Errorf("Rank on empty = %d, want 0", m.Rank(5))
	}
	if len(m.Keys()) != 0 || len(m.DescendingKeys()) != 0 {
		t.Error("iterators on empty must be empty")
	}
	if len(m.RangeKeys(rangev.All())) != 0 {
		t.Error("range on empty must be empty")
	}
}

func TestEmptySetIsValid(t *testing.T) {
	s := FromSortedInt32([]int32{})
	if !s.IsEmpty() {
		t.Fatal("empty set not empty")
	}
	if _, ok := s.First(); ok {
		t.Error("First on empty must be absent")
	}
	if _, ok := s.Floor(0); ok {
		t.Error("Floor on empty must be absent")
	}
	if s.Rank(0) != 0 {
		t.Error("Rank on empty must be 0")
	}
	if _, ok := s.Select(0); ok {
		t.Error("Select on empty must be absent")
	}
}

func TestSingleElementIsValid(t *testing.T) {
	m := FromSortedInt32Int32([]int32{7}, []int32{700})
	if v, ok := m.Get(7); !ok || v != 700 {
		t.Errorf("Get(7) = (%d,%v)", v, ok)
	}
	if v, ok := m.FloorKey(7); !ok || v != 7 {
		t.Errorf("FloorKey(7) = (%d,%v)", v, ok)
	}
	if v, ok := m.CeilingKey(7); !ok || v != 7 {
		t.Errorf("CeilingKey(7) = (%d,%v)", v, ok)
	}
	if _, ok := m.LowerKey(7); ok {
		t.Error("LowerKey(7) must be absent")
	}
	if _, ok := m.HigherKey(7); ok {
		t.Error("HigherKey(7) must be absent")
	}
	if m.Rank(6) != 0 || m.Rank(7) != 0 || m.Rank(8) != 1 {
		t.Errorf("Rank: 6=%d 7=%d 8=%d", m.Rank(6), m.Rank(7), m.Rank(8))
	}
	if v, ok := m.SelectKey(0); !ok || v != 7 {
		t.Errorf("SelectKey(0) = (%d,%v)", v, ok)
	}
	if _, ok := m.SelectKey(1); ok {
		t.Error("SelectKey(1) must be absent")
	}
}

// ── Values()/Entries() key-order pairing (native-only obligation) ────────

func TestValuesAndEntriesPairWithKeysNotValueSorted(t *testing.T) {
	// Deliberately NON-monotonic values: a port that sorts values independently
	// would mis-pair. keys ascending {10,20,30}; values {300,100,200}.
	m := FromSortedInt32Int32([]int32{10, 20, 30}, []int32{300, 100, 200})

	keys := m.Keys()
	values := m.Values()
	if !slices.Equal(keys, []int32{10, 20, 30}) {
		t.Errorf("Keys() = %v", keys)
	}
	if !slices.Equal(values, []int32{300, 100, 200}) { // NOT [100,200,300]
		t.Errorf("Values() = %v, want key-ordered [300,100,200]", values)
	}
	// Zip-and-assert: values[i] is the value of keys[i].
	for i, k := range keys {
		if v, ok := m.Get(k); !ok || v != values[i] {
			t.Errorf("Get(%d) = (%d,%v), want %d", k, v, ok, values[i])
		}
	}
	// Entries() carries the same pairing.
	got := m.Entries()
	want := []Int32Int32Entry{{10, 300}, {20, 100}, {30, 200}}
	if !slices.Equal(got, want) {
		t.Errorf("Entries() = %v, want %v", got, want)
	}
}

// ── Snapshot independence from a mutated source buffer ───────────────────

func TestMapConstructionTakesIndependentSnapshot(t *testing.T) {
	keys := []int32{10, 20, 30}
	values := []int32{100, 200, 300}
	m := FromSortedInt32Int32(keys, values)

	// Mutate the caller's source buffers AFTER construction.
	keys[0] = 999
	values[1] = -1

	if m.Size() != 3 {
		t.Errorf("Size after source mutation = %d", m.Size())
	}
	if v, ok := m.Get(10); !ok || v != 100 {
		t.Errorf("Get(10) after source mutation = (%d,%v)", v, ok)
	}
	if v, ok := m.Get(20); !ok || v != 200 {
		t.Errorf("Get(20) after source mutation = (%d,%v)", v, ok)
	}
	if k, ok := m.FirstKey(); !ok || k != 10 {
		t.Errorf("FirstKey after source mutation = (%d,%v)", k, ok)
	}
	if m.ContainsKey(999) {
		t.Error("must not contain mutated-in 999")
	}
}

func TestSetSnapshotIndependence(t *testing.T) {
	elems := []int32{1, 2, 3}
	s := FromSortedInt32(elems)
	elems[0] = 99
	if s.Size() != 3 {
		t.Errorf("Size after source mutation = %d", s.Size())
	}
	if !s.Contains(1) {
		t.Error("must still contain 1")
	}
	if s.Contains(99) {
		t.Error("must not contain mutated-in 99")
	}
}

// Returned slices must also be independent snapshots (no aliasing of storage).
func TestReturnedSlicesAreSnapshots(t *testing.T) {
	m := FromSortedInt32Int32([]int32{1, 2, 3}, []int32{10, 20, 30})
	ks := m.Keys()
	ks[0] = -1
	if k, _ := m.FirstKey(); k != 1 {
		t.Error("mutating returned Keys() slice corrupted the map")
	}
	vs := m.Values()
	vs[0] = -1
	if v, _ := m.Get(1); v != 10 {
		t.Error("mutating returned Values() slice corrupted the map")
	}
}

// ── Select(Rank(k)) == k round-trip identity ─────────────────────────────

func TestSelectRankRoundTrip(t *testing.T) {
	keys := []int32{-100, -1, 0, 1, 42, 1000}
	m := FromSortedInt32Int32(keys, []int32{1, 2, 3, 4, 5, 6})
	for _, k := range keys {
		r := m.Rank(k)
		if got, ok := m.SelectKey(r); !ok || got != k {
			t.Errorf("SelectKey(Rank(%d)) = (%d,%v), want %d", k, got, ok, k)
		}
		sel, _ := m.SelectKey(r)
		if m.Rank(sel) != r {
			t.Errorf("Rank(SelectKey(%d)) = %d, want %d", r, m.Rank(sel), r)
		}
	}
	// rank on absent keys is the lower-bound index.
	if m.Rank(-101) != 0 {
		t.Errorf("Rank(-101) = %d, want 0", m.Rank(-101))
	}
	if m.Rank(500) != 5 {
		t.Errorf("Rank(500) = %d, want 5", m.Rank(500))
	}
	if m.Rank(100000) != 6 {
		t.Errorf("Rank(100000) = %d, want 6", m.Rank(100000))
	}
}

// ── Sortedness / parallel-array invariants post-build ────────────────────

func TestStoredArraysAreStrictlyAscendingAndAligned(t *testing.T) {
	m := FromSortedInt32Int32([]int32{10, 20, 30, 40, 50}, []int32{1, 2, 3, 4, 5})
	keys := m.Keys()
	for i := 1; i < len(keys); i++ {
		if keys[i-1] >= keys[i] {
			t.Errorf("stored keys not strictly ascending at %d", i)
		}
	}
	for i, k := range keys {
		ek, ev, ok := m.SelectEntry(i)
		gv, gok := m.Get(k)
		if !ok || !gok || ek != k || ev != gv {
			t.Errorf("alignment at %d: SelectEntry=(%d,%d,%v) Get(%d)=(%d,%v)", i, ek, ev, ok, k, gv, gok)
		}
	}
}

// ── Signed extremes (INT_MIN / INT_MAX) ──────────────────────────────────

func TestSignedExtremesLookupNavRankSelect(t *testing.T) {
	keys := []int32{intMin, -1, 0, 1, intMax}
	m := FromSortedInt32Int32(keys, []int32{10, 20, 30, 40, 50})

	if v, ok := m.Get(intMin); !ok || v != 10 {
		t.Errorf("Get(INT_MIN) = (%d,%v)", v, ok)
	}
	if v, ok := m.Get(intMax); !ok || v != 50 {
		t.Errorf("Get(INT_MAX) = (%d,%v)", v, ok)
	}
	if v, ok := m.FloorKey(intMin); !ok || v != intMin {
		t.Errorf("FloorKey(INT_MIN) = (%d,%v)", v, ok)
	}
	if _, ok := m.LowerKey(intMin); ok {
		t.Error("LowerKey(INT_MIN) must be absent")
	}
	if v, ok := m.HigherKey(-1); !ok || v != 0 {
		t.Errorf("HigherKey(-1) = (%d,%v)", v, ok)
	}
	if v, ok := m.CeilingKey(intMax); !ok || v != intMax {
		t.Errorf("CeilingKey(INT_MAX) = (%d,%v)", v, ok)
	}
	if _, ok := m.HigherKey(intMax); ok {
		t.Error("HigherKey(INT_MAX) must be absent")
	}
	if m.Rank(0) != 2 || m.Rank(intMin) != 0 || m.Rank(intMax) != 4 {
		t.Errorf("Rank: 0=%d MIN=%d MAX=%d", m.Rank(0), m.Rank(intMin), m.Rank(intMax))
	}
	if v, ok := m.SelectKey(0); !ok || v != intMin {
		t.Errorf("SelectKey(0) = (%d,%v)", v, ok)
	}
	if v, ok := m.SelectKey(4); !ok || v != intMax {
		t.Errorf("SelectKey(4) = (%d,%v)", v, ok)
	}
	if _, ok := m.SelectKey(5); ok {
		t.Error("SelectKey(5) must be absent")
	}
	if !slices.Equal(m.DescendingKeys(), []int32{intMax, 1, 0, -1, intMin}) {
		t.Errorf("DescendingKeys = %v", m.DescendingKeys())
	}
}

func TestRangeBracketsAtSignedExtremesDoNotOverflow(t *testing.T) {
	keys := []int32{intMin, -1, 0, 1, intMax}
	m := FromSortedInt32Int32(keys, []int32{10, 20, 30, 40, 50})

	// Open bound at INT_MIN: GreaterThan(MIN) excludes MIN, no MIN-1.
	if got := m.RangeKeys(rangev.GreaterThan(intMin)); !slices.Equal(got, []int32{-1, 0, 1, intMax}) {
		t.Errorf("RangeKeys(GreaterThan(INT_MIN)) = %v", got)
	}
	// Open bound at INT_MAX: LessThan(MAX) excludes MAX, no MAX+1.
	if got := m.RangeKeys(rangev.LessThan(intMax)); !slices.Equal(got, []int32{intMin, -1, 0, 1}) {
		t.Errorf("RangeKeys(LessThan(INT_MAX)) = %v", got)
	}
	// Closed both ends spanning the full signed range.
	if got := m.RangeKeys(rangev.Closed(intMin, intMax)); !slices.Equal(got, keys) {
		t.Errorf("RangeKeys(Closed(MIN,MAX)) = %v", got)
	}
	// Singleton at the extreme.
	if got := m.RangeKeys(rangev.Singleton(intMax)); !slices.Equal(got, []int32{intMax}) {
		t.Errorf("RangeKeys(Singleton(INT_MAX)) = %v", got)
	}
}

// ── Range membership == range.Contains (discrete-empty is NOT an error) ──

func TestOpenRangeOverAdjacentIntsIsEmptyNotError(t *testing.T) {
	m := FromSortedInt32Int32([]int32{1, 2}, []int32{10, 20})
	if got := m.RangeKeys(rangev.Open(1, 2)); len(got) != 0 {
		t.Errorf("RangeKeys(Open(1,2)) = %v, want []", got)
	}
	if got := m.RangeKeys(rangev.ClosedOpen(5, 5)); len(got) != 0 {
		t.Errorf("RangeKeys(ClosedOpen(5,5)) = %v, want []", got)
	}
}

func TestRangeQueryContiguousSlice(t *testing.T) {
	keys := make([]int32, 10)
	vals := make([]int32, 10)
	for i := range keys {
		keys[i] = int32((i + 1) * 10)
		vals[i] = keys[i] * 10
	}
	m := FromSortedInt32Int32(keys, vals)
	if got := m.RangeKeys(rangev.ClosedOpen(30, 70)); !slices.Equal(got, []int32{30, 40, 50, 60}) {
		t.Errorf("RangeKeys(ClosedOpen(30,70)) = %v", got)
	}
	if got := m.DescendingRangeKeys(rangev.ClosedOpen(30, 70)); !slices.Equal(got, []int32{60, 50, 40, 30}) {
		t.Errorf("DescendingRangeKeys = %v", got)
	}
	if got := m.RangeEntries(rangev.Closed(40, 50)); !slices.Equal(got, []Int32Int32Entry{{40, 400}, {50, 500}}) {
		t.Errorf("RangeEntries(Closed(40,50)) = %v", got)
	}
	if got := m.DescendingRangeEntries(rangev.Closed(40, 50)); !slices.Equal(got, []Int32Int32Entry{{50, 500}, {40, 400}}) {
		t.Errorf("DescendingRangeEntries(Closed(40,50)) = %v", got)
	}
	if got := m.RangeKeys(rangev.AtLeast(80)); !slices.Equal(got, []int32{80, 90, 100}) {
		t.Errorf("RangeKeys(AtLeast(80)) = %v", got)
	}
	if got := m.RangeKeys(rangev.AtMost(30)); !slices.Equal(got, []int32{10, 20, 30}) {
		t.Errorf("RangeKeys(AtMost(30)) = %v", got)
	}
	if got := m.RangeKeys(rangev.All()); len(got) != 10 {
		t.Errorf("RangeKeys(All()) len = %d, want 10", len(got))
	}
}

// ── Large flat-array parity (paging-invariance is trivial for flat) ──────

func TestLargeFlatLookupParity(t *testing.T) {
	keys := make([]int32, 10000)
	vals := make([]int32, 10000)
	for i := range keys {
		keys[i] = int32(i)
		vals[i] = int32(i) * 7
	}
	m := FromSortedInt32Int32(keys, vals)
	if m.Size() != 10000 {
		t.Fatalf("Size = %d", m.Size())
	}
	for _, p := range []int32{0, 1023, 1024, 1025, 4095, 4096, 4097, 8191, 8192, 9999} {
		if v, ok := m.Get(p); !ok || v != p*7 {
			t.Errorf("Get(%d) = (%d,%v)", p, v, ok)
		}
		if m.Rank(p) != int(p) {
			t.Errorf("Rank(%d) = %d", p, m.Rank(p))
		}
		if v, ok := m.SelectKey(int(p)); !ok || v != p {
			t.Errorf("SelectKey(%d) = (%d,%v)", p, v, ok)
		}
		if v, ok := m.FloorKey(p); !ok || v != p {
			t.Errorf("FloorKey(%d) = (%d,%v)", p, v, ok)
		}
		if v, ok := m.CeilingKey(p); !ok || v != p {
			t.Errorf("CeilingKey(%d) = (%d,%v)", p, v, ok)
		}
	}
	if _, ok := m.Get(10000); ok {
		t.Error("Get(10000) must be absent")
	}
	if m.Rank(10000) != 10000 {
		t.Errorf("Rank(10000) = %d", m.Rank(10000))
	}
	if _, ok := m.SelectKey(10000); ok {
		t.Error("SelectKey(10000) must be absent")
	}
	if got := m.RangeKeys(rangev.ClosedOpen(4090, 4100)); len(got) != 10 {
		t.Errorf("mid-range query len = %d, want 10", len(got))
	}
}

// ── Set surface mirrors the map ──────────────────────────────────────────

func TestSetFullSurface(t *testing.T) {
	s := FromSortedInt32([]int32{10, 20, 30, 40, 50})
	if s.Size() != 5 {
		t.Fatalf("Size = %d", s.Size())
	}
	if !s.Contains(30) || s.Contains(25) {
		t.Error("Contains")
	}
	if v, ok := s.First(); !ok || v != 10 {
		t.Errorf("First = (%d,%v)", v, ok)
	}
	if v, ok := s.Last(); !ok || v != 50 {
		t.Errorf("Last = (%d,%v)", v, ok)
	}
	if v, ok := s.Floor(25); !ok || v != 20 {
		t.Errorf("Floor(25) = (%d,%v)", v, ok)
	}
	if v, ok := s.Ceiling(25); !ok || v != 30 {
		t.Errorf("Ceiling(25) = (%d,%v)", v, ok)
	}
	if _, ok := s.Lower(10); ok {
		t.Error("Lower(10) must be absent")
	}
	if _, ok := s.Higher(50); ok {
		t.Error("Higher(50) must be absent")
	}
	if s.Rank(30) != 2 {
		t.Errorf("Rank(30) = %d", s.Rank(30))
	}
	if v, ok := s.Select(0); !ok || v != 10 {
		t.Errorf("Select(0) = (%d,%v)", v, ok)
	}
	if _, ok := s.Select(5); ok {
		t.Error("Select(5) must be absent")
	}
	if !slices.Equal(s.Elements(), []int32{10, 20, 30, 40, 50}) {
		t.Errorf("Elements = %v", s.Elements())
	}
	if !slices.Equal(s.DescendingElements(), []int32{50, 40, 30, 20, 10}) {
		t.Errorf("DescendingElements = %v", s.DescendingElements())
	}
	if !slices.Equal(s.RangeElements(rangev.ClosedOpen(20, 50)), []int32{20, 30, 40}) {
		t.Errorf("RangeElements = %v", s.RangeElements(rangev.ClosedOpen(20, 50)))
	}
	if !slices.Equal(s.DescendingRangeElements(rangev.ClosedOpen(20, 50)), []int32{40, 30, 20}) {
		t.Errorf("DescendingRangeElements = %v", s.DescendingRangeElements(rangev.ClosedOpen(20, 50)))
	}
}

func TestSetSelectRankRoundTrip(t *testing.T) {
	elems := []int32{intMin, -7, 0, 13, intMax}
	s := FromSortedInt32(elems)
	for _, e := range elems {
		r := s.Rank(e)
		if got, ok := s.Select(r); !ok || got != e {
			t.Errorf("Select(Rank(%d)) = (%d,%v), want %d", e, got, ok, e)
		}
	}
}
