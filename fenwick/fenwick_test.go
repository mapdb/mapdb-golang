// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package fenwick

import (
	"math"
	"reflect"
	"testing"
)

const (
	int32Min = math.MinInt32
	int32Max = math.MaxInt32
)

// brute is a brute-force int64 reference: a flat array of per-index int64
// values, with the same wrapping arithmetic the Fenwick tree must match. Go's
// native int64 wraps two's-complement, so the brute force needs no explicit
// wrap calls.
type brute struct {
	vals []int64
}

func newBrute(n int) *brute            { return &brute{vals: make([]int64, n)} }
func (b *brute) update(i int, d int32) { b.vals[i] += int64(d) }
func (b *brute) set(i int, v int32)    { b.vals[i] = int64(v) }
func (b *brute) get(i int) int64       { return b.vals[i] }
func (b *brute) prefixSum(i int) int64 {
	var acc int64
	for k := 0; k <= i; k++ {
		acc += b.vals[k]
	}
	return acc
}
func (b *brute) rangeSum(lo, hi int) int64 {
	if lo > hi {
		return 0
	}
	var acc int64
	for k := lo; k <= hi; k++ {
		acc += b.vals[k]
	}
	return acc
}
func (b *brute) total() int64 {
	var acc int64
	for _, v := range b.vals {
		acc += v
	}
	return acc
}

// lcg is a tiny deterministic LCG so the property tests need no external dep.
type lcg struct{ s uint64 }

func (l *lcg) nextU64() uint64 {
	l.s = l.s*6364136223846793005 + 1442695040888963407
	return l.s
}
func (l *lcg) nextI32() int32 { return int32(uint32(l.nextU64())) }
func (l *lcg) nextN(bound int) int {
	return int(l.nextU64() % uint64(bound))
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, got none", name)
		}
	}()
	f()
}

