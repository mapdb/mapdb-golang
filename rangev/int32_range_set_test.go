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

func rs(ranges ...Int32Range) *Int32RangeSet {
	s := NewInt32RangeSet()
	s.AddAll(ranges...)
	return s
}

func assertRanges(t *testing.T, s *Int32RangeSet, want ...Int32Range) {
	t.Helper()
	got := s.AsRanges()
	if want == nil {
		want = []Int32Range{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
}

func TestRangeSetCoalesceOverlap(t *testing.T) {
	s := rs(Closed(1, 5), Closed(3, 9))
	assertRanges(t, s, Closed(1, 9))
	if !s.Contains(4) {
		t.Error("expected contains 4")
	}
	if s.Contains(10) {
		t.Error("expected !contains 10")
	}
	if sp, ok := s.Span(); !ok || sp != Closed(1, 9) {
		t.Errorf("span = %v %v", sp, ok)
	}
}

func TestRangeSetCoalesceAbutCutTouch(t *testing.T) {
	// [1,3) & [3,5) touch at Below(3) -> single [1,5).
	s := rs(ClosedOpen(1, 3), ClosedOpen(3, 5))
	assertRanges(t, s, ClosedOpen(1, 5))
	if !s.Contains(3) {
		t.Error("expected contains 3")
	}
	if s.Contains(5) {
		t.Error("expected !contains 5")
	}
	bt, ok := s.AsRanges()[0].UpperBoundType()
	if !ok || bt != BoundOpen {
		t.Errorf("merged upper bound = %v %v, want open", bt, ok)
	}
}

func TestRangeSetOpenGapNoMerge(t *testing.T) {
	// (1,3) & (3,5): value 3 is the gap -> TWO ranges.
	s := rs(Open(1, 3), Open(3, 5))
	assertRanges(t, s, Open(1, 3), Open(3, 5))
	if s.Contains(3) {
		t.Error("expected !contains 3")
	}
}

func TestRangeSetAdjacentClosedNoIntegerAdjacencyMerge(t *testing.T) {
	// [1,3] & [4,5]: cut model has no integer adjacency (Below(4) > Above(3)).
	s := rs(Closed(1, 3), Closed(4, 5))
	assertRanges(t, s, Closed(1, 3), Closed(4, 5))
}

func TestRangeSetAddEmptyIsNoop(t *testing.T) {
	s := NewInt32RangeSet()
	s.Add(ClosedOpen(5, 5))
	if !s.IsEmpty() {
		t.Error("ClosedOpen(5,5) add should be no-op")
	}
	s.Add(OpenClosed(5, 5))
	if !s.IsEmpty() {
		t.Error("OpenClosed(5,5) add should be no-op")
	}
}

func TestRangeSetAddOpenNoIntegerStores(t *testing.T) {
	// Open(1,2) is cut-non-empty -> stored, though it contains no int32.
	s := rs(Open(1, 2))
	if s.IsEmpty() {
		t.Error("Open(1,2) must be stored (cut-non-empty)")
	}
	assertRanges(t, s, Open(1, 2))
	if s.Contains(1) || s.Contains(2) {
		t.Error("Open(1,2) contains no integer")
	}
}

func TestRangeSetRemoveSplits(t *testing.T) {
	s := rs(Closed(1, 9))
	s.Remove(ClosedOpen(4, 7))
	assertRanges(t, s, ClosedOpen(1, 4), Closed(7, 9))
}

func TestRangeSetRemoveEmptyIsNoop(t *testing.T) {
	s := rs(Closed(1, 9))
	s.Remove(ClosedOpen(5, 5))
	assertRanges(t, s, Closed(1, 9))
}

func TestRangeSetRemoveAbutmentDoesNotSplit(t *testing.T) {
	// remove([5,9)) abuts [1,5) at Below(5) -> no change to [1,5).
	s := rs(ClosedOpen(1, 5))
	s.Remove(ClosedOpen(5, 9))
	assertRanges(t, s, ClosedOpen(1, 5))
}

func TestRangeSetContainsAndRangeContaining(t *testing.T) {
	s := rs(ClosedOpen(1, 5), Closed(8, 9))
	if !s.Contains(3) || s.Contains(6) {
		t.Error("contains failed")
	}
	if r, ok := s.RangeContaining(3); !ok || r != ClosedOpen(1, 5) {
		t.Errorf("rangeContaining(3) = %v %v", r, ok)
	}
	if _, ok := s.RangeContaining(6); ok {
		t.Error("rangeContaining(6) should be absent")
	}
}

func TestRangeSetEnclosesSingleRange(t *testing.T) {
	s := rs(ClosedOpen(1, 3), ClosedOpen(5, 9))
	if s.Encloses(ClosedOpen(2, 6)) {
		t.Error("no single stored range encloses [2,6)")
	}
	if !s.Encloses(ClosedOpen(1, 2)) {
		t.Error("expected encloses [1,2)")
	}
	if !s.EnclosesAll(ClosedOpen(1, 2), ClosedOpen(5, 8)) {
		t.Error("expected enclosesAll")
	}
	if s.EnclosesAll(ClosedOpen(1, 2), ClosedOpen(2, 6)) {
		t.Error("expected !enclosesAll (2,6 not enclosed)")
	}
}

func TestRangeSetIntersectsCutAlgebra(t *testing.T) {
	s := rs(ClosedOpen(1, 3), ClosedOpen(5, 9))
	if !s.Intersects(ClosedOpen(2, 6)) {
		t.Error("cut-non-empty overlap with both -> true")
	}
	if s.Intersects(ClosedOpen(5, 5)) {
		t.Error("cut-empty query -> false")
	}
	s2 := rs(ClosedOpen(5, 9))
	if s2.Intersects(ClosedOpen(3, 5)) {
		t.Error("abutment -> false")
	}
}

func TestRangeSetIntersectsOpenCutNonEmptyNoInteger(t *testing.T) {
	// intersects(open(1,2)) vs stored (1,2) is TRUE (cut-non-empty), even
	// though no int32 lies in it.
	s := rs(Open(1, 2))
	if !s.Intersects(Open(1, 2)) {
		t.Error("expected intersects open(1,2)")
	}
}

func TestRangeSetComplementBasic(t *testing.T) {
	s := rs(Closed(1, 5))
	assertRanges(t, s.Complement(), LessThan(1), GreaterThan(5))
}

func TestRangeSetComplementAllIsEmpty(t *testing.T) {
	s := rs(All())
	if !s.Complement().IsEmpty() {
		t.Error("complement(all) should be empty")
	}
}

func TestRangeSetComplementEmptyIsAll(t *testing.T) {
	s := NewInt32RangeSet()
	assertRanges(t, s.Complement(), All())
}

func TestRangeSetComplementUnboundedNoSpuriousGap(t *testing.T) {
	// complement(lessThan(10)) = {[10,+inf)}, no leading (-inf,..) gap.
	s := rs(LessThan(10))
	assertRanges(t, s.Complement(), AtLeast(10))
}

func TestRangeSetComplementInvolution(t *testing.T) {
	cases := [][]Int32Range{
		{Closed(1, 5)},
		{Open(1, 3), Open(3, 5)},
		{LessThan(10)},
		{Closed(math.MinInt32, 0), OpenClosed(0, math.MaxInt32)},
		{},
		{All()},
	}
	for _, ranges := range cases {
		s := rs(ranges...)
		cc := s.Complement().Complement()
		if !reflect.DeepEqual(s.AsRanges(), cc.AsRanges()) {
			t.Errorf("involution failed for %v: got %v", ranges, cc.AsRanges())
		}
	}
}

func TestRangeSetSubRangeSetClips(t *testing.T) {
	s := rs(ClosedOpen(1, 5), Closed(8, 9))
	sub := s.SubRangeSet(ClosedOpen(3, 6))
	assertRanges(t, sub, ClosedOpen(3, 5))
}

func TestRangeSetSubRangeSetIndependentSnapshot(t *testing.T) {
	s := rs(ClosedOpen(1, 5))
	sub := s.SubRangeSet(ClosedOpen(2, 4))
	sub.Add(Closed(100, 200))
	// mutating the snapshot must not touch the parent.
	assertRanges(t, s, ClosedOpen(1, 5))
	// mutating the parent must not touch the snapshot.
	s.Add(Closed(300, 400))
	assertRanges(t, sub, ClosedOpen(2, 4), Closed(100, 200))
}

func TestRangeSetSignedExtremesNoPlusMinusOne(t *testing.T) {
	s := NewInt32RangeSet()
	s.Add(Closed(math.MinInt32, 0))
	s.Add(OpenClosed(0, math.MaxInt32))
	// [MIN,0] and (0,MAX] abut at Above(0) -> coalesce to [MIN, MAX].
	assertRanges(t, s, Closed(math.MinInt32, math.MaxInt32))
	if !s.Contains(math.MinInt32) || !s.Contains(math.MaxInt32) {
		t.Error("contains at extremes failed")
	}
	if sp, ok := s.Span(); !ok || sp != Closed(math.MinInt32, math.MaxInt32) {
		t.Errorf("span = %v %v", sp, ok)
	}
	// [MIN, MAX] is NOT all(); complement is the two flanking gaps.
	assertRanges(t, s.Complement(), LessThan(math.MinInt32), GreaterThan(math.MaxInt32))

	whole := NewInt32RangeSet()
	whole.Add(All())
	if !whole.Complement().IsEmpty() {
		t.Error("complement(all) over the whole domain should be empty")
	}
}

func TestRangeSetClearEmpties(t *testing.T) {
	s := rs(Closed(1, 9))
	s.Clear()
	if !s.IsEmpty() {
		t.Error("clear should empty")
	}
	assertRanges(t, s)
}

func TestRangeSetNormalFormAfterSequence(t *testing.T) {
	s := NewInt32RangeSet()
	ops := []Int32Range{
		Closed(1, 5),
		ClosedOpen(10, 12),
		ClosedOpen(12, 15),
		Open(20, 25),
		Closed(4, 11),
	}
	for _, r := range ops {
		s.Add(r)
	}
	v := s.AsRanges()
	for i := 0; i+1 < len(v); i++ {
		if v[i].lower.cmp(v[i+1].lower) >= 0 {
			t.Errorf("not ascending at %d: %v", i, v)
		}
		if v[i].IsConnected(v[i+1]) {
			t.Errorf("connected pair at %d: %v", i, v)
		}
	}
	for _, r := range v {
		if r.IsEmpty() {
			t.Errorf("empty range stored: %v", v)
		}
	}
}

// A set has no values, so a single ascending pass is enough: every connected
// run collapses no matter which piece lands last.
func TestRangeSetAddCoalescesWholeRunFromEitherDirection(t *testing.T) {
	asc := rs(ClosedOpen(1, 2), ClosedOpen(2, 3), ClosedOpen(3, 4))
	assertRanges(t, asc, ClosedOpen(1, 4))

	desc := rs(ClosedOpen(2, 3), ClosedOpen(3, 4), ClosedOpen(1, 2))
	assertRanges(t, desc, ClosedOpen(1, 4))
	if !reflect.DeepEqual(asc.AsRanges(), desc.AsRanges()) {
		t.Fatalf("direction dependent: asc = %v, desc = %v", asc.AsRanges(), desc.AsRanges())
	}

	// Bridging piece lands between two existing ranges -> one range.
	s := rs(ClosedOpen(1, 3), ClosedOpen(5, 7))
	if n := len(s.AsRanges()); n != 2 {
		t.Fatalf("pre-state should hold 2 ranges, got %v", s.AsRanges())
	}
	s.Add(ClosedOpen(3, 5))
	assertRanges(t, s, ClosedOpen(1, 7))
}

func TestRangeSetAddRejoinsFragmentsLeftBehindByRemove(t *testing.T) {
	s := rs(ClosedOpen(0, 10))
	s.Remove(ClosedOpen(3, 7))
	assertRanges(t, s, ClosedOpen(0, 3), ClosedOpen(7, 10))
	s.Add(ClosedOpen(3, 7))
	assertRanges(t, s, ClosedOpen(0, 10))
}
