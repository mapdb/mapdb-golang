package hashmap

import "testing"

func TestImmutableFloat32Int16_Generated_GetAndSize(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	im := m.ToImmutable()
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
	}
	if v, ok := im.Get(1.0); !ok || v != 1 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if v, ok := im.Get(99.0); ok || v != 0 {
		t.Errorf("Get missing = (%v,%v)", v, ok)
	}
}
func TestImmutableFloat32Int16_Generated_ContainsKey(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	im := m.ToImmutable()
	if !im.ContainsKey(1.0) {
		t.Error("Should contain key")
	}
	if im.ContainsKey(99.0) {
		t.Error("Should not contain missing")
	}
}
func TestImmutableFloat32Int16_Generated_ContainsValue(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	im := m.ToImmutable()
	if !im.ContainsValue(1) {
		t.Error("Should contain value")
	}
}
func TestImmutableFloat32Int16_Generated_GetOrDefault(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	im := m.ToImmutable()
	if v := im.GetOrDefault(1.0, 3); v != 1 {
		t.Errorf("got %v", v)
	}
	if v := im.GetOrDefault(99.0, 3); v != 3 {
		t.Errorf("got %v", v)
	}
}
func TestImmutableFloat32Int16_Generated_IsEmpty(t *testing.T) {
	im := NewFloat32Int16().ToImmutable()
	if im.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestImmutableFloat32Int16_Generated_All(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	im := m.ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}
func TestImmutableFloat32Int16_Generated_Keys(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	count := 0
	for range m.ToImmutable().Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d", count)
	}
}
func TestImmutableFloat32Int16_Generated_Values(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	count := 0
	for range m.ToImmutable().Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d", count)
	}
}
func TestImmutableFloat32Int16_Generated_Select(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	sel := m.ToImmutable().Select(func(k float32, v int16) bool { return v > 1 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Len())
	}
}
func TestImmutableFloat32Int16_Generated_Reject(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	rej := m.ToImmutable().Reject(func(k float32, v int16) bool { return v > 1 })
	if rej.Len() != 1 {
		t.Errorf("Reject size = %d, want 1", rej.Len())
	}
}
func TestImmutableFloat32Int16_Generated_Equals(t *testing.T) {
	m1 := NewFloat32Int16()
	m1.Put(1.0, 1)
	m1.Put(2.0, 2)
	m2 := NewFloat32Int16()
	m2.Put(2.0, 2)
	m2.Put(1.0, 1)
	if !m1.ToImmutable().Equals(m2.ToImmutable()) {
		t.Error("Should be equal")
	}
}
func TestImmutableFloat32Int16_Generated_ToMutable(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	im := m.ToImmutable()
	m2 := im.ToMutable()
	m2.Put(2.0, 2)
	if im.Len() != 1 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableFloat32Int16_Generated_String(t *testing.T) {
	m := NewFloat32Int16()
	m.Put(1.0, 1)
	if m.ToImmutable().String() == "" {
		t.Error("String empty")
	}
}
