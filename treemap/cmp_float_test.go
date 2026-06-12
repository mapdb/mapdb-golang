package treemap

import (
	"math"
	"testing"
)

// Regression for the intransitive-comparator bug: the old comparator placed
// NaN between negative and positive floats (raw bit compare), so a tree's
// binary search lost keys when NaN coexisted with negatives.
func TestFloat32_NaNWithNegativesFindable(t *testing.T) {
	nan := float32(math.NaN())
	m := NewFloat32Int32()
	m.Put(nan, 1)
	m.Put(2.0, 2)
	m.Put(-0.5, 3)
	m.Put(-1.0, 4)

	for _, tc := range []struct {
		k float32
		v int32
	}{{nan, 1}, {2.0, 2}, {-0.5, 3}, {-1.0, 4}} {
		got, ok := m.Get(tc.k)
		if !ok || got != tc.v {
			t.Fatalf("Get(%v) = (%v, %v), want (%v, true) — key lost by intransitive comparator", tc.k, got, ok, tc.v)
		}
	}
	if m.Len() != 4 {
		t.Fatalf("Size() = %d, want 4", m.Len())
	}
}

// +0.0 and -0.0 must be distinct keys (total order places -0.0 < +0.0).
func TestFloat32_SignedZeroDistinct(t *testing.T) {
	m := NewFloat32Int32()
	m.Put(float32(math.Copysign(0, 1)), 10)  // +0.0
	m.Put(float32(math.Copysign(0, -1)), 20) // -0.0
	if m.Len() != 2 {
		t.Fatalf("Size() = %d, want 2 (+0.0 and -0.0 distinct)", m.Len())
	}
	if v, _ := m.Get(float32(math.Copysign(0, 1))); v != 10 {
		t.Fatalf("Get(+0.0) = %d, want 10", v)
	}
	if v, _ := m.Get(float32(math.Copysign(0, -1))); v != 20 {
		t.Fatalf("Get(-0.0) = %d, want 20", v)
	}
}
