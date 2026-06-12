package hashmap

import (
	"testing"
)

func TestObjectInt32_PutGet(t *testing.T) {
	m := NewObjectInt32[string]()
	m.Put("hello", 1)
	m.Put("world", 2)

	if v, ok := m.Get("hello"); !ok || v != 1 {
		t.Errorf("Get(hello) = (%d, %v), want (1, true)", v, ok)
	}
	if m.Len() != 2 {
		t.Errorf("Size() = %d, want 2", m.Len())
	}
}

func TestObjectInt32_Remove(t *testing.T) {
	m := NewObjectInt32[string]()
	m.Put("a", 10)
	m.Put("b", 20)
	old, ok := m.Remove("a")
	if !ok || old != 10 {
		t.Errorf("Remove(a) = (%d, %v), want (10, true)", old, ok)
	}
	if m.Len() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Len())
	}
}

func TestObjectInt32_All(t *testing.T) {
	m := NewObjectInt32[string]()
	m.Put("x", 100)
	m.Put("y", 200)

	count := 0
	for range m.All() {
		count++
	}
	if count != 2 {
		t.Errorf("All count = %d, want 2", count)
	}
}
