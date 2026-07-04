// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package interval_test

import (
	"context"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/interval"
	"github.com/mapdb/mapdb-golang/par"
)

// An interval is a par.Segmenter[int32] via its index-computed Segments — the
// load-bearing on-ramp to the parallel layer for a virtual (unmaterialized)
// source.
var _ par.Segmenter[int32] = (*interval.Int32)(nil)

// TestSegmentsConcatInOrder pins the Segments law for the ORDERED, virtual
// interval: concat(Segments(n)) equals All() as a SEQUENCE (intervals are
// ordered and Segments follow the index space), ∀ n including n>len and n=1,
// across ascending, descending, stepped, and single-element intervals. (An empty
// interval is not publicly constructible — see the "single" case note; SplitIndex's
// total=0 path is covered in internal/segment.)
func TestSegmentsConcatInOrder(t *testing.T) {
	cases := []struct {
		name           string
		from, to, step int32
	}{
		{"ascending", 0, 99, 1},
		{"descending", 50, 1, -1},
		{"stepped", 0, 100, 7},
		{"single", 5, 5, 1}, // from==to → Len 1 (the smallest constructible interval;
		// the constructor forbids a from/to-vs-step direction mismatch, so Len is always >= 1)
	}
	for _, c := range cases {
		iv := interval.NewInt32(c.from, c.to, c.step)
		want := slices.Collect(iv.All())
		for _, n := range []int{1, 2, 7, len(want) + 1} {
			var got []int32
			for _, s := range iv.Segments(n) {
				got = append(got, slices.Collect(s)...)
			}
			if len(got) == 0 && len(want) == 0 {
				continue
			}
			if !slices.Equal(got, want) {
				t.Fatalf("%s n=%d: concat = %v, want %v", c.name, n, got, want)
			}
		}
	}
}

// TestSegmentsCountMatchesLen checks the segment count is min(n, Len) clamped to
// >= 1 — no empty padding, no over-splitting past Len.
func TestSegmentsCountMatchesLen(t *testing.T) {
	iv := interval.NewInt32(0, 4, 1) // Len 5
	cases := []struct{ n, want int }{
		{-1, 1}, {0, 1}, {1, 1}, {3, 3}, {5, 5}, {6, 5}, {100, 5},
	}
	for _, c := range cases {
		if got := len(iv.Segments(c.n)); got != c.want {
			t.Errorf("Segments(%d): %d segments, want %d", c.n, got, c.want)
		}
	}
}

// TestSegmentsFeedParFrom is the payoff: a virtual interval drives a real
// parallel reduction through par.From with no materialization of the range.
func TestSegmentsFeedParFrom(t *testing.T) {
	iv := interval.NewInt32(1, 10_000, 1)
	got, err := par.Fold(context.Background(), par.From[int32](iv, par.Workers(8)),
		func() int64 { return 0 },
		func(acc int64, v int32) int64 { return acc + int64(v) },
		func(a, b int64) int64 { return a + b },
	)
	if err != nil {
		t.Fatalf("par.Fold over interval: %v", err)
	}
	const want = int64(10_000) * 10_001 / 2 // sum 1..10000
	if got != want {
		t.Fatalf("parallel sum = %d, want %d", got, want)
	}
}
