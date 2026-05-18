
package treeset

import "testing"

func TestFloat32TreeSet_Generated_AddContains(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(3.0)
	s.Add(1.0)
	s.Add(2.0)
	if s.Size() != 3 {
		t.Errorf("Size = %d", s.Size())
	}
	if !s.Contains(2.0) {
		t.Error("Contains should be true")
	}
	if s.Contains(99.0) {
		t.Error("Contains should be false")
	}
}
func TestFloat32TreeSet_Generated_AddDuplicate(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	added := s.Add(1.0)
	if added {
		t.Error("Duplicate should return false")
	}
}
func TestFloat32TreeSet_Generated_Remove(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	s.Add(2.0)
	if !s.Remove(1.0) {
		t.Error("Remove should return true")
	}
	if s.Contains(1.0) {
		t.Error("Should not contain")
	}
}
func TestFloat32TreeSet_Generated_MinMax(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(3.0)
	s.Add(1.0)
	s.Add(2.0)
	if min, ok := s.Min(); !ok || min != 1.0 {
		t.Errorf("Min = %v", min)
	}
	if max, ok := s.Max(); !ok || max != 3.0 {
		t.Errorf("Max = %v", max)
	}
}
func TestFloat32TreeSet_Generated_FloorCeiling(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	s.Add(3.0)
	if f, ok := s.Floor(2.0); !ok || f != 1.0 {
		t.Errorf("Floor = %v", f)
	}
	if c, ok := s.Ceiling(2.0); !ok || c != 3.0 {
		t.Errorf("Ceiling = %v", c)
	}
}
func TestFloat32TreeSet_Generated_Clear(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestFloat32TreeSet_Generated_SortedIteration(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(3.0)
	s.Add(1.0)
	s.Add(2.0)
	var vals []float32
	for v := range s.All() {
		vals = append(vals, v)
	}
	for i := 1; i < len(vals); i++ {
		if vals[i] < vals[i-1] {
			t.Errorf("Not sorted at %d: %v", i, vals)
		}
	}
}
func TestFloat32TreeSet_Generated_Select(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	s.Add(2.0)
	s.Add(3.0)
	sel := s.Select(func(v float32) bool { return v > 1.0 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d", sel.Size())
	}
}
func TestFloat32TreeSet_Generated_Union(t *testing.T) {
	a := NewFloat32TreeSet()
	a.Add(1.0)
	a.Add(2.0)
	b := NewFloat32TreeSet()
	b.Add(2.0)
	b.Add(3.0)
	if a.Union(b).Size() != 3 {
		t.Error("Union wrong")
	}
}
func TestFloat32TreeSet_Generated_Intersect(t *testing.T) {
	a := NewFloat32TreeSet()
	a.Add(1.0)
	a.Add(2.0)
	b := NewFloat32TreeSet()
	b.Add(2.0)
	b.Add(3.0)
	if a.Intersect(b).Size() != 1 {
		t.Error("Intersect wrong")
	}
}
func TestFloat32TreeSet_Generated_Difference(t *testing.T) {
	a := NewFloat32TreeSet()
	a.Add(1.0)
	a.Add(2.0)
	b := NewFloat32TreeSet()
	b.Add(2.0)
	b.Add(3.0)
	if a.Difference(b).Size() != 1 {
		t.Error("Difference wrong")
	}
}
func TestFloat32TreeSet_Generated_ToSlice(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	s.Add(2.0)
	if len(s.ToSlice()) != 2 {
		t.Error("wrong len")
	}
}
func TestFloat32TreeSet_Generated_String(t *testing.T) {
	s := NewFloat32TreeSet()
	s.Add(1.0)
	if s.String() == "" {
		t.Error("empty")
	}
}
