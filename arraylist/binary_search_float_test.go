package arraylist

import (
	"math"
	"testing"
)

// BinarySearch on a float list must navigate by the SAME IEEE totalOrder that
// Sort() imposes (cmpFloatNN), not a raw `<`. A raw `<` is incoherent with the
// bit-pattern equality check and total-order sort: it never advances past a
// signed-zero twin or a NaN (every `mid < NaN` is false), so a present +0.0 or
// NaN is reported missing. These are the regression cases.

func TestFloat64ArrayListBinarySearchSignedZeroAndNaN(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	nan := math.NaN()

	// Total order: -0.0 < +0.0 < 1.0 < NaN.
	l := Float64ArrayListOf(nan, 1.0, posZero, negZero)
	l.Sort()

	get := func(i int) float64 { v, _ := l.Get(i); return v }
	// +0.0 is present and distinct from -0.0 under bit-pattern equality.
	if idx, ok := l.BinarySearch(posZero); !ok || math.Float64bits(get(idx)) != math.Float64bits(posZero) {
		t.Errorf("BinarySearch(+0.0) = (%d,%v), want the +0.0 slot found", idx, ok)
	}
	if idx, ok := l.BinarySearch(negZero); !ok || math.Float64bits(get(idx)) != math.Float64bits(negZero) {
		t.Errorf("BinarySearch(-0.0) = (%d,%v), want the -0.0 slot found", idx, ok)
	}
	// NaN sorts to the total-order maximum and must be findable.
	if idx, ok := l.BinarySearch(nan); !ok || !math.IsNaN(get(idx)) {
		t.Errorf("BinarySearch(NaN) = (%d,%v), want NaN found at the top", idx, ok)
	}
	// A genuinely absent value is still reported missing.
	if _, ok := l.BinarySearch(2.0); ok {
		t.Errorf("BinarySearch(2.0) reported found, want missing")
	}
}

func TestFloat32ArrayListBinarySearchSignedZeroAndNaN(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0.0)
	nan := float32(math.NaN())

	l := Float32ArrayListOf(nan, 1.0, posZero, negZero)
	l.Sort()

	get := func(i int) float32 { v, _ := l.Get(i); return v }
	if idx, ok := l.BinarySearch(posZero); !ok || math.Float32bits(get(idx)) != math.Float32bits(posZero) {
		t.Errorf("BinarySearch(+0.0) = (%d,%v), want the +0.0 slot found", idx, ok)
	}
	if idx, ok := l.BinarySearch(negZero); !ok || math.Float32bits(get(idx)) != math.Float32bits(negZero) {
		t.Errorf("BinarySearch(-0.0) = (%d,%v), want the -0.0 slot found", idx, ok)
	}
	if idx, ok := l.BinarySearch(nan); !ok || !math.IsNaN(float64(get(idx))) {
		t.Errorf("BinarySearch(NaN) = (%d,%v), want NaN found at the top", idx, ok)
	}
	if _, ok := l.BinarySearch(2.0); ok {
		t.Errorf("BinarySearch(2.0) reported found, want missing")
	}
}
