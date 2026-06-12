package hashmap

import (
	"math"
	"testing"
)

func TestInt16Float32_Generated_PutGet(t *testing.T) {
	m := NewInt16Float32()
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

func TestInt16Float32_Generated_PutOverwrite(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	old, existed := m.Put(1, 2.0)
	if !existed || old != 1.0 {
		t.Errorf("Put overwrite = (%v, %v), want (1.0, true)", old, existed)
	}
	if v, _ := m.Get(1); v != 2.0 {
		t.Errorf("Get after overwrite = %v, want 2.0", v)
	}
}

func TestInt16Float32_Generated_Remove(t *testing.T) {
	m := NewInt16Float32()
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

func TestInt16Float32_Generated_ContainsKey(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	if !m.ContainsKey(1) {
		t.Error("ContainsKey(1) should be true")
	}
	if m.ContainsKey(99) {
		t.Error("ContainsKey(99) should be false")
	}
}

func TestInt16Float32_Generated_ContainsValue(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	if !m.ContainsValue(1.0) {
		t.Error("ContainsValue(1.0) should be true")
	}
	if m.ContainsValue(99.0) {
		t.Error("ContainsValue for missing should be false")
	}
}

func TestInt16Float32_Generated_GetOrDefault(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	if v := m.GetOrDefault(1, 3.0); v != 1.0 {
		t.Errorf("GetOrDefault existing = %v, want 1.0", v)
	}
	if v := m.GetOrDefault(99, 3.0); v != 3.0 {
		t.Errorf("GetOrDefault missing = %v, want 3.0", v)
	}
}

func TestInt16Float32_Generated_IsEmpty(t *testing.T) {
	m := NewInt16Float32()
	if m.Len() != 0 {
		t.Error("New map should be empty")
	}
	m.Put(1, 1.0)
	if m.Len() == 0 {
		t.Error("Map with entry should not be empty")
	}
}

func TestInt16Float32_Generated_Clear(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", m.Len(), m.Len() == 0)
	}
}

func TestInt16Float32_Generated_All(t *testing.T) {
	m := NewInt16Float32()
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

func TestInt16Float32_Generated_Keys(t *testing.T) {
	m := NewInt16Float32()
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

func TestInt16Float32_Generated_Values(t *testing.T) {
	m := NewInt16Float32()
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

func TestInt16Float32_Generated_Select(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	selected := m.Select(func(k int16, v float32) bool { return v > 1.0 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestInt16Float32_Generated_Reject(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	rejected := m.Reject(func(k int16, v float32) bool { return v > 1.0 })
	if rejected.Len() != 1 {
		t.Errorf("Reject size = %d, want 1", rejected.Len())
	}
}

func TestInt16Float32_Generated_Detect(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	_, val, found := m.Detect(func(k int16, v float32) bool { return k == 2 })
	if !found || val != 2.0 {
		t.Errorf("Detect = (%v, %v), want (2.0, true)", val, found)
	}
}

func TestInt16Float32_Generated_AnySatisfy(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if !m.AnySatisfy(func(k int16, v float32) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if m.AnySatisfy(func(k int16, v float32) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestInt16Float32_Generated_AllSatisfy(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if !m.AllSatisfy(func(k int16, v float32) bool { return v > 0.0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
	if m.AllSatisfy(func(k int16, v float32) bool { return v > 1.0 }) {
		t.Error("AllSatisfy should be false")
	}
}

func TestInt16Float32_Generated_NoneSatisfy(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)

	if !m.NoneSatisfy(func(k int16, v float32) bool { return v == 99.0 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestInt16Float32_Generated_Count(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)

	if c := m.Count(func(k int16, v float32) bool { return v > 1.0 }); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestInt16Float32_Generated_Entry(t *testing.T) {
	m := NewInt16Float32()
	v := m.Entry(1).OrInsert(1.0)
	if !(math.Float32bits(v) == math.Float32bits(1.0)) {
		t.Errorf("OrInsert = %v, want 1.0", v)
	}
	v = m.Entry(1).OrInsert(2.0)
	if !(math.Float32bits(v) == math.Float32bits(1.0)) {
		t.Errorf("OrInsert existing = %v, want 1.0 (original)", v)
	}
}

// TestInt16Float32_Generated_AndModify_ResizeDetection forces a resize from
// within the AndModify callback and verifies the template's
// "do not mutate the map from within AndModify" guard fires, so that
// silent data loss through a dangling pointer into the pre-resize
// entries slice cannot happen.
func TestInt16Float32_Generated_AndModify_ResizeDetection(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when AndModify callback mutates the map")
		}
	}()
	m.Entry(1).AndModify(func(_ *float32) {
		// Flood the map to force a resize mid-callback.
		for i := 0; i < 128; i++ {
			m.Put(int16(i)+2, 2.0)
		}
	})
}

func TestInt16Float32_Generated_PutReturning(t *testing.T) {
	m := NewInt16Float32()
	m2 := m.PutReturning(1, 1.0)
	if m2.Len() != 1 {
		t.Errorf("PutReturning: size = %d, want 1", m2.Len())
	}
}

func TestInt16Float32_Generated_RemoveKeyReturning(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m2 := m.RemoveKeyReturning(1)
	if m2.ContainsKey(1) {
		t.Error("RemoveKeyReturning: should not contain removed key")
	}
}

func TestInt16Float32_Generated_Equals(t *testing.T) {
	m1 := NewInt16Float32()
	m1.Put(1, 1.0)
	m1.Put(2, 2.0)

	m2 := NewInt16Float32()
	m2.Put(2, 2.0)
	m2.Put(1, 1.0)

	if !m1.Equals(m2) {
		t.Error("Equal maps should be equal")
	}

	m3 := NewInt16Float32()
	m3.Put(1, 1.0)
	if m1.Equals(m3) {
		t.Error("Different maps should not be equal")
	}
}

func TestInt16Float32_Generated_String(t *testing.T) {
	m := NewInt16Float32()
	m.Put(1, 1.0)
	if m.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestInt16Float32_Generated_Resize(t *testing.T) {
	m := NewInt16Float32()
	for i := int16(0); i < 100; i++ {
		m.Put(i, float32(i))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d, want 100", m.Len())
	}
	for i := int16(0); i < 100; i++ {
		v, ok := m.Get(i)
		if !ok || v != float32(i) {
			t.Fatalf("Get(%v) = (%v, %v), want (%v, true)", i, v, ok, float32(i))
		}
	}
}

func TestInt16Float32_NaNValue_ContainsValue(t *testing.T) {
	m := NewInt16Float32()
	nan := float32(math.NaN())
	m.Put(1, nan)
	if !m.ContainsValue(nan) {
		t.Error("ContainsValue(NaN) should be true (bit-level comparison)")
	}
}

func TestInt16Float32_NaNValue_GetReturnsNaN(t *testing.T) {
	m := NewInt16Float32()
	nan := float32(math.NaN())
	m.Put(1, nan)
	v, ok := m.Get(1)
	if !ok {
		t.Fatal("expected Get to find the key")
	}
	if !math.IsNaN(float64(v)) {
		t.Errorf("Get returned %v, want NaN", v)
	}
}
