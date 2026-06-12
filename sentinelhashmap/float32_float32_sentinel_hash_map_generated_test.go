package sentinelhashmap

import "testing"

func TestFloat32Float32_Generated_PutGet(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)
	if m.Len() != 3 {
		t.Errorf("Size = %d", m.Len())
	}
	if v, ok := m.Get(1.0); !ok || v != 1.0 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if _, ok := m.Get(99.0); ok {
		t.Error("Get missing should be false")
	}
}
func TestFloat32Float32_Generated_PutOverwrite(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	old, existed := m.Put(1.0, 2.0)
	if !existed || old != 1.0 {
		t.Errorf("Overwrite = (%v,%v)", old, existed)
	}
}
func TestFloat32Float32_Generated_Remove(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	old, ok := m.Remove(1.0)
	if !ok || old != 1.0 {
		t.Errorf("Remove = (%v,%v)", old, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Size = %d", m.Len())
	}
}
func TestFloat32Float32_Generated_ContainsKey(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	if !m.ContainsKey(1.0) {
		t.Error("Should contain")
	}
	if m.ContainsKey(99.0) {
		t.Error("Should not contain")
	}
}
func TestFloat32Float32_Generated_ContainsValue(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	if !m.ContainsValue(1.0) {
		t.Error("Should contain value")
	}
}
func TestFloat32Float32_Generated_Clear(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	m.Clear()
	if m.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestFloat32Float32_Generated_All(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestFloat32Float32_Generated_Select(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	m.Put(3.0, 3.0)
	sel := m.Select(func(k float32, v float32) bool { return v > 1.0 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d", sel.Len())
	}
}
func TestFloat32Float32_Generated_AnySatisfy(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	m.Put(2.0, 2.0)
	if !m.AnySatisfy(func(k float32, v float32) bool { return v == 2.0 }) {
		t.Error("Should be true")
	}
}
func TestFloat32Float32_Generated_String(t *testing.T) {
	m := NewFloat32Float32()
	m.Put(1.0, 1.0)
	if m.String() == "" {
		t.Error("empty")
	}
}
func TestFloat32Float32_Generated_Resize(t *testing.T) {
	m := NewFloat32Float32()
	for i := float32(0); i < 100; i++ {
		m.Put(i, float32(i))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d", m.Len())
	}
}
