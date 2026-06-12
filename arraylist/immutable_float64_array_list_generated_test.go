package arraylist

import "testing"

func TestImmutableFloat64_Generated_GetAndSize(t *testing.T) {
	m := Float64Of(1.0, 2.0, 3.0)
	im := m.ToImmutable()
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
	}
	if v := im.Get(1); v != 2.0 {
		t.Errorf("Get(1) = %v", v)
	}
}
func TestImmutableFloat64_Generated_Contains(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0).ToImmutable()
	if !im.Contains(2.0) {
		t.Error("Contains should be true")
	}
	if im.Contains(99.0) {
		t.Error("Contains should be false")
	}
}
func TestImmutableFloat64_Generated_IndexOf(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0).ToImmutable()
	if idx := im.IndexOf(2.0); idx != 1 {
		t.Errorf("IndexOf = %d, want 1", idx)
	}
	if idx := im.IndexOf(99.0); idx != -1 {
		t.Errorf("IndexOf missing = %d, want -1", idx)
	}
}
func TestImmutableFloat64_Generated_IsEmpty(t *testing.T) {
	im := NewFloat64().ToImmutable()
	if im.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestImmutableFloat64_Generated_All(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0).ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
func TestImmutableFloat64_Generated_Select(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0, 4.0, 5.0).ToImmutable()
	sel := im.Select(func(v float64) bool { return v > 3.0 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Len())
	}
}
func TestImmutableFloat64_Generated_Reject(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0, 4.0, 5.0).ToImmutable()
	rej := im.Reject(func(v float64) bool { return v > 3.0 })
	if rej.Len() != 3 {
		t.Errorf("Reject size = %d, want 3", rej.Len())
	}
}
func TestImmutableFloat64_Generated_Detect(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0).ToImmutable()
	val, found := im.Detect(func(v float64) bool { return v == 2.0 })
	if !found || val != 2.0 {
		t.Errorf("Detect = (%v, %v)", val, found)
	}
}
func TestImmutableFloat64_Generated_AnySatisfy(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0).ToImmutable()
	if !im.AnySatisfy(func(v float64) bool { return v == 2.0 }) {
		t.Error("Should be true")
	}
	if im.AnySatisfy(func(v float64) bool { return v == 99.0 }) {
		t.Error("Should be false")
	}
}
func TestImmutableFloat64_Generated_Reversed(t *testing.T) {
	im := Float64Of(1.0, 2.0, 3.0).ToImmutable()
	r := im.Reversed()
	v0 := r.Get(0)
	v2 := r.Get(2)
	if v0 != 3.0 || v2 != 1.0 {
		t.Errorf("Reversed wrong")
	}
}
func TestImmutableFloat64_Generated_Distinct(t *testing.T) {
	im := Float64Of(1.0, 2.0, 1.0).ToImmutable()
	d := im.Distinct()
	if d.Len() != 2 {
		t.Errorf("Distinct size = %d, want 2", d.Len())
	}
}
func TestImmutableFloat64_Generated_ToSlice(t *testing.T) {
	im := Float64Of(1.0, 2.0).ToImmutable()
	sl := im.ToSlice()
	if len(sl) != 2 {
		t.Errorf("ToSlice len = %d", len(sl))
	}
}
func TestImmutableFloat64_Generated_Equals(t *testing.T) {
	im1 := Float64Of(1.0, 2.0).ToImmutable()
	im2 := Float64Of(1.0, 2.0).ToImmutable()
	im3 := Float64Of(1.0).ToImmutable()
	if !im1.Equals(im2) {
		t.Error("Equal lists should be equal")
	}
	if im1.Equals(im3) {
		t.Error("Different lists should not be equal")
	}
}
func TestImmutableFloat64_Generated_ToMutable(t *testing.T) {
	im := Float64Of(1.0, 2.0).ToImmutable()
	m := im.ToMutable()
	m.Add(3.0)
	if m.Len() != 3 {
		t.Errorf("Mutable size = %d", m.Len())
	}
	if im.Len() != 2 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableFloat64_Generated_String(t *testing.T) {
	im := Float64Of(1.0).ToImmutable()
	if im.String() == "" {
		t.Error("String empty")
	}
}
