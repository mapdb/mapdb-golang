
package hashset

import (
	"math"
	"testing"
)

func TestFloat64HashSet_Generated_AddContains(t *testing.T) {
	s := NewFloat64HashSet()
	s.Add(1.0)
	s.Add(2.0)
	s.Add(3.0)

	if s.Size() != 3 {
		t.Errorf("Size() = %d, want 3", s.Size())
	}
	if !s.Contains(2.0) {
		t.Error("Contains(2.0) should be true")
	}
	if s.Contains(99.0) {
		t.Error("Contains(99.0) should be false")
	}
}

func TestFloat64HashSet_Generated_AddDuplicate(t *testing.T) {
	s := NewFloat64HashSet()
	added1 := s.Add(1.0)
	added2 := s.Add(1.0)
	if !added1 || added2 {
		t.Errorf("Add duplicate: first=%v second=%v, want true, false", added1, added2)
	}
	if s.Size() != 1 {
		t.Errorf("Size after duplicate = %d, want 1", s.Size())
	}
}

func TestFloat64HashSet_Generated_AddAll(t *testing.T) {
	s := NewFloat64HashSet()
	s.AddAll(1.0, 2.0, 3.0)
	if s.Size() != 3 {
		t.Errorf("AddAll: Size() = %d, want 3", s.Size())
	}
}

func TestFloat64HashSet_Generated_Remove(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	removed := s.Remove(2.0)
	if !removed || s.Size() != 2 || s.Contains(2.0) {
		t.Errorf("After Remove: removed=%v size=%d contains=%v", removed, s.Size(), s.Contains(2.0))
	}
	if s.Remove(99.0) {
		t.Error("Remove missing value should return false")
	}
}

func TestFloat64HashSet_Generated_IsEmpty(t *testing.T) {
	s := NewFloat64HashSet()
	if !s.IsEmpty() {
		t.Error("New set should be empty")
	}
	s.Add(1.0)
	if s.IsEmpty() {
		t.Error("Set with element should not be empty")
	}
}

func TestFloat64HashSet_Generated_Clear(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	s.Clear()
	if s.Size() != 0 || !s.IsEmpty() {
		t.Errorf("After Clear: size=%d, empty=%v", s.Size(), s.IsEmpty())
	}
}

func TestFloat64HashSet_Generated_All(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	count := 0
	for range s.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}

func TestFloat64HashSet_Generated_Select(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0, 4.0, 5.0)
	selected := s.Select(func(v float64) bool { return v > 3.0 })
	if selected.Size() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Size())
	}
}

func TestFloat64HashSet_Generated_Reject(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0, 4.0, 5.0)
	rejected := s.Reject(func(v float64) bool { return v > 3.0 })
	if rejected.Size() != 3 {
		t.Errorf("Reject size = %d, want 3", rejected.Size())
	}
}

func TestFloat64HashSet_Generated_Detect(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	val, found := s.Detect(func(v float64) bool { return v == 2.0 })
	if !found || val != 2.0 {
		t.Errorf("Detect = (%v, %v), want (2.0, true)", val, found)
	}
}

