// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package rangev

import (
	"math"
	"reflect"
	"testing"
)

func entry(r Int32Range, v int32) Int32Int32Entry {
	return Int32Int32Entry{Range: r, Value: v}
}

func assertEntries(t *testing.T, m *Int32Int32RangeMap, want ...Int32Int32Entry) {
	t.Helper()
	got := m.AsMapOfRanges()
	if want == nil {
		want = []Int32Int32Entry{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %v, want %v", got, want)
	}
}

func getOr(m *Int32Int32RangeMap, k int32) (int32, bool) { return m.Get(k) }

func TestRangeMapPutBasic(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(Closed(8, 9), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(Closed(8, 9), 200))
	if v, ok := getOr(m, 3); !ok || v != 100 {
		t.Errorf("get(3) = %v %v", v, ok)
	}
	if _, ok := getOr(m, 6); ok {
		t.Error("get(6) should be absent")
	}
	if v, ok := getOr(m, 8); !ok || v != 200 {
		t.Errorf("get(8) = %v %v", v, ok)
	}
}

func TestRangeMapPutOverwriteClips(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(3, 9), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 3), 100), entry(ClosedOpen(3, 9), 200))
	if v, _ := getOr(m, 2); v != 100 {
		t.Error("get(2) want 100")
	}
	if v, _ := getOr(m, 4); v != 200 {
		t.Error("get(4) want 200")
	}
	if v, _ := getOr(m, 8); v != 200 {
		t.Error("get(8) want 200")
	}
}

func TestRangeMapPutSplitStraddle(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Put(ClosedOpen(3, 5), 200)
	assertEntries(t, m,
		entry(ClosedOpen(1, 3), 100),
		entry(ClosedOpen(3, 5), 200),
		entry(ClosedOpen(5, 9), 100),
	)
	if v, _ := getOr(m, 2); v != 100 {
		t.Error("get(2) want 100")
	}
	if v, _ := getOr(m, 4); v != 200 {
		t.Error("get(4) want 200")
	}
	if v, _ := getOr(m, 6); v != 100 {
		t.Error("get(6) want 100")
	}
}

func TestRangeMapPutDoesNotCoalesce(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(5, 9), 100)
	// TWO entries even though value equal and they abut.
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(ClosedOpen(5, 9), 100))
	if v, _ := getOr(m, 5); v != 100 {
		t.Error("get(5) want 100")
	}
}

func TestRangeMapPutCoalescingEqualValueAbut(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.PutCoalescing(ClosedOpen(5, 9), 100)
	assertEntries(t, m, entry(ClosedOpen(1, 9), 100))
}

func TestRangeMapPutCoalescingDifferentValueNoMerge(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.PutCoalescing(ClosedOpen(5, 9), 200)
	assertEntries(t, m, entry(ClosedOpen(1, 5), 100), entry(ClosedOpen(5, 9), 200))
}

func TestRangeMapPutCoalescingBothSides(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(ClosedOpen(9, 12), 100)
	m.PutCoalescing(ClosedOpen(5, 9), 100)
	assertEntries(t, m, entry(ClosedOpen(1, 12), 100))
}

func TestRangeMapRemoveSplits(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Remove(ClosedOpen(4, 7))
	assertEntries(t, m, entry(ClosedOpen(1, 4), 100), entry(ClosedOpen(7, 9), 100))
	if _, ok := getOr(m, 5); ok {
		t.Error("get(5) should be absent after remove")
	}
}

func TestRangeMapGetEntryLookup(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	if r, v, ok := m.GetEntry(3); !ok || r != ClosedOpen(1, 5) || v != 100 {
		t.Errorf("getEntry(3) = %v %v %v", r, v, ok)
	}
	if _, _, ok := m.GetEntry(6); ok {
		t.Error("getEntry(6) should be absent")
	}
}

func TestRangeMapSpanOverEntries(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(Closed(8, 9), 200)
	if sp, ok := m.Span(); !ok || sp != Closed(1, 9) {
		t.Errorf("span = %v %v", sp, ok)
	}
}

func TestRangeMapEmptyPutIsNoop(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(5, 5), 100)
	if !m.IsEmpty() {
		t.Error("cut-empty put should be no-op")
	}
	assertEntries(t, m)
}

func TestRangeMapSubRangeMapClipsSnapshot(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 5), 100)
	m.Put(Closed(8, 9), 200)
	sub := m.SubRangeMap(ClosedOpen(3, 6))
	assertEntries(t, sub, entry(ClosedOpen(3, 5), 100))
	// snapshot independence: mutate the parent, sub unchanged.
	m.Put(Closed(3, 3), 999)
	assertEntries(t, sub, entry(ClosedOpen(3, 5), 100))
	// mutating the snapshot does not touch the parent.
	sub.Put(Closed(50, 60), 7)
	if _, ok := m.Get(55); ok {
		t.Error("parent must not see snapshot mutation")
	}
}

func TestRangeMapSignedExtremesNoPlusMinusOne(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(math.MinInt32, 0), 1)
	m.Put(Closed(0, math.MaxInt32), 2)
	if v, _ := getOr(m, math.MinInt32); v != 1 {
		t.Error("get(MIN) want 1")
	}
	if v, _ := getOr(m, 0); v != 2 {
		t.Error("get(0) want 2")
	}
	if v, _ := getOr(m, math.MaxInt32); v != 2 {
		t.Error("get(MAX) want 2")
	}
}

func TestRangeMapNormalFormDisjointAfterSequence(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 10), 1)
	m.Put(ClosedOpen(3, 5), 2)
	m.Put(ClosedOpen(7, 20), 3)
	m.PutCoalescing(ClosedOpen(20, 25), 3)
	v := m.AsMapOfRanges()
	for i := 0; i+1 < len(v); i++ {
		if v[i].Range.lower.cmp(v[i+1].Range.lower) >= 0 {
			t.Errorf("not ascending at %d: %v", i, v)
		}
		inter, ok := v[i].Range.Intersection(v[i+1].Range)
		if ok && !inter.IsEmpty() {
			t.Errorf("overlapping entries at %d: %v", i, v)
		}
	}
	for _, e := range v {
		if e.Range.IsEmpty() {
			t.Errorf("empty entry range: %v", v)
		}
	}
}

func TestRangeMapClear(t *testing.T) {
	m := NewInt32Int32RangeMap()
	m.Put(ClosedOpen(1, 9), 100)
	m.Clear()
	if !m.IsEmpty() {
		t.Error("clear should empty")
	}
	assertEntries(t, m)
}
