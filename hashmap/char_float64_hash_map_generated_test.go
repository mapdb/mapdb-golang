package hashmap

import (
	"math"
	"testing"
)

func TestCharFloat64_Generated_PutGet(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	if v, ok := m.Get(1); !ok || v != 1.0 {
		t.Errorf("Get(1) = (%v, %v), want (1.0, true)", v, ok)
	}
	if v, ok := m.Get(99); ok || v != 0.0 {
		t.Errorf("Get(99) = (%v, %v), want (0.0, false)", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("Size() = %d, want 3", m.Len())
	}
}

func TestCharFloat64_Generated_PutOverwrite(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	old, existed := m.Put(1, 2.0)
	if !existed || old != 1.0 {
		t.Errorf("Put overwrite = (%v, %v), want (1.0, true)", old, existed)
	}
	if v, _ := m.Get(1); v != 2.0 {
		t.Errorf("Get after overwrite = %v, want 2.0", v)
	}
}

func TestCharFloat64_Generated_Remove(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	old, ok := m.Remove(1)
	if !ok || old != 1.0 {
		t.Errorf("Remove(1) = (%v, %v), want (1.0, true)", old, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Len())
	}
	if m.ContainsKey(1) {
		t.Error("ContainsKey(1) should be false after remove")
	}

	_, ok = m.Remove(99)
	if ok {
		t.Error("Remove missing key should return false")
	}
}

func TestCharFloat64_Generated_ContainsKey(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	if !m.ContainsKey(1) {
		t.Error("ContainsKey(1) should be true")
	}
	if m.ContainsKey(99) {
		t.Error("ContainsKey(99) should be false")
	}
}

func TestCharFloat64_Generated_ContainsValue(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	if !m.ContainsValue(1.0) {
		t.Error("ContainsValue(1.0) should be true")
	}
	if m.ContainsValue(99.0) {
		t.Error("ContainsValue for missing should be false")
	}
}

func TestCharFloat64_Generated_GetOrDefault(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	if v := m.GetOrDefault(1, 3.0); v != 1.0 {
		t.Errorf("GetOrDefault existing = %v, want 1.0", v)
	}
	if v := m.GetOrDefault(99, 3.0); v != 3.0 {
		t.Errorf("GetOrDefault missing = %v, want 3.0", v)
	}
}

func TestCharFloat64_Generated_IsEmpty(t *testing.T) {
	m := NewCharFloat64()
	if m.Len() != 0 {
		t.Error("New map should be empty")
	}
	m.Put(1, 1.0)
	if m.Len() == 0 {
		t.Error("Map with entry should not be empty")
	}
}

func TestCharFloat64_Generated_Clear(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", m.Len(), m.Len() == 0)
	}
}

func TestCharFloat64_Generated_All(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}

func TestCharFloat64_Generated_Keys(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	count := 0
	for range m.Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d, want 2", count)
	}
}

func TestCharFloat64_Generated_Values(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	count := 0
	for range m.Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d, want 2", count)
	}
}

