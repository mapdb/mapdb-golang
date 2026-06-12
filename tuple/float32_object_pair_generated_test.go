package tuple

import (
	"math"
	"testing"
)

func TestFloat32ObjectPair_Generated_OneTwo(t *testing.T) {
	p := NewFloat32ObjectPair[string](1.0, "alice")
	if !(math.Float32bits(p.One()) == math.Float32bits(1.0)) {
		t.Errorf("One = %v, want 1.0", p.One())
	}
	if p.Two() != "alice" {
		t.Errorf("Two = %q, want alice", p.Two())
	}
}

func TestFloat32ObjectPair_Generated_String(t *testing.T) {
	p := NewFloat32ObjectPair[string](2.0, "bob")
	if p.String() == "" {
		t.Error("String empty")
	}
}

func TestFloat32ObjectPair_Generated_UsableAsMapKey(t *testing.T) {
	p1 := NewFloat32ObjectPair[string](1.0, "a")
	p2 := NewFloat32ObjectPair[string](1.0, "a")
	m := map[Float32ObjectPair[string]]int{p1: 1}
	if m[p2] != 1 {
		t.Errorf("same-content pair should map to same key")
	}
}
