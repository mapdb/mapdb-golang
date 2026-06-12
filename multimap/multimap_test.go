package multimap

import (
	"testing"
)

func TestMultimap_PutGet(t *testing.T) {
	m := NewMultimap[string, int]()
	m.Put("a", 1)
	m.Put("a", 2)
	m.Put("b", 3)

	if m.Len() != 3 {
		t.Errorf("Size = %d, want 3", m.Len())
	}
	if m.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", m.SizeDistinct())
	}

	vals := m.Get("a")
	if len(vals) != 2 || vals[0] != 1 || vals[1] != 2 {
		t.Errorf("Get(a) = %v, want [1,2]", vals)
	}
}

func TestMultimap_RemoveAll(t *testing.T) {
	m := NewMultimap[string, int]()
	m.PutAll("x", 1, 2, 3)
	removed := m.RemoveAll("x")
	if len(removed) != 3 {
		t.Errorf("removed = %v", removed)
	}
	if m.Len() != 0 {
		t.Errorf("Size = %d", m.Len())
	}
	if m.ContainsKey("x") {
		t.Error("should not contain x")
	}
}

func TestMultimap_All(t *testing.T) {
	m := NewMultimap[int, string]()
	m.Put(1, "a")
	m.Put(1, "b")
	m.Put(2, "c")

	count := 0
	for range m.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestMultimap_ForEachKey(t *testing.T) {
	m := NewMultimap[string, int]()
	m.PutAll("x", 1, 2, 3)
	m.PutAll("y", 4, 5)

	total := 0
	m.ForEachKey(func(k string, vals []int) {
		total += len(vals)
	})
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
}
