package bag

import (
	"errors"
	"math"
	"testing"

	"github.com/mapdb/mapdb-golang/pump"
)

// TestTreeInt32Of_EqualsFromSorted verifies the O(n log n) Of path and the O(n)
// FromSorted path produce identical entries/counts, and that they match a naive
// repeated-Add bag.
func TestTreeInt32Of_EqualsFromSorted(t *testing.T) {
	raw := []int32{5, 1, 3, 3, 5, 5, 2, 1}
	of := TreeInt32Of(raw...)

	sorted := []int32{1, 1, 2, 3, 3, 5, 5, 5}
	fs, err := NewTreeInt32FromSorted(sorted)
	if err != nil {
		t.Fatal(err)
	}

	naive := NewTreeInt32()
	for _, v := range raw {
		naive.Add(v)
	}

	for _, b := range []*TreeInt32{of, fs, naive} {
		if b.Len() != 8 {
			t.Fatalf("Len=%d want 8", b.Len())
		}
	}
	// compare counts per value across all three
	for _, v := range []int32{1, 2, 3, 5} {
		c1 := of.OccurrencesOf(v)
		c2 := fs.OccurrencesOf(v)
		c3 := naive.OccurrencesOf(v)
		if c1 != c2 || c2 != c3 {
			t.Fatalf("value %d counts differ: of=%d fs=%d naive=%d", v, c1, c2, c3)
		}
	}
	if of.OccurrencesOf(1) != 2 || of.OccurrencesOf(5) != 3 {
		t.Fatalf("unexpected counts: 1=%d 5=%d", of.OccurrencesOf(1), of.OccurrencesOf(5))
	}
}

func TestTreeInt32FromSorted_OrderAndCoalesce(t *testing.T) {
	// sorted runs coalesce into counts
	fs, err := NewTreeInt32FromSorted([]int32{1, 1, 1, 4, 4, 9})
	if err != nil {
		t.Fatal(err)
	}
	if fs.SizeDistinct() != 3 {
		t.Fatalf("distinct=%d want 3", fs.SizeDistinct())
	}
	if fs.OccurrencesOf(1) != 3 || fs.OccurrencesOf(4) != 2 || fs.OccurrencesOf(9) != 1 {
		t.Fatal("bad coalesced counts")
	}
	// out of order errors
	if _, err := NewTreeInt32FromSorted([]int32{3, 1}); !errors.Is(err, pump.ErrNotSorted) {
		t.Fatalf("got %v", err)
	}
	// empty ok
	if b, err := NewTreeInt32FromSorted(nil); err != nil || b.Len() != 0 {
		t.Fatalf("empty: len=%d err=%v", b.Len(), err)
	}
}

func TestTreeFloat64FromSorted_FloatEdges(t *testing.T) {
	negZero := math.Copysign(0, -1)
	// total order: -Inf < -0 < +0 < +Inf < NaN ; duplicates coalesce
	vals := []float64{math.Inf(-1), negZero, negZero, 0, math.Inf(1), math.NaN()}
	b, err := NewTreeFloat64FromSorted(vals)
	if err != nil {
		t.Fatal(err)
	}
	if b.OccurrencesOf(negZero) != 2 {
		t.Fatalf("-0 count=%d want 2", b.OccurrencesOf(negZero))
	}
	// -0 and +0 distinct
	if b.OccurrencesOf(0) != 1 {
		t.Fatalf("+0 count=%d want 1", b.OccurrencesOf(0))
	}
	if b.Len() != 6 {
		t.Fatalf("len=%d want 6", b.Len())
	}
}
