package hashmap

import (
	"testing"
)

func TestObjectInt32HashMap_PutGet(t *testing.T) {
	m := NewObjectInt32HashMap[string]()
	m.Put("hello", 1)
	m.Put("world", 2)

	if v, ok := m.Get("hello"); !ok || v != 1 {
		t.Errorf("Get(hello) = (%d, %v), want (1, true)", v, ok)
	}
	if m.Size() != 2 {
		t.Errorf("Size() = %d, want 2", m.Size())
	}
}

func TestObjectInt32HashMap_Remove(t *testing.T) {
	m := NewObjectInt32HashMap[string]()
	m.Put("a", 10)
	m.Put("b", 20)
	old, ok := m.Remove("a")
	if !ok || old != 10 {
		t.Errorf("Remove(a) = (%d, %v), want (10, true)", old, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Size())
	}
}

func TestObjectInt32HashMap_All(t *testing.T) {
	m := NewObjectInt32HashMap[string]()
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
