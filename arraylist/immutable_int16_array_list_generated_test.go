package arraylist

import "testing"

func TestImmutableInt16_Generated_GetAndSize(t *testing.T) {
	m := Int16Of(1, 2, 3)
	im := m.ToImmutable()
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
	}
	if v := im.Get(1); v != 2 {
		t.Errorf("Get(1) = %v", v)
	}
}
func TestImmutableInt16_Generated_Contains(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	if !im.Contains(2) {
		t.Error("Contains should be true")
	}
	if im.Contains(99) {
		t.Error("Contains should be false")
	}
}
func TestImmutableInt16_Generated_IndexOf(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	if idx := im.IndexOf(2); idx != 1 {
		t.Errorf("IndexOf = %d, want 1", idx)
	}
	if idx := im.IndexOf(99); idx != -1 {
		t.Errorf("IndexOf missing = %d, want -1", idx)
	}
}
func TestImmutableInt16_Generated_IsEmpty(t *testing.T) {
	im := NewInt16().ToImmutable()
	if im.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestImmutableInt16_Generated_All(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
func TestImmutableInt16_Generated_Select(t *testing.T) {
	im := Int16Of(1, 2, 3, 4, 5).ToImmutable()
	sel := im.Select(func(v int16) bool { return v > 3 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Len())
	}
}
func TestImmutableInt16_Generated_Reject(t *testing.T) {
	im := Int16Of(1, 2, 3, 4, 5).ToImmutable()
	rej := im.Reject(func(v int16) bool { return v > 3 })
	if rej.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", rej.Len())
	}
}
func TestImmutableInt16_Generated_Detect(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	val, found := im.Detect(func(v int16) bool { return v == 2 })
	if !found || val != 2 {
		t.Errorf("Detect = (%v, %v)", val, found)
	}
}
func TestImmutableInt16_Generated_AnySatisfy(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	if !im.AnySatisfy(func(v int16) bool { return v == 2 }) {
		t.Error("Should be true")
	}
	if im.AnySatisfy(func(v int16) bool { return v == 99 }) {
		t.Error("Should be false")
	}
}
func TestImmutableInt16_Generated_Reversed(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	r := im.Reversed()
	v0 := r.Get(0)
	v2 := r.Get(2)
	if v0 != 3 || v2 != 1 {
		t.Errorf("Reversed wrong")
	}
}
func TestImmutableInt16_Generated_Distinct(t *testing.T) {
	im := Int16Of(1, 2, 1).ToImmutable()
	d := im.Distinct()
	if d.Len() != 2 {
		t.Errorf("Distinct size = %d, want 2", d.Len())
	}
}
func TestImmutableInt16_Generated_ToSlice(t *testing.T) {
	im := Int16Of(1, 2).ToImmutable()
	sl := im.ToSlice()
	if len(sl) != 2 {
		t.Errorf("ToSlice len = %d", len(sl))
	}
}
func TestImmutableInt16_Generated_Equals(t *testing.T) {
	im1 := Int16Of(1, 2).ToImmutable()
	im2 := Int16Of(1, 2).ToImmutable()
	im3 := Int16Of(1).ToImmutable()
	if !im1.Equals(im2) {
		t.Error("Equal lists should be equal")
	}
	if im1.Equals(im3) {
		t.Error("Different lists should not be equal")
	}
}
func TestImmutableInt16_Generated_ToMutable(t *testing.T) {
	im := Int16Of(1, 2).ToImmutable()
	m := im.ToMutable()
	m.Add(3)
	if m.Len() != 3 {
		t.Errorf("Mutable size = %d", m.Len())
	}
	if im.Len() != 2 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableInt16_Generated_String(t *testing.T) {
	im := Int16Of(1).ToImmutable()
	if im.String() == "" {
		t.Error("String empty")
	}
}
