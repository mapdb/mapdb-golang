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

func TestMultimap_EmptyPutAllDoesNotCreateKey(t *testing.T) {
	for _, newMap := range []func() *Multimap[string, int]{
		NewMultimap[string, int],
		func() *Multimap[string, int] { return new(Multimap[string, int]) },
	} {
		m := newMap()
		m.PutAll("absent")
		if m.Len() != 0 || m.SizeDistinct() != 0 || m.ContainsKey("absent") {
			t.Fatalf("empty PutAll created a key: len=%d distinct=%d", m.Len(), m.SizeDistinct())
		}
		for range m.Keys() {
			t.Fatal("empty PutAll exposed a key through Keys")
		}

		m.Put("present", 7)
		m.PutAll("present")
		if m.Len() != 1 || m.SizeDistinct() != 1 || len(m.Get("present")) != 1 {
			t.Fatal("empty PutAll changed an existing key")
		}
	}
}

func TestMultimap_RemoveAllClearsEmptyStoredKey(t *testing.T) {
	m := NewMultimap[string, int]()
	m.store("empty", nil)
	m.RemoveAll("empty")
	if m.SizeDistinct() != 0 {
		t.Fatalf("RemoveAll left an empty stored key: distinct=%d", m.SizeDistinct())
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
