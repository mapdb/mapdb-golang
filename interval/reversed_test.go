package interval

import (
	"math"
	"testing"
)

// intervalOps is the read surface shared by every signed-integer interval,
// used by the width-parametrised Reversed tests below.
type intervalOps[T comparable] interface {
	Len() int
	Get(int) T
	Contains(T) bool
	ToSlice() []T
}

// checkReversed asserts that rev holds exactly want (in order), that source
// and reverse agree on Len and on Contains for every value in probe, and that
// twice (the source reversed twice) has the source's element sequence.
func checkReversed[T comparable](t *testing.T, name string, src, rev, twice intervalOps[T], want, probe []T) {
	t.Helper()
	got := rev.ToSlice()
	if len(got) != len(want) {
		t.Fatalf("%s: reversed = %v, want %v", name, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s: reversed[%d] = %v, want %v", name, i, got[i], want[i])
		}
	}
	if rev.Len() != src.Len() {
		t.Errorf("%s: reversed Len = %d, source Len = %d", name, rev.Len(), src.Len())
	}
	if rev.Get(0) != want[0] || rev.Get(rev.Len()-1) != want[len(want)-1] {
		t.Errorf("%s: reversed Get(0)/Get(Len-1) = %v/%v, want %v/%v",
			name, rev.Get(0), rev.Get(rev.Len()-1), want[0], want[len(want)-1])
	}
	for _, v := range probe {
		if src.Contains(v) != rev.Contains(v) {
			t.Errorf("%s: Contains(%v): source %v, reversed %v", name, v, src.Contains(v), rev.Contains(v))
		}
	}
	srcSeq, twiceSeq := src.ToSlice(), twice.ToSlice()
	if len(srcSeq) != len(twiceSeq) {
		t.Fatalf("%s: reversed twice = %v, want source %v", name, twiceSeq, srcSeq)
	}
	for i := range srcSeq {
		if srcSeq[i] != twiceSeq[i] {
			t.Errorf("%s: reversed twice[%d] = %v, want %v", name, i, twiceSeq[i], srcSeq[i])
		}
	}
}

// Spec (algorithms.md "Reversed() starts from the last element"): the reverse
// starts from the last element actually produced, not from `to`, which is
// only an inclusive bound and may sit off the step grid.

func TestInt8_ReversedOffGridKeepsElements(t *testing.T) {
	const maxT, minT = math.MaxInt8, math.MinInt8
	probe := []int8{minT, minT + 1, minT + 4, minT + 7, -7, -2, -1, 0, 1, 3, 5, 6, 8, 9, 10, maxT - 7, maxT - 4, maxT - 1, maxT}
	cases := []struct {
		name           string
		from, to, step int8
		want           []int8
	}{
		{"off_grid", 0, 10, 3, []int8{9, 6, 3, 0}},
		{"descending_off_grid", 10, 0, -3, []int8{1, 4, 7, 10}},
		{"step_exceeds_range", 0, 5, maxT, []int8{0}},
		{"negative_off_grid", -7, 9, 5, []int8{8, 3, -2, -7}},
		{"max_boundary", maxT - 7, maxT, 3, []int8{maxT - 1, maxT - 4, maxT - 7}},
		{"min_boundary_descending", minT + 7, minT, -3, []int8{minT + 1, minT + 4, minT + 7}},
		{"near_min_step", 0, minT + 1, minT + 1, []int8{minT + 1, 0}},
		{"single", 5, 5, 1, []int8{5}},
	}
	for _, c := range cases {
		src := NewInt8(c.from, c.to, c.step)
		checkReversed[int8](t, c.name, src, src.Reversed(), src.Reversed().Reversed(), c.want, probe)
	}
}

func TestInt16_ReversedOffGridKeepsElements(t *testing.T) {
	const maxT, minT = math.MaxInt16, math.MinInt16
	probe := []int16{minT, minT + 1, minT + 4, minT + 7, -7, -2, -1, 0, 1, 3, 5, 6, 8, 9, 10, maxT - 7, maxT - 4, maxT - 1, maxT}
	cases := []struct {
		name           string
		from, to, step int16
		want           []int16
	}{
		{"off_grid", 0, 10, 3, []int16{9, 6, 3, 0}},
		{"descending_off_grid", 10, 0, -3, []int16{1, 4, 7, 10}},
		{"step_exceeds_range", 0, 5, maxT, []int16{0}},
		{"negative_off_grid", -7, 9, 5, []int16{8, 3, -2, -7}},
		{"max_boundary", maxT - 7, maxT, 3, []int16{maxT - 1, maxT - 4, maxT - 7}},
		{"min_boundary_descending", minT + 7, minT, -3, []int16{minT + 1, minT + 4, minT + 7}},
		{"near_min_step", 0, minT + 1, minT + 1, []int16{minT + 1, 0}},
		{"single", 5, 5, 1, []int16{5}},
	}
	for _, c := range cases {
		src := NewInt16(c.from, c.to, c.step)
		checkReversed[int16](t, c.name, src, src.Reversed(), src.Reversed().Reversed(), c.want, probe)
	}
}

