
package tuple

import (
	"testing"
)

func TestInt8ObjectPair_Generated_OneTwo(t *testing.T) {
	p := NewInt8ObjectPair[string](1, "alice")
	if !(p.One() == 1) {
		t.Errorf("One = %v, want 1", p.One())
	}
	if p.Two() != "alice" {
		t.Errorf("Two = %q, want alice", p.Two())
	}
}

func TestInt8ObjectPair_Generated_String(t *testing.T) {
	p := NewInt8ObjectPair[string](2, "bob")
	if p.String() == "" {
		t.Error("String empty")
	}
}

func TestInt8ObjectPair_Generated_UsableAsMapKey(t *testing.T) {
	p1 := NewInt8ObjectPair[string](1, "a")
	p2 := NewInt8ObjectPair[string](1, "a")
	m := map[Int8ObjectPair[string]]int{p1: 1}
	if m[p2] != 1 {
		t.Errorf("same-content pair should map to same key")
	}
}
