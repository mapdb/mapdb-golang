// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object_test

import (
	"context"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/object"
	"github.com/mapdb/mapdb-golang/par"
)

// The generic ArrayList[T] is a par.Segmenter[T] — the object-typed on-ramp to
// the parallel layer (load-bearing: drops or signature drift break compilation).
var _ par.Segmenter[int] = (*object.ArrayList[int])(nil)

// TestArrayListSegmentsCoverAll pins the Segments law for the generic list:
// concat(Segments(n)) ≡ All() as an (in-order) sequence, since a list's Segments
// follow its ordered backing array.
func TestArrayListSegmentsCoverAll(t *testing.T) {
	for _, size := range []int{0, 1, 7, 8, 100} {
		l := object.NewArrayList[int]()
		for i := 0; i < size; i++ {
			l.Add(i)
		}
		want := slices.Collect(l.All())

		for _, n := range []int{1, 2, 7, size + 1} {
			var got []int
			for _, s := range l.Segments(n) {
				got = append(got, slices.Collect(s)...)
			}
			if size == 0 {
				if len(got) != 0 {
					t.Fatalf("size=0 n=%d: got %v, want empty", n, got)
				}
				continue
			}
			// A list is ordered, so Segments concatenate back in order (not just
			// as a multiset).
			if !slices.Equal(got, want) {
				t.Fatalf("size=%d n=%d: concat = %v, want %v", size, n, got, want)
			}
		}
	}
}

// TestArrayListSegmentsFeedParFrom is the payoff: an object.ArrayList drives a
// real parallel map+reduce through par.From.
func TestArrayListSegmentsFeedParFrom(t *testing.T) {
	l := object.NewArrayList[int]()
	sum := 0
	for i := 0; i < 10_000; i++ {
		l.Add(i)
		sum += i
	}
	// Fold splits into per-segment accumulators then merges — exercises the
	// Segmenter across multiple workers.
	got, err := par.Fold(context.Background(), par.From[int](l, par.Workers(8)),
		func() int { return 0 },
		func(acc, v int) int { return acc + v },
		func(a, b int) int { return a + b },
	)
	if err != nil {
		t.Fatalf("par.Fold over list: %v", err)
	}
	if got != sum {
		t.Fatalf("parallel sum = %d, want %d", got, sum)
	}
}
