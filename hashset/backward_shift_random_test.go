package hashset

import (
	"math/rand/v2"
	"testing"
)

// Randomized Add/Remove/Contains against a reference Go map. The set's
// backward-shift deletion (hashSetTmpl rehashFrom) was already correct; this
// guards it and the top-bit index with a small table so probe runs wrap.
func TestBackwardShiftRandom_Int32Set(t *testing.T) {
	for seed := uint64(1); seed <= 8; seed++ {
		r := rand.New(rand.NewPCG(seed, 0x5e7))
		s := NewInt32WithCapacity(16)
		ref := map[int32]bool{}
		for op := 0; op < 5000; op++ {
			k := int32(r.IntN(40)) * 16
			switch r.IntN(3) {
			case 0:
				if got, want := s.Add(k), !ref[k]; got != want {
					t.Fatalf("seed %d op %d Add(%d): got %v want %v", seed, op, k, got, want)
				}
				ref[k] = true
			case 1:
				if got, want := s.Remove(k), ref[k]; got != want {
					t.Fatalf("seed %d op %d Remove(%d): got %v want %v", seed, op, k, got, want)
				}
				delete(ref, k)
			default:
				if got, want := s.Contains(k), ref[k]; got != want {
					t.Fatalf("seed %d op %d Contains(%d): got %v want %v", seed, op, k, got, want)
				}
			}
			if s.Len() != len(ref) {
				t.Fatalf("seed %d op %d: Len %d want %d", seed, op, s.Len(), len(ref))
			}
		}
	}
}
