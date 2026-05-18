
package tuple

import (
	"math"
	"testing"
)

func TestObjectFloat64Pair_Generated_OneTwo(t *testing.T) {
	p := NewObjectFloat64Pair[string]("alice", 2.0)
	if p.One() != "alice" {
		t.Errorf("One = %q, want alice", p.One())
	}
	if !(math.Float64bits(p.Two()) == math.Float64bits(2.0)) {
		t.Errorf("Two = %v, want 2.0", p.Two())
	}
}

func TestObjectFloat64Pair_Generated_String(t *testing.T) {
	p := NewObjectFloat64Pair[string]("bob", 1.0)
	s := p.String()
	if s == "" {
		t.Error("String empty")
	}
}

// TestObjectFloat64Pair_Generated_UsableAsMapKey: Go generic type arguments can make
// a struct uncomparable. This test fails to compile if the underlying
// pair struct ever loses == semantics for a fully-concrete T (here: string).
func TestObjectFloat64Pair_Generated_UsableAsMapKey(t *testing.T) {
	p1 := NewObjectFloat64Pair[string]("a", 1.0)
	p2 := NewObjectFloat64Pair[string]("a", 1.0)
	m := map[ObjectFloat64Pair[string]]int{p1: 1}
	if m[p2] != 1 {
		t.Errorf("same-content pair should map to same key")
	}
}
