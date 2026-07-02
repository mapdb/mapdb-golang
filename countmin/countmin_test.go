// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package countmin

import (
	"math"
	"reflect"
	"testing"

	"github.com/mapdb/mapdb-golang/hash"
)

// encode is the LE-4-byte encoding of an i32, the input the byte positions path
// consumes (reinterpret, not sign-extend).
func encode(item int32) []byte {
	u := uint32(item)
	return []byte{byte(u), byte(u >> 8), byte(u >> 16), byte(u >> 24)}
}

func TestRowHashMatchesPositions(t *testing.T) {
	// The d columns are EXACTLY positions(encode_i32(item), w, d) in order.
	c := NewCountMinWithParams(4, 16)
	cols := c.columns(7)
	if want := hash.Positions(encode(7), 16, 4); !reflect.DeepEqual(cols, want) {
		t.Fatalf("columns(7) = %v, want %v", cols, want)
	}
	// Pinned worked example: add(7) over (d=4, w=16) touches [7,0,9,2]
	// (the 12-hash-pipeline/positions_basic.json vector).
	if want := []uint32{7, 0, 9, 2}; !reflect.DeepEqual(cols, want) {
		t.Fatalf("columns(7) = %v, want %v", cols, want)
	}
}

func TestElementEncodingBytePath(t *testing.T) {
	// Reuses the byte positions path (length fold), NOT the scalar word.
	c := NewCountMinWithParams(4, 16)
	if got, want := c.columns(7), hash.Positions(encode(7), 16, 4); !reflect.DeepEqual(got, want) {
		t.Fatalf("columns(7) = %v, want %v", got, want)
	}
	// -1 reinterprets to 0xffffffff -> LE [ff,ff,ff,ff].
	if got, want := encode(-1), []byte{0xff, 0xff, 0xff, 0xff}; !reflect.DeepEqual(got, want) {
		t.Fatalf("encode(-1) = %v, want %v", got, want)
	}
	if got, want := encode(math.MinInt32), []byte{0x00, 0x00, 0x00, 0x80}; !reflect.DeepEqual(got, want) {
		t.Fatalf("encode(INT_MIN) = %v, want %v", got, want)
	}
}

func TestAddOneTouchesTheFourColumns(t *testing.T) {
	c := NewCountMinWithParams(4, 16)
	c.AddOne(7)
	m := c.ToCounters()
	// Rows 0..3 touch columns [7,0,9,2]; index r*16+col.
	if m[7] != 1 || m[16] != 1 || m[2*16+9] != 1 || m[3*16+2] != 1 {
		t.Fatalf("expected the four [7,0,9,2] cells set, got %v", m)
	}
	ones := 0
	for _, v := range m {
		if v == 1 {
			ones++
		}
	}
	if ones != 4 {
		t.Fatalf("expected exactly 4 cells = 1, got %d", ones)
	}
	if c.Estimate(7) != 1 {
		t.Fatalf("estimate(7) = %d, want 1", c.Estimate(7))
	}
	if c.Total() != 1 {
		t.Fatalf("total = %d, want 1", c.Total())
	}
	if len(m) != 64 {
		t.Fatalf("len = %d, want 64", len(m))
	}
}

func TestAddByCountEqualsRepeatedAddOne(t *testing.T) {
	a := NewCountMinWithParams(3, 13)
	b := NewCountMinWithParams(3, 13)
	a.Add(42, 5)
	for i := 0; i < 5; i++ {
		b.AddOne(42)
	}
	if !reflect.DeepEqual(a.ToCounters(), b.ToCounters()) {
		t.Fatal("add(42,5) != 5x add_one(42)")
	}
	if a.Estimate(42) != 5 || a.Total() != 5 {
		t.Fatalf("estimate=%d total=%d, want 5/5", a.Estimate(42), a.Total())
	}
}

func TestAddCountAccumulates(t *testing.T) {
	c := NewCountMinWithParams(4, 16)
	c.Add(7, 5)
	c.Add(7, 3)
	if c.Estimate(7) != 8 || c.Total() != 8 {
		t.Fatalf("estimate=%d total=%d, want 8/8", c.Estimate(7), c.Total())
	}
}

func TestCountZeroIsCounterNoopButUpdatesTotal(t *testing.T) {
	c := NewCountMinWithParams(3, 7)
	c.Add(1, 0)
	for _, v := range c.ToCounters() {
		if v != 0 {
			t.Fatal("count=0 add must not touch counters")
		}
	}
	if c.Total() != 0 {
		t.Fatalf("total = %d, want 0", c.Total())
	}
	c.Add(1, 4)
	c.Add(1, 0)
	if c.Estimate(1) != 4 || c.Total() != 4 {
		t.Fatalf("estimate=%d total=%d, want 4/4", c.Estimate(1), c.Total())
	}
}

func TestCollisionAcrossRowsNotDeduped(t *testing.T) {
	// Find an item + (w,d) whose positions repeat a column across rows; both
	// same-numbered counters in DIFFERENT rows must be incremented.
	const d = uint32(3)
	for w := uint32(2); w < 32; w++ {
		for item := int32(0); item < 256; item++ {
			cols := hash.Positions(encode(item), w, d)
			seen := map[uint32]int{}
			r0, r1, col, found := -1, -1, uint32(0), false
			for r, cc := range cols {
				if prev, ok := seen[cc]; ok {
					r0, r1, col, found = prev, r, cc, true
					break
				}
				seen[cc] = r
			}
			if !found {
				continue
			}
			c := NewCountMinWithParams(d, w)
			c.AddOne(item)
			m := c.ToCounters()
			if m[r0*int(w)+int(col)] != 1 || m[r1*int(w)+int(col)] != 1 {
				t.Fatalf("cross-row collision col %d: both rows must be 1", col)
			}
			if c.Estimate(item) != 1 {
				t.Fatalf("estimate = %d, want 1", c.Estimate(item))
			}
			return
		}
	}
	t.Fatal("no cross-row column collision found in the search space")
}

