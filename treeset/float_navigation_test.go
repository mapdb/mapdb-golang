package treeset

import (
	"math"
	"testing"
)

// These tests exercise the total-order routing of Floor/Ceiling exact-match
// short-circuits and RangeValues bounds for float sets. With the previous raw
// ==/</>= comparisons, the +0.0/-0.0 exact-match short-circuits conflated the
// two zeroes and NaN could not participate in navigation.

func TestFloat32FloorCeilingSignedZero(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	s := Float32Of(negZero, posZero)
	if s.Len() != 2 {
		t.Fatalf("expected -0.0 and +0.0 to be distinct (size 2), got size %d", s.Len())
	}

	// Floor(+0.0) is an exact hit on +0.0 (not conflated with -0.0).
	if got, ok := s.Floor(posZero); !ok || math.Float32bits(got) != math.Float32bits(posZero) {
		t.Fatalf("Floor(+0.0) = %v (bits %x), want +0.0", got, math.Float32bits(got))
	}
	// Ceiling(-0.0) is an exact hit on -0.0.
	if got, ok := s.Ceiling(negZero); !ok || math.Float32bits(got) != math.Float32bits(negZero) {
		t.Fatalf("Ceiling(-0.0) = %v (bits %x), want -0.0", got, math.Float32bits(got))
	}
}

func TestFloat64FloorCeilingSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	s := Float64Of(negZero, posZero)
	if s.Len() != 2 {
		t.Fatalf("expected size 2, got %d", s.Len())
	}
	if got, ok := s.Floor(posZero); !ok || math.Float64bits(got) != math.Float64bits(posZero) {
		t.Fatalf("Floor(+0.0) = %v, want +0.0", got)
	}
	if got, ok := s.Ceiling(negZero); !ok || math.Float64bits(got) != math.Float64bits(negZero) {
		t.Fatalf("Ceiling(-0.0) = %v, want -0.0", got)
	}
}

func TestFloat64NaNIsMax(t *testing.T) {
	nan := math.NaN()
	s := Float64Of(1.0, math.Inf(1), nan)
	// In total order NaN is the maximum, so Floor(NaN) returns NaN itself.
	if got, ok := s.Floor(nan); !ok || !math.IsNaN(got) {
		t.Fatalf("Floor(NaN) = %v, want NaN (NaN is total-order max)", got)
	}
	if got, ok := s.Max(); !ok || !math.IsNaN(got) {
		t.Fatalf("Max() = %v, want NaN", got)
	}
}

func TestFloat32RangeValuesTotalOrder(t *testing.T) {
	negZero := float32(math.Copysign(0, -1))
	posZero := float32(0)
	s := Float32Of(-1, negZero, posZero, 1)
	// Range [+0.0, 1) must EXCLUDE -0.0 (which is < +0.0 in total order) and
	// include +0.0. With raw bounds -0.0 >= +0.0 was true, wrongly including it.
	var got []float32
	for v := range s.RangeValues(posZero, 1) {
		got = append(got, v)
	}
	if len(got) != 1 || math.Float32bits(got[0]) != math.Float32bits(posZero) {
		t.Fatalf("RangeValues(+0.0, 1) = %v, want exactly [+0.0]", got)
	}
}
