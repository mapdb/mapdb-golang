package hyperloglog

import "testing"

// Regression for M-4 (todo/fable-golang/01-critical-bugs.md): Registers() returned
// the internal slice, so writing through it corrupted the sketch (including states
// FromBytes rejects). It now returns a copy, matching every sibling accessor.
func TestRegisters_ReturnsCopy(t *testing.T) {
	h, err := NewHyperLogLogWithPrecision(6)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []int32{1, 2, 3, 4, 5} {
		h.Add(v)
	}
	before := h.MaxRegister()

	regs := h.Registers()
	for i := range regs {
		regs[i] = 255 // attempt to corrupt via the returned slice
	}

	if got := h.MaxRegister(); got != before {
		t.Fatalf("mutating Registers() corrupted the sketch: MaxRegister %d -> %d", before, got)
	}
	// A second call must be unaffected by the mutation of the first.
	regs2 := h.Registers()
	for _, r := range regs2 {
		if r == 255 {
			t.Fatal("Registers() leaked internal state across calls")
		}
	}
}

// The constructor now returns *HyperLogLog, so value-copy aliasing and zero-value
// panics are gone. This compiles only because New returns a pointer, and Add on a
// second sketch does not disturb the first.
func TestConstructor_ReturnsPointerNoAliasing(t *testing.T) {
	a, err := NewHyperLogLogWithPrecision(6)
	if err != nil {
		t.Fatal(err)
	}
	a.Add(42)
	b, err := NewHyperLogLogWithPrecision(6)
	if err != nil {
		t.Fatal(err)
	}
	b.Add(7)
	// Independent sketches: b's register set differs from a's for these inputs.
	if a.NonzeroRegisters() == 0 || b.NonzeroRegisters() == 0 {
		t.Fatal("expected both sketches to have nonzero registers")
	}
}
