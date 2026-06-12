package arraylist

import (
	"testing"
)

func TestInt8_Generated_AddAndGet(t *testing.T) {
	l := NewInt8()
	l.Add(1)
	l.Add(2)
	l.Add(3)

	if l.Len() != 3 {
		t.Errorf("Size() = %d, want 3", l.Len())
	}
	if v := l.Get(0); v != 1 {
		t.Errorf("Get(0) = %v, want 1", v)
	}
	if v := l.Get(2); v != 3 {
		t.Errorf("Get(2) = %v, want 3", v)
	}
}

func TestInt8_Generated_GetOutOfBounds(t *testing.T) {
	l := Int8Of(1, 2)
	assertPanics(t, func() { l.Get(-1) })
	assertPanics(t, func() { l.Get(2) })
}

func TestInt8_Generated_Of(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if l.Len() != 3 {
		t.Errorf("Of: Size() = %d, want 3", l.Len())
	}
	if v := l.Get(1); v != 2 {
		t.Errorf("Of: Get(1) = %v, want 2", v)
	}
}

func TestInt8_Generated_Set(t *testing.T) {
	l := Int8Of(1, 2, 3)
	old := l.Set(1, 4)
	if old != 2 {
		t.Errorf("Set returned %v, want 2", old)
	}
	if v := l.Get(1); v != 4 {
		t.Errorf("After Set: Get(1) = %v, want 4", v)
	}
	assertPanics(t, func() { l.Set(99, 1) })
}

func TestInt8_Generated_RemoveAtIndex(t *testing.T) {
	l := Int8Of(1, 2, 3)
	removed := l.RemoveAtIndex(1)
	if removed != 2 {
		t.Errorf("RemoveAtIndex(1) = %v, want 2", removed)
	}
	if l.Len() != 2 {
		t.Errorf("Size after remove = %d, want 2", l.Len())
	}
	assertPanics(t, func() { l.RemoveAtIndex(99) })
}

func TestInt8_Generated_Remove(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if !l.Remove(2) {
		t.Error("Remove(2) should return true")
	}
	if l.Contains(2) {
		t.Error("Should not contain 2 after remove")
	}
	if l.Remove(99) {
		t.Error("Remove(99) should return false")
	}
}

func TestInt8_Generated_Contains(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if !l.Contains(2) {
		t.Error("Contains(2) should be true")
	}
	if l.Contains(99) {
		t.Error("Contains(99) should be false")
	}
}

func TestInt8_Generated_IndexOf(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if idx := l.IndexOf(2); idx != 1 {
		t.Errorf("IndexOf(2) = %d, want 1", idx)
	}
	if idx := l.IndexOf(99); idx != -1 {
		t.Errorf("IndexOf(99) = %d, want -1", idx)
	}
}

func TestInt8_Generated_IsEmpty(t *testing.T) {
	l := NewInt8()
	if l.Len() != 0 {
		t.Error("New list should be empty")
	}
	l.Add(1)
	if l.Len() == 0 {
		t.Error("List with element should not be empty")
	}
}

func TestInt8_Generated_Clear(t *testing.T) {
	l := Int8Of(1, 2, 3)
	l.Clear()
	if l.Len() != 0 {
		t.Errorf("After Clear: size=%d, empty=%v", l.Len(), l.Len() == 0)
	}
}

func TestInt8_Generated_AddAll(t *testing.T) {
	l := NewInt8()
	l.AddAll(1, 2, 3)
	if l.Len() != 3 {
		t.Errorf("AddAll: Size() = %d, want 3", l.Len())
	}
}

func TestInt8_Generated_All(t *testing.T) {
	l := Int8Of(1, 2, 3)
	sum := int8(0)
	for v := range l.All() {
		sum += v
	}
	expected := int8(1) + int8(2) + int8(3)
	if sum != expected {
		t.Errorf("All sum = %v, want %v", sum, expected)
	}
}

func TestInt8_Generated_AllWithIndex(t *testing.T) {
	l := Int8Of(1, 2, 3)
	indices := 0
	for i, _ := range l.AllWithIndex() {
		indices += i
	}
	if indices != 3 { // 0+1+2 = 3
		t.Errorf("AllWithIndex index sum = %d, want 3", indices)
	}
}

func TestInt8_Generated_Select(t *testing.T) {
	l := Int8Of(1, 2, 3, 4, 5)
	selected := l.Select(func(v int8) bool { return v > 3 })
	if selected.Len() != 2 {
		t.Errorf("Select size = %d, want 2", selected.Len())
	}
}

func TestInt8_Generated_Reject(t *testing.T) {
	l := Int8Of(1, 2, 3, 4, 5)
	rejected := l.Reject(func(v int8) bool { return v > 3 })
	if rejected.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", rejected.Len())
	}
}

func TestInt8_Generated_Detect(t *testing.T) {
	l := Int8Of(1, 2, 3)
	val, found := l.Detect(func(v int8) bool { return v == 2 })
	if !found || val != 2 {
		t.Errorf("Detect = (%v, %v), want (2, true)", val, found)
	}
	_, found = l.Detect(func(v int8) bool { return v == 99 })
	if found {
		t.Error("Detect for missing should return false")
	}
}

