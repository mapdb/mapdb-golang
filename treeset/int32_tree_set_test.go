package treeset

import (
	"testing"
)

func TestInt32_AddContains(t *testing.T) {
	s := NewInt32()
	s.Add(3)
	s.Add(1)
	s.Add(2)
	if s.Len() != 3 {
		t.Errorf("Size = %d, want 3", s.Len())
	}
	if !s.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if s.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt32_SortedIteration(t *testing.T) {
	s := Int32Of(50, 10, 30, 20, 40)
	values := s.ToSlice()
	expected := []int32{10, 20, 30, 40, 50}
	for i, v := range values {
		if v != expected[i] {
			t.Fatalf("got %v, want %v", values, expected)
		}
	}
}

func TestInt32_MinMax(t *testing.T) {
	s := Int32Of(30, 10, 50)
	min, ok := s.Min()
	if !ok || min != 10 {
		t.Errorf("Min = (%d, %v)", min, ok)
	}
	max, ok2 := s.Max()
	if !ok2 || max != 50 {
		t.Errorf("Max = (%d, %v)", max, ok2)
	}
}

func TestInt32_FloorCeiling(t *testing.T) {
	s := Int32Of(10, 20, 30)
	f, ok := s.Floor(25)
	if !ok || f != 20 {
		t.Errorf("Floor(25) = (%d, %v)", f, ok)
	}
	c, ok2 := s.Ceiling(25)
	if !ok2 || c != 30 {
		t.Errorf("Ceiling(25) = (%d, %v)", c, ok2)
	}
}

func TestInt32_UnionIntersect(t *testing.T) {
	a := Int32Of(1, 2, 3)
	b := Int32Of(3, 4, 5)
	if a.Union(b).Len() != 5 {
		t.Error("Union size should be 5")
	}
	if a.Intersect(b).Len() != 1 {
		t.Error("Intersect size should be 1")
	}
	if a.Difference(b).Len() != 2 {
		t.Error("Difference size should be 2")
	}
}

func TestInt32_Remove(t *testing.T) {
	s := Int32Of(1, 2, 3, 4, 5)
	s.Remove(3)
	if s.Len() != 4 || s.Contains(3) {
		t.Error("Remove(3) failed")
	}
	values := s.ToSlice()
	for i := 1; i < len(values); i++ {
		if values[i] <= values[i-1] {
			t.Fatal("Not sorted after remove")
		}
	}
}

func TestInt32_LargeInsertDelete(t *testing.T) {
	s := NewInt32()
	for i := int32(0); i < 500; i++ {
		s.Add(i)
	}
	for i := int32(0); i < 250; i++ {
		s.Remove(i)
	}
	if s.Len() != 250 {
		t.Fatalf("Size = %d, want 250", s.Len())
	}
	prev := int32(-1)
	for v := range s.All() {
		if v <= prev {
			t.Fatalf("Not sorted: %d after %d", v, prev)
		}
		prev = v
	}
}