func TestWorkedExampleFromSpec(t *testing.T) {
	f := NewFenwickTreeWithSize(8)
	f.Update(0, 5)
	f.Update(3, 2)
	f.Update(7, 9)
	checks := []struct {
		name string
		got  int64
		want int64
	}{
		{"prefix_sum(0)", f.PrefixSum(0), 5},
		{"prefix_sum(3)", f.PrefixSum(3), 7},
		{"prefix_sum(6)", f.PrefixSum(6), 7},
		{"prefix_sum(7)", f.PrefixSum(7), 16},
		{"total", f.Total(), 16},
		{"range_sum(1,7)", f.RangeSum(1, 7), 11},
		{"get(3)", f.Get(3), 2},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
	if f.Len() != 8 || f.IsEmpty() {
		t.Errorf("Len=%d IsEmpty=%v, want 8,false", f.Len(), f.IsEmpty())
	}
}

func TestInclusiveConventions(t *testing.T) {
	f := NewFenwickTreeFromValues([]int32{3, 1, 4, 1, 5, 9, 2, 6})
	// prefix_sum(0) is the first value (NOT 0 -- inclusive).
	if f.PrefixSum(0) != 3 {
		t.Errorf("PrefixSum(0) = %d, want 3", f.PrefixSum(0))
	}
	// single-element inclusive range == that value (NOT 0).
	if f.RangeSum(2, 2) != 4 {
		t.Errorf("RangeSum(2,2) = %d, want 4", f.RangeSum(2, 2))
	}
	if f.Get(2) != 4 {
		t.Errorf("Get(2) = %d, want 4", f.Get(2))
	}
	if f.PrefixSum(7) != 31 || f.Total() != 31 {
		t.Errorf("PrefixSum(7)=%d Total=%d, want 31,31", f.PrefixSum(7), f.Total())
	}
	// total == prefix_sum(n-1) == range_sum(0, n-1).
	if f.Total() != f.PrefixSum(7) || f.Total() != f.RangeSum(0, 7) {
		t.Errorf("total identity broken")
	}
}

func TestFromValuesMatchesUpdates(t *testing.T) {
	cases := [][]int32{
		{},
		{42},
		{3, 1, 4, 1, 5, 9, 2, 6},
		{int32Min, int32Max, -1, 0, 7},
		{-5, -5, -5, -5, -5, -5, -5},
	}
	for _, vals := range cases {
		built := NewFenwickTreeFromValues(vals)
		updated := NewFenwickTreeWithSize(len(vals))
		for i, v := range vals {
			updated.Update(i, v)
		}
		if !reflect.DeepEqual(built.CanonicalTree(), updated.CanonicalTree()) {
			t.Errorf("canonical tree mismatch for %v: %v vs %v",
				vals, built.CanonicalTree(), updated.CanonicalTree())
		}
		for i := range vals {
			if built.PrefixSum(i) != updated.PrefixSum(i) || built.Get(i) != updated.Get(i) {
				t.Errorf("query mismatch at %d for %v", i, vals)
			}
		}
		if built.Total() != updated.Total() {
			t.Errorf("total mismatch for %v", vals)
		}
	}
}

func TestSetReplacesNotAdds(t *testing.T) {
	f := NewFenwickTreeWithSize(4)
	f.Update(1, 5)
	f.Set(1, 3) // replace, NOT add: get(1) must be 3, not 8.
	f.Update(2, 7)
	if f.Get(1) != 3 {
		t.Errorf("Get(1) = %d, want 3 (set replaces)", f.Get(1))
	}
	if f.Get(2) != 7 {
		t.Errorf("Get(2) = %d, want 7", f.Get(2))
	}
	if f.PrefixSum(1) != 3 || f.PrefixSum(3) != 10 || f.Total() != 10 {
		t.Errorf("PrefixSum(1)=%d PrefixSum(3)=%d Total=%d, want 3,10,10",
			f.PrefixSum(1), f.PrefixSum(3), f.Total())
	}
}

func TestNegativeDeltasCrossZero(t *testing.T) {
	f := NewFenwickTreeWithSize(5)
	f.Update(0, 10)
	f.Update(1, -4)
	f.Update(2, -20)
	f.Update(3, 7)
	want := map[string]int64{
		"prefix_sum(0)": 10, "prefix_sum(1)": 6, "prefix_sum(2)": -14,
		"prefix_sum(4)": -7, "total": -7, "range_sum(1,3)": -17,
	}
	got := map[string]int64{
		"prefix_sum(0)": f.PrefixSum(0), "prefix_sum(1)": f.PrefixSum(1),
		"prefix_sum(2)": f.PrefixSum(2), "prefix_sum(4)": f.PrefixSum(4),
		"total": f.Total(), "range_sum(1,3)": f.RangeSum(1, 3),
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("%s = %d, want %d", k, got[k], w)
		}
	}
}

func TestSignedExtremesWidenToI64(t *testing.T) {
	f := NewFenwickTreeWithSize(3)
	f.Set(0, int32Max) // 2147483647
	f.Set(1, int32Min) // -2147483648
	f.Update(2, int32Max)
	f.Update(2, 1) // value becomes 2147483648 as int64 (NOT int32-wrapped).
	if f.Get(0) != 2147483647 {
		t.Errorf("Get(0) = %d, want 2147483647", f.Get(0))
	}
	if f.Get(1) != -2147483648 {
		t.Errorf("Get(1) = %d, want -2147483648", f.Get(1))
	}
	if f.Get(2) != 2147483648 {
		t.Errorf("Get(2) = %d, want 2147483648 (int64, not int32-wrapped)", f.Get(2))
	}
	if f.PrefixSum(1) != -1 {
		t.Errorf("PrefixSum(1) = %d, want -1", f.PrefixSum(1))
	}
	if f.Total() != 2147483647 {
		t.Errorf("Total = %d, want 2147483647", f.Total())
	}
}

