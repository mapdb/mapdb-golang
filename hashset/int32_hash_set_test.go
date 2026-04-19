package hashset

import (
	"testing"
)

func TestInt32HashSet_AddContains(t *testing.T) {
	s := NewInt32HashSet()
	s.Add(1)
	s.Add(2)
	s.Add(3)

	if s.Size() != 3 {
		t.Errorf("Size() = %d, want 3", s.Size())
	}
	if !s.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if s.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32HashSet_AddDuplicate(t *testing.T) {
	s := NewInt32HashSet()
	added1 := s.Add(1)
	added2 := s.Add(1)
	if !added1 || added2 {
		t.Errorf("Add(1) twice: first=%v second=%v, want true, false", added1, added2)
	}
	if s.Size() != 1 {
		t.Errorf("Size after duplicate = %d, want 1", s.Size())
	}
}

func TestInt32HashSet_Remove(t *testing.T) {
	s := Int32HashSetOf(1, 2, 3)
	removed := s.Remove(2)
	if !removed || s.Size() != 2 || s.Contains(2) {
		t.Errorf("After Remove(2): removed=%v size=%d contains=%v", removed, s.Size(), s.Contains(2))
	}
}

func TestInt32HashSet_Union(t *testing.T) {
	a := Int32HashSetOf(1, 2, 3)
	b := Int32HashSetOf(3, 4, 5)
	u := a.Union(b)
	if u.Size() != 5 {
		t.Errorf("Union size = %d, want 5", u.Size())
	}
}

func TestInt32HashSet_Intersect(t *testing.T) {
	a := Int32HashSetOf(1, 2, 3)
	b := Int32HashSetOf(2, 3, 4)
	i := a.Intersect(b)
	if i.Size() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Size())
	}
}

func TestInt32HashSet_Difference(t *testing.T) {
	a := Int32HashSetOf(1, 2, 3)
	b := Int32HashSetOf(2, 3, 4)
	d := a.Difference(b)
	if d.Size() != 1 || !d.Contains(1) {
		t.Errorf("Difference size = %d, contains(1) = %v", d.Size(), d.Contains(1))
	}
}

func TestInt32HashSet_All(t *testing.T) {
	s := Int32HashSetOf(1, 2, 3)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
