package tuple

import (
	"math"
	"testing"
)

// These tests exercise the total-order routing of CompareTo for pairs with a
// float field. The previous raw < / > comparison was NaN-unsafe (every
// comparison with NaN returned 0, making the order intransitive) and conflated
// +0.0/-0.0.

func TestFloat32Float32PairCompareToNaNIsMax(t *testing.T) {
	nan := float32(math.NaN())
	inf := float32(math.Inf(1))
	a := NewFloat32Float32Pair(nan, 0)
	b := NewFloat32Float32Pair(inf, 0)
	// NaN sorts above +Inf in total order.
	if a.CompareTo(b) <= 0 {
		t.Fatalf("Pair(NaN,_).CompareTo(Pair(+Inf,_)) = %d, want > 0", a.CompareTo(b))
	}
	if b.CompareTo(a) >= 0 {
		t.Fatalf("Pair(+Inf,_).CompareTo(Pair(NaN,_)) = %d, want < 0", b.CompareTo(a))
	}
	// Reflexive: NaN-keyed pair equals itself.
	if a.CompareTo(a) != 0 {
		t.Fatalf("Pair(NaN,_).CompareTo(self) = %d, want 0", a.CompareTo(a))
	}
}

func TestFloat32Float32PairCompareToTransitive(t *testing.T) {
	nan := float32(math.NaN())
	x := NewFloat32Float32Pair(1, 0)
	y := NewFloat32Float32Pair(2, 0)
	z := NewFloat32Float32Pair(nan, 0)
	// x < y < z must hold consistently (a proper total order).
	if !(x.CompareTo(y) < 0 && y.CompareTo(z) < 0 && x.CompareTo(z) < 0) {
		t.Fatal("CompareTo is not transitive across NaN-keyed pairs")
	}
}

func TestInt32Float32PairCompareToSecondFieldNaN(t *testing.T) {
	nan := float32(math.NaN())
	// Equal first field; second field decides via total order (NaN is max).
	a := NewInt32Float32Pair(5, nan)
	b := NewInt32Float32Pair(5, float32(math.Inf(1)))
	if a.CompareTo(b) <= 0 {
		t.Fatalf("second-field NaN should sort above +Inf, got %d", a.CompareTo(b))
	}
}

func TestFloat64Float64PairCompareToSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	a := NewFloat64Float64Pair(negZero, 0)
	b := NewFloat64Float64Pair(posZero, 0)
	// -0.0 sorts strictly below +0.0 in total order.
	if a.CompareTo(b) >= 0 {
		t.Fatalf("Pair(-0.0,_).CompareTo(Pair(+0.0,_)) = %d, want < 0", a.CompareTo(b))
	}
}
