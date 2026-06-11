package bag

import (
	"math"
	"testing"
)

// These tests exercise the tree-bag Equals bit-pattern fix. The previous raw
// != comparison made two [NaN] tree bags unequal and conflated +0.0/-0.0.
// (The hash bag was already correct — it keys its map by the bit pattern.)

func TestFloat32TreeBagEqualsNaN(t *testing.T) {
	nan := float32(math.NaN())
	a := NewFloat32TreeBag()
	a.Add(nan)
	b := NewFloat32TreeBag()
	b.Add(nan)
	if !a.Equals(b) {
		t.Fatal("two tree bags {NaN} should be Equals by bit pattern")
	}
}

func TestFloat64TreeBagEqualsNaN(t *testing.T) {
	nan := math.NaN()
	a := NewFloat64TreeBag()
	a.Add(nan)
	b := NewFloat64TreeBag()
	b.Add(nan)
	if !a.Equals(b) {
		t.Fatal("two tree bags {NaN} should be Equals by bit pattern")
	}
}

func TestFloat32TreeBagEqualsSignedZero(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	a := NewFloat32TreeBag()
	a.Add(negZero)
	b := NewFloat32TreeBag()
	b.Add(posZero)
	if a.Equals(b) {
		t.Fatal("tree bag {-0.0} must not Equals {+0.0}")
	}
}

func TestFloat64TreeBagEqualsSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	a := NewFloat64TreeBag()
	a.Add(negZero)
	b := NewFloat64TreeBag()
	b.Add(posZero)
	if a.Equals(b) {
		t.Fatal("tree bag {-0.0} must not Equals {+0.0}")
	}
}
