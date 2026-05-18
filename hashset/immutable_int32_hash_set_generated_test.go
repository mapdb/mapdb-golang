
package hashset

import "testing"

func TestImmutableInt32HashSet_Generated_Contains(t *testing.T) {
	im := Int32HashSetOf(1, 2, 3).ToImmutable()
	if !im.Contains(2) {
		t.Error("Contains should be true")
	}
	if im.Contains(99) {
		t.Error("Contains should be false")
	}
	if im.Size() != 3 {
		t.Errorf("Size = %d, want 3", im.Size())
	}
}
func TestImmutableInt32HashSet_Generated_IsEmpty(t *testing.T) {
	im := NewInt32HashSet().ToImmutable()
	if !im.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestImmutableInt32HashSet_Generated_All(t *testing.T) {
	im := Int32HashSetOf(1, 2, 3).ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
func TestImmutableInt32HashSet_Generated_Select(t *testing.T) {
	im := Int32HashSetOf(1, 2, 3, 4, 5).ToImmutable()
	sel := im.Select(func(v int32) bool { return v > 3 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Size())
	}
}
func TestImmutableInt32HashSet_Generated_Union(t *testing.T) {
	a := Int32HashSetOf(1, 2, 3).ToImmutable()
	b := Int32HashSetOf(3, 4, 5).ToImmutable()
	u := a.Union(b)
	if u.Size() != 5 {
		t.Errorf("Union size = %d, want 5", u.Size())
	}
}
func TestImmutableInt32HashSet_Generated_Intersect(t *testing.T) {
	a := Int32HashSetOf(1, 2, 3).ToImmutable()
	b := Int32HashSetOf(2, 3, 4).ToImmutable()
	i := a.Intersect(b)
	if i.Size() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Size())
	}
}
func TestImmutableInt32HashSet_Generated_Difference(t *testing.T) {
	a := Int32HashSetOf(1, 2, 3).ToImmutable()
	b := Int32HashSetOf(2, 3, 4).ToImmutable()
	d := a.Difference(b)
	if d.Size() != 1 {
		t.Errorf("Difference size = %d, want 1", d.Size())
	}
}
func TestImmutableInt32HashSet_Generated_ToSlice(t *testing.T) {
	im := Int32HashSetOf(1, 2).ToImmutable()
	if len(im.ToSlice()) != 2 {
		t.Error("ToSlice wrong len")
	}
}
func TestImmutableInt32HashSet_Generated_Equals(t *testing.T) {
	im1 := Int32HashSetOf(1, 2).ToImmutable()
	im2 := Int32HashSetOf(2, 1).ToImmutable()
	if !im1.Equals(im2) {
		t.Error("Should be equal")
	}
}
func TestImmutableInt32HashSet_Generated_ToMutable(t *testing.T) {
	im := Int32HashSetOf(1, 2).ToImmutable()
	m := im.ToMutable()
	m.Add(3)
	if im.Size() != 2 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableInt32HashSet_Generated_String(t *testing.T) {
	im := Int32HashSetOf(1).ToImmutable()
	if im.String() == "" {
		t.Error("String empty")
	}
}
