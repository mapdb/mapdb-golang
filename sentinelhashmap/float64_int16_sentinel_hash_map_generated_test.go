package sentinelhashmap

import "testing"

func TestFloat64Int16_Generated_PutGet(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	if m.Len() != 3 {
		t.Errorf("Size = %d", m.Len())
	}
	if v, ok := m.Get(1.0); !ok || v != 1 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if _, ok := m.Get(99.0); ok {
		t.Error("Get missing should be false")
	}
}
func TestFloat64Int16_Generated_PutOverwrite(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	old, existed := m.Put(1.0, 2)
	if !existed || old != 1 {
		t.Errorf("Overwrite = (%v,%v)", old, existed)
	}
}
func TestFloat64Int16_Generated_Remove(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	old, ok := m.Remove(1.0)
	if !ok || old != 1 {
		t.Errorf("Remove = (%v,%v)", old, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Size = %d", m.Len())
	}
}
func TestFloat64Int16_Generated_ContainsKey(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	if !m.ContainsKey(1.0) {
		t.Error("Should contain")
	}
	if m.ContainsKey(99.0) {
		t.Error("Should not contain")
	}
}
func TestFloat64Int16_Generated_ContainsValue(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	if !m.ContainsValue(1) {
		t.Error("Should contain value")
	}
}
func TestFloat64Int16_Generated_Clear(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	m.Clear()
	if m.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestFloat64Int16_Generated_All(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestFloat64Int16_Generated_Select(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	m.Put(3.0, 3)
	sel := m.Select(func(k float64, v int16) bool { return v > 1 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d", sel.Len())
	}
}
func TestFloat64Int16_Generated_AnySatisfy(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	m.Put(2.0, 2)
	if !m.AnySatisfy(func(k float64, v int16) bool { return v == 2 }) {
		t.Error("Should be true")
	}
}
func TestFloat64Int16_Generated_String(t *testing.T) {
	m := NewFloat64Int16()
	m.Put(1.0, 1)
	if m.String() == "" {
		t.Error("empty")
	}
}
func TestFloat64Int16_Generated_Resize(t *testing.T) {
	m := NewFloat64Int16()
	for i := float64(0); i < 100; i++ {
		m.Put(i, int16(i))
	}
	if m.Len() != 100 {
		t.Errorf("Size = %d", m.Len())
	}
}
