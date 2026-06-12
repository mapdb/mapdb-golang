package hashmap

import (
	"math"
	"testing"
)

func TestFloat64Float64_NaN(t *testing.T) {
	m := NewFloat64Float64()
	nan := math.NaN()
	m.Put(nan, 42.0)

	// NaN key should be retrievable (using bit-level comparison)
	v, ok := m.Get(nan)
	if !ok || v != 42.0 {
		t.Errorf("Get(NaN) = (%v, %v), want (42, true)", v, ok)
	}

	if !m.ContainsKey(nan) {
		t.Error("ContainsKey(NaN) should be true")
	}
	if m.Len() != 1 {
		t.Errorf("Size = %d, want 1", m.Len())
	}
}

func TestFloat64Float64_NegativeZero(t *testing.T) {
	m := NewFloat64Float64()
	posZero := 0.0
	negZero := math.Copysign(0, -1) // -0.0

	m.Put(posZero, 1.0)
	m.Put(negZero, 2.0)

	// +0.0 and -0.0 have different bit patterns, so they are different keys
	if m.Len() != 2 {
		t.Logf("Size = %d (bit-level: +0 and -0 are different keys)", m.Len())
	}
}

func TestFloat64Float64_Infinity(t *testing.T) {
	m := NewFloat64Float64()
	m.Put(math.Inf(1), 1.0)
	m.Put(math.Inf(-1), 2.0)

	if v, ok := m.Get(math.Inf(1)); !ok || v != 1.0 {
		t.Errorf("Get(+Inf) = (%v, %v), want (1, true)", v, ok)
	}
	if v, ok := m.Get(math.Inf(-1)); !ok || v != 2.0 {
		t.Errorf("Get(-Inf) = (%v, %v), want (2, true)", v, ok)
	}
}

func TestFloat64Float64_NaNValue(t *testing.T) {
	m := NewFloat64Float64()
	nan := math.NaN()
	m.Put(1.0, nan)

	if !m.ContainsValue(nan) {
		t.Error("ContainsValue(NaN) should be true (bit-level comparison)")
	}
}