func TestFloat64HashSet_Generated_AnySatisfy(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	if !s.AnySatisfy(func(v float64) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if s.AnySatisfy(func(v float64) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestFloat64HashSet_Generated_AllSatisfy(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	if !s.AllSatisfy(func(v float64) bool { return v > 0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
}

func TestFloat64HashSet_Generated_NoneSatisfy(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	if !s.NoneSatisfy(func(v float64) bool { return v == 99.0 }) {
		t.Error("NoneSatisfy should be true")
	}
}

func TestFloat64HashSet_Generated_Union(t *testing.T) {
	a := Float64HashSetOf(1.0, 2.0, 3.0)
	b := Float64HashSetOf(3.0, 4.0, 5.0)
	u := a.Union(b)
	if u.Size() != 5 {
		t.Errorf("Union size = %d, want 5", u.Size())
	}
}

func TestFloat64HashSet_Generated_Intersect(t *testing.T) {
	a := Float64HashSetOf(1.0, 2.0, 3.0)
	b := Float64HashSetOf(2.0, 3.0, 4.0)
	i := a.Intersect(b)
	if i.Size() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Size())
	}
}

func TestFloat64HashSet_Generated_Difference(t *testing.T) {
	a := Float64HashSetOf(1.0, 2.0, 3.0)
	b := Float64HashSetOf(2.0, 3.0, 4.0)
	d := a.Difference(b)
	if d.Size() != 1 || !d.Contains(1.0) {
		t.Errorf("Difference size=%d, contains(1.0)=%v", d.Size(), d.Contains(1.0))
	}
}

func TestFloat64HashSet_Generated_SymmetricDifference(t *testing.T) {
	a := Float64HashSetOf(1.0, 2.0, 3.0)
	b := Float64HashSetOf(2.0, 3.0, 4.0)
	sd := a.SymmetricDifference(b)
	if sd.Size() != 2 {
		t.Errorf("SymmetricDifference size = %d, want 2", sd.Size())
	}
}

func TestFloat64HashSet_Generated_ToSlice(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	sl := s.ToSlice()
	if len(sl) != 3 {
		t.Errorf("ToSlice len = %d, want 3", len(sl))
	}
}

func TestFloat64HashSet_Generated_With(t *testing.T) {
	s := Float64HashSetOf(1.0)
	s2 := s.With(2.0)
	if s2.Size() != 2 || !s2.Contains(2.0) {
		t.Errorf("With: size=%d, contains=%v", s2.Size(), s2.Contains(2.0))
	}
}

func TestFloat64HashSet_Generated_Without(t *testing.T) {
	s := Float64HashSetOf(1.0, 2.0, 3.0)
	s2 := s.Without(2.0)
	if s2.Contains(2.0) {
		t.Error("Without should remove the value")
	}
}

func TestFloat64HashSet_Generated_Equals(t *testing.T) {
	s1 := Float64HashSetOf(1.0, 2.0, 3.0)
	s2 := Float64HashSetOf(3.0, 1.0, 2.0)
	s3 := Float64HashSetOf(1.0, 2.0)
	if !s1.Equals(s2) {
		t.Error("Equal sets should be equal regardless of insertion order")
	}
	if s1.Equals(s3) {
		t.Error("Different sets should not be equal")
	}
}

func TestFloat64HashSet_Generated_String(t *testing.T) {
	s := Float64HashSetOf(1.0)
	if s.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestFloat64HashSet_NaN_Findable(t *testing.T) {
	s := NewFloat64HashSet()
	s.Add(float64(math.NaN()))
	if !s.Contains(float64(math.NaN())) {
		t.Errorf("expected NaN element to be findable")
	}
	if s.Size() != 1 {
		t.Errorf("expected size 1, got %d", s.Size())
	}
}

func TestFloat64HashSet_NaN_AddDuplicate(t *testing.T) {
	s := NewFloat64HashSet()
	added1 := s.Add(float64(math.NaN()))
	added2 := s.Add(float64(math.NaN()))
	added3 := s.Add(float64(math.NaN()))
	if !added1 {
		t.Errorf("expected first Add(NaN) to return true")
	}
	if added2 || added3 {
		t.Errorf("expected duplicate Add(NaN) to return false, got %v %v", added2, added3)
	}
	if s.Size() != 1 {
		t.Errorf("expected size 1 after 3 NaN adds, got %d", s.Size())
	}
}

func TestFloat64HashSet_NaN_Remove(t *testing.T) {
	s := NewFloat64HashSet()
	s.Add(float64(math.NaN()))
	if !s.Remove(float64(math.NaN())) {
		t.Errorf("expected NaN element to be removable")
	}
	if s.Size() != 0 {
		t.Errorf("expected size 0 after remove, got %d", s.Size())
	}
	if s.Contains(float64(math.NaN())) {
		t.Errorf("expected Contains(NaN) to be false after remove")
	}
}

func TestFloat64HashSet_NegativeZeroDistinct(t *testing.T) {
	s := NewFloat64HashSet()
	s.Add(float64(0.0))
	s.Add(float64(math.Copysign(0, -1)))
	if s.Size() != 2 {
		t.Errorf("expected size 2 with +0 and -0 distinct, got %d", s.Size())
	}
	if !s.Contains(float64(0.0)) {
		t.Errorf("expected Contains(+0) to be true")
	}
	if !s.Contains(float64(math.Copysign(0, -1))) {
		t.Errorf("expected Contains(-0) to be true")
	}
}

func TestFloat64HashSet_Infinity(t *testing.T) {
	s := NewFloat64HashSet()
	s.Add(float64(math.Inf(1)))
	s.Add(float64(math.Inf(-1)))
	if s.Size() != 2 {
		t.Errorf("expected size 2, got %d", s.Size())
	}
	if !s.Contains(float64(math.Inf(1))) {
		t.Errorf("expected Contains(+Inf) to be true")
	}
	if !s.Contains(float64(math.Inf(-1))) {
		t.Errorf("expected Contains(-Inf) to be true")
	}
}
