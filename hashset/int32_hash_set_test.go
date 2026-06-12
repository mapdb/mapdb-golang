package hashset

import (
	"testing"
)

func TestInt32_AddContains(t *testing.T) {
	s := NewInt32()
	s.Add(1)
	s.Add(2)
	s.Add(3)

	if s.Len() != 3 {
		t.Errorf("Size() = %d, want 3", s.Len())
	}
	if !s.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if s.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32_AddDuplicate(t *testing.T) {
	s := NewInt32()
	added1 := s.Add(1)
	added2 := s.Add(1)
	if !added1 || added2 {
		t.Errorf("Add(1) twice: first=%v second=%v, want true, false", added1, added2)
	}
	if s.Len() != 1 {
		t.Errorf("Size after duplicate = %d, want 1", s.Len())
	}
}

func TestInt32_Remove(t *testing.T) {
	s := Int32Of(1, 2, 3)
	removed := s.Remove(2)
	if !removed || s.Len() != 2 || s.Contains(2) {
		t.Errorf("After Remove(2): removed=%v size=%d contains=%v", removed, s.Len(), s.Contains(2))
	}
}

func TestInt32_Union(t *testing.T) {
	a := Int32Of(1, 2, 3)
	b := Int32Of(3, 4, 5)
	u := a.Union(b)
	if u.Len() != 5 {
		t.Errorf("Union size = %d, want 5", u.Len())
	}
}

func TestInt32_Intersect(t *testing.T) {
	a := Int32Of(1, 2, 3)
	b := Int32Of(2, 3, 4)
	i := a.Intersect(b)
	if i.Len() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Len())
	}
}

func TestInt32_Difference(t *testing.T) {
	a := Int32Of(1, 2, 3)
	b := Int32Of(2, 3, 4)
	d := a.Difference(b)
	if d.Len() != 1 || !d.Contains(1) {
		t.Errorf("Difference size = %d, contains(1) = %v", d.Len(), d.Contains(1))
	}
}

func TestInt32_All(t *testing.T) {
	s := Int32Of(1, 2, 3)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
