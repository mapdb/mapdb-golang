package hashset

import (
	"testing"
)

func TestInt16_Generated_AddContains(t *testing.T) {
	s := NewInt16()
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

func TestInt16_Generated_AddDuplicate(t *testing.T) {
	s := NewInt16()
	added1 := s.Add(1)
	added2 := s.Add(1)
	if !added1 || added2 {
		t.Errorf("Add duplicate: first=%v second=%v, want true, false", added1, added2)
	}
	if s.Len() != 1 {
		t.Errorf("Size after duplicate = %d, want 1", s.Len())
	}
}

func TestInt16_Generated_AddAll(t *testing.T) {
	s := NewInt16()
	s.AddAll(1, 2, 3)
	if s.Len() != 3 {
		t.Errorf("AddAll: Size() = %d, want 3", s.Len())
	}
}

func TestInt16_Generated_Remove(t *testing.T) {
	s := Int16Of(1, 2, 3)
	removed := s.Remove(2)
	if !removed || s.Len() != 2 || s.Contains(2) {
		t.Errorf("After Remove: removed=%v size=%d contains=%v", removed, s.Len(), s.Contains(2))
	}
	if s.Remove(99) {
		t.Error("Remove missing value should return false")
	}
}

func TestInt16_Generated_IsEmpty(t *testing.T) {
	s := NewInt16()
	if s.Len() != 0 {
		t.Error("New set should be empty")
	}
	s.Add(1)
	if s.Len() == 0 {
		t.Error("Set with element should not be empty")
	}
}

func TestInt16_Generated_Clear(t *testing.T) {
	s := Int16Of(1, 2, 3)
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", s.Len(), s.Len() == 0)
	}
}

func TestInt16_Generated_All(t *testing.T) {
	s := Int16Of(1, 2, 3)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestInt16_Generated_Select(t *testing.T) {
	s := Int16Of(1, 2, 3, 4, 5)
	selected := s.Select(func(v int16) bool { return v > 3 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestInt16_Generated_Reject(t *testing.T) {
	s := Int16Of(1, 2, 3, 4, 5)
	rejected := s.Reject(func(v int16) bool { return v > 3 })
	if rejected.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", rejected.Len())
	}
}

func TestInt16_Generated_Detect(t *testing.T) {
	s := Int16Of(1, 2, 3)
	val, found := s.Detect(func(v int16) bool { return v == 2 })
	if !found || val != 2 {
		t.Errorf("Detect = (%v, %v), want (2, true)", val, found)
	}
}

func TestInt16_Generated_AnySatisfy(t *testing.T) {
	s := Int16Of(1, 2, 3)
	if !s.AnySatisfy(func(v int16) bool { return v == 2 }) {
		t.Error("AnySatisfy should be true")
	}
	if s.AnySatisfy(func(v int16) bool { return v == 99 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestInt16_Generated_AllSatisfy(t *testing.T) {
	s := Int16Of(1, 2, 3)
	if !s.AllSatisfy(func(v int16) bool { return v > 0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
}

func TestInt16_Generated_NoneSatisfy(t *testing.T) {
	s := Int16Of(1, 2, 3)
	if !s.NoneSatisfy(func(v int16) bool { return v == 99 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestInt16_Generated_Union(t *testing.T) {
	a := Int16Of(1, 2, 3)
	b := Int16Of(3, 4, 5)
	u := a.Union(b)
	if u.Len() != 5 {
		t.Errorf("Union size = %d, want 5", u.Len())
	}
}

func TestInt16_Generated_Intersect(t *testing.T) {
	a := Int16Of(1, 2, 3)
	b := Int16Of(2, 3, 4)
	i := a.Intersect(b)
	if i.Len() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Len())
	}
}

func TestInt16_Generated_Difference(t *testing.T) {
	a := Int16Of(1, 2, 3)
	b := Int16Of(2, 3, 4)
	d := a.Difference(b)
	if d.Len() != 1 || !d.Contains(1) {
		t.Errorf("Difference size=%d, contains(1)=%v", d.Len(), d.Contains(1))
	}
}

func TestInt16_Generated_SymmetricDifference(t *testing.T) {
	a := Int16Of(1, 2, 3)
	b := Int16Of(2, 3, 4)
	sd := a.SymmetricDifference(b)
	if sd.Len() != 2 {
		t.Errorf("SymmetricDifference size = %d, want 2", sd.Len())
	}
}

func TestInt16_Generated_ToSlice(t *testing.T) {
	s := Int16Of(1, 2, 3)
	sl := s.ToSlice()
	if len(sl) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(sl))
	}
}

func TestInt16_Generated_With(t *testing.T) {
	s := Int16Of(1)
	s2 := s.AddReturning(2)
	if s2.Len() != 2 || !s2.Contains(2) {
		t.Errorf("With: size=%d, contains=%v", s2.Len(), s2.Contains(2))
	}
}

func TestInt16_Generated_Without(t *testing.T) {
	s := Int16Of(1, 2, 3)
	s2 := s.RemoveReturning(2)
	if s2.Contains(2) {
		t.Error("Without should remove the value")
	}
}

func TestInt16_Generated_Equals(t *testing.T) {
	s1 := Int16Of(1, 2, 3)
	s2 := Int16Of(3, 1, 2)
	s3 := Int16Of(1, 2)
	if !s1.Equals(s2) {
		t.Error("Equal sets should be equal regardless of insertion order")
	}
	if s1.Equals(s3) {
		t.Error("Different sets should not be equal")
	}
}

func TestInt16_Generated_String(t *testing.T) {
	s := Int16Of(1)
	if s.String() == "" {
		t.Error("String should not be empty")
	}
}
