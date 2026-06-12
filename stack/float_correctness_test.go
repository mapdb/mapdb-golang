package stack

import (
	"math"
	"testing"
)

// These tests exercise the float bit-pattern equality fix for Contains and
// Equals on the stack family. They would FAIL with the previous raw ==
// comparison: NaN == NaN is false, and +0.0 == -0.0 is true.

func TestFloat32ContainsNaN(t *testing.T) {
	nan := float32(math.NaN())
	s := Float32Of(1, nan, 2)
	if !s.Contains(nan) {
		t.Fatal("Contains(NaN) should be true when NaN was pushed")
	}
}

func TestFloat64ContainsNaN(t *testing.T) {
	nan := math.NaN()
	s := Float64Of(1, nan, 2)
	if !s.Contains(nan) {
		t.Fatal("Contains(NaN) should be true when NaN was pushed")
	}
}

func TestFloat32EqualsNaN(t *testing.T) {
	nan := float32(math.NaN())
	a := Float32Of(nan)
	b := Float32Of(nan)
	if !a.Equals(b) {
		t.Fatal("two stacks holding [NaN] should be Equals by bit pattern")
	}
}

func TestFloat64EqualsNaN(t *testing.T) {
	nan := math.NaN()
	a := Float64Of(nan)
	b := Float64Of(nan)
	if !a.Equals(b) {
		t.Fatal("two stacks holding [NaN] should be Equals by bit pattern")
	}
}

func TestFloat32SignedZeroDistinct(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	s := Float32Of(negZero)
	if s.Contains(posZero) {
		t.Fatal("stack holding -0.0 must not Contains(+0.0)")
	}
	if !s.Contains(negZero) {
		t.Fatal("stack holding -0.0 must Contains(-0.0)")
	}
	if Float32Of(negZero).Equals(Float32Of(posZero)) {
		t.Fatal("[-0.0] must not Equals [+0.0]")
	}
}

func TestFloat64SignedZeroDistinct(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	s := Float64Of(negZero)
	if s.Contains(posZero) {
		t.Fatal("stack holding -0.0 must not Contains(+0.0)")
	}
	if Float64Of(negZero).Equals(Float64Of(posZero)) {
		t.Fatal("[-0.0] must not Equals [+0.0]")
	}
}
