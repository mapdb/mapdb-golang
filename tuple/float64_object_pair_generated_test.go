package tuple

import (
	"math"
	"testing"
)

func TestFloat64ObjectPair_Generated_OneTwo(t *testing.T) {
	p := NewFloat64ObjectPair[string](1.0, "alice")
	if !(math.Float64bits(p.One()) == math.Float64bits(1.0)) {
		t.Errorf("One = %v, want 1.0", p.One())
	}
	if p.Two() != "alice" {
		t.Errorf("Two = %q, want alice", p.Two())
	}
}

func TestFloat64ObjectPair_Generated_String(t *testing.T) {
	p := NewFloat64ObjectPair[string](2.0, "bob")
	if p.String() == "" {
		t.Error("String empty")
	}
}

func TestFloat64ObjectPair_Generated_UsableAsMapKey(t *testing.T) {
	p1 := NewFloat64ObjectPair[string](1.0, "a")
	p2 := NewFloat64ObjectPair[string](1.0, "a")
	m := map[Float64ObjectPair[string]]int{p1: 1}
	if m[p2] != 1 {
		t.Errorf("same-content pair should map to same key")
	}
}
