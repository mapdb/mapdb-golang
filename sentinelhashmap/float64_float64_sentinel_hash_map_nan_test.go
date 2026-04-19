package sentinelhashmap

import (
	"math"
	"testing"
)

func TestFloat64Float64SentinelHashMap_NaNKey(t *testing.T) {
	m := NewFloat64Float64SentinelHashMap()
	nan := math.NaN()
	m.Put(nan, 42.0)

	// NaN key must be retrievable via bit-level comparison.
	if v, ok := m.Get(nan); !ok || v != 42.0 {
		t.Errorf("Get(NaN) = (%v, %v), want (42, true)", v, ok)
	}
	if !m.ContainsKey(nan) {
		t.Error("ContainsKey(NaN) should be true")
	}
	if m.Size() != 1 {
		t.Errorf("Size = %d, want 1", m.Size())
	}
}

func TestFloat64Float64SentinelHashMap_NaNValue(t *testing.T) {
	m := NewFloat64Float64SentinelHashMap()
	nan := math.NaN()
	m.Put(2.5, nan)
	if !m.ContainsValue(nan) {
		t.Error("ContainsValue(NaN) should be true (bit-level comparison)")
	}
}

func TestFloat64Float64SentinelHashMap_NaNRemove(t *testing.T) {
	m := NewFloat64Float64SentinelHashMap()
	nan := math.NaN()
	m.Put(nan, 99.0)
	if old, ok := m.Remove(nan); !ok || old != 99.0 {
		t.Errorf("Remove(NaN) = (%v, %v), want (99, true)", old, ok)
	}
	if m.ContainsKey(nan) {
		t.Error("NaN should be gone after Remove")
	}
}
