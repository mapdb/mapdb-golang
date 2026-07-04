// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package priorityqueue_test

import (
	"context"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/par"
	"github.com/mapdb/mapdb-golang/priorityqueue"
)

// Both priority-queue variants are par.Segmenter[int32] via the generated
// Segments (the load-bearing par.From on-ramp).
var (
	_ par.Segmenter[int32] = (*priorityqueue.Int32)(nil)
	_ par.Segmenter[int32] = (*priorityqueue.SynchronizedInt32)(nil)
)

func multiset(seq func(func(int32) bool)) []int32 {
	got := slices.Collect(seq)
	slices.Sort(got)
	return got
}

// TestSegmentsCoverAll pins the Segments law: concat(Segments(n)) ≡ All() as a
// multiset (both walk the heap array, so they coincide, but the Segmenter
// contract only promises the multiset). ∀ n ∈ {1,2,7,len+1}, over sizes that
// stress the balanced-remainder split.
func TestSegmentsCoverAll(t *testing.T) {
	for _, size := range []int{0, 1, 7, 8, 100} {
		vals := make([]int32, size)
		for i := range vals {
			vals[i] = int32(size - i) // reverse, so heapify actually reorders
		}
		base := priorityqueue.Int32Of(vals...)
		sync := priorityqueue.NewSynchronizedInt32()
		for _, v := range vals {
			sync.Push(v)
		}

		wantBase := multiset(base.All())
		wantSync := multiset(sync.All())

		for _, n := range []int{1, 2, 7, size + 1} {
			var gb, gs []int32
			for _, s := range base.Segments(n) {
				gb = append(gb, slices.Collect(s)...)
			}
			for _, s := range sync.Segments(n) {
				gs = append(gs, slices.Collect(s)...)
			}
			slices.Sort(gb)
			slices.Sort(gs)
			if size == 0 {
				if len(gb) != 0 || len(gs) != 0 {
					t.Fatalf("size=0 n=%d: base=%v sync=%v, want empty", n, gb, gs)
				}
				continue
			}
			if !slices.Equal(gb, wantBase) {
				t.Fatalf("base size=%d n=%d: multiset = %v, want %v", size, n, gb, wantBase)
			}
			if !slices.Equal(gs, wantSync) {
				t.Fatalf("sync size=%d n=%d: multiset = %v, want %v", size, n, gs, wantSync)
			}
		}
	}
}

// TestSegmentsCountMatchesLen: segment count is min(n, Len) clamped to >= 1.
func TestSegmentsCountMatchesLen(t *testing.T) {
	q := priorityqueue.Int32Of(1, 2, 3, 4, 5)
	cases := []struct{ n, want int }{
		{-1, 1}, {0, 1}, {1, 1}, {3, 3}, {5, 5}, {6, 5}, {100, 5},
	}
	for _, c := range cases {
		if got := len(q.Segments(c.n)); got != c.want {
			t.Errorf("Segments(%d): %d segments, want %d", c.n, got, c.want)
		}
	}
}

// TestSegmentsFeedParFrom: a priority queue drives a real parallel reduction —
// the elements are covered exactly once regardless of heap layout.
func TestSegmentsFeedParFrom(t *testing.T) {
	q := priorityqueue.NewInt32()
	var want int64
	for i := int32(1); i <= 10_000; i++ {
		q.Push(i)
		want += int64(i)
	}
	got, err := par.Fold(context.Background(), par.From[int32](q, par.Workers(8)),
		func() int64 { return 0 },
		func(acc int64, v int32) int64 { return acc + int64(v) },
		func(a, b int64) int64 { return a + b },
	)
	if err != nil {
		t.Fatalf("par.Fold over priority queue: %v", err)
	}
	if got != want {
		t.Fatalf("parallel sum = %d, want %d", got, want)
	}
}
