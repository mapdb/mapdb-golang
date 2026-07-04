// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package deque_test

import (
	"context"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/deque"
	"github.com/mapdb/mapdb-golang/par"
)

// Both deque variants are par.Segmenter[int32] via the generated Segments (the
// load-bearing par.From on-ramp).
var (
	_ par.Segmenter[int32] = (*deque.Int32)(nil)
	_ par.Segmenter[int32] = (*deque.SynchronizedInt32)(nil)
)

// buildWrapped returns a deque holding logical [0,1,...,size-1] whose physical
// head is non-zero and whose elements WRAP around the ring — the case that
// distinguishes a deque from a slice, and the one Segments must get right. size
// must be <= 8 so the default cap-16 ring never grow()s (grow repacks head to 0,
// which would undo the wrap and make the test trivial). Built by filling the back
// half via AddLast (head stays 0) then prepending the front half via AddFirst
// (which drives head backward past 0, wrapping it to the top of the buffer).
func buildWrapped(t *testing.T, size int) *deque.Int32 {
	t.Helper()
	if size > 8 {
		t.Fatalf("buildWrapped size %d > 8 would grow the ring and unwrap head", size)
	}
	d := deque.NewInt32()
	half := size / 2
	for i := half; i < size; i++ {
		d.AddLast(int32(i))
	}
	for i := half - 1; i >= 0; i-- {
		d.AddFirst(int32(i))
	}
	return d
}

// TestSegmentsWrappedRingInOrder is the load-bearing deque test: over a ring
// whose head is non-zero and whose storage wraps, concat(Segments(n)) reproduces
// the LOGICAL front-to-back order exactly. A naive slice-range split of the
// backing array (instead of the logical SplitIndex mapping) would yield
// stale/wrapped slots here and fail.
func TestSegmentsWrappedRingInOrder(t *testing.T) {
	for _, size := range []int{1, 2, 3, 5, 8} {
		want := make([]int32, size)
		for i := range want {
			want[i] = int32(i)
		}
		d := buildWrapped(t, size)

		// Sanity: the deque really is in logical order (ToSlice is the trusted
		// wrap-aware oracle) and All() agrees.
		if got := d.ToSlice(); !slices.Equal(got, want) {
			t.Fatalf("size=%d: ToSlice = %v, want %v (construction wrong?)", size, got, want)
		}
		if got := slices.Collect(d.All()); !slices.Equal(got, want) {
			t.Fatalf("size=%d: All = %v, want %v", size, got, want)
		}

		for _, n := range []int{1, 2, 3, 7, size + 1} {
			var got []int32
			for _, s := range d.Segments(n) {
				got = append(got, slices.Collect(s)...)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("size=%d n=%d: wrapped-ring concat = %v, want %v", size, n, got, want)
			}
		}
	}
}

// TestSegmentsCountMatchesLen: segment count is min(n, Len) clamped to >= 1.
func TestSegmentsCountMatchesLen(t *testing.T) {
	d := deque.NewInt32()
	for i := 0; i < 5; i++ {
		d.AddLast(int32(i))
	}
	cases := []struct{ n, want int }{
		{-1, 1}, {0, 1}, {1, 1}, {3, 3}, {5, 5}, {6, 5}, {100, 5},
	}
	for _, c := range cases {
		if got := len(d.Segments(c.n)); got != c.want {
			t.Errorf("Segments(%d): %d segments, want %d", c.n, got, c.want)
		}
	}
	// Empty deque → exactly one (empty) segment.
	if got := len(deque.NewInt32().Segments(4)); got != 1 {
		t.Errorf("empty Segments(4): %d segments, want 1", got)
	}
}

// TestSegmentsSyncInOrder checks the synchronized variant (splits a logical-order
// ToSlice snapshot) over a wrapped source.
func TestSegmentsSyncInOrder(t *testing.T) {
	sd := deque.NewSynchronizedInt32()
	sd.AddLast(4)
	sd.AddLast(5)
	sd.AddFirst(3)
	sd.AddFirst(2)
	sd.AddFirst(1)
	sd.AddFirst(0) // logical [0 1 2 3 4 5], head wrapped
	want := []int32{0, 1, 2, 3, 4, 5}
	for _, n := range []int{1, 2, 3, 7} {
		var got []int32
		for _, s := range sd.Segments(n) {
			got = append(got, slices.Collect(s)...)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("sync n=%d: concat = %v, want %v", n, got, want)
		}
	}
}

// TestSegmentsFeedParFrom: a deque drives a real parallel reduction, covering
// every element exactly once regardless of ring layout.
func TestSegmentsFeedParFrom(t *testing.T) {
	d := deque.NewInt32()
	var want int64
	for i := int32(1); i <= 10_000; i++ {
		if i%2 == 0 {
			d.AddLast(i)
		} else {
			d.AddFirst(i) // mix to exercise a non-trivial ring layout
		}
		want += int64(i)
	}
	got, err := par.Fold(context.Background(), par.From[int32](d, par.Workers(8)),
		func() int64 { return 0 },
		func(acc int64, v int32) int64 { return acc + int64(v) },
		func(a, b int64) int64 { return a + b },
	)
	if err != nil {
		t.Fatalf("par.Fold over deque: %v", err)
	}
	if got != want {
		t.Fatalf("parallel sum = %d, want %d", got, want)
	}
}
