
package tuple

import "testing"

func TestCharCharPair_Generated_OneTwo(t *testing.T) {
	p := NewCharCharPair(1, 2)
	if p.One() != 1 {
		t.Errorf("One = %v", p.One())
	}
	if p.Two() != 2 {
		t.Errorf("Two = %v", p.Two())
	}
}
func TestCharCharPair_Generated_Equals(t *testing.T) {
	p1 := NewCharCharPair(1, 2)
	p2 := NewCharCharPair(1, 2)
	p3 := NewCharCharPair(2, 1)
	if !p1.Equals(p2) {
		t.Error("Equal pairs should be equal")
	}
	if p1.Equals(p3) {
		t.Error("Different pairs should not be equal")
	}
}
func TestCharCharPair_Generated_String(t *testing.T) {
	p := NewCharCharPair(1, 2)
	if p.String() == "" {
		t.Error("String empty")
	}
}
