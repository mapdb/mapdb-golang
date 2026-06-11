package bag

import "testing"

// Zero-value (nil-map) usability: a freshly declared map-backed bag must accept
// mutations without panicking ("assignment to entry in nil map"). Phase 7a.

func TestZeroValueInt32HashBag(t *testing.T) {
	var b Int32HashBag
	b.Add(1)
	b.AddOccurrences(2, 3)
	if got := b.Size(); got != 4 {
		t.Fatalf("Size() = %d, want 4", got)
	}
	if got := b.OccurrencesOf(2); got != 3 {
		t.Fatalf("OccurrencesOf(2) = %d, want 3", got)
	}
}

func TestZeroValueFloat64HashBag(t *testing.T) {
	var b Float64HashBag
	b.Add(1.5)
	if got := b.OccurrencesOf(1.5); got != 1 {
		t.Fatalf("OccurrencesOf(1.5) = %d, want 1", got)
	}
}

func TestZeroValueInt32TreeBag(t *testing.T) {
	var b Int32TreeBag
	b.Add(7)
	if got := b.Size(); got != 1 {
		t.Fatalf("Size() = %d, want 1", got)
	}
}
