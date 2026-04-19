package hashmap

import (
	"testing"
)

func TestInt32ObjectHashMap_PutGet(t *testing.T) {
	m := NewInt32ObjectHashMap[string]()
	m.Put(1, "hello")
	m.Put(2, "world")

	if v, ok := m.Get(1); !ok || v != "hello" {
		t.Errorf("Get(1) = (%s, %v), want (hello, true)", v, ok)
	}
	if m.Size() != 2 {
		t.Errorf("Size() = %d, want 2", m.Size())
	}
}

func TestInt32ObjectHashMap_Remove(t *testing.T) {
	m := NewInt32ObjectHashMap[string]()
	m.Put(1, "a")
	m.Put(2, "b")
	old, ok := m.Remove(1)
	if !ok || old != "a" {
		t.Errorf("Remove(1) = (%s, %v), want (a, true)", old, ok)
	}
	if m.Size() != 1 {
		t.Errorf("Size after remove = %d, want 1", m.Size())
	}
}
