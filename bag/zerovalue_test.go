package bag

import "testing"

// Zero-value (nil-map) usability: a freshly declared map-backed bag must accept
// mutations without panicking ("assignment to entry in nil map"). Phase 7a.

func TestZeroValueHashInt32(t *testing.T) {
	var b HashInt32
	b.Add(1)
	b.AddOccurrences(2, 3)
	if got := b.Len(); got != 4 {
		t.Fatalf("Size() = %d, want 4", got)
	}
	if got := b.OccurrencesOf(2); got != 3 {
		t.Fatalf("OccurrencesOf(2) = %d, want 3", got)
	}
}

func TestZeroValueHashFloat64(t *testing.T) {
	var b HashFloat64
	b.Add(1.5)
	if got := b.OccurrencesOf(1.5); got != 1 {
		t.Fatalf("OccurrencesOf(1.5) = %d, want 1", got)
	}
}

func TestZeroValueTreeInt32(t *testing.T) {
	var b TreeInt32
	b.Add(7)
	if got := b.Len(); got != 1 {
		t.Fatalf("Size() = %d, want 1", got)
	}
}
