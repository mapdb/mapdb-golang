package tuple

import (
	"testing"
)

func TestObjectInt8Pair_Generated_OneTwo(t *testing.T) {
	p := NewObjectInt8Pair[string]("alice", 2)
	if p.One() != "alice" {
		t.Errorf("One = %q, want alice", p.One())
	}
	if !(p.Two() == 2) {
		t.Errorf("Two = %v, want 2", p.Two())
	}
}

func TestObjectInt8Pair_Generated_String(t *testing.T) {
	p := NewObjectInt8Pair[string]("bob", 1)
	s := p.String()
	if s == "" {
		t.Error("String empty")
	}
}

// TestObjectInt8Pair_Generated_UsableAsMapKey: Go generic type arguments can make
// a struct uncomparable. This test fails to compile if the underlying
// pair struct ever loses == semantics for a fully-concrete T (here: string).
func TestObjectInt8Pair_Generated_UsableAsMapKey(t *testing.T) {
	p1 := NewObjectInt8Pair[string]("a", 1)
	p2 := NewObjectInt8Pair[string]("a", 1)
	m := map[ObjectInt8Pair[string]]int{p1: 1}
	if m[p2] != 1 {
		t.Errorf("same-content pair should map to same key")
	}
}