func TestInt32_ReversedOffGridKeepsElements(t *testing.T) {
	const maxT, minT = math.MaxInt32, math.MinInt32
	probe := []int32{minT, minT + 1, minT + 4, minT + 7, -7, -2, -1, 0, 1, 3, 5, 6, 8, 9, 10, maxT - 7, maxT - 4, maxT - 1, maxT}
	cases := []struct {
		name           string
		from, to, step int32
		want           []int32
	}{
		{"off_grid", 0, 10, 3, []int32{9, 6, 3, 0}},
		{"descending_off_grid", 10, 0, -3, []int32{1, 4, 7, 10}},
		{"step_exceeds_range", 0, 5, maxT, []int32{0}},
		{"negative_off_grid", -7, 9, 5, []int32{8, 3, -2, -7}},
		{"max_boundary", 2147483640, 2147483647, 3, []int32{2147483646, 2147483643, 2147483640}},
		{"min_boundary_descending", -2147483641, -2147483648, -3, []int32{-2147483647, -2147483644, -2147483641}},
		{"near_min_step", 0, minT + 1, minT + 1, []int32{minT + 1, 0}},
		{"single", 5, 5, 1, []int32{5}},
	}
	for _, c := range cases {
		src := NewInt32(c.from, c.to, c.step)
		checkReversed[int32](t, c.name, src, src.Reversed(), src.Reversed().Reversed(), c.want, probe)
	}
}

func TestInt64_ReversedOffGridKeepsElements(t *testing.T) {
	const maxT, minT = math.MaxInt64, math.MinInt64
	probe := []int64{minT, minT + 1, minT + 4, minT + 7, -7, -2, -1, 0, 1, 3, 5, 6, 8, 9, 10, maxT - 7, maxT - 4, maxT - 1, maxT}
	cases := []struct {
		name           string
		from, to, step int64
		want           []int64
	}{
		{"off_grid", 0, 10, 3, []int64{9, 6, 3, 0}},
		{"descending_off_grid", 10, 0, -3, []int64{1, 4, 7, 10}},
		{"step_exceeds_range", 0, 5, maxT, []int64{0}},
		{"negative_off_grid", -7, 9, 5, []int64{8, 3, -2, -7}},
		{"max_boundary", maxT - 7, maxT, 3, []int64{maxT - 1, maxT - 4, maxT - 7}},
		{"min_boundary_descending", minT + 7, minT, -3, []int64{minT + 1, minT + 4, minT + 7}},
		{"near_min_step", 0, minT + 1, minT + 1, []int64{minT + 1, 0}},
		{"single", 5, 5, 1, []int64{5}},
	}
	for _, c := range cases {
		src := NewInt64(c.from, c.to, c.step)
		checkReversed[int64](t, c.name, src, src.Reversed(), src.Reversed().Reversed(), c.want, probe)
	}
}

// TestInt64_ReversedFullRangeNoSizeCap pins the remainder form: Len() caps at
// MaxInt for the full int64 range, so deriving the last element from
// Get(Len()-1) would be wrong; the remainder form has no cap. The full
// range is not iterable to completion, so only the shape, the capped Len(),
// Get(0) and the first element of All() are checked. Len() must reach the
// cap rather than wrap to 0: the quotient distance/absStep is MaxUint64
// there and a naive +1 wrapped before the cap check.
func TestInt64_ReversedFullRangeNoSizeCap(t *testing.T) {
	src := NewInt64(math.MinInt64, math.MaxInt64, 1)
	rev := src.Reversed()
	if rev.From() != math.MaxInt64 || rev.To() != math.MinInt64 || rev.Step() != -1 {
		t.Errorf("full-range reversed = [%d, %d] step %d, want [MaxInt64, MinInt64] step -1",
			rev.From(), rev.To(), rev.Step())
	}
	if src.Len() != math.MaxInt {
		t.Errorf("full-range Len() = %d, want MaxInt", src.Len())
	}
	if rev.Len() != math.MaxInt {
		t.Errorf("full-range reversed Len() = %d, want MaxInt", rev.Len())
	}
	if got := src.Get(0); got != math.MinInt64 {
		t.Errorf("full-range Get(0) = %d, want MinInt64", got)
	}
	if got := rev.Get(0); got != math.MaxInt64 {
		t.Errorf("full-range reversed Get(0) = %d, want MaxInt64", got)
	}
	seen := false
	for v := range rev.All() {
		if v != math.MaxInt64 {
			t.Errorf("full-range reversed first element from All() = %d, want MaxInt64", v)
		}
		seen = true
		break
	}
	if !seen {
		t.Errorf("full-range reversed All() produced no elements")
	}
	off := NewInt64(math.MinInt64, math.MaxInt64, 7).Reversed() // 2^64-1 = 1 (mod 7): MaxInt64 is off the grid
	if off.From() != math.MaxInt64-1 || off.To() != math.MinInt64 || off.Step() != -7 {
		t.Errorf("off-grid full-range reversed = [%d, %d] step %d, want [MaxInt64-1, MinInt64] step -7",
			off.From(), off.To(), off.Step())
	}
}
