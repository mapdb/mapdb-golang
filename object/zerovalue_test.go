package object

import "testing"

// Zero-value (nil-map) usability: freshly declared generic object collections
// must accept mutations without panicking ("assignment to entry in nil map").
// Phase 7a.

func TestZeroValueHashMap(t *testing.T) {
	var m HashMap[int, int]
	m.Put(1, 2)
	if v, ok := m.Get(1); !ok || v != 2 {
		t.Fatalf("Get(1) = %d,%v, want 2,true", v, ok)
	}
}

func TestZeroValueHashSet(t *testing.T) {
	var s HashSet[int]
	if !s.Add(1) || s.Add(1) {
		t.Fatal("Add semantics wrong on zero value")
	}
	if s.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", s.Size())
	}
}

func TestZeroValueHashBag(t *testing.T) {
	var b HashBag[int]
	b.Add(1)
	b.AddOccurrences(2, 3)
	if b.Size() != 4 {
		t.Fatalf("Size() = %d, want 4", b.Size())
	}
}

func TestZeroValueLinkedHashMap(t *testing.T) {
	var m LinkedHashMap[int, int]
	m.Put(1, 2)
	if m.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", m.Size())
	}
}

func TestZeroValueLinkedHashSet(t *testing.T) {
	var s LinkedHashSet[int]
	s.Add(1)
	if s.Size() != 1 {
		t.Fatalf("Size() = %d, want 1", s.Size())
	}
}

func TestZeroValueHashBiMap(t *testing.T) {
	var m HashBiMap[int, int]
	m.Put(1, 2)
	if v, ok := m.Get(1); !ok || v != 2 {
		t.Fatalf("Get(1) = %d,%v, want 2,true", v, ok)
	}
}

func TestZeroValueHashMultimap(t *testing.T) {
	var m HashMultimap[int, int]
	m.Put(1, 10)
	m.PutAll(1, 20, 30)
	if m.Size() != 3 {
		t.Fatalf("Size() = %d, want 3", m.Size())
	}
}
