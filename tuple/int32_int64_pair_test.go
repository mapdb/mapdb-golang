package tuple

import (
	"testing"
)

func TestInt32Int64Pair_Basic(t *testing.T) {
	p := NewInt32Int64Pair(42, 100)
	if p.One() != 42 {
		t.Errorf("One = %d, want 42", p.One())
	}
	if p.Two() != 100 {
		t.Errorf("Two = %d, want 100", p.Two())
	}
}

func TestInt32Int64Pair_Equals(t *testing.T) {
	a := NewInt32Int64Pair(1, 2)
	b := NewInt32Int64Pair(1, 2)
	c := NewInt32Int64Pair(1, 3)
	if !a.Equals(b) {
		t.Error("Equal pairs should be equal")
	}
	if a.Equals(c) {
		t.Error("Different pairs should not be equal")
	}
}

func TestInt32Int64Pair_CompareTo(t *testing.T) {
	a := NewInt32Int64Pair(1, 10)
	b := NewInt32Int64Pair(2, 5)
	c := NewInt32Int64Pair(1, 20)
	if a.CompareTo(b) >= 0 {
		t.Error("(1,10) should be < (2,5)")
	}
	if a.CompareTo(c) >= 0 {
		t.Error("(1,10) should be < (1,20)")
	}
	if a.CompareTo(a) != 0 {
		t.Error("pair should equal itself")
	}
}

func TestInt32Int64Pair_Swap(t *testing.T) {
	p := NewInt32Int64Pair(1, 2)
	swapped := p.Swap()
	if swapped.One() != 2 || swapped.Two() != 1 {
		t.Errorf("Swap = (%d, %d), want (2, 1)", swapped.One(), swapped.Two())
	}
}

func TestInt32Int64Pair_String(t *testing.T) {
	p := NewInt32Int64Pair(42, 100)
	if s := p.String(); s != "(42, 100)" {
		t.Errorf("String = %q, want (42, 100)", s)
	}
}

func TestObjectInt32Pair(t *testing.T) {
	p := NewObjectInt32Pair[string]("hello", 42)
	if p.One() != "hello" {
		t.Errorf("One = %q, want hello", p.One())
	}
	if p.Two() != 42 {
		t.Errorf("Two = %d, want 42", p.Two())
	}
	if s := p.String(); s != "(hello, 42)" {
		t.Errorf("String = %q, want (hello, 42)", s)
	}
}

func TestInt32ObjectPair(t *testing.T) {
	p := NewInt32ObjectPair[string](42, "world")
	if p.One() != 42 {
		t.Errorf("One = %d, want 42", p.One())
	}
	if p.Two() != "world" {
		t.Errorf("Two = %q, want world", p.Two())
	}
}
