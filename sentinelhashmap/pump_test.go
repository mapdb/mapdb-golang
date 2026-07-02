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

func TestInt32Int32BulkLoadExact_CollisionHeavyLayout(t *testing.T) {
	keys := []int32{11, 32, 43, 64, 75, 96, 107, 128, 139, 160, 171, 181}
	vals := make([]int32, len(keys))
	for i := range vals {
		vals[i] = int32(i * 10)
	}
	bulk, err := Int32Int32BulkLoadExact(keys, vals, len(keys), pump.ErrorOnDuplicate)
	if err != nil {
		t.Fatal(err)
	}
	incr := NewInt32Int32WithCapacity(len(bulk.keys))
	for i, k := range keys {
		incr.Put(k, vals[i])
	}
	if len(incr.keys) != len(bulk.keys) {
		t.Fatalf("incremental resized: %d vs %d", len(incr.keys), len(bulk.keys))
	}
	for i := range bulk.keys {
		if bulk.keys[i] != incr.keys[i] || bulk.values[i] != incr.values[i] {
			t.Fatalf("slot %d differs", i)
		}
	}
	if bulk.zeroKeyPresent != incr.zeroKeyPresent || bulk.zeroKeyValue != incr.zeroKeyValue ||
		bulk.oneKeyPresent != incr.oneKeyPresent || bulk.oneKeyValue != incr.oneKeyValue {
		t.Fatal("sentinel fields differ")
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

func TestInt32Int32BulkLoadExact_TooManyCountsConsumedDuplicates(t *testing.T) {
	keys := []int32{2, 2, 2}
	vals := []int32{20, 21, 22}
	if _, err := Int32Int32BulkLoadExact(keys, vals, 2, pump.IgnoreDuplicates); !errors.Is(err, pump.ErrTooManyElements) {
		t.Fatalf("got %v, want ErrTooManyElements", err)
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
