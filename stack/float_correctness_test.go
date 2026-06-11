package stack

import (
	"math"
	"testing"
)

// These tests exercise the float bit-pattern equality fix for Contains and
// Equals on the stack family. They would FAIL with the previous raw ==
// comparison: NaN == NaN is false, and +0.0 == -0.0 is true.

func TestFloat32ArrayStackContainsNaN(t *testing.T) {
	nan := float32(math.NaN())
	s := Float32ArrayStackOf(1, nan, 2)
	if !s.Contains(nan) {
		t.Fatal("Contains(NaN) should be true when NaN was pushed")
	}
}

func TestFloat64ArrayStackContainsNaN(t *testing.T) {
	nan := math.NaN()
	s := Float64ArrayStackOf(1, nan, 2)
	if !s.Contains(nan) {
		t.Fatal("Contains(NaN) should be true when NaN was pushed")
	}
}

func TestFloat32ArrayStackEqualsNaN(t *testing.T) {
	nan := float32(math.NaN())
	a := Float32ArrayStackOf(nan)
	b := Float32ArrayStackOf(nan)
	if !a.Equals(b) {
		t.Fatal("two stacks holding [NaN] should be Equals by bit pattern")
	}
}

func TestFloat64ArrayStackEqualsNaN(t *testing.T) {
	nan := math.NaN()
	a := Float64ArrayStackOf(nan)
	b := Float64ArrayStackOf(nan)
	if !a.Equals(b) {
		t.Fatal("two stacks holding [NaN] should be Equals by bit pattern")
	}
}

func TestFloat32ArrayStackSignedZeroDistinct(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	s := Float32ArrayStackOf(negZero)
	if s.Contains(posZero) {
		t.Fatal("stack holding -0.0 must not Contains(+0.0)")
	}
	if !s.Contains(negZero) {
		t.Fatal("stack holding -0.0 must Contains(-0.0)")
	}
	if Float32ArrayStackOf(negZero).Equals(Float32ArrayStackOf(posZero)) {
		t.Fatal("[-0.0] must not Equals [+0.0]")
	}
}

func TestFloat64ArrayStackSignedZeroDistinct(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	s := Float64ArrayStackOf(negZero)
	if s.Contains(posZero) {
		t.Fatal("stack holding -0.0 must not Contains(+0.0)")
	}
	if Float64ArrayStackOf(negZero).Equals(Float64ArrayStackOf(posZero)) {
		t.Fatal("[-0.0] must not Equals [+0.0]")
	}
}
