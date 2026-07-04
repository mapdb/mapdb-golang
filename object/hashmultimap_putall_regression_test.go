package object

import "testing"

// Regression for M-1 (todo/fable-golang/01-critical-bugs.md): PutAll with zero
// values appended an empty slice, creating a phantom key — ContainsKey=true and
// SizeDistinct=1 while Len=0 and All never yields it. TreeMultimap has the guard;
// HashMultimap now matches.
func TestHashMultimap_PutAllZeroValuesNoPhantomKey(t *testing.T) {
	h := NewHashMultimap[string, int]()
	h.PutAll("k") // zero values

	if h.ContainsKey("k") {
		t.Error("ContainsKey(k) = true after PutAll with no values; want false")
	}
	if h.SizeDistinct() != 0 {
		t.Errorf("SizeDistinct = %d, want 0", h.SizeDistinct())
	}
	if h.Len() != 0 {
		t.Errorf("Len = %d, want 0", h.Len())
	}
	for k := range h.All() {
		t.Errorf("All yielded phantom key %q", k)
	}

	// A subsequent real PutAll still works.
	h.PutAll("k", 1, 2)
	if !h.ContainsKey("k") || h.Len() != 2 {
		t.Errorf("after real PutAll: ContainsKey=%v Len=%d, want true 2", h.ContainsKey("k"), h.Len())
	}
}
