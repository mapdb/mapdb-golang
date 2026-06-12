package bag

import "testing"

func TestImmutableHashFloat32_Generated_OccurrencesOf(t *testing.T) {
	im := HashFloat32Of(1.0, 1.0, 2.0).ToImmutable()
	if im.OccurrencesOf(1.0) != 2 {
		t.Errorf("OccurrencesOf = %d, want 2", im.OccurrencesOf(1.0))
	}
	if im.Len() != 3 {
		t.Errorf("Size = %d, want 3", im.Len())
	}
	if im.SizeDistinct() != 2 {
		t.Errorf("SizeDistinct = %d, want 2", im.SizeDistinct())
	}
}
func TestImmutableHashFloat32_Generated_Contains(t *testing.T) {
	im := HashFloat32Of(1.0, 2.0).ToImmutable()
	if !im.Contains(1.0) {
		t.Error("Contains should be true")
	}
}
func TestImmutableHashFloat32_Generated_IsEmpty(t *testing.T) {
	im := NewHashFloat32().ToImmutable()
	if im.Len() != 0 {
		t.Error("Should be empty")
	}
}
func TestImmutableHashFloat32_Generated_All(t *testing.T) {
	im := HashFloat32Of(1.0, 1.0, 2.0).ToImmutable()
	count := 0
	for range im.All() {
		count++
	}
	if count != 3 {
		t.Errorf("All count = %d, want 3", count)
	}
}
func TestImmutableHashFloat32_Generated_Select(t *testing.T) {
	im := HashFloat32Of(1.0, 2.0, 3.0).ToImmutable()
	sel := im.Select(func(v float32) bool { return v > 1.0 })
	if sel.Len() != 2 {
		t.Errorf("Select size = %d, want 2", sel.Len())
	}
}
func TestImmutableHashFloat32_Generated_ForEachWithOccurrences(t *testing.T) {
	im := HashFloat32Of(1.0, 1.0, 2.0).ToImmutable()
	total := 0
	im.ForEachWithOccurrences(func(v float32, c int) { total += c })
	if total != 3 {
		t.Errorf("Total = %d, want 3", total)
	}
}
func TestImmutableHashFloat32_Generated_Equals(t *testing.T) {
	im1 := HashFloat32Of(1.0, 1.0, 2.0).ToImmutable()
	im2 := HashFloat32Of(2.0, 1.0, 1.0).ToImmutable()
	if !im1.Equals(im2) {
		t.Error("Should be equal")
	}
}
func TestImmutableHashFloat32_Generated_ToMutable(t *testing.T) {
	im := HashFloat32Of(1.0).ToImmutable()
	m := im.ToMutable()
	m.Add(2.0)
	if im.Len() != 1 {
		t.Error("Immutable should not change")
	}
}
func TestImmutableHashFloat32_Generated_String(t *testing.T) {
	im := HashFloat32Of(1.0).ToImmutable()
	if im.String() == "" {
		t.Error("String empty")
	}
}
