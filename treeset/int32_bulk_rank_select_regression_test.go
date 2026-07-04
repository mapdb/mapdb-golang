package treeset

import (
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

// Regression for HIGH-3 (todo/fable-golang/01-critical-bugs.md): the FromSorted
// bulk builder created nodes without setting the subtree-size augmentation, so
// Rank/Select returned wrong answers after a bulk load and subsequent mutations
// propagated the corruption. Before the fix, NewInt32FromSorted([10..50]) gave
// Rank(45)=2 (want 4) and Select(3)=(0,false) (want 40).
func TestInt32FromSorted_RankSelectAugmentation(t *testing.T) {
	vals := []int32{10, 20, 30, 40, 50}
	s, err := NewInt32FromSorted(vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatalf("NewInt32FromSorted: %v", err)
	}

	// The size augmentation must be intact at every node immediately after build.
	assertSetSizeInvariant(t, s)

	for i, want := range vals {
		if got := s.Rank(want); got != i {
			t.Errorf("Rank(%d) = %d, want %d", want, got, i)
		}
		if v, ok := s.Select(i); !ok || v != want {
			t.Errorf("Select(%d) = (%d, %v), want (%d, true)", i, v, ok, want)
		}
	}
	if got := s.Rank(45); got != 4 {
		t.Errorf("Rank(45) = %d, want 4", got)
	}

	// Mutation after a bulk load must keep the augmentation consistent.
	s.Add(25)
	s.Remove(20)
	assertSetSizeInvariant(t, s)
	if got := s.Rank(30); got != 2 { // {10,25,30,40,50} -> index of 30 is 2
		t.Errorf("after mutation Rank(30) = %d, want 2", got)
	}
}
