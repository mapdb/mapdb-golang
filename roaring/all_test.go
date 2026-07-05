// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package roaring

import (
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/internal/conformance"
)

// TestAllYieldsAscendingAcrossChunks pins the All() law-1 method: it yields every
// added value exactly once, in unsigned-ascending order, across multiple chunks
// (values chosen so they land in different high-16-bit chunks and both container
// kinds), and the count matches Cardinality. A duplicate Add must not double-yield.
func TestAllYieldsAscendingAcrossChunks(t *testing.T) {
	s := NewRoaringU32()
	// Values spanning three chunks (high bits 0, 1, 0xFFFF) plus a duplicate and
	// the max uint32 (sorts last under unsigned order).
	added := []uint32{5, 1, 1 << 16, 3, 65535, 0xFFFFFFFF, 5, 1<<16 + 7}
	for _, v := range added {
		s.Add(v)
	}

	got := slices.Collect(s.All())

	want := []uint32{1, 3, 5, 65535, 1 << 16, 1<<16 + 7, 0xFFFFFFFF}
	if !slices.Equal(got, want) {
		t.Fatalf("All() = %v, want %v", got, want)
	}
	if uint64(len(got)) != s.Cardinality() {
		t.Fatalf("All() yielded %d values, Cardinality() = %d", len(got), s.Cardinality())
	}
}

// TestAllEmpty confirms All() over an empty set yields nothing (and does not
// panic on the zero value path).
func TestAllEmpty(t *testing.T) {
	if got := slices.Collect(NewRoaringU32().All()); len(got) != 0 {
		t.Fatalf("All() over empty set = %v, want none", got)
	}
}

// TestAllEarlyBreak confirms All() honors an early break — the yield function
// returning false stops iteration (Iterate's short-circuit path).
func TestAllEarlyBreak(t *testing.T) {
	s := NewRoaringU32()
	for _, v := range []uint32{10, 20, 30, 40} {
		s.Add(v)
	}
	var seen []uint32
	for v := range s.All() {
		seen = append(seen, v)
		if v == 20 {
			break
		}
	}
	if !slices.Equal(seen, []uint32{10, 20}) {
		t.Fatalf("early-break All() = %v, want [10 20]", seen)
	}
}

// TestSegmentsConformance stamps the Segments partition law: for n∈{1,2,7,len+1}
// the chunk-boundary split covers All() exactly and each segment is re-runnable.
// The fixture spans several chunks so the split is non-trivial.
func TestSegmentsConformance(t *testing.T) {
	s := NewRoaringU32()
	// Values across 5 distinct chunks (high bits 0,1,2,7,0xFFFF), some chunks
	// holding multiple values, so #chunks (5) exceeds several split counts.
	for _, v := range []uint32{
		1, 3, 5, 65535, // chunk 0
		1 << 16, 1<<16 + 9, // chunk 1
		2 << 16,                       // chunk 2
		7 << 16, 7<<16 + 1, 7<<16 + 2, // chunk 7
		0xFFFFFFFF, // chunk 0xFFFF
	} {
		s.Add(v)
	}
	conformance.SegmentsCoverAll(t, s.All(), s.Segments)
}

// TestSegmentsEmptyAndClamp checks the boundary behaviors: no segments for an
// empty set, and n clamped to the chunk count (n larger than #chunks does not
// produce empty trailing segments that break coverage).
func TestSegmentsEmptyAndClamp(t *testing.T) {
	if segs := NewRoaringU32().Segments(4); segs != nil {
		t.Fatalf("empty set Segments = %v, want nil", segs)
	}
	s := NewRoaringU32()
	s.Add(1 << 16)
	s.Add(2 << 16) // 2 chunks
	// n far larger than #chunks: still covers everything.
	var got []uint32
	for _, seg := range s.Segments(100) {
		got = append(got, slices.Collect(seg)...)
	}
	slices.Sort(got)
	if !slices.Equal(got, []uint32{1 << 16, 2 << 16}) {
		t.Fatalf("Segments(100) concat = %v, want the 2 values", got)
	}
}
