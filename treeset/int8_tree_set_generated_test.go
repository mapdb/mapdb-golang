
package treeset

import "testing"

func TestInt8TreeSet_Generated_AddContains(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(3)
	s.Add(1)
	s.Add(2)
	if s.Size() != 3 {
		t.Errorf("Size = %d", s.Size())
	}
	if !s.Contains(2) {
		t.Error("Contains should be true")
	}
	if s.Contains(99) {
		t.Error("Contains should be false")
	}
}
func TestInt8TreeSet_Generated_AddDuplicate(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	added := s.Add(1)
	if added {
		t.Error("Duplicate should return false")
	}
}
func TestInt8TreeSet_Generated_Remove(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	s.Add(2)
	if !s.Remove(1) {
		t.Error("Remove should return true")
	}
	if s.Contains(1) {
		t.Error("Should not contain")
	}
}
func TestInt8TreeSet_Generated_MinMax(t *testing.T) {
	s := NewInt8TreeSet()
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
func TestInt8TreeSet_Generated_FloorCeiling(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	s.Add(3)
	if f, ok := s.Floor(2); !ok || f != 1 {
		t.Errorf("Floor = %v", f)
	}
	if c, ok := s.Ceiling(2); !ok || c != 3 {
		t.Errorf("Ceiling = %v", c)
	}
}
func TestInt8TreeSet_Generated_Clear(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestInt8TreeSet_Generated_SortedIteration(t *testing.T) {
	s := NewInt8TreeSet()
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
func TestInt8TreeSet_Generated_Select(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	sel := s.Select(func(v int8) bool { return v > 1 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d", sel.Size())
	}
}
func TestInt8TreeSet_Generated_Union(t *testing.T) {
	a := NewInt8TreeSet()
	a.Add(1)
	a.Add(2)
	b := NewInt8TreeSet()
	b.Add(2)
	b.Add(3)
	if a.Union(b).Size() != 3 {
		t.Error("Union wrong")
	}
}
func TestInt8TreeSet_Generated_Intersect(t *testing.T) {
	a := NewInt8TreeSet()
	a.Add(1)
	a.Add(2)
	b := NewInt8TreeSet()
	b.Add(2)
	b.Add(3)
	if a.Intersect(b).Size() != 1 {
		t.Error("Intersect wrong")
	}
}
func TestInt8TreeSet_Generated_Difference(t *testing.T) {
	a := NewInt8TreeSet()
	a.Add(1)
	a.Add(2)
	b := NewInt8TreeSet()
	b.Add(2)
	b.Add(3)
	if a.Difference(b).Size() != 1 {
		t.Error("Difference wrong")
	}
}
func TestInt8TreeSet_Generated_ToSlice(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	s.Add(2)
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}
func TestInt8TreeSet_Generated_String(t *testing.T) {
	s := NewInt8TreeSet()
	s.Add(1)
	if s.String() == "" {
		t.Error("empty")
	}
}
