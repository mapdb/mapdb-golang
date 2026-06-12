package tuple

import "testing"

func TestFloat32Int16Pair_Generated_OneTwo(t *testing.T) {
	p := NewFloat32Int16Pair(1.0, 2)
	if p.One() != 1.0 {
		t.Errorf("One = %v", p.One())
	}
	if p.Two() != 2 {
		t.Errorf("Two = %v", p.Two())
	}
}
func TestFloat32Int16Pair_Generated_Equals(t *testing.T) {
	p1 := NewFloat32Int16Pair(1.0, 2)
	p2 := NewFloat32Int16Pair(1.0, 2)
	p3 := NewFloat32Int16Pair(2.0, 1)
	if !p1.Equals(p2) {
		t.Error("Equal pairs should be equal")
	}
	if p1.Equals(p3) {
		t.Error("Different pairs should not be equal")
	}
}
func TestFloat32Int16Pair_Generated_String(t *testing.T) {
	p := NewFloat32Int16Pair(1.0, 2)
	if p.String() == "" {
		t.Error("String empty")
	}
}
