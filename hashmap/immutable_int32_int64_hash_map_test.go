package hashmap

import (
	"testing"
)

func TestImmutableInt32Int64HashMap_Get(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	m.Put(2, 200)

	im := m.ToImmutable()

	if v, ok := im.Get(1); !ok || v != 100 {
		t.Errorf("Get(1) = (%d, %v), want (100, true)", v, ok)
	}
	if im.Size() != 2 {
		t.Errorf("Size = %d, want 2", im.Size())
	}
}

func TestImmutableInt32Int64HashMap_Select(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 10)
	m.Put(2, 20)
	m.Put(3, 30)

	im := m.ToImmutable()
	selected := im.Select(func(k int32, v int64) bool { return v > 15 })
	if selected.Size() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Size())
	}
}

func TestImmutableInt32Int64HashMap_ToMutable(t *testing.T) {
	m := NewInt32Int64HashMap()
	m.Put(1, 100)
	im := m.ToImmutable()
	mut := im.ToMutable()
	mut.Put(2, 200)

	if mut.Size() != 2 {
		t.Errorf("Mutable copy size = %d, want 2", mut.Size())
	}
	if im.Size() != 1 {
		t.Errorf("Immutable size should still be 1, got %d", im.Size())
	}
}
