package deque

import (
	"math"
	"testing"
)

// These tests exercise the synchronized-deque Equals bit-pattern fix (the
// synchronized Equals previously duplicated a raw != loop instead of routing
// floats through the bit pattern). Contains delegates to the base deque, which
// already used the bit pattern; it is checked here for completeness.

func TestSynchronizedFloat32EqualsNaN(t *testing.T) {
	nan := float32(math.NaN())
	a := NewSynchronizedFloat32()
	a.AddLast(nan)
	b := NewSynchronizedFloat32()
	b.AddLast(nan)
	if !a.Equals(b) {
		t.Fatal("two sync deques [NaN] should be Equals by bit pattern")
	}
}

func TestSynchronizedFloat64EqualsNaN(t *testing.T) {
	nan := math.NaN()
	a := NewSynchronizedFloat64()
	a.AddLast(nan)
	b := NewSynchronizedFloat64()
	b.AddLast(nan)
	if !a.Equals(b) {
		t.Fatal("two sync deques [NaN] should be Equals by bit pattern")
	}
}

func TestSynchronizedFloat32EqualsSignedZero(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	a := NewSynchronizedFloat32()
	a.AddLast(negZero)
	b := NewSynchronizedFloat32()
	b.AddLast(posZero)
	if a.Equals(b) {
		t.Fatal("sync deque [-0.0] must not Equals [+0.0]")
	}
}

func TestSynchronizedFloat64EqualsSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	a := NewSynchronizedFloat64()
	a.AddLast(negZero)
	b := NewSynchronizedFloat64()
	b.AddLast(posZero)
	if a.Equals(b) {
		t.Fatal("sync deque [-0.0] must not Equals [+0.0]")
	}
}

func TestSynchronizedFloat32ContainsNaN(t *testing.T) {
	nan := float32(math.NaN())
	d := NewSynchronizedFloat32()
	d.AddLast(nan)
	if !d.Contains(nan) {
		t.Fatal("sync deque Contains(NaN) should be true")
	}
}
