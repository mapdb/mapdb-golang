
package tuple

import (
	"testing"
)

func TestObjectInt64Pair_Generated_OneTwo(t *testing.T) {
	p := NewObjectInt64Pair[string]("alice", 2)
	if p.One() != "alice" {
		t.Errorf("One = %q, want alice", p.One())
	}
	if !(p.Two() == 2) {
		t.Errorf("Two = %v, want 2", p.Two())
	}
}

func TestObjectInt64Pair_Generated_String(t *testing.T) {
	p := NewObjectInt64Pair[string]("bob", 1)
	s := p.String()
	if s == "" {
		t.Error("String empty")
	}
}

// TestObjectInt64Pair_Generated_UsableAsMapKey: Go generic type arguments can make
// a struct uncomparable. This test fails to compile if the underlying
// pair struct ever loses == semantics for a fully-concrete T (here: string).
func TestObjectInt64Pair_Generated_UsableAsMapKey(t *testing.T) {
	p1 := NewObjectInt64Pair[string]("a", 1)
	p2 := NewObjectInt64Pair[string]("a", 1)
	m := map[ObjectInt64Pair[string]]int{p1: 1}
	if m[p2] != 1 {
		t.Errorf("same-content pair should map to same key")
	}
}
