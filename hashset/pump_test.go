package hashset

import (
	"errors"
	"math"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

func TestInt32BulkLoad_EqualsIncremental(t *testing.T) {
	for _, n := range []int{0, 1, 3, 6, 12, 24, 48, 100} {
		vals := make([]int32, n)
		for i := range vals {
			vals[i] = int32(i*5 + 1)
		}
		bulk, err := Int32BulkLoad(vals, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		incr := NewInt32()
		for _, v := range vals {
			incr.Add(v)
		}
		if bulk.Len() != incr.Len() {
			t.Fatalf("n=%d len differs", n)
		}
		for _, v := range vals {
			if !bulk.Contains(v) {
				t.Fatalf("n=%d missing %d", n, v)
			}
		}
	}
}

func TestInt32BulkLoadExact_ZeroRehash(t *testing.T) {
	for _, n := range []int{3, 6, 12, 24, 48, 96} {
		vals := make([]int32, n)
		for i := range vals {
			vals[i] = int32(i*11 + 1)
		}
		s, err := Int32BulkLoadExact(vals, n, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		if len(s.entries) != Int32bulkCap(n) {
			t.Fatalf("n=%d: cap %d != %d (rehash happened)", n, len(s.entries), Int32bulkCap(n))
		}
	}
}

func TestInt32BulkLoadExact_TooMany(t *testing.T) {
	if _, err := Int32BulkLoadExact([]int32{1, 2, 3}, 2, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrTooManyElements) {
		t.Fatalf("got %v", err)
	}
}

func TestInt32BulkLoadExact_TooManyCountsConsumedDuplicates(t *testing.T) {
	if _, err := Int32BulkLoadExact([]int32{1, 1, 1}, 2, pump.IgnoreDuplicates); !errors.Is(err, pump.ErrTooManyElements) {
		t.Fatalf("got %v, want ErrTooManyElements", err)
	}
}

func TestInt32BulkLoad_Duplicates(t *testing.T) {
	if _, err := Int32BulkLoad([]int32{1, 2, 1}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("got %v", err)
	}
	s, err := Int32BulkLoad([]int32{1, 2, 1, 3}, pump.IgnoreDuplicates)
	if err != nil || s.Len() != 3 {
		t.Fatalf("len=%d err=%v", s.Len(), err)
	}
}

func TestFloat64BulkLoad_FloatEdges(t *testing.T) {
	negZero := math.Copysign(0, -1)
	nan := math.NaN()
	s, err := Float64BulkLoad([]float64{negZero, 0, math.Inf(1), math.Inf(-1), nan}, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Contains(negZero) || !s.Contains(0) {
		t.Fatal("-0/+0 distinct membership broken")
	}
	if !s.Contains(nan) {
		t.Fatal("NaN membership broken")
	}
	if s.Len() != 5 {
		t.Fatalf("len=%d want 5", s.Len())
	}
}

func TestBoolBulkLoad(t *testing.T) {
	s, err := BoolBulkLoad([]bool{true, false, true}, pump.IgnoreDuplicates)
	if err != nil {
		t.Fatal(err)
	}
	if s.Len() != 2 || !s.Contains(true) || !s.Contains(false) {
		t.Fatalf("bool bulk load broken: len=%d", s.Len())
	}
}