func TestInt8_Generated_AnySatisfy(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if !l.AnySatisfy(func(v int8) bool { return v == 2 }) {
		t.Error("AnySatisfy should be true")
	}
	if l.AnySatisfy(func(v int8) bool { return v == 99 }) {
		t.Error("AnySatisfy should be false")
	}
}

func TestInt8_Generated_AllSatisfy(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if !l.AllSatisfy(func(v int8) bool { return v > 0 }) {
		t.Error("AllSatisfy should be true for > 0")
	}
	if l.AllSatisfy(func(v int8) bool { return v > 2 }) {
		t.Error("AllSatisfy should be false")
	}
}

func TestInt8_Generated_NoneSatisfy(t *testing.T) {
	l := Int8Of(1, 2, 3)
	if !l.NoneSatisfy(func(v int8) bool { return v == 99 }) {
		t.Error("NoneSatisfy should be true for missing")
	}
	if l.NoneSatisfy(func(v int8) bool { return v == 1 }) {
		t.Error("NoneSatisfy should be false")
	}
}

func TestInt8_Generated_Count(t *testing.T) {
	l := Int8Of(1, 2, 3, 4, 5)
	if c := l.Count(func(v int8) bool { return v > 3 }); c != 2 {
		t.Errorf("Count = %d, want 2", c)
	}
}

func TestInt8_Generated_Sum(t *testing.T) {
	l := Int8Of(1, 2, 3)
	expected := int64(1) + int64(2) + int64(3)
	if s := l.Sum(); s != expected {
		t.Errorf("Sum = %v, want %v", s, expected)
	}
}

func TestInt8_Generated_MinMax(t *testing.T) {
	l := Int8Of(3, 1, 2)
	if min, ok := l.Min(); !ok || min != 1 {
		t.Errorf("Min = (%v, %v), want (1, true)", min, ok)
	}
	if max, ok := l.Max(); !ok || max != 3 {
		t.Errorf("Max = (%v, %v), want (3, true)", max, ok)
	}

	empty := NewInt8()
	if _, ok := empty.Min(); ok {
		t.Error("Min on empty should return false")
	}
	if _, ok := empty.Max(); ok {
		t.Error("Max on empty should return false")
	}
}

func TestInt8_Generated_Sort(t *testing.T) {
	l := Int8Of(3, 1, 2)
	l.Sort()
	v0 := l.Get(0)
	v1 := l.Get(1)
	v2 := l.Get(2)
	if v0 != 1 || v1 != 2 || v2 != 3 {
		t.Errorf("After sort: %v", l.ToSlice())
	}
}

func TestInt8_Generated_Reversed(t *testing.T) {
	l := Int8Of(1, 2, 3)
	r := l.Reversed()
	r0 := r.Get(0)
	r2 := r.Get(2)
	if r0 != 3 || r2 != 1 {
		t.Errorf("Reversed: %v", r.ToSlice())
	}
}

func TestInt8_Generated_Distinct(t *testing.T) {
	l := Int8Of(1, 2, 1, 3, 2)
	d := l.Distinct()
	if d.Len() != 3 {
		t.Errorf("Distinct size = %d, want 3", d.Len())
	}
}

func TestInt8_Generated_ToSlice(t *testing.T) {
	l := Int8Of(1, 2, 3)
	s := l.ToSlice()
	if len(s) != 3 || s[0] != 1 || s[1] != 2 || s[2] != 3 {
		t.Errorf("ToSlice = %v", s)
	}
}

func TestInt8_Generated_With(t *testing.T) {
	l := Int8Of(1)
	l2 := l.AddReturning(2)
	if l2.Len() != 2 {
		t.Errorf("With: Size = %d, want 2", l2.Len())
	}
}

func TestInt8_Generated_Without(t *testing.T) {
	l := Int8Of(1, 2, 3)
	l2 := l.RemoveReturning(2)
	if l2.Contains(2) {
		t.Error("Without: should not contain removed value")
	}
}

func TestInt8_Generated_Equals(t *testing.T) {
	l1 := Int8Of(1, 2, 3)
	l2 := Int8Of(1, 2, 3)
	l3 := Int8Of(1, 2)
	if !l1.Equals(l2) {
		t.Error("Equal lists should be equal")
	}
	if l1.Equals(l3) {
		t.Error("Different lists should not be equal")
	}
}

func TestInt8_Generated_String(t *testing.T) {
	l := Int8Of(1, 2)
	s := l.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}

func TestInt8_Generated_Resize(t *testing.T) {
	l := NewInt8()
	for i := int8(0); i < 100; i++ {
		l.Add(i)
	}
	if l.Len() != 100 {
		t.Errorf("Size after 100 adds = %d", l.Len())
	}
	if v := l.Get(99); v != int8(99) {
		t.Errorf("Get(99) = %v, want 99", v)
	}
}

func TestInt8_Generated_InjectInto(t *testing.T) {
	l := Int8Of(1, 2, 3)
	result := l.InjectInto(int8(0), func(acc, v int8) int8 { return acc + v })
	expected := int8(1) + int8(2) + int8(3)
	if result != expected {
		t.Errorf("InjectInto = %v, want %v", result, expected)
	}
}
