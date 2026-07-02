package bitset

import "testing"

func TestBitSetOfIndices_EqualsRepeatedSet(t *testing.T) {
	indices := []int{0, 5, 63, 64, 200, 1, 5} // unsorted, with a duplicate
	bulk := BitSetOfIndices(indices...)

	naive := NewBitSet()
	for _, i := range indices {
		naive.Set(i)
	}

	if bulk.Cardinality() != naive.Cardinality() {
		t.Fatalf("cardinality %d vs %d", bulk.Cardinality(), naive.Cardinality())
	}
	for i := 0; i <= 200; i++ {
		if bulk.Get(i) != naive.Get(i) {
			t.Fatalf("bit %d differs", i)
		}
	}
}

func TestBitSetSetAll(t *testing.T) {
	b := NewBitSetOfLength(10)
	b.SetAll(2, 100, 4)
	if !b.Get(2) || !b.Get(4) || !b.Get(100) {
		t.Fatal("SetAll missed a bit")
	}
	if b.Get(3) {
		t.Fatal("SetAll set an extra bit")
	}
	if b.Cardinality() != 3 {
		t.Fatalf("cardinality %d want 3", b.Cardinality())
	}
}

func TestBitSetOfIndices_NegativePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	BitSetOfIndices(1, -1)
}