func TestCharFloat64_Generated_Select(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	selected := m.Select(func(k uint16, v float64) bool { return v > 1.0 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestCharFloat64_Generated_Reject(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	rejected := m.Reject(func(k uint16, v float64) bool { return v > 1.0 })
	if rejected.Len() != 1 {
		t.Errorf("Reject size = %d, want 1", rejected.Len())
	}
}

func TestCharFloat64_Generated_Detect(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	_, val, found := m.Detect(func(k uint16, v float64) bool { return k == 2 })
	if !found || val != 2.0 {
		t.Errorf("Detect = (%v, %v), want (2.0, true)", val, found)
	}
}

func TestCharFloat64_Generated_AnySatisfy(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if !m.AnySatisfy(func(k uint16, v float64) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if m.AnySatisfy(func(k uint16, v float64) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestCharFloat64_Generated_AllSatisfy(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if !m.AllSatisfy(func(k uint16, v float64) bool { return v > 0.0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
	if m.AllSatisfy(func(k uint16, v float64) bool { return v > 1.0 }) {
		t.Error("AllSatisfy should be false")
	}
}

func TestCharFloat64_Generated_NoneSatisfy(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if !m.NoneSatisfy(func(k uint16, v float64) bool { return v == 99.0 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestCharFloat64_Generated_Count(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	if c := m.Count(func(k uint16, v float64) bool { return v > 1.0 }); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestCharFloat64_Generated_Entry(t *testing.T) {
	m := NewCharFloat64()
	v := m.Entry(1).OrInsert(1.0)
	if !(math.Float64bits(v) == math.Float64bits(1.0)) {
		t.Errorf("OrInsert = %v, want 1.0", v)
	}
	v = m.Entry(1).OrInsert(2.0)
	if !(math.Float64bits(v) == math.Float64bits(1.0)) {
		t.Errorf("OrInsert existing = %v, want 1.0 (original)", v)
	}
}

// TestCharFloat64_Generated_AndModify_ResizeDetection forces a resize from
// within the AndModify callback and verifies the template's
// "do not mutate the map from within AndModify" guard fires, so that
// silent data loss through a dangling pointer into the pre-resize
// entries slice cannot happen.
func TestCharFloat64_Generated_AndModify_ResizeDetection(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when AndModify callback mutates the map")
		}
	}()
	m.Entry(1).AndModify(func(_ *float64) {
		// Flood the map to force a resize mid-callback.
		for i := 0; i < 128; i++ {
			m.Put(uint16(i)+2, 2.0)
		}
	})
}

func TestCharFloat64_Generated_PutReturning(t *testing.T) {
	m := NewCharFloat64()
	m2 := m.PutReturning(1, 1.0)
	if m2.Len() != 1 {
		t.Errorf("PutReturning: size = %d, want 1", m2.Len())
	}
}

func TestCharFloat64_Generated_RemoveKeyReturning(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m2 := m.RemoveKeyReturning(1)
	if m2.ContainsKey(1) {
		t.Error("RemoveKeyReturning: should not contain removed key")
	}
}

func TestCharFloat64_Generated_Equals(t *testing.T) {
	m1 := NewCharFloat64()
	m1.Put(1, 1.0)
	m1.Put(2, 2.0)

	m2 := NewCharFloat64()
	m2.Put(2, 2.0)
	m2.Put(1, 1.0)

	if !m1.Equals(m2) {
		t.Error("Equal maps should be equal")
	}

	m3 := NewCharFloat64()
	m3.Put(1, 1.0)
	if m1.Equals(m3) {
		t.Error("Different maps should not be equal")
	}
}

func TestCharFloat64_Generated_String(t *testing.T) {
	m := NewCharFloat64()
	m.Put(1, 1.0)
	if m.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestCharFloat64_Generated_Resize(t *testing.T) {
	m := NewCharFloat64()
	for i := uint16(0); i < 100; i++ {
		m.Put(i, float64(i))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d, want 100", m.Len())
	}
	for i := uint16(0); i < 100; i++ {
		v, ok := m.Get(i)
		if !ok || v != float64(i) {
			t.Fatalf("Get(%v) = (%v, %v), want (%v, true)", i, v, ok, float64(i))
		}
	}
}

func TestCharFloat64_NaNValue_ContainsValue(t *testing.T) {
	m := NewCharFloat64()
	nan := float64(math.NaN())
	m.Put(1, nan)
	if !m.ContainsValue(nan) {
		t.Error("ContainsValue(NaN) should be true (bit-level comparison)")
	}
}

func TestCharFloat64_NaNValue_GetReturnsNaN(t *testing.T) {
	m := NewCharFloat64()
	nan := float64(math.NaN())
	m.Put(1, nan)
	v, ok := m.Get(1)
	if !ok {
		t.Fatal("expected Get to find the key")
	}
	if !math.IsNaN(float64(v)) {
		t.Errorf("Get returned %v, want NaN", v)
	}
}