func TestLargeI64SumExceeds2_53(t *testing.T) {
	f := NewFenwickTreeWithSize(4)
	for i := 0; i < 4; i++ {
		f.Set(i, int32Max)
	}
	if f.Total() != 8589934588 { // 4 * (2^31 - 1)
		t.Errorf("Total = %d, want 8589934588", f.Total())
	}
	if f.PrefixSum(3) != 8589934588 {
		t.Errorf("PrefixSum(3) = %d, want 8589934588", f.PrefixSum(3))
	}
	if f.RangeSum(1, 2) != 4294967294 {
		t.Errorf("RangeSum(1,2) = %d, want 4294967294", f.RangeSum(1, 2))
	}
}

// TestI64WrapIsTwoComplementNotSaturating seeds a slot near math.MaxInt64 via
// the production add path (white-box addInternal -- reaching MaxInt64 with int32
// deltas is infeasible), then adds past it and verifies the result wraps
// two's-complement to a negative int64, NOT saturates.
func TestI64WrapIsTwoComplementNotSaturating(t *testing.T) {
	f := NewFenwickTreeWithSize(1)
	f.addInternal(0, math.MaxInt64-1)
	if f.Get(0) != math.MaxInt64-1 {
		t.Fatalf("Get(0) = %d, want %d", f.Get(0), int64(math.MaxInt64-1))
	}
	f.addInternal(0, 5) // (MaxInt64-1) + 5 wraps to negative.
	// Compute the expected wrap at runtime (a constant expression would be a
	// compile-time overflow error; runtime int64 arithmetic wraps).
	base := int64(math.MaxInt64 - 1)
	expected := base + 5
	if expected >= 0 {
		t.Fatalf("test premise wrong: expected wrap to negative, got %d", expected)
	}
	if f.Get(0) != expected {
		t.Errorf("Get(0) = %d, want %d (wrapped)", f.Get(0), expected)
	}
	if f.Total() != expected || f.PrefixSum(0) != expected {
		t.Errorf("Total/PrefixSum did not wrap consistently")
	}
}

// TestRangeSumEqualsPrefixDiffAfterWrap proves invertibility holds even after
// the running total has wrapped (wrapping subtraction is the exact inverse).
func TestRangeSumEqualsPrefixDiffAfterWrap(t *testing.T) {
	f := NewFenwickTreeWithSize(3)
	f.addInternal(0, math.MaxInt64-10)
	f.addInternal(1, 100) // prefix_sum(1) wraps.
	f.addInternal(2, -7)
	// Each per-index logical value is exact (a single value never overflows).
	if f.Get(0) != math.MaxInt64-10 || f.Get(1) != 100 || f.Get(2) != -7 {
		t.Fatalf("per-index values diverged: %d %d %d", f.Get(0), f.Get(1), f.Get(2))
	}
	if f.RangeSum(1, 1) != 100 || f.RangeSum(2, 2) != -7 {
		t.Errorf("single-index range sums wrong")
	}
	base := int64(math.MaxInt64 - 10)
	total := base + 100 + (-7)
	if total >= 0 {
		t.Fatalf("test premise wrong: expected wrapped negative total, got %d", total)
	}
	if f.RangeSum(0, 2) != total || f.Total() != total {
		t.Errorf("wrapped total mismatch: range=%d total=%d want %d",
			f.RangeSum(0, 2), f.Total(), total)
	}
	// Invertibility for every sub-range after the wrap.
	for lo := 0; lo < 3; lo++ {
		for hi := lo; hi < 3; hi++ {
			direct := f.RangeSum(lo, hi)
			var lower int64
			if lo > 0 {
				lower = f.PrefixSum(lo - 1)
			}
			via := f.PrefixSum(hi) - lower
			if direct != via {
				t.Errorf("invertibility lo=%d hi=%d: direct=%d via=%d", lo, hi, direct, via)
			}
		}
	}
}

