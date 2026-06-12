package hashmap

import (
	"testing"
)

func TestInt8Int8_Generated_PutGet(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)

	if v, ok := m.Get(1); !ok || v != 1 {
		t.Errorf("Get(1) = (%v, %v), want (1, true)", v, ok)
	}
	if v, ok := m.Get(99); ok || v != 0 {
		t.Errorf("Get(99) = (%v, %v), want (0, false)", v, ok)
	}
	if m.Len() != 3 {
		t.Errorf("Size() = %d, want 3", m.Len())
	}
}

func TestInt8Int8_Generated_PutOverwrite(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	old, existed := m.Put(1, 2)
	if !existed || old != 1 {
		t.Errorf("Put overwrite = (%v, %v), want (1, true)", old, existed)
	}
	if v, _ := m.Get(1); v != 2 {
		t.Errorf("Get after overwrite = %v, want 2", v)
	}
}

func TestInt8Int8_Generated_Remove(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	old, ok := m.Remove(1)
	if !ok || old != 1 {
		t.Errorf("Remove(1) = (%v, %v), want (1, true)", old, ok)
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

func TestInt8Int8_Generated_ContainsKey(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	if !m.ContainsKey(1) {
		t.Error("ContainsKey(1) should be true")
	}
	if m.ContainsKey(99) {
		t.Error("ContainsKey(99) should be false")
	}
}

func TestInt8Int8_Generated_ContainsValue(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	if !m.ContainsValue(1) {
		t.Error("ContainsValue(1) should be true")
	}
	if m.ContainsValue(99) {
		t.Error("ContainsValue for missing should be false")
	}
}

func TestInt8Int8_Generated_GetOrDefault(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	if v := m.GetOrDefault(1, 3); v != 1 {
		t.Errorf("GetOrDefault existing = %v, want 1", v)
	}
	if v := m.GetOrDefault(99, 3); v != 3 {
		t.Errorf("GetOrDefault missing = %v, want 3", v)
	}
}

func TestInt8Int8_Generated_IsEmpty(t *testing.T) {
	m := NewInt8Int8()
	if m.Len() != 0 {
		t.Error("New map should be empty")
	}
	m.Put(1, 1)
	if m.Len() == 0 {
		t.Error("Map with entry should not be empty")
	}
}

func TestInt8Int8_Generated_Clear(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", m.Len(), m.Len() == 0)
	}
}

func TestInt8Int8_Generated_All(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}

func TestInt8Int8_Generated_Keys(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	count := 0
	for range m.Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d, want 2", count)
	}
}

func TestInt8Int8_Generated_Values(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	count := 0
	for range m.Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d, want 2", count)
	}
}

func TestInt8Int8_Generated_Select(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)

	selected := m.Select(func(k int8, v int8) bool { return v > 1 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestInt8Int8_Generated_Reject(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)

	rejected := m.Reject(func(k int8, v int8) bool { return v > 1 })
	if rejected.Len() != 1 {
		t.Errorf("Reject size = %d, want 1", rejected.Len())
	}
}

func TestInt8Int8_Generated_Detect(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	_, val, found := m.Detect(func(k int8, v int8) bool { return k == 2 })
	if !found || val != 2 {
		t.Errorf("Detect = (%v, %v), want (2, true)", val, found)
	}
}

func TestInt8Int8_Generated_AnySatisfy(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	if !m.AnySatisfy(func(k int8, v int8) bool { return v == 2 }) {
		t.Error("AnySatisfy should be true")
	}
	if m.AnySatisfy(func(k int8, v int8) bool { return v == 99 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestInt8Int8_Generated_AllSatisfy(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	if !m.AllSatisfy(func(k int8, v int8) bool { return v > 0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
	if m.AllSatisfy(func(k int8, v int8) bool { return v > 1 }) {
		t.Error("AllSatisfy should be false")
	}
}

func TestInt8Int8_Generated_NoneSatisfy(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)

	if !m.NoneSatisfy(func(k int8, v int8) bool { return v == 99 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestInt8Int8_Generated_Count(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)

	if c := m.Count(func(k int8, v int8) bool { return v > 1 }); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestInt8Int8_Generated_Entry(t *testing.T) {
	m := NewInt8Int8()
	v := m.Entry(1).OrInsert(1)
	if !(v == 1) {
		t.Errorf("OrInsert = %v, want 1", v)
	}
	v = m.Entry(1).OrInsert(2)
	if !(v == 1) {
		t.Errorf("OrInsert existing = %v, want 1 (original)", v)
	}
}

// TestInt8Int8_Generated_AndModify_ResizeDetection forces a resize from
// within the AndModify callback and verifies the template's
// "do not mutate the map from within AndModify" guard fires, so that
// silent data loss through a dangling pointer into the pre-resize
// entries slice cannot happen.
func TestInt8Int8_Generated_AndModify_ResizeDetection(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when AndModify callback mutates the map")
		}
	}()
	m.Entry(1).AndModify(func(_ *int8) {
		// Flood the map to force a resize mid-callback.
		for i := 0; i < 128; i++ {
			m.Put(int8(i)+2, 2)
		}
	})
}

func TestInt8Int8_Generated_PutReturning(t *testing.T) {
	m := NewInt8Int8()
	m2 := m.PutReturning(1, 1)
	if m2.Len() != 1 {
		t.Errorf("PutReturning: size = %d, want 1", m2.Len())
	}
}

func TestInt8Int8_Generated_RemoveKeyReturning(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	m.Put(2, 2)
	m2 := m.RemoveKeyReturning(1)
	if m2.ContainsKey(1) {
		t.Error("RemoveKeyReturning: should not contain removed key")
	}
}

func TestInt8Int8_Generated_Equals(t *testing.T) {
	m1 := NewInt8Int8()
	m1.Put(1, 1)
	m1.Put(2, 2)

	m2 := NewInt8Int8()
	m2.Put(2, 2)
	m2.Put(1, 1)

	if !m1.Equals(m2) {
		t.Error("Equal maps should be equal")
	}

	m3 := NewInt8Int8()
	m3.Put(1, 1)
	if m1.Equals(m3) {
		t.Error("Different maps should not be equal")
	}
}

func TestInt8Int8_Generated_String(t *testing.T) {
	m := NewInt8Int8()
	m.Put(1, 1)
	if m.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestInt8Int8_Generated_Resize(t *testing.T) {
	m := NewInt8Int8()
	for i := int8(0); i < 100; i++ {
		m.Put(i, int8(i))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d, want 100", m.Len())
	}
	for i := int8(0); i < 100; i++ {
		v, ok := m.Get(i)
		if !ok || v != int8(i) {
			t.Fatalf("Get(%v) = (%v, %v), want (%v, true)", i, v, ok, int8(i))
		}
	}
}
