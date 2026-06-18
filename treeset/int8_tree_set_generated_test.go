package treeset

import "testing"

func TestInt8_Generated_AddContains(t *testing.T) {
	s := NewInt8()
	s.Add(3)
	s.Add(1)
	s.Add(2)
	if s.Len() != 3 {
		t.Errorf("Size = %d", s.Len())
	}
	if !s.Contains(2) {
		t.Error("Contains should be true")
	}
	if s.Contains(99) {
		t.Error("Contains should be false")
	}
}
func TestInt8_Generated_AddDuplicate(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	added := s.Add(1)
	if added {
		t.Error("Duplicate should return false")
	}
}
func TestInt8_Generated_Remove(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	s.Add(2)
	if !s.Remove(1) {
		t.Error("Remove should return true")
	}
	if s.Contains(1) {
		t.Error("Should not contain")
	}
}
func TestInt8_Generated_MinMax(t *testing.T) {
	s := NewInt8()
	s.Add(3)
	s.Add(1)
	s.Add(2)
	if min, ok := s.Min(); !ok || min != 1 {
		t.Errorf("Min = %v", min)
	}
	if max, ok := s.Max(); !ok || max != 3 {
		t.Errorf("Max = %v", max)
	}
}
func TestInt8_Generated_FloorCeiling(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	s.Add(3)
	if f, ok := s.Floor(2); !ok || f != 1 {
		t.Errorf("Floor = %v", f)
	}
	if c, ok := s.Ceiling(2); !ok || c != 3 {
		t.Errorf("Ceiling = %v", c)
	}
}
func TestInt8_Generated_Clear(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	s.Clear()
	if s.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestInt8_Generated_SortedIteration(t *testing.T) {
	s := NewInt8()
	s.Add(3)
	s.Add(1)
	s.Add(2)
	var vals []int8
	for v := range s.All() {
		vals = append(vals, v)
	}
	for i := 1; i < len(vals); i++ {
		if vals[i] < vals[i-1] {
			t.Errorf("Not sorted at %d: %v", i, vals)
		}
	}
}
func TestInt8_Generated_SelectWhere(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	sel := s.SelectWhere(func(v int8) bool { return v > 1 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d", sel.Len())
	}
}
func TestInt8_Generated_Union(t *testing.T) {
	a := NewInt8()
	a.Add(1)
	a.Add(2)
	b := NewInt8()
	b.Add(2)
	b.Add(3)
	if a.Union(b).Len() != 3 {
		t.Error("Union wrong")
	}
}
func TestInt8_Generated_Intersect(t *testing.T) {
	a := NewInt8()
	a.Add(1)
	a.Add(2)
	b := NewInt8()
	b.Add(2)
	b.Add(3)
	if a.Intersect(b).Len() != 1 {
		t.Error("Intersect wrong")
	}
}
func TestInt8_Generated_Difference(t *testing.T) {
	a := NewInt8()
	a.Add(1)
	a.Add(2)
	b := NewInt8()
	b.Add(2)
	b.Add(3)
	if a.Difference(b).Len() != 1 {
		t.Error("Difference wrong")
	}
}
func TestInt8_Generated_ToSlice(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	s.Add(2)
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}
func TestInt8_Generated_String(t *testing.T) {
	s := NewInt8()
	s.Add(1)
	if s.String() == "" {
		t.Error("empty")
	}
}
