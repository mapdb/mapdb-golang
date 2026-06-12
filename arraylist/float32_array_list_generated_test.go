package arraylist

import (
	"testing"
)

func TestFloat32_Generated_AddAndGet(t *testing.T) {
	l := NewFloat32()
	l.Add(1.0)
	l.Add(2.0)
	l.Add(3.0)

	if l.Len() != 3 {
		t.Errorf("Size() = %d, want 3", l.Len())
	}
	if v := l.Get(0); v != 1.0 {
		t.Errorf("Get(0) = %v, want 1.0", v)
	}
	if v := l.Get(2); v != 3.0 {
		t.Errorf("Get(2) = %v, want 3.0", v)
	}
}

func TestFloat32_Generated_GetOutOfBounds(t *testing.T) {
	l := Float32Of(1.0, 2.0)
	assertPanics(t, func() { l.Get(-1) })
	assertPanics(t, func() { l.Get(2) })
}

func TestFloat32_Generated_Of(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if l.Len() != 3 {
		t.Errorf("Of: Size() = %d, want 3", l.Len())
	}
	if v := l.Get(1); v != 2.0 {
		t.Errorf("Of: Get(1) = %v, want 2.0", v)
	}
}

func TestFloat32_Generated_Set(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	old := l.Set(1, 4.0)
	if old != 2.0 {
		t.Errorf("Set returned %v, want 2.0", old)
	}
	if v := l.Get(1); v != 4.0 {
		t.Errorf("After Set: Get(1) = %v, want 4.0", v)
	}
	assertPanics(t, func() { l.Set(99, 1.0) })
}

func TestFloat32_Generated_RemoveAtIndex(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	removed := l.RemoveAtIndex(1)
	if removed != 2.0 {
		t.Errorf("RemoveAtIndex(1) = %v, want 2.0", removed)
	}
	if l.Len() != 2 {
		t.Errorf("Size after remove = %d, want 2", l.Len())
	}
	assertPanics(t, func() { l.RemoveAtIndex(99) })
}

func TestFloat32_Generated_Remove(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if !l.Remove(2.0) {
		t.Error("Remove(2.0) should return true")
	}
	if l.Contains(2.0) {
		t.Error("Should not contain 2.0 after remove")
	}
	if l.Remove(99.0) {
		t.Error("Remove(99.0) should return false")
	}
}

func TestFloat32_Generated_Contains(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if !l.Contains(2.0) {
		t.Error("Contains(2.0) should be true")
	}
	if l.Contains(99.0) {
		t.Error("Contains(99.0) should be false")
	}
}

func TestFloat32_Generated_IndexOf(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if idx := l.IndexOf(2.0); idx != 1 {
		t.Errorf("IndexOf(2.0) = %d, want 1", idx)
	}
	if idx := l.IndexOf(99.0); idx != -1 {
		t.Errorf("IndexOf(99.0) = %d, want -1", idx)
	}
}

func TestFloat32_Generated_IsEmpty(t *testing.T) {
	l := NewFloat32()
	if l.Len() != 0 {
		t.Error("New list should be empty")
	}
	l.Add(1.0)
	if l.Len() == 0 {
		t.Error("List with element should not be empty")
	}
}

func TestFloat32_Generated_Clear(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	l.Clear()
	if l.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", l.Len(), l.Len() == 0)
	}
}

func TestFloat32_Generated_AddAll(t *testing.T) {
	l := NewFloat32()
	l.AddAll(1.0, 2.0, 3.0)
	if l.Len() != 3 {
		t.Errorf("AddAll: Size() = %d, want 3", l.Len())
	}
}

func TestFloat32_Generated_All(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	sum := float32(0)
	for v := range l.All() {
		sum += v
	}
	expected := float32(1.0) + float32(2.0) + float32(3.0)
	if sum != expected {
		t.Errorf("All sum = %v, want %v", sum, expected)
	}
}

func TestFloat32_Generated_AllWithIndex(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	indices := 0
	for i, _ := range l.AllWithIndex() {
		indices += i
	}
	if indices != 3 { // 0+1+2 = 3
		t.Errorf("AllWithIndex index sum = %d, want 3", indices)
	}
}

func TestFloat32_Generated_Select(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0, 4.0, 5.0)
	selected := l.Select(func(v float32) bool { return v > 3.0 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestFloat32_Generated_Reject(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0, 4.0, 5.0)
	rejected := l.Reject(func(v float32) bool { return v > 3.0 })
	if rejected.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", rejected.Len())
	}
}

func TestFloat32_Generated_Detect(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	val, found := l.Detect(func(v float32) bool { return v == 2.0 })
	if !found || val != 2.0 {
		t.Errorf("Detect = (%v, %v), want (2.0, true)", val, found)
	}
	_, found = l.Detect(func(v float32) bool { return v == 99.0 })
	if found {
		t.Error("Detect for missing should return false")
	}
}

