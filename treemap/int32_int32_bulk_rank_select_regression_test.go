package treemap

import (
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

// Regression for HIGH-3 (todo/fable-golang/01-critical-bugs.md): the FromSorted
// bulk builder created nodes without setting the subtree-size augmentation, so
// Rank/SelectKey returned wrong answers after a bulk load and subsequent
// mutations propagated the corruption.
func TestInt32Int32FromSorted_RankSelectAugmentation(t *testing.T) {
	keys := []int32{10, 20, 30, 40, 50}
	vals := []int32{100, 200, 300, 400, 500}
	m, err := NewInt32Int32FromSorted(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatalf("NewInt32Int32FromSorted: %v", err)
	}

	assertSizeInvariant(t, m)

	for i, k := range keys {
		if got := m.Rank(k); got != i {
			t.Errorf("Rank(%d) = %d, want %d", k, got, i)
		}
		if sk, ok := m.SelectKey(i); !ok || sk != k {
			t.Errorf("SelectKey(%d) = (%d, %v), want (%d, true)", i, sk, ok, k)
		}
	}
	if got := m.Rank(45); got != 4 {
		t.Errorf("Rank(45) = %d, want 4", got)
	}

	m.Put(25, 250)
	m.Remove(20)
	assertSizeInvariant(t, m)
	if got := m.Rank(30); got != 2 { // {10,25,30,40,50} -> index of 30 is 2
		t.Errorf("after mutation Rank(30) = %d, want 2", got)
	}
}
