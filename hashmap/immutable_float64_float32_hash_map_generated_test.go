package hashmap

import "testing"

func TestImmutableFloat64Float32_Generated_GetAndSize(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)
	im := m.ToImmutable()
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
	}
	if v, ok := im.Get(1.0); !ok || v != 1.0 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if v, ok := im.Get(99.0); ok || v != 0.0 {
		t.Errorf("Get missing = (%v,%v)", v, ok)
	}
}
func TestImmutableFloat64Float32_Generated_ContainsKey(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	im := m.ToImmutable()
	if !im.ContainsKey(1.0) {
		t.Error("Should contain key")
	}
	if im.ContainsKey(99.0) {
		t.Error("Should not contain missing")
	}
}
func TestImmutableFloat64Float32_Generated_ContainsValue(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	im := m.ToImmutable()
	if !im.ContainsValue(1.0) {
		t.Error("Should contain value")
	}
}
func TestImmutableFloat64Float32_Generated_GetOrDefault(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	im := m.ToImmutable()
	if v := im.GetOrDefault(1.0, 3.0); v != 1.0 {
		t.Errorf("got %v", v)
	}
	if v := im.GetOrDefault(99.0, 3.0); v != 3.0 {
		t.Errorf("got %v", v)
	}
}
func TestImmutableFloat64Float32_Generated_IsEmpty(t *testing.T) {
	im := NewFloat64Float32().ToImmutable()
	if im.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestImmutableFloat64Float32_Generated_All(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	im := m.ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}
func TestImmutableFloat64Float32_Generated_Keys(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	count := 0
	for range m.ToImmutable().Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d", count)
	}
}
func TestImmutableFloat64Float32_Generated_Values(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	count := 0
	for range m.ToImmutable().Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d", count)
	}
}
func TestImmutableFloat64Float32_Generated_Select(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)
	sel := m.ToImmutable().Select(func(k float64, v float32) bool { return v > 1.0 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Len())
	}
}
func TestImmutableFloat64Float32_Generated_Reject(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)
	rej := m.ToImmutable().Reject(func(k float64, v float32) bool { return v > 1.0 })
	if rej.Len() != 1 {
		t.Errorf("Reject size = %d, want 1", rej.Len())
	}
}
func TestImmutableFloat64Float32_Generated_Equals(t *testing.T) {
	m1 := NewFloat64Float32()
	m1.Put(1.0, 1.0)
	m1.Put(2.0, 2.0)
	m2 := NewFloat64Float32()
	m2.Put(2.0, 2.0)
	m2.Put(1.0, 1.0)
	if !m1.ToImmutable().Equals(m2.ToImmutable()) {
		t.Error("Should be equal")
	}
}
func TestImmutableFloat64Float32_Generated_ToMutable(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	im := m.ToImmutable()
	m2 := im.ToMutable()
	m2.Put(2.0, 2.0)
	if im.Len() != 1 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableFloat64Float32_Generated_String(t *testing.T) {
	m := NewFloat64Float32()
	m.Put(1.0, 1.0)
	if m.ToImmutable().String() == "" {
		t.Error("String empty")
	}
}
