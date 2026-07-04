// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package treeset_test

import (
	"context"
	"math/rand"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/par"
	"github.com/mapdb/mapdb-golang/treeset"
)

// A sorted set is a par.Segmenter[int32] via the rank-pruned Segments (the
// load-bearing par.From on-ramp, delivering ORDERED segments).
var _ par.Segmenter[int32] = (*treeset.Int32)(nil)

// buildSet returns a set of {0..size-1} inserted in the given order, so the tree
// SHAPE varies (ascending, descending, shuffled) while the sorted contents are
// identical — stressing the rank-pruned walk against different internal layouts.
func buildSet(size int, order string) *treeset.Int32 {
	keys := make([]int32, size)
	for i := range keys {
		keys[i] = int32(i)
	}
	switch order {
	case "asc":
		// already ascending
	case "desc":
		slices.Reverse(keys)
	case "shuf":
		r := rand.New(rand.NewSource(int64(size) * 2654435761))
		r.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	}
	s := treeset.NewInt32()
	for _, k := range keys {
		s.Add(k)
	}
	return s
}

// TestSegmentsConcatEqualsAllInOrder is the load-bearing differential test: for
// every tree shape, size, and split count, concat(Segments(n)) must equal
// slices.Collect(All()) EXACTLY — same elements, ascending, no gap, overlap, drop,
// duplicate, or misattributed boundary. All() is the trusted in-order oracle, so
// any bug in the rank-pruned walk (wrong prune guard, off-by-one at a rank
// boundary, shared closure range) surfaces as an inequality.
func TestSegmentsConcatEqualsAllInOrder(t *testing.T) {
	for _, order := range []string{"asc", "desc", "shuf"} {
		for _, size := range []int{0, 1, 2, 3, 7, 8, 31, 100, 257} {
			s := buildSet(size, order)
			want := slices.Collect(s.All())
			for _, n := range []int{1, 2, 3, 7, 8, size + 1, 1000} {
				var got []int32
				for _, seg := range s.Segments(n) {
					got = append(got, slices.Collect(seg)...)
				}
				if size == 0 {
					if len(got) != 0 {
						t.Fatalf("%s size=0 n=%d: got %v, want empty", order, n, got)
					}
					continue
				}
				if !slices.Equal(got, want) {
					t.Fatalf("%s size=%d n=%d: concat mismatch\n got=%v\nwant=%v", order, size, n, got, want)
				}
			}
		}
	}
}

// TestSegmentsCountMatchesLen: segment count is min(n, Len) clamped to >= 1.
func TestSegmentsCountMatchesLen(t *testing.T) {
	s := buildSet(5, "asc")
	cases := []struct{ n, want int }{
		{-1, 1}, {0, 1}, {1, 1}, {3, 3}, {5, 5}, {6, 5}, {100, 5},
	}
	for _, c := range cases {
		if got := len(s.Segments(c.n)); got != c.want {
			t.Errorf("Segments(%d): %d segments, want %d", c.n, got, c.want)
		}
	}
	if got := len(treeset.NewInt32().Segments(4)); got != 1 {
		t.Errorf("empty Segments(4): %d segments, want 1", got)
	}
}

// TestSegmentsAreDisjointRankRanges verifies each segment covers a CONTIGUOUS,
// ascending, non-overlapping rank range (the ordered-segment guarantee): segment
// i's elements are all less than segment i+1's, and each is internally ascending.
func TestSegmentsAreDisjointRankRanges(t *testing.T) {
	s := buildSet(100, "shuf")
	segs := s.Segments(7)
	var prevMax int32 = -1
	total := 0
	for i, seg := range segs {
		elems := slices.Collect(seg)
		if !slices.IsSorted(elems) {
			t.Fatalf("segment %d not ascending: %v", i, elems)
		}
		if len(elems) > 0 {
			if elems[0] <= prevMax {
				t.Fatalf("segment %d starts at %d, not strictly after previous max %d", i, elems[0], prevMax)
			}
			prevMax = elems[len(elems)-1]
		}
		total += len(elems)
	}
	if total != 100 {
		t.Fatalf("segments covered %d elements, want 100", total)
	}
}

// TestSegmentsEarlyBreak: a consumer that stops mid-segment must not panic and the
// walk's stop signal must propagate (partial consumption is safe and re-runnable).
func TestSegmentsEarlyBreak(t *testing.T) {
	s := buildSet(50, "shuf")
	seg := s.Segments(1)[0] // one segment over all 50, ascending 0..49
	var got []int32
	for v := range seg {
		got = append(got, v)
		if len(got) == 10 {
			break
		}
	}
	want := []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !slices.Equal(got, want) {
		t.Fatalf("early break got %v, want %v", got, want)
	}
	// Re-run the same view fully — early break must not have poisoned it.
	if full := slices.Collect(seg); len(full) != 50 {
		t.Fatalf("re-run after early break yielded %d, want 50", len(full))
	}
}

// TestSegmentsFeedParFrom: a sorted set drives a real ordered parallel reduction.
func TestSegmentsFeedParFrom(t *testing.T) {
	s := treeset.NewInt32()
	var want int64
	for i := int32(1); i <= 10_000; i++ {
		s.Add(i)
		want += int64(i)
	}
	got, err := par.Fold(context.Background(), par.From[int32](s, par.Workers(8)),
		func() int64 { return 0 },
		func(acc int64, v int32) int64 { return acc + int64(v) },
		func(a, b int64) int64 { return a + b },
	)
	if err != nil {
		t.Fatalf("par.Fold over treeset: %v", err)
	}
	if got != want {
		t.Fatalf("parallel sum = %d, want %d", got, want)
	}
}
