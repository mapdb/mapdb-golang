// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package segment

import (
	"slices"
	"testing"
)

// collect drains one view into a slice.
func collect[T any](seq func(func(T) bool)) []T {
	var out []T
	for v := range seq {
		out = append(out, v)
	}
	return out
}

// TestSplitCoversExactlyOnceInOrder pins the core law: for any n, the segments
// concatenate back to the original slice, in order (contiguous ranges preserve
// order), with no gap or overlap.
func TestSplitCoversExactlyOnceInOrder(t *testing.T) {
	for _, total := range []int{0, 1, 2, 3, 7, 8, 100} {
		xs := make([]int, total)
		for i := range xs {
			xs[i] = i
		}
		for _, n := range []int{1, 2, 3, 7, total, total + 1, 1000} {
			segs := Split(xs, n)

			// k = min(n, len) clamped to >= 1.
			wantK := n
			if wantK > total {
				wantK = total
			}
			if wantK < 1 {
				wantK = 1
			}
			if len(segs) != wantK {
				t.Fatalf("total=%d n=%d: got %d segments, want %d", total, n, len(segs), wantK)
			}

			var got []int
			for _, s := range segs {
				got = append(got, collect(s)...)
			}
			if total == 0 {
				if len(got) != 0 {
					t.Fatalf("total=0 n=%d: concat = %v, want empty", n, got)
				}
				continue
			}
			if !slices.Equal(got, xs) {
				t.Fatalf("total=%d n=%d: concat = %v, want %v", total, n, got, xs)
			}
		}
	}
}

// TestSplitBalanced checks segment lengths differ by at most one (the remainder
// is spread across the leading segments).
func TestSplitBalanced(t *testing.T) {
	xs := make([]int, 10)
	segs := Split(xs, 3) // 10 = 4 + 3 + 3
	lens := []int{len(collect(segs[0])), len(collect(segs[1])), len(collect(segs[2]))}
	if !slices.Equal(lens, []int{4, 3, 3}) {
		t.Fatalf("lengths = %v, want [4 3 3]", lens)
	}
}

// TestSplitViewsAreReRunnable proves each returned view is re-iterable (a
// re-runnable Segmenter requirement), and that the views are independent.
func TestSplitViewsAreReRunnable(t *testing.T) {
	xs := []int{10, 20, 30, 40}
	segs := Split(xs, 2)
	first := collect(segs[0])
	again := collect(segs[0])
	if !slices.Equal(first, again) {
		t.Fatalf("view not re-runnable: %v then %v", first, again)
	}
}

// TestSplitIndexMatchesSplit proves the index-addressed form produces the same
// segmentation as the slice form (Split is defined in terms of SplitIndex, so
// this pins that they cannot drift), and that at is called only with in-range
// indices (an at panicking out of range would be triggered otherwise).
func TestSplitIndexMatchesSplit(t *testing.T) {
	for _, total := range []int{0, 1, 7, 8, 100} {
		xs := make([]int, total)
		for i := range xs {
			xs[i] = i * 10
		}
		for _, n := range []int{1, 2, 7, total + 1} {
			at := func(i int) int {
				if i < 0 || i >= total {
					t.Fatalf("SplitIndex called at with out-of-range index %d (total=%d)", i, total)
				}
				return xs[i]
			}
			viaIndex := SplitIndex(total, n, at)
			viaSlice := Split(xs, n)
			if len(viaIndex) != len(viaSlice) {
				t.Fatalf("total=%d n=%d: SplitIndex %d segs, Split %d segs", total, n, len(viaIndex), len(viaSlice))
			}
			for i := range viaIndex {
				if !slices.Equal(collect(viaIndex[i]), collect(viaSlice[i])) {
					t.Fatalf("total=%d n=%d seg=%d: SplitIndex %v != Split %v",
						total, n, i, collect(viaIndex[i]), collect(viaSlice[i]))
				}
			}
		}
	}
}

// TestSplitEarlyBreakStops confirms a consumer that stops early does not panic
// and the range simply ends (yield returning false is honored).
func TestSplitEarlyBreakStops(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5}
	seg := Split(xs, 1)[0]
	n := 0
	for range seg {
		n++
		if n == 2 {
			break
		}
	}
	if n != 2 {
		t.Fatalf("early break visited %d, want 2", n)
	}
}
