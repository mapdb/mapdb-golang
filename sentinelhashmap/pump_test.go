package sentinelhashmap

import (
	"errors"
	"math"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

func TestInt32Int32BulkLoad_EqualsIncremental(t *testing.T) {
	for _, n := range []int{0, 1, 3, 6, 12, 24, 48, 100} {
		keys := make([]int32, n)
		vals := make([]int32, n)
		for i := 0; i < n; i++ {
			// include the sentinel keys 0 and 1 as real data points
			keys[i] = int32(i)
			vals[i] = int32(i * 3)
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
			t.Fatalf("n=%d len differs", n)
		}
		for i := 0; i < n; i++ {
			bv, _ := bulk.Get(keys[i])
			iv, _ := incr.Get(keys[i])
			if bv != iv {
				t.Fatalf("n=%d: value differs for key %d", n, keys[i])
			}
		}
	}
}

func TestInt32Int32BulkLoadExact_ZeroRehash(t *testing.T) {
	for _, n := range []int{3, 6, 12, 24, 48, 96} {
		keys := make([]int32, n)
		vals := make([]int32, n)
		for i := 0; i < n; i++ {
			keys[i] = int32(i + 2) // avoid sentinels for a pure table load
		}
		m, err := Int32Int32BulkLoadExact(keys, vals, n, pump.ErrorOnDuplicate)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		if len(m.keys) != Int32Int32bulkCap(n) {
			t.Fatalf("n=%d: cap %d != %d (rehash happened)", n, len(m.keys), Int32Int32bulkCap(n))
		}
	}
}

func TestInt32Int32BulkLoad_SentinelKeys(t *testing.T) {
	// keys 0 and 1 are the sentinels — they must route to dedicated fields
	keys := []int32{0, 1, 5}
	vals := []int32{100, 101, 105}
	m, err := Int32Int32BulkLoad(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	for i, k := range keys {
		if v, ok := m.Get(k); !ok || v != vals[i] {
			t.Fatalf("sentinel key %d: got %v %v", k, v, ok)
		}
	}
	if m.Len() != 3 {
		t.Fatalf("len %d want 3", m.Len())
	}
	// duplicate sentinel key errors
	if _, err := Int32Int32BulkLoad([]int32{0, 0}, []int32{1, 2}, pump.ErrorOnDuplicate); !errors.Is(err, pump.ErrDuplicateKey) {
		t.Fatalf("got %v", err)
	}
}

func TestFloat64Float64BulkLoad_FloatEdges(t *testing.T) {
	negZero := math.Copysign(0, -1)
	keys := []float64{0, negZero, 1, math.Inf(1), math.NaN()}
	vals := []float64{1, 2, 3, 4, 5}
	m, err := Float64Float64BulkLoad(keys, vals, pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	// +0 and -0 distinct via the dedicated sentinel fields
	if v, ok := m.Get(0); !ok || v != 1 {
		t.Fatalf("+0: %v %v", v, ok)
	}
	if v, ok := m.Get(negZero); !ok || v != 2 {
		t.Fatalf("-0: %v %v", v, ok)
	}
	if m.Len() != 5 {
		t.Fatalf("len %d want 5", m.Len())
	}
}
