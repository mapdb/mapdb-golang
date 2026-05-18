
package sentinelhashmap

import "testing"

func TestInt8Float64SentinelHashMap_Generated_PutGet(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)
	if m.Size() != 3 {
		t.Errorf("Size = %d", m.Size())
	}
	if v, ok := m.Get(1); !ok || v != 1.0 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if _, ok := m.Get(99); ok {
		t.Error("Get missing should be false")
	}
}
func TestInt8Float64SentinelHashMap_Generated_PutOverwrite(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	old, existed := m.Put(1, 2.0)
	if !existed || old != 1.0 {
		t.Errorf("Overwrite = (%v,%v)", old, existed)
	}
}
func TestInt8Float64SentinelHashMap_Generated_Remove(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	old, ok := m.Remove(1)
	if !ok || old != 1.0 {
		t.Errorf("Remove = (%v,%v)", old, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size = %d", m.Size())
	}
}
func TestInt8Float64SentinelHashMap_Generated_ContainsKey(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	if !m.ContainsKey(1) {
		t.Error("Should contain")
	}
	if m.ContainsKey(99) {
		t.Error("Should not contain")
	}
}
func TestInt8Float64SentinelHashMap_Generated_ContainsValue(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	if !m.ContainsValue(1.0) {
		t.Error("Should contain value")
	}
}
func TestInt8Float64SentinelHashMap_Generated_Clear(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	m.Clear()
	if !m.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestInt8Float64SentinelHashMap_Generated_All(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestInt8Float64SentinelHashMap_Generated_Select(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	m.Put(3, 3.0)
	sel := m.Select(func(k int8, v float64) bool { return v > 1.0 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d", sel.Size())
	}
}
func TestInt8Float64SentinelHashMap_Generated_AnySatisfy(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	m.Put(2, 2.0)
	if !m.AnySatisfy(func(k int8, v float64) bool { return v == 2.0 }) {
		t.Error("Should be true")
	}
}
func TestInt8Float64SentinelHashMap_Generated_String(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	m.Put(1, 1.0)
	if m.String() == "" {
		t.Error("empty")
	}
}
func TestInt8Float64SentinelHashMap_Generated_Resize(t *testing.T) {
	m := NewInt8Float64SentinelHashMap()
	for i := int8(0); i < 100; i++ {
		m.Put(i, float64(i))
	}
	if m.Size() != 100 {
		t.Errorf("Size = %d", m.Size())
	}
}
