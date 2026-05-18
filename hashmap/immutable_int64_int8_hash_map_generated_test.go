
package hashmap

import "testing"

func TestImmutableInt64Int8HashMap_Generated_GetAndSize(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)
	im := m.ToImmutable()
	if im.Size() != 3 {
		t.Errorf("Size = %d, want 3", im.Size())
	}
	if v, ok := im.Get(1); !ok || v != 1 {
		t.Errorf("Get = (%v,%v)", v, ok)
	}
	if v, ok := im.Get(99); ok || v != 0 {
		t.Errorf("Get missing = (%v,%v)", v, ok)
	}
}
func TestImmutableInt64Int8HashMap_Generated_ContainsKey(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	im := m.ToImmutable()
	if !im.ContainsKey(1) {
		t.Error("Should contain key")
	}
	if im.ContainsKey(99) {
		t.Error("Should not contain missing")
	}
}
func TestImmutableInt64Int8HashMap_Generated_ContainsValue(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	im := m.ToImmutable()
	if !im.ContainsValue(1) {
		t.Error("Should contain value")
	}
}
func TestImmutableInt64Int8HashMap_Generated_GetOrDefault(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	im := m.ToImmutable()
	if v := im.GetOrDefault(1, 3); v != 1 {
		t.Errorf("got %v", v)
	}
	if v := im.GetOrDefault(99, 3); v != 3 {
		t.Errorf("got %v", v)
	}
}
func TestImmutableInt64Int8HashMap_Generated_IsEmpty(t *testing.T) {
	im := NewInt64Int8HashMap().ToImmutable()
	if !im.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestImmutableInt64Int8HashMap_Generated_All(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	im := m.ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}
func TestImmutableInt64Int8HashMap_Generated_Keys(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	count := 0
	for range m.ToImmutable().Keys() {
		count++
	}
	if count != 2 {
		t.Errorf("Keys count = %d", count)
	}
}
func TestImmutableInt64Int8HashMap_Generated_Values(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	count := 0
	for range m.ToImmutable().Values() {
		count++
	}
	if count != 2 {
		t.Errorf("Values count = %d", count)
	}
}
func TestImmutableInt64Int8HashMap_Generated_Select(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)
	sel := m.ToImmutable().Select(func(k int64, v int8) bool { return v > 1 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Size())
	}
}
func TestImmutableInt64Int8HashMap_Generated_Reject(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	m.Put(2, 2)
	m.Put(3, 3)
	rej := m.ToImmutable().Reject(func(k int64, v int8) bool { return v > 1 })
	if rej.Size() != 1 {
		t.Errorf("Reject size = %d, want 1", rej.Size())
	}
}
func TestImmutableInt64Int8HashMap_Generated_Equals(t *testing.T) {
	m1 := NewInt64Int8HashMap()
	m1.Put(1, 1)
	m1.Put(2, 2)
	m2 := NewInt64Int8HashMap()
	m2.Put(2, 2)
	m2.Put(1, 1)
	if !m1.ToImmutable().Equals(m2.ToImmutable()) {
		t.Error("Should be equal")
	}
}
func TestImmutableInt64Int8HashMap_Generated_ToMutable(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	im := m.ToImmutable()
	m2 := im.ToMutable()
	m2.Put(2, 2)
	if im.Size() != 1 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableInt64Int8HashMap_Generated_String(t *testing.T) {
	m := NewInt64Int8HashMap()
	m.Put(1, 1)
	if m.ToImmutable().String() == "" {
		t.Error("String empty")
	}
}
