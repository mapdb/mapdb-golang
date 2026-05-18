
package tuple

import "testing"

func TestFloat64Float64Pair_Generated_OneTwo(t *testing.T) {
	p := NewFloat64Float64Pair(1.0, 2.0)
	if p.One() != 1.0 {
		t.Errorf("One = %v", p.One())
	}
	if p.Two() != 2.0 {
		t.Errorf("Two = %v", p.Two())
	}
}
func TestFloat64Float64Pair_Generated_Equals(t *testing.T) {
	p1 := NewFloat64Float64Pair(1.0, 2.0)
	p2 := NewFloat64Float64Pair(1.0, 2.0)
	p3 := NewFloat64Float64Pair(2.0, 1.0)
	if !p1.Equals(p2) {
		t.Error("Equal pairs should be equal")
	}
	if p1.Equals(p3) {
		t.Error("Different pairs should not be equal")
	}
}
func TestFloat64Float64Pair_Generated_String(t *testing.T) {
	p := NewFloat64Float64Pair(1.0, 2.0)
	if p.String() == "" {
		t.Error("String empty")
	}
}