func TestSingleElement(t *testing.T) {
	f := NewFenwickTreeWithSize(1)
	f.Update(0, 42)
	if f.Len() != 1 || f.Get(0) != 42 || f.PrefixSum(0) != 42 ||
		f.RangeSum(0, 0) != 42 || f.Total() != 42 {
		t.Errorf("single-element tree wrong")
	}
}

func TestEmptyTreeEdges(t *testing.T) {
	f := NewFenwickTreeWithSize(0)
	if f.Len() != 0 || !f.IsEmpty() || f.Total() != 0 {
		t.Errorf("with_size(0) not a valid empty tree")
	}
	g := NewFenwickTreeFromValues(nil)
	if g.Len() != 0 || !g.IsEmpty() || g.Total() != 0 || len(g.CanonicalTree()) != 0 {
		t.Errorf("from_values([]) not a valid empty tree")
	}
	h := NewFenwickTreeFromValues([]int32{})
	if h.Len() != 0 || !h.IsEmpty() {
		t.Errorf("from_values(empty slice) not a valid empty tree")
	}
}

func TestLoGtHiReturnsZero(t *testing.T) {
	f := NewFenwickTreeFromValues([]int32{3, 1, 4, 1, 5, 9, 2, 6})
	if f.RangeSum(5, 2) != 0 { // both endpoints valid, lo > hi.
		t.Errorf("RangeSum(5,2) = %d, want 0", f.RangeSum(5, 2))
	}
	if f.RangeSum(7, 0) != 0 {
		t.Errorf("RangeSum(7,0) = %d, want 0", f.RangeSum(7, 0))
	}
}

func TestOutOfRangeTraps(t *testing.T) {
	f := NewFenwickTreeWithSize(4)
	// i == n
	mustPanic(t, "Update(4)", func() { f.Update(4, 1) })
	mustPanic(t, "Set(4)", func() { f.Set(4, 1) })
	mustPanic(t, "Get(4)", func() { f.Get(4) })
	mustPanic(t, "PrefixSum(4)", func() { f.PrefixSum(4) })
	mustPanic(t, "RangeSum(0,4)", func() { f.RangeSum(0, 4) }) // hi == n traps
	mustPanic(t, "RangeSum(4,0)", func() { f.RangeSum(4, 0) }) // lo == n traps
}

// TestNegativeIndexTraps verifies a negative index deliberately TRAPS (Go uses a
// signed int) rather than wrapping to a huge index / valid slot.
func TestNegativeIndexTraps(t *testing.T) {
	f := NewFenwickTreeWithSize(4)
	mustPanic(t, "Update(-1)", func() { f.Update(-1, 1) })
	mustPanic(t, "Set(-1)", func() { f.Set(-1, 1) })
	mustPanic(t, "Get(-1)", func() { f.Get(-1) })
	mustPanic(t, "PrefixSum(-1)", func() { f.PrefixSum(-1) })
	mustPanic(t, "RangeSum(-1,2)", func() { f.RangeSum(-1, 2) })
	mustPanic(t, "RangeSum(0,-1)", func() { f.RangeSum(0, -1) })
}

func TestNegativeSizeTraps(t *testing.T) {
	mustPanic(t, "NewFenwickTreeWithSize(-1)", func() { NewFenwickTreeWithSize(-1) })
}

func TestEmptyTreeQueriesTrap(t *testing.T) {
	f := NewFenwickTreeWithSize(0)
	mustPanic(t, "empty Get(0)", func() { f.Get(0) })
	mustPanic(t, "empty PrefixSum(0)", func() { f.PrefixSum(0) })
	mustPanic(t, "empty RangeSum(0,0)", func() { f.RangeSum(0, 0) })
}

