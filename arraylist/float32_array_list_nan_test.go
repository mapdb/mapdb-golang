package arraylist

import (
	"math"
	"testing"
)

// negZeroF32 / posZeroF32 give bit-distinct -0 and +0 so the tests can assert
// the total-order comparator and bit-keyed dedup keep them apart.
var (
	posZeroF32 = float32(0.0)
	negZeroF32 = float32(math.Copysign(0, -1))
	nanF32     = float32(math.NaN())
)

func bitsF32(v float32) uint32 { return math.Float32bits(v) }

// TestFloat32ArrayList_SortTotalOrder verifies Sort uses the IEEE total order:
// -Inf < negatives < -0 < +0 < positives < +Inf < (positive) NaN, and that the
// two zeroes are kept distinct rather than collapsed.
func TestFloat32ArrayList_SortTotalOrder(t *testing.T) {
	l := NewFloat32ArrayList()
	// Deliberately scrambled, including both zeroes and a NaN.
	l.AddAll(2.0, nanF32, -1.0, float32(math.Inf(1)), float32(math.Inf(-1)), posZeroF32, negZeroF32)
	l.Sort()
	got := l.ToSlice()

	want := []float32{
		float32(math.Inf(-1)),
		-1.0,
		negZeroF32,
		posZeroF32,
		2.0,
		float32(math.Inf(1)),
		nanF32,
	}
	if len(got) != len(want) {
		t.Fatalf("length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if bitsF32(got[i]) != bitsF32(want[i]) {
			t.Fatalf("Sort order at %d: got %v (bits %#x) want %v (bits %#x); full=%v",
				i, got[i], bitsF32(got[i]), want[i], bitsF32(want[i]), got)
		}
	}
	// Explicit -0/+0 ordering check: -0 must precede +0.
	if bitsF32(got[2]) != bitsF32(negZeroF32) || bitsF32(got[3]) != bitsF32(posZeroF32) {
		t.Fatalf("expected -0 before +0, got %v", got)
	}
}

// TestFloat32ArrayList_MinMaxWithNaN verifies that with total ordering NaN is
// the maximum (sorts above +Inf) and is never selected as the minimum.
func TestFloat32ArrayList_MinMaxWithNaN(t *testing.T) {
	l := NewFloat32ArrayList()
	l.AddAll(3.0, nanF32, 1.0, 2.0)

	mn, ok := l.Min()
	if !ok {
		t.Fatal("Min: expected ok")
	}
	if mn != 1.0 {
		t.Fatalf("Min: got %v want 1.0", mn)
	}

	mx, ok := l.Max()
	if !ok {
		t.Fatal("Max: expected ok")
	}
	if !math.IsNaN(float64(mx)) {
		t.Fatalf("Max: got %v want NaN", mx)
	}

	// Empty list: no min/max.
	empty := NewFloat32ArrayList()
	if _, ok := empty.Min(); ok {
		t.Fatal("Min on empty: expected !ok")
	}
	if _, ok := empty.Max(); ok {
		t.Fatal("Max on empty: expected !ok")
	}
}

// TestFloat32ArrayList_DistinctNaNAndZeros verifies Distinct dedupes NaN against
// itself (bit-keyed) and keeps -0 and +0 as separate values.
func TestFloat32ArrayList_DistinctNaNAndZeros(t *testing.T) {
	l := NewFloat32ArrayList()
	l.AddAll(1.0, nanF32, 1.0, nanF32, posZeroF32, negZeroF32, posZeroF32)
	got := l.Distinct().ToSlice()

	// Expected first-occurrence order: 1.0, NaN, +0, -0.
	want := []float32{1.0, nanF32, posZeroF32, negZeroF32}
	if len(got) != len(want) {
		t.Fatalf("Distinct length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if bitsF32(got[i]) != bitsF32(want[i]) {
			t.Fatalf("Distinct at %d: got bits %#x want bits %#x; full=%v",
				i, bitsF32(got[i]), bitsF32(want[i]), got)
		}
	}
}

// TestFloat32ArrayList_WithoutAllNaNAndZeros verifies WithoutAll removes NaN
// entries (bit-keyed) and removes only the matching zero sign.
func TestFloat32ArrayList_WithoutAllNaNAndZeros(t *testing.T) {
	// Removing NaN must actually drop every NaN.
	l := NewFloat32ArrayList()
	l.AddAll(1.0, nanF32, 2.0, nanF32)
	l.WithoutAll(nanF32)
	got := l.ToSlice()
	want := []float32{1.0, 2.0}
	if len(got) != len(want) {
		t.Fatalf("WithoutAll(NaN) length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if bitsF32(got[i]) != bitsF32(want[i]) {
			t.Fatalf("WithoutAll(NaN) at %d: got %v want %v", i, got[i], want[i])
		}
	}

	// Removing -0 must leave +0 intact (distinct bit patterns).
	l2 := NewFloat32ArrayList()
	l2.AddAll(posZeroF32, negZeroF32, 5.0)
	l2.WithoutAll(negZeroF32)
	got2 := l2.ToSlice()
	want2 := []float32{posZeroF32, 5.0}
	if len(got2) != len(want2) {
		t.Fatalf("WithoutAll(-0) length: got %d want %d (%v)", len(got2), len(want2), got2)
	}
	if bitsF32(got2[0]) != bitsF32(posZeroF32) {
		t.Fatalf("WithoutAll(-0) removed +0 by mistake: got %v", got2)
	}
}