func TestFloat32_Generated_AnySatisfy(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if !l.AnySatisfy(func(v float32) bool { return v == 2.0 }) {
		t.Error("AnySatisfy should be true")
	}
	if l.AnySatisfy(func(v float32) bool { return v == 99.0 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestFloat32_Generated_AllSatisfy(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if !l.AllSatisfy(func(v float32) bool { return v > 0.0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
	if l.AllSatisfy(func(v float32) bool { return v > 2.0 }) {
		t.Error("AllSatisfy should be false")
	}
}

func TestFloat32_Generated_NoneSatisfy(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	if !l.NoneSatisfy(func(v float32) bool { return v == 99.0 }) {
		t.Error("NoneSatisfy should be true for missing")
	}
	if l.NoneSatisfy(func(v float32) bool { return v == 1.0 }) {
		t.Error("NoneSatisfy should be false")
	}
}

func TestFloat32_Generated_Count(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0, 4.0, 5.0)
	if c := l.Count(func(v float32) bool { return v > 3.0 }); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestFloat32_Generated_Sum(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	expected := float32(1.0) + float32(2.0) + float32(3.0)
	if s := l.Sum(); s != expected {
		t.Errorf("Sum = %v, want %v", s, expected)
	}
}

func TestFloat32_Generated_MinMax(t *testing.T) {
	l := Float32Of(3.0, 1.0, 2.0)
	if min, ok := l.Min(); !ok || min != 1.0 {
		t.Errorf("Min = (%v, %v), want (1.0, true)", min, ok)
	}
	if max, ok := l.Max(); !ok || max != 3.0 {
		t.Errorf("Max = (%v, %v), want (3.0, true)", max, ok)
	}

	empty := NewFloat32()
	if _, ok := empty.Min(); ok {
		t.Error("Min on empty should return false")
	}
	if _, ok := empty.Max(); ok {
		t.Error("Max on empty should return false")
	}
}

func TestFloat32_Generated_Sort(t *testing.T) {
	l := Float32Of(3.0, 1.0, 2.0)
	l.Sort()
	v0 := l.Get(0)
	v1 := l.Get(1)
	v2 := l.Get(2)
	if v0 != 1.0 || v1 != 2.0 || v2 != 3.0 {
		t.Errorf("After sort: %v", l.ToSlice())
	}
}

func TestFloat32_Generated_Reversed(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	r := l.Reversed()
	r0 := r.Get(0)
	r2 := r.Get(2)
	if r0 != 3.0 || r2 != 1.0 {
		t.Errorf("Reversed: %v", r.ToSlice())
	}
}

func TestFloat32_Generated_Distinct(t *testing.T) {
	l := Float32Of(1.0, 2.0, 1.0, 3.0, 2.0)
	d := l.Distinct()
	if d.Len() != 3 {
		t.Errorf("Distinct size = %d, want 3", d.Len())
	}
}

func TestFloat32_Generated_ToSlice(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	s := l.ToSlice()
	if len(s) != 3 || s[0] != 1.0 || s[1] != 2.0 || s[2] != 3.0 {
		t.Errorf("ToSlice = %v", s)
	}
}

func TestFloat32_Generated_With(t *testing.T) {
	l := Float32Of(1.0)
	l2 := l.AddReturning(2.0)
	if l2.Len() != 2 {
		t.Errorf("With: Size = %d, want 2", l2.Len())
	}
}

func TestFloat32_Generated_Without(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	l2 := l.RemoveReturning(2.0)
	if l2.Contains(2.0) {
		t.Error("Without: should not contain removed value")
	}
}

func TestFloat32_Generated_Equals(t *testing.T) {
	l1 := Float32Of(1.0, 2.0, 3.0)
	l2 := Float32Of(1.0, 2.0, 3.0)
	l3 := Float32Of(1.0, 2.0)
	if !l1.Equals(l2) {
		t.Error("Equal lists should be equal")
	}
	if l1.Equals(l3) {
		t.Error("Different lists should not be equal")
	}
}

func TestFloat32_Generated_String(t *testing.T) {
	l := Float32Of(1.0, 2.0)
	s := l.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}

func TestFloat32_Generated_Resize(t *testing.T) {
	l := NewFloat32()
	for i := float32(0); i < 100; i++ {
		l.Add(i)
	}
	if l.Len() != 100 {
		t.Errorf("Size after 100 adds = %d", l.Len())
	}
	if v := l.Get(99); v != float32(99) {
		t.Errorf("Get(99) = %v, want 99", v)
	}
}

func TestFloat32_Generated_InjectInto(t *testing.T) {
	l := Float32Of(1.0, 2.0, 3.0)
	result := l.InjectInto(float32(0), func(acc, v float32) float32 { return acc + v })
	expected := float32(1.0) + float32(2.0) + float32(3.0)
	if result != expected {
		t.Errorf("InjectInto = %v, want %v", result, expected)
	}
}
