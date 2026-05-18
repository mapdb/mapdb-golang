
package sentinelhashmap

import "testing"

func TestCharCharSentinelHashMap_Generated_PutGet(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)
	if m.Size() != 3 {
		t.Errorf("Size = %d", m.Size())
	}
	if v, ok := m.Get(1); !ok || v != 1 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if _, ok := m.Get(99); ok {
		t.Error("Get missing should be false")
	}
}
func TestCharCharSentinelHashMap_Generated_PutOverwrite(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	old, existed := m.Put(1, 2)
	if !existed || old != 1 {
		t.Errorf("Overwrite = (%v,%v)", old, existed)
	}
}
func TestCharCharSentinelHashMap_Generated_Remove(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	old, ok := m.Remove(1)
	if !ok || old != 1 {
		t.Errorf("Remove = (%v,%v)", old, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size = %d", m.Size())
	}
}
func TestCharCharSentinelHashMap_Generated_ContainsKey(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	if !m.ContainsKey(1) {
		t.Error("Should contain")
	}
	if m.ContainsKey(99) {
		t.Error("Should not contain")
	}
}
func TestCharCharSentinelHashMap_Generated_ContainsValue(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	if !m.ContainsValue(1) {
		t.Error("Should contain value")
	}
}
func TestCharCharSentinelHashMap_Generated_Clear(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	m.Clear()
	if !m.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestCharCharSentinelHashMap_Generated_All(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d", count)
	}
}
func TestCharCharSentinelHashMap_Generated_Select(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)
	sel := m.Select(func(k uint16, v uint16) bool { return v > 1 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d", sel.Size())
	}
}
func TestCharCharSentinelHashMap_Generated_AnySatisfy(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	if !m.AnySatisfy(func(k uint16, v uint16) bool { return v == 2 }) {
		t.Error("Should be true")
	}
}
func TestCharCharSentinelHashMap_Generated_String(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	m.Put(1, 1)
	if m.String() == "" {
		t.Error("empty")
	}
}
func TestCharCharSentinelHashMap_Generated_Resize(t *testing.T) {
	m := NewCharCharSentinelHashMap()
	for i := uint16(0); i < 100; i++ {
		m.Put(i, uint16(i))
	}
	if m.Size() != 100 {
		t.Errorf("Size = %d", m.Size())
	}
}
