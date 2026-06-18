package hashmap

import (
	"errors"
	"math"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

func TestInt32Int32BulkLoad_EqualsIncremental(t *testing.T) {
	for _, n := range []int{0, 1, 3, 6, 12, 24, 48, 100, 257} {
		keys := make([]int32, n)
		vals := make([]int32, n)
		for i := 0; i < n; i++ {
			keys[i] = int32(i*7 + 1)
			vals[i] = int32(i)
		}
		bulk, err := Int32Int32BulkLoad(keys, vals, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		incr := NewInt32Int32()
		for i := 0; i < n; i++ {
			incr.Put(keys[i], vals[i])
		}
		if bulk.Len() != incr.Len() {
			t.Fatalf("n=%d: len differs %d vs %d", n, bulk.Len(), incr.Len())
		}
		for i := 0; i < n; i++ {
			bv, bok := bulk.Get(keys[i])
			iv, iok := incr.Get(keys[i])
			if bok != iok || bv != iv {
				t.Fatalf("n=%d: lookup differs for key %d", n, keys[i])
			}
		}
	}
}

// TestInt32Int32BulkLoadExact_ByteIdenticalLayout asserts the pumped table is the
// exact same slot layout as an incremental put-loop into a table of the same
// final capacity (the equivalence anchor for the zero-rehash contract). Uses the
// exact loader so the capacity is tight and the incremental map starts at the
// same capacity without resizing.
func TestInt32Int32BulkLoadExact_ByteIdenticalLayout(t *testing.T) {
	for _, n := range []int{12, 24, 48, 100, 257} {
		keys := make([]int32, n)
		vals := make([]int32, n)
		for i := 0; i < n; i++ {
			keys[i] = int32(i*7 + 1)
			vals[i] = int32(i)
		}
		bulk, err := Int32Int32BulkLoadExact(keys, vals, n, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		incr := NewInt32Int32WithCapacity(len(bulk.entries))
		if len(incr.entries) != len(bulk.entries) {
			t.Fatalf("n=%d: starting cap differs", n)
		}
		for i := 0; i < n; i++ {
			incr.Put(keys[i], vals[i])
		}
		if len(incr.entries) != len(bulk.entries) {
			t.Fatalf("n=%d: incremental resized (cap %d vs %d)", n, len(incr.entries), len(bulk.entries))
		}
		for i := range bulk.entries {
			if bulk.entries[i] != incr.entries[i] {
				t.Fatalf("n=%d: slot %d differs", n, i)
			}
		}
	}
}

// TestInt32Int32BulkLoadExact_ZeroRehash asserts the table is presized so that no
// rehash happens during the load, at the n = 3*2^k boundary that catches the
// off-by-one in a naive ceil(n/0.75).
func TestInt32Int32BulkLoadExact_ZeroRehash(t *testing.T) {
	for _, n := range []int{3, 6, 12, 24, 48, 96} {
		keys := make([]int32, n)
		vals := make([]int32, n)
		for i := 0; i < n; i++ {
			keys[i] = int32(i*13 + 1)
		}
		m, err := Int32Int32BulkLoadExact(keys, vals, n, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		wantCap := Int32Int32bulkCap(n)
		if len(m.entries) != wantCap {
			t.Fatalf("n=%d: capacity %d != presized %d (a rehash occurred)", n, len(m.entries), wantCap)
		}
		// confirm the table did NOT need resizing for the final insert
		if m.needsResize() {
			// after loading n, the next insert WOULD resize, but the load itself
			// must not have. Verify the predicate held false for the n-th insert:
			if (n)*4 >= len(m.entries)*3 {
				t.Fatalf("n=%d: load crossed the resize threshold", n)
			}
		}
	}
}

func TestInt32Int32BulkLoadExact_TooMany(t *testing.T) {
	keys := []int32{1, 2, 3, 4}
	vals := []int32{0, 0, 0, 0}
	if _, err := Int32Int32BulkLoadExact(keys, vals, 3, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrTooManyElements) {
		t.Fatalf("got %v", err)
	}
	// exactly n is fine
	if _, err := Int32Int32BulkLoadExact(keys, vals, 4, pump.ErrorOnDuplicate); err != nil {
		t.Fatalf("got %v", err)
	}
}

func TestInt32Int32BulkLoad_Duplicates(t *testing.T) {
	keys := []int32{1, 2, 1, 3}
	vals := []int32{10, 20, 99, 30}
	if _, err := Int32Int32BulkLoad(keys, vals, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("got %v", err)
	}
	m, err := Int32Int32BulkLoad(keys, vals, pump.IgnoreDuplicates)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Get(1); v != 10 {
		t.Fatalf("first-wins broken: got %d", v)
	}
	if m.Len() != 3 {
		t.Fatalf("len %d want 3", m.Len())
	}
}

func TestInt32Int32BulkLoad_LengthMismatchPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	Int32Int32BulkLoad([]int32{1}, []int32{1, 2}, pump.ErrorOnDuplicate)
}

func TestFloat64Float64BulkLoad_FloatEdges(t *testing.T) {
	negZero := math.Copysign(0, -1)
	nan := math.NaN()
	keys := []float64{negZero, 0, math.Inf(1), math.Inf(-1), nan}
	vals := []float64{1, 2, 3, 4, 5}
	m, err := Float64Float64BulkLoad(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	// -0 and +0 are distinct keys (bit-pattern equality)
	if v, ok := m.Get(negZero); !ok || v != 1 {
		t.Fatalf("-0 key: %v %v", v, ok)
	}
	if v, ok := m.Get(0); !ok || v != 2 {
		t.Fatalf("+0 key: %v %v", v, ok)
	}
	// NaN key findable
	if v, ok := m.Get(nan); !ok || v != 5 {
		t.Fatalf("NaN key: %v %v", v, ok)
	}
}

func TestInt32Int32BiMapBulkLoad(t *testing.T) {
	keys := []int32{1, 2, 3}
	vals := []int32{10, 20, 30}
	m, err := Int32Int32BiMapBulkLoad(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := m.Get(2); v != 20 {
		t.Fatalf("forward broken")
	}
	if k, _ := m.GetKey(30); k != 3 {
		t.Fatalf("reverse broken")
	}
	// duplicate key
	if _, err := Int32Int32BiMapBulkLoad([]int32{1, 1}, []int32{10, 20}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("dup key: got %v", err)
	}
	// duplicate value
	if _, err := Int32Int32BiMapBulkLoad([]int32{1, 2}, []int32{10, 10}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrDuplicateValue) {
		t.Fatalf("dup value: got %v", err)
	}
}

// TestInt32Int32BiMapBulkLoad_IgnoreDuplicatesStillErrors verifies the BiMap
// ignores the duplicate policy entirely: a duplicate key OR value is always an
// error, even under IgnoreDuplicates and even for a fully identical (k, v) pair
// (a repeated key breaks the single-pass bijection build).
func TestInt32Int32BiMapBulkLoad_IgnoreDuplicatesStillErrors(t *testing.T) {
	// identical duplicate pair must still error (regression: previously skipped)
	if _, err := Int32Int32BiMapBulkLoad([]int32{1, 1}, []int32{10, 10}, pump.IgnoreDuplicates); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("identical pair under IgnoreDuplicates: got %v, want ErrDuplicateKey", err)
	}
	// duplicate key under IgnoreDuplicates errors
	if _, err := Int32Int32BiMapBulkLoad([]int32{1, 1}, []int32{10, 20}, pump.IgnoreDuplicates); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("dup key under IgnoreDuplicates: got %v, want ErrDuplicateKey", err)
	}
	// duplicate value under IgnoreDuplicates errors
	if _, err := Int32Int32BiMapBulkLoad([]int32{1, 2}, []int32{10, 10}, pump.IgnoreDuplicates); !errors.Is(err, pump.ErrDuplicateValue) {
		t.Fatalf("dup value under IgnoreDuplicates: got %v, want ErrDuplicateValue", err)
	}
}
