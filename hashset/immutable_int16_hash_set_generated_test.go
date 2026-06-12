package hashset

import "testing"

func TestImmutableInt16_Generated_Contains(t *testing.T) {
	im := Int16Of(1, 2, 3).ToImmutable()
	if !im.Contains(2) {
		t.Error("Contains should be true")
	}
	if im.Contains(99) {
		t.Error("Contains should be false")
	}
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
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
func TestImmutableInt16_Generated_Union(t *testing.T) {
	a := Int16Of(1, 2, 3).ToImmutable()
	b := Int16Of(3, 4, 5).ToImmutable()
	u := a.Union(b)
	if u.Len() != 5 {
		t.Errorf("Union size = %d, want 5", u.Len())
	}
}
func TestImmutableInt16_Generated_Intersect(t *testing.T) {
	a := Int16Of(1, 2, 3).ToImmutable()
	b := Int16Of(2, 3, 4).ToImmutable()
	i := a.Intersect(b)
	if i.Len() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Len())
	}
}
func TestImmutableInt16_Generated_Difference(t *testing.T) {
	a := Int16Of(1, 2, 3).ToImmutable()
	b := Int16Of(2, 3, 4).ToImmutable()
	d := a.Difference(b)
	if d.Len() != 1 {
		t.Errorf("Difference size = %d, want 1", d.Len())
	}
}
func TestImmutableInt16_Generated_ToSlice(t *testing.T) {
	im := Int16Of(1, 2).ToImmutable()
	if len(im.ToSlice()) != 2 {
		t.Error("ToSlice wrong len")
	}
}
func TestImmutableInt16_Generated_Equals(t *testing.T) {
	im1 := Int16Of(1, 2).ToImmutable()
	im2 := Int16Of(2, 1).ToImmutable()
	if !im1.Equals(im2) {
		t.Error("Should be equal")
	}
}
func TestImmutableInt16_Generated_ToMutable(t *testing.T) {
	im := Int16Of(1, 2).ToImmutable()
	m := im.ToMutable()
	m.Add(3)
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
