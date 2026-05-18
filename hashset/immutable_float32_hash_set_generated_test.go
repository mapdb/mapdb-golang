
package hashset

import "testing"

func TestImmutableFloat32HashSet_Generated_Contains(t *testing.T) {
	im := Float32HashSetOf(1.0, 2.0, 3.0).ToImmutable()
	if !im.Contains(2.0) {
		t.Error("Contains should be true")
	}
	if im.Contains(99.0) {
		t.Error("Contains should be false")
	}
	if im.Size() != 3 {
		t.Errorf("Size = %d, want 3", im.Size())
	}
}
func TestImmutableFloat32HashSet_Generated_IsEmpty(t *testing.T) {
	im := NewFloat32HashSet().ToImmutable()
	if !im.IsEmpty() {
		t.Error("Should be empty")
	}
}
func TestImmutableFloat32HashSet_Generated_All(t *testing.T) {
	im := Float32HashSetOf(1.0, 2.0, 3.0).ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
func TestImmutableFloat32HashSet_Generated_Select(t *testing.T) {
	im := Float32HashSetOf(1.0, 2.0, 3.0, 4.0, 5.0).ToImmutable()
	sel := im.Select(func(v float32) bool { return v > 3.0 })
	if sel.Size() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Size())
	}
}
func TestImmutableFloat32HashSet_Generated_Union(t *testing.T) {
	a := Float32HashSetOf(1.0, 2.0, 3.0).ToImmutable()
	b := Float32HashSetOf(3.0, 4.0, 5.0).ToImmutable()
	u := a.Union(b)
	if u.Size() != 5 {
		t.Errorf("Union size = %d, want 5", u.Size())
	}
}
func TestImmutableFloat32HashSet_Generated_Intersect(t *testing.T) {
	a := Float32HashSetOf(1.0, 2.0, 3.0).ToImmutable()
	b := Float32HashSetOf(2.0, 3.0, 4.0).ToImmutable()
	i := a.Intersect(b)
	if i.Size() != 2 {
		t.Errorf("Intersect size = %d, want 2", i.Size())
	}
}
func TestImmutableFloat32HashSet_Generated_Difference(t *testing.T) {
	a := Float32HashSetOf(1.0, 2.0, 3.0).ToImmutable()
	b := Float32HashSetOf(2.0, 3.0, 4.0).ToImmutable()
	d := a.Difference(b)
	if d.Size() != 1 {
		t.Errorf("Difference size = %d, want 1", d.Size())
	}
}
func TestImmutableFloat32HashSet_Generated_ToSlice(t *testing.T) {
	im := Float32HashSetOf(1.0, 2.0).ToImmutable()
	if len(im.ToSlice()) != 2 {
		t.Error("ToSlice wrong len")
	}
}
func TestImmutableFloat32HashSet_Generated_Equals(t *testing.T) {
	im1 := Float32HashSetOf(1.0, 2.0).ToImmutable()
	im2 := Float32HashSetOf(2.0, 1.0).ToImmutable()
	if !im1.Equals(im2) {
		t.Error("Should be equal")
	}
}
func TestImmutableFloat32HashSet_Generated_ToMutable(t *testing.T) {
	im := Float32HashSetOf(1.0, 2.0).ToImmutable()
	m := im.ToMutable()
	m.Add(3.0)
	if im.Size() != 2 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableFloat32HashSet_Generated_String(t *testing.T) {
	im := Float32HashSetOf(1.0).ToImmutable()
	if im.String() == "" {
		t.Error("String empty")
	}
}