// TestCanonicalTreeIsCopy locks down that CanonicalTree returns a fresh copy,
// not an alias of the backing slice.
func TestCanonicalTreeIsCopy(t *testing.T) {
	f := NewFenwickTreeFromValues([]int32{3, 1, 4, 1, 5})
	ct := f.CanonicalTree()
	ct[0] = 9999 // mutate the returned slice
	if f.CanonicalTree()[0] == 9999 {
		t.Errorf("CanonicalTree aliases the backing slice")
	}
	// Underlying queries are unaffected too.
	if f.Get(0) != 3 {
		t.Errorf("Get(0) = %d, want 3 (backing must be untouched)", f.Get(0))
	}
}

// TestRangeSumValidatesEndpointsBeforeLoGtHi pins the order: even when lo > hi,
// an out-of-domain endpoint TRAPS (emptiness is never inferred from an
// out-of-domain endpoint).
func TestRangeSumValidatesEndpointsBeforeLoGtHi(t *testing.T) {
	f := NewFenwickTreeWithSize(4)
	mustPanic(t, "RangeSum(4,-1)", func() { f.RangeSum(4, -1) }) // both invalid, lo>hi
	mustPanic(t, "RangeSum(4,1)", func() { f.RangeSum(4, 1) })   // lo invalid, lo>hi
	mustPanic(t, "RangeSum(2,4)", func() { f.RangeSum(2, 4) })   // hi invalid, lo<hi
}

func TestFenwickIdentityVsBruteForceRandomized(t *testing.T) {
	rng := &lcg{s: 0x1234_5678_9abc_def0}
	for trial := 0; trial < 200; trial++ {
		n := 1 + rng.nextN(20)
		f := NewFenwickTreeWithSize(n)
		b := newBrute(n)
		ops := 5 + rng.nextN(40)
		for o := 0; o < ops; o++ {
			i := rng.nextN(n)
			var v int32
			switch rng.nextU64() % 5 {
			case 0:
				v = int32Min
			case 1:
				v = int32Max
			default:
				v = rng.nextI32()
			}
			if rng.nextU64()%2 == 0 {
				f.Update(i, v)
				b.update(i, v)
			} else {
				f.Set(i, v)
				b.set(i, v)
			}
		}
		for i := 0; i < n; i++ {
			if f.Get(i) != b.get(i) {
				t.Fatalf("trial %d get %d: %d vs %d", trial, i, f.Get(i), b.get(i))
			}
			if f.PrefixSum(i) != b.prefixSum(i) {
				t.Fatalf("trial %d prefix %d: %d vs %d", trial, i, f.PrefixSum(i), b.prefixSum(i))
			}
		}
		for lo := 0; lo < n; lo++ {
			for hi := 0; hi < n; hi++ {
				if f.RangeSum(lo, hi) != b.rangeSum(lo, hi) {
					t.Fatalf("trial %d range %d..%d: %d vs %d",
						trial, lo, hi, f.RangeSum(lo, hi), b.rangeSum(lo, hi))
				}
			}
		}
		if f.Total() != b.total() {
			t.Fatalf("trial %d total: %d vs %d", trial, f.Total(), b.total())
		}
	}
}

func TestBuildDeterminismRandomized(t *testing.T) {
	rng := &lcg{s: 0xdead_beef_cafe_babe}
	for trial := 0; trial < 200; trial++ {
		n := rng.nextN(20)
		vals := make([]int32, n)
		for i := range vals {
			switch rng.nextU64() % 4 {
			case 0:
				vals[i] = int32Min
			case 1:
				vals[i] = int32Max
			default:
				vals[i] = rng.nextI32()
			}
		}
		built := NewFenwickTreeFromValues(vals)
		updated := NewFenwickTreeWithSize(n)
		for i, v := range vals {
			updated.Update(i, v)
		}
		if !reflect.DeepEqual(built.CanonicalTree(), updated.CanonicalTree()) {
			t.Fatalf("trial %d build determinism broken: %v vs %v",
				trial, built.CanonicalTree(), updated.CanonicalTree())
		}
	}
}
