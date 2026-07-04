package countmin

import "testing"

// Regression for M-3 (todo/fable-golang/01-critical-bugs.md): NewCountMinOptimal
// checked only isFiniteGE1(w) before uint32(w). For tiny epsilon, w = ceil(e/eps)
// exceeds MaxUint32, and float→uint32 conversion of an out-of-range value is
// implementation-defined per the Go spec (arbitrary width instead of a clear
// panic). It now rejects the out-of-range width explicitly, mirroring bloom.Optimal.
func TestNewCountMinOptimal_TinyEpsilonPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewCountMinOptimal(1e-10, 0.01) should panic (w exceeds uint32), got no panic")
		}
	}()
	NewCountMinOptimal(1e-10, 0.01) // w ≈ 2.7e10 > MaxUint32
}

// A normal epsilon/delta still constructs fine.
func TestNewCountMinOptimal_NormalOK(t *testing.T) {
	c := NewCountMinOptimal(0.01, 0.01)
	if c.Width() == 0 || c.Depth() == 0 {
		t.Fatalf("degenerate sketch: w=%d d=%d", c.Width(), c.Depth())
	}
}
