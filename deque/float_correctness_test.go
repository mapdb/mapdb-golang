package deque

import (
	"math"
	"testing"
)

// These tests exercise the synchronized-deque Equals bit-pattern fix (the
// synchronized Equals previously duplicated a raw != loop instead of routing
// floats through the bit pattern). Contains delegates to the base deque, which
// already used the bit pattern; it is checked here for completeness.

func TestSynchronizedFloat32ArrayDequeEqualsNaN(t *testing.T) {
	nan := float32(math.NaN())
	a := NewSynchronizedFloat32ArrayDeque()
	a.AddLast(nan)
	b := NewSynchronizedFloat32ArrayDeque()
	b.AddLast(nan)
	if !a.Equals(b) {
		t.Fatal("two sync deques [NaN] should be Equals by bit pattern")
	}
}

func TestSynchronizedFloat64ArrayDequeEqualsNaN(t *testing.T) {
	nan := math.NaN()
	a := NewSynchronizedFloat64ArrayDeque()
	a.AddLast(nan)
	b := NewSynchronizedFloat64ArrayDeque()
	b.AddLast(nan)
	if !a.Equals(b) {
		t.Fatal("two sync deques [NaN] should be Equals by bit pattern")
	}
}

func TestSynchronizedFloat32ArrayDequeEqualsSignedZero(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	a := NewSynchronizedFloat32ArrayDeque()
	a.AddLast(negZero)
	b := NewSynchronizedFloat32ArrayDeque()
	b.AddLast(posZero)
	if a.Equals(b) {
		t.Fatal("sync deque [-0.0] must not Equals [+0.0]")
	}
}

func TestSynchronizedFloat64ArrayDequeEqualsSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	a := NewSynchronizedFloat64ArrayDeque()
	a.AddLast(negZero)
	b := NewSynchronizedFloat64ArrayDeque()
	b.AddLast(posZero)
	if a.Equals(b) {
		t.Fatal("sync deque [-0.0] must not Equals [+0.0]")
	}
}

func TestSynchronizedFloat32ArrayDequeContainsNaN(t *testing.T) {
	nan := float32(math.NaN())
	d := NewSynchronizedFloat32ArrayDeque()
	d.AddLast(nan)
	if !d.Contains(nan) {
		t.Fatal("sync deque Contains(NaN) should be true")
	}
}
