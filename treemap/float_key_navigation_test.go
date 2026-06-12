package treemap

import (
	"math"
	"testing"
)

// These tests exercise the total-order routing of key navigation for
// float-keyed maps: Floor/Ceiling exact-match short-circuits and the
// RangeKeys/HeadMap/TailMap bounds. The previous raw ==/</>= comparisons
// conflated +0.0/-0.0 keys and excluded NaN from navigation.

func TestFloat64Int32FloorCeilingSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	m := NewFloat64Int32()
	m.Put(negZero, 10)
	m.Put(posZero, 20)
	if m.Len() != 2 {
		t.Fatalf("expected distinct -0.0/+0.0 keys (size 2), got %d", m.Len())
	}
	// Floor(+0.0) hits the +0.0 key exactly -> value 20.
	if k, v, ok := m.Floor(posZero); !ok || math.Float64bits(k) != math.Float64bits(posZero) || v != 20 {
		t.Fatalf("Floor(+0.0) = (%v,%d), want (+0.0,20)", k, v)
	}
	// Ceiling(-0.0) hits the -0.0 key exactly -> value 10.
	if k, v, ok := m.Ceiling(negZero); !ok || math.Float64bits(k) != math.Float64bits(negZero) || v != 10 {
		t.Fatalf("Ceiling(-0.0) = (%v,%d), want (-0.0,10)", k, v)
	}
}

func TestFloat32Int32NaNIsMaxKey(t *testing.T) {
	nan := float32(math.NaN())
	m := NewFloat32Int32()
	m.Put(1, 1)
	m.Put(float32(math.Inf(1)), 2)
	m.Put(nan, 3)
	if k, v, ok := m.Floor(nan); !ok || !math.IsNaN(float64(k)) || v != 3 {
		t.Fatalf("Floor(NaN) = (%v,%d), want (NaN,3) (NaN is total-order max)", k, v)
	}
	if k, _, ok := m.Max(); !ok || !math.IsNaN(float64(k)) {
		t.Fatalf("Max() key = %v, want NaN", k)
	}
}

func TestFloat64Int32RangeKeysTotalOrder(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	m := NewFloat64Int32()
	m.Put(-1, 1)
	m.Put(negZero, 2)
	m.Put(posZero, 3)
	m.Put(1, 4)
	// Range [+0.0, 1) must exclude -0.0 (total-order less than +0.0). Collect
	// into slices, not a map[float64] — a Go map key conflates -0.0/+0.0 and
	// would mask a wrongly-included -0.0 entry.
	var keyBits []uint64
	var vals []int32
	for k, v := range m.RangeKeys(posZero, 1) {
		keyBits = append(keyBits, math.Float64bits(k))
		vals = append(vals, v)
	}
	if len(keyBits) != 1 || keyBits[0] != math.Float64bits(posZero) || vals[0] != 3 {
		t.Fatalf("RangeKeys(+0.0,1) = keys%v vals%v, want exactly {+0.0:3}", keyBits, vals)
	}
}

// HeadMap/TailMap route their bounds through the same total-order comparator as
// RangeKeys; assert the signed-zero split directly through those entry points.
func TestFloat64Int32HeadTailMapSignedZero(t *testing.T) {
	negZero := math.Copysign(0, -1)
	posZero := 0.0
	m := NewFloat64Int32()
	m.Put(-1, 1)
	m.Put(negZero, 2)
	m.Put(posZero, 3)
	m.Put(1, 4)

	// HeadMap(+0.0) is keys strictly < +0.0 in total order: {-1, -0.0}.
	var headBits []uint64
	for k := range m.HeadMap(posZero) {
		headBits = append(headBits, math.Float64bits(k))
	}
	if len(headBits) != 2 ||
		headBits[0] != math.Float64bits(-1) ||
		headBits[1] != math.Float64bits(negZero) {
		t.Fatalf("HeadMap(+0.0) keys = %v, want {-1, -0.0}", headBits)
	}

	// TailMap(+0.0) is keys >= +0.0 in total order: {+0.0, 1} (excludes -0.0).
	var tailBits []uint64
	for k := range m.TailMap(posZero) {
		tailBits = append(tailBits, math.Float64bits(k))
	}
	if len(tailBits) != 2 ||
		tailBits[0] != math.Float64bits(posZero) ||
		tailBits[1] != math.Float64bits(1) {
		t.Fatalf("TailMap(+0.0) keys = %v, want {+0.0, 1}", tailBits)
	}
}
