package hashmap

import (
	"math"
	"testing"
)

func TestFloat64Float32_Generated_PutGet(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)

	if v, ok := m.Get(1.0); !ok || v != 1.0 {
		t.Errorf("Get(1.0) = (%v, %v), want (1.0, true)", v, ok)
	}
	if v, ok := m.Get(99.0); ok || v != 0.0 {
		t.Errorf("Get(99.0) = (%v, %v), want (0.0, false)", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("Size() = %d, want 3", m.Len())
	}
}

func TestFloat64Float32_Generated_PutOverwrite(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	old, existed := m.Put(1.0, 2.0)
	if !existed || old != 1.0 {
		t.Errorf("Put overwrite = (%v, %v), want (1.0, true)", old, existed)
	}
	if v, _ := m.Get(1.0); v != 2.0 {
		t.Errorf("Get after overwrite = %v, want 2.0", v)
	}
}

func TestFloat64Float32_Generated_Remove(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	old, ok := m.Remove(1.0)
	if !ok || old != 1.0 {
		t.Errorf("Remove(1.0) = (%v, %v), want (1.0, true)", old, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Len())
	}
	if m.ContainsKey(1.0) {
		t.Error("ContainsKey(1.0) should be false after remove")
	}

	_, ok = m.Remove(99.0)
	if ok {
		t.Error("Remove missing key should return false")
	}
}

func TestFloat64Float32_Generated_ContainsKey(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	if !m.ContainsKey(1.0) {
		t.Error("ContainsKey(1.0) should be true")
	}
	if m.ContainsKey(99.0) {
		t.Error("ContainsKey(99.0) should be false")
	}
}

func TestFloat64Float32_Generated_ContainsValue(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	if !m.ContainsValue(1.0) {
		t.Error("ContainsValue(1.0) should be true")
	}
	if m.ContainsValue(99.0) {
		t.Error("ContainsValue for missing should be false")
	}
}

func TestFloat64Float32_Generated_GetOrDefault(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	if v := m.GetOrDefault(1.0, 3.0); v != 1.0 {
		t.Errorf("GetOrDefault existing = %v, want 1.0", v)
	}
	if v := m.GetOrDefault(99.0, 3.0); v != 3.0 {
		t.Errorf("GetOrDefault missing = %v, want 3.0", v)
	}
}

func TestFloat64Float32_Generated_IsEmpty(t *testing.T) {
	m := NewFloat64Float32()
	if m.Len() != 0 {
		t.Error("New map should be empty")
	}
	m.Put(1.0, 1.0)
	if m.Len() == 0 {
		t.Error("Map with entry should not be empty")
	}
}

func TestFloat64Float32_Generated_Clear(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", m.Len(), m.Len() == 0)
	}
}

func TestFloat64Float32_Generated_All(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}

func TestFloat64Float32_Generated_Keys(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	count := 0
	for range m.Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d, want 2", count)
	}
}

func TestFloat64Float32_Generated_Values(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	count := 0
	for range m.Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d, want 2", count)
	}
}

