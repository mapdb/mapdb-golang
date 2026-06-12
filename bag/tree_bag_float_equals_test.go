package bag

import (
	"math"
	"testing"
)

// These tests exercise the tree-bag Equals bit-pattern fix. The previous raw
// != comparison made two [NaN] tree bags unequal and conflated +0.0/-0.0.
// (The hash bag was already correct — it keys its map by the bit pattern.)

func TestTreeFloat32EqualsNaN(t *testing.T) {
	nan := float32(math.NaN())
	a := NewTreeFloat32()
	a.Add(nan)
	b := NewTreeFloat32()
	b.Add(nan)
	if !a.Equals(b) {
		t.Fatal("two tree bags {NaN} should be Equals by bit pattern")
	}
}

func TestTreeFloat64EqualsNaN(t *testing.T) {
	nan := math.NaN()
	a := NewTreeFloat64()
	a.Add(nan)
	b := NewTreeFloat64()
	b.Add(nan)
	if !a.Equals(b) {
		t.Fatal("two tree bags {NaN} should be Equals by bit pattern")
	}
}

func TestTreeFloat32EqualsSignedZero(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	a := NewTreeFloat32()
	a.Add(negZero)
	b := NewTreeFloat32()
	b.Add(posZero)
	if a.Equals(b) {
		t.Fatal("tree bag {-0.0} must not Equals {+0.0}")
	}
}

func TestTreeFloat64EqualsSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	a := NewTreeFloat64()
	a.Add(negZero)
	b := NewTreeFloat64()
	b.Add(posZero)
	if a.Equals(b) {
		t.Fatal("tree bag {-0.0} must not Equals {+0.0}")
	}
}