func TestEstimateIsMinNotAverageOrRow0(t *testing.T) {
	c := NewCountMinWithParams(4, 8)
	const target = int32(5)
	c.Add(target, 1)
	cols := c.columns(target)
	for other := int32(0); other < 200; other++ {
		if other == target {
			continue
		}
		c.Add(other, 7)
	}
	m := c.ToCounters()
	min := uint64(math.MaxUint64)
	max := uint64(0)
	for r, col := range cols {
		v := m[r*8+int(col)]
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if c.Estimate(target) != min {
		t.Fatalf("estimate = %d, want MIN %d", c.Estimate(target), min)
	}
	if c.Estimate(target) < 1 {
		t.Fatal("under-estimate of true count 1")
	}
	if max < min {
		t.Fatal("max < min impossible")
	}
}

func TestOverflowSaturatesNotWraps(t *testing.T) {
	c := NewCountMinWithParams(2, 4)
	c.Add(9, math.MaxUint64)
	c.Add(9, 5)
	if c.Estimate(9) != math.MaxUint64 {
		t.Fatalf("estimate = %d, want MaxUint64", c.Estimate(9))
	}
	if c.Total() != math.MaxUint64 {
		t.Fatalf("total = %d, want MaxUint64", c.Total())
	}
	m := c.ToCounters()
	for r, col := range c.columns(9) {
		if m[r*4+int(col)] != math.MaxUint64 {
			t.Fatal("counter did not saturate")
		}
	}
}

func TestNoUnderEstimate(t *testing.T) {
	c := NewCountMinWithParams(5, 64)
	c.Add(-1, 3)
	c.Add(math.MinInt32, 10)
	if c.Estimate(-1) < 3 {
		t.Fatalf("estimate(-1) = %d, want >= 3", c.Estimate(-1))
	}
	if c.Estimate(math.MinInt32) < 10 {
		t.Fatalf("estimate(INT_MIN) = %d, want >= 10", c.Estimate(math.MinInt32))
	}
}

func TestOrderIndependence(t *testing.T) {
	a := NewCountMinWithParams(4, 16)
	b := NewCountMinWithParams(4, 16)
	type pair struct {
		it int32
		ct uint64
	}
	seq := []pair{{1, 3}, {2, 5}, {1, 2}, {-7, 9}, {math.MaxInt32, 1}}
	for _, p := range seq {
		a.Add(p.it, p.ct)
	}
	for i := len(seq) - 1; i >= 0; i-- {
		b.Add(seq[i].it, seq[i].ct)
	}
	if !reflect.DeepEqual(a.ToCounters(), b.ToCounters()) {
		t.Fatal("add order changed the matrix")
	}
	if a.Total() != b.Total() {
		t.Fatal("add order changed total")
	}
}

func TestDZeroIsLegalVacuousMax(t *testing.T) {
	c := NewCountMinWithParams(0, 16)
	c.Add(5, 1)
	if len(c.ToCounters()) != 0 {
		t.Fatal("d=0 matrix must be empty")
	}
	if c.Total() != 1 {
		t.Fatalf("total = %d, want 1", c.Total())
	}
	if c.Estimate(5) != math.MaxUint64 {
		t.Fatalf("estimate = %d, want MaxUint64 (empty MIN)", c.Estimate(5))
	}
}

func TestEmptyMatrixIsAllZeroDense(t *testing.T) {
	c := NewCountMinWithParams(4, 16)
	m := c.ToCounters()
	if len(m) != 64 {
		t.Fatalf("len = %d, want 64", len(m))
	}
	for _, v := range m {
		if v != 0 {
			t.Fatal("empty matrix must be all zero")
		}
	}
	if c.Estimate(7) != 0 || c.Total() != 0 {
		t.Fatal("empty estimate/total must be 0")
	}
}

func TestWZeroTraps(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("w=0 must panic")
		}
	}()
	NewCountMinWithParams(4, 0)
}

func TestOptimalIntegerTable(t *testing.T) {
	// (epsilon, delta) -> (w, d) per spec/features/count-min.md.
	cases := []struct {
		eps, delta float64
		w, d       uint32
	}{
		{0.01, 0.01, 272, 5},
		{0.001, 0.001, 2719, 7},
		{0.1, 0.05, 28, 3},
		{0.01, 0.001, 272, 7},
		{0.5, 0.5, 6, 1},
	}
	for _, tc := range cases {
		c := NewCountMinOptimal(tc.eps, tc.delta)
		if c.Width() != tc.w {
			t.Errorf("w for (%v,%v) = %d, want %d", tc.eps, tc.delta, c.Width(), tc.w)
		}
		if c.Depth() != tc.d {
			t.Errorf("d for (%v,%v) = %d, want %d", tc.eps, tc.delta, c.Depth(), tc.d)
		}
	}
}

func TestOptimalRejectsBadInputs(t *testing.T) {
	for _, tc := range []struct {
		name       string
		eps, delta float64
	}{
		{"epsilon=0", 0.0, 0.5},
		{"delta=1", 0.5, 1.0},
		{"epsilon=NaN", math.NaN(), 0.5},
		{"delta=Inf", 0.5, math.Inf(1)},
		{"epsilon>=1", 1.5, 0.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("%s must panic", tc.name)
				}
			}()
			NewCountMinOptimal(tc.eps, tc.delta)
		})
	}
}
