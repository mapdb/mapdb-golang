package tuple

import "testing"

func TestCharFloat32Pair_Generated_OneTwo(t *testing.T) {
	p := NewCharFloat32Pair(1, 2.0)
	if p.One() != 1 {
		t.Errorf("One = %v", p.One())
	}
	if p.Two() != 2.0 {
		t.Errorf("Two = %v", p.Two())
	}
}
func TestCharFloat32Pair_Generated_Equals(t *testing.T) {
	p1 := NewCharFloat32Pair(1, 2.0)
	p2 := NewCharFloat32Pair(1, 2.0)
	p3 := NewCharFloat32Pair(2, 1.0)
	if !p1.Equals(p2) {
		t.Error("Equal pairs should be equal")
	}
	if p1.Equals(p3) {
		t.Error("Different pairs should not be equal")
	}
}
func TestCharFloat32Pair_Generated_String(t *testing.T) {
	p := NewCharFloat32Pair(1, 2.0)
	if p.String() == "" {
		t.Error("String empty")
	}
}