func TestFloat64Float32_Generated_Select(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)

	selected := m.Select(func(k float64, v float32) bool { return v > 1.0 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestFloat64Float32_Generated_Reject(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)

	rejected := m.Reject(func(k float64, v float32) bool { return v > 1.0 })
	if rejected.Len() != 1 {
		t.Errorf("Reject size = %d, want 1", rejected.Len())
	}
}

func TestFloat64Float32_Generated_Detect(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	_, val, found := m.Detect(func(k float64, v float32) bool { return k == 2.0 })
	if !found || val != 2.0 {
		t.Errorf("Detect = (%v, %v), want (2.0, true)", val, found)
	}
}

func TestFloat64Float32_Generated_AnySatisfy(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	if !m.AnySatisfy(func(k float64, v float32) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if m.AnySatisfy(func(k float64, v float32) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestFloat64Float32_Generated_AllSatisfy(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	if !m.AllSatisfy(func(k float64, v float32) bool { return v > 0.0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
	if m.AllSatisfy(func(k float64, v float32) bool { return v > 1.0 }) {
		t.Error("AllSatisfy should be false")
	}
}

func TestFloat64Float32_Generated_NoneSatisfy(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)

	if !m.NoneSatisfy(func(k float64, v float32) bool { return v == 99.0 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestFloat64Float32_Generated_Count(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)

	if c := m.Count(func(k float64, v float32) bool { return v > 1.0 }); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestFloat64Float32_Generated_Entry(t *testing.T) {
	m := NewFloat64Float32()
	v := m.Entry(1.0).OrInsert(1.0)
	if !(math.Float32bits(v) == math.Float32bits(1.0)) {
		t.Errorf("OrInsert = %v, want 1.0", v)
	}
	v = m.Entry(1.0).OrInsert(2.0)
	if !(math.Float32bits(v) == math.Float32bits(1.0)) {
		t.Errorf("OrInsert existing = %v, want 1.0 (original)", v)
	}
}

// TestFloat64Float32_Generated_AndModify_ResizeDetection forces a resize from
// within the AndModify callback and verifies the template's
// "do not mutate the map from within AndModify" guard fires, so that
// silent data loss through a dangling pointer into the pre-resize
// entries slice cannot happen.
func TestFloat64Float32_Generated_AndModify_ResizeDetection(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when AndModify callback mutates the map")
		}
	}()
	m.Entry(1.0).AndModify(func(_ *float32) {
		// Flood the map to force a resize mid-callback.
		for i := 0; i < 128; i++ {
			m.Put(float64(i)+2.0, 2.0)
		}
	})
}

func TestFloat64Float32_Generated_PutReturning(t *testing.T) {
	m := NewFloat64Float32()
	m2 := m.PutReturning(1.0, 1.0)
	if m2.Len() != 1 {
		t.Errorf("PutReturning: size = %d, want 1", m2.Len())
	}
}

func TestFloat64Float32_Generated_RemoveKeyReturning(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m2 := m.RemoveKeyReturning(1.0)
	if m2.ContainsKey(1.0) {
		t.Error("RemoveKeyReturning: should not contain removed key")
	}
}

func TestFloat64Float32_Generated_Equals(t *testing.T) {
	m1 := NewFloat64Float32()
	m1.Put(1.0, 1.0)
	m1.Put(2.0, 2.0)

	m2 := NewFloat64Float32()
	m2.Put(2.0, 2.0)
	m2.Put(1.0, 1.0)

	if !m1.Equals(m2) {
		t.Error("Equal maps should be equal")
	}

	m3 := NewFloat64Float32()
	m3.Put(1.0, 1.0)
	if m1.Equals(m3) {
		t.Error("Different maps should not be equal")
	}
}

func TestFloat64Float32_Generated_String(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	if m.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestFloat64Float32_Generated_Resize(t *testing.T) {
	m := NewFloat64Float32()
	for i := float64(0); i < 100; i++ {
		m.Put(i, float32(i))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d, want 100", m.Len())
	}
	for i := float64(0); i < 100; i++ {
		v, ok := m.Get(i)
		if !ok || v != float32(i) {
			t.Fatalf("Get(%v) = (%v, %v), want (%v, true)", i, v, ok, float32(i))
		}
	}
}

func TestFloat64Float32_NaNKey_Findable(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(float64(math.NaN()), float32(1))
	if !m.ContainsKey(float64(math.NaN())) {
		t.Errorf("expected NaN key to be findable")
	}
	if v, ok := m.Get(float64(math.NaN())); !ok || v != float32(1) {
		t.Errorf("expected get(NaN) to return (1, true), got (%v, %v)", v, ok)
	}
	if m.Len() != 1 {
		t.Errorf("expected size 1, got %d", m.Len())
	}
}

func TestFloat64Float32_NaNKey_Replaces(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(float64(math.NaN()), float32(1))
	m.Put(float64(math.NaN()), float32(2))
	m.Put(float64(math.NaN()), float32(3))
	if m.Len() != 1 {
		t.Errorf("expected size 1 after 3 NaN puts, got %d", m.Len())
	}
	if v, _ := m.Get(float64(math.NaN())); v != float32(3) {
		t.Errorf("expected last put to win (3), got %v", v)
	}
}

func TestFloat64Float32_NaNKey_Remove(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(float64(math.NaN()), float32(1))
	if _, removed := m.Remove(float64(math.NaN())); !removed {
		t.Errorf("expected NaN key to be removable")
	}
	if m.Len() != 0 {
		t.Errorf("expected size 0 after remove, got %d", m.Len())
	}
	if m.ContainsKey(float64(math.NaN())) {
		t.Errorf("expected ContainsKey(NaN) to be false after remove")
	}
}

func TestFloat64Float32_NegativeZeroDistinct(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(float64(0.0), float32(10))
	m.Put(float64(math.Copysign(0, -1)), float32(20))
	if m.Len() != 2 {
		t.Errorf("expected size 2 with +0 and -0 distinct, got %d", m.Len())
	}
	if v, _ := m.Get(float64(0.0)); v != float32(10) {
		t.Errorf("get(+0) = %v, want 10", v)
	}
	if v, _ := m.Get(float64(math.Copysign(0, -1))); v != float32(20) {
		t.Errorf("get(-0) = %v, want 20", v)
	}
}

func TestFloat64Float32_InfinityKeys(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(float64(math.Inf(1)), float32(11))
	m.Put(float64(math.Inf(-1)), float32(22))
	if m.Len() != 2 {
		t.Errorf("expected size 2, got %d", m.Len())
	}
	if v, _ := m.Get(float64(math.Inf(1))); v != float32(11) {
		t.Errorf("get(+Inf) = %v, want 11", v)
	}
	if v, _ := m.Get(float64(math.Inf(-1))); v != float32(22) {
		t.Errorf("get(-Inf) = %v, want 22", v)
	}
}

func TestFloat64Float32_NaNValue_ContainsValue(t *testing.T) {
	m := NewFloat64Float32()
	nan := float32(math.NaN())
	m.Put(1.0, nan)
	if !m.ContainsValue(nan) {
		t.Error("ContainsValue(NaN) should be true (bit-level comparison)")
	}
}

func TestFloat64Float32_NaNValue_GetReturnsNaN(t *testing.T) {
	m := NewFloat64Float32()
	nan := float32(math.NaN())
	m.Put(1.0, nan)
	v, ok := m.Get(1.0)
	if !ok {
		t.Fatal("expected Get to find the key")
	}
	if !math.IsNaN(float64(v)) {
		t.Errorf("Get returned %v, want NaN", v)
	}
}
