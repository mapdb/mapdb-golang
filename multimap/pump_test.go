package multimap

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

func sortedInt32(s []int32) []int32 {
	c := append([]int32(nil), s...)
	sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	return c
}

func TestListMultimapFromSortedKeys_EqualsIncremental(t *testing.T) {
	keys := []int64{1, 1, 2, 5, 5, 5, 9}
	vals := []int32{10, 11, 20, 50, 51, 52, 90}

	grouped, err := NewInt64Int32ListFromSortedKeys(keys, vals)
	if err != nil {
		t.Fatal(err)
	}
	incr := NewInt64Int32List()
	for i := range keys {
		incr.Put(keys[i], vals[i])
	}
	if grouped.Len() != incr.Len() {
		t.Fatalf("len differs %d vs %d", grouped.Len(), incr.Len())
	}
	for _, k := range []int64{1, 2, 5, 9} {
		if !reflect.DeepEqual(grouped.GetAll(k), incr.GetAll(k)) {
			t.Fatalf("values for key %d differ: %v vs %v", k, grouped.GetAll(k), incr.GetAll(k))
		}
	}
}

func TestListMultimapBulkLoad_Unsorted(t *testing.T) {
	keys := []int64{5, 1, 5, 2, 1}
	vals := []int32{50, 10, 51, 20, 11}
	m := Int64Int32ListBulkLoad(keys, vals)
	incr := NewInt64Int32List()
	for i := range keys {
		incr.Put(keys[i], vals[i])
	}
	if m.Len() != incr.Len() {
		t.Fatalf("len differs")
	}
	for _, k := range []int64{1, 2, 5} {
		if !reflect.DeepEqual(m.GetAll(k), incr.GetAll(k)) {
			t.Fatalf("key %d: %v vs %v", k, m.GetAll(k), incr.GetAll(k))
		}
	}
}

func TestListMultimapFromSortedKeys_OrderError(t *testing.T) {
	// non-monotonic / interleaved keys error
	if _, err := NewInt64Int32ListFromSortedKeys([]int64{1, 2, 1}, []int32{0, 0, 0}); !errors.Is(err, pump.ErrNotSorted) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewInt64Int32ListFromSortedKeys([]int64{3, 1}, []int32{0, 0}); !errors.Is(err, pump.ErrNotSorted) {
		t.Fatalf("got %v", err)
	}
}

func TestSetMultimapFromSortedKeyValues_DedupesAndEqualsIncremental(t *testing.T) {
	// sorted by key then value; value 50 repeated within key 5 must dedupe
	keys := []int64{1, 2, 2, 5, 5, 5}
	vals := []int32{10, 20, 21, 50, 50, 51}

	built, err := NewInt64Int32SetFromSortedKeyValues(keys, vals)
	if err != nil {
		t.Fatal(err)
	}
	incr := NewInt64Int32Set()
	for i := range keys {
		incr.Put(keys[i], vals[i])
	}
	if built.Len() != incr.Len() {
		t.Fatalf("len differs %d vs %d", built.Len(), incr.Len())
	}
	for _, k := range []int64{1, 2, 5} {
		if !reflect.DeepEqual(sortedInt32(built.GetAll(k)), sortedInt32(incr.GetAll(k))) {
			t.Fatalf("key %d: %v vs %v", k, built.GetAll(k), incr.GetAll(k))
		}
	}
	// key 5 should have {50, 51} (one 50 dropped)
	if got := sortedInt32(built.GetAll(5)); !reflect.DeepEqual(got, []int32{50, 51}) {
		t.Fatalf("dedupe broken: %v", got)
	}
}

func TestSetMultimapBulkLoad_Unsorted(t *testing.T) {
	keys := []int64{5, 5, 1}
	vals := []int32{50, 50, 10} // duplicate (5,50) deduped
	m := Int64Int32SetBulkLoad(keys, vals)
	if m.Len() != 2 {
		t.Fatalf("len %d want 2", m.Len())
	}
	if got := m.GetAll(5); len(got) != 1 || got[0] != 50 {
		t.Fatalf("set dedupe broken: %v", got)
	}
}

func TestFloatKeyMultimapFromSorted(t *testing.T) {
	// float keys go through the IEEE total-order comparator
	keys := []float64{-1.0, -1.0, 0.0, 2.5}
	vals := []int32{1, 2, 3, 4}
	m, err := NewFloat64Int32ListFromSortedKeys(keys, vals)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.GetAll(-1.0), []int32{1, 2}) {
		t.Fatalf("float key grouping broken: %v", m.GetAll(-1.0))
	}
	if m.Len() != 4 {
		t.Fatalf("len %d", m.Len())
	}
}
