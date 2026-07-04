// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package arraylist_test

import (
	"context"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/arraylist"
	"github.com/mapdb/mapdb-golang/par"
)

// The generated Segments makes each array-list variant a par.Segmenter[int32] —
// the load-bearing on-ramp to the parallel layer (these fail to compile if the
// method is dropped or its signature drifts).
var (
	_ par.Segmenter[int32] = (*arraylist.Int32)(nil)
	_ par.Segmenter[int32] = (*arraylist.ImmutableInt32)(nil)
	_ par.Segmenter[int32] = (*arraylist.SynchronizedInt32)(nil)
)

// segmentsMultiset drains every segment of a Segmenter into one sorted slice —
// the "up to permutation" view the Segmenter contract guarantees.
func segmentsMultiset(seg par.Segmenter[int32], n int) []int32 {
	var got []int32
	for _, s := range seg.Segments(n) {
		for v := range s {
			got = append(got, v)
		}
	}
	slices.Sort(got)
	return got
}

// TestSegmentsCoverAllInt32 pins the Segments law across variants and split
// counts: concat(Segments(n)) ≡ All() as a multiset, for every n including
// n>len and n=1, over sizes that stress the balanced-remainder split.
func TestSegmentsCoverAllInt32(t *testing.T) {
	for _, size := range []int{0, 1, 7, 8, 100} {
		want := make([]int32, size)
		for i := range want {
			want[i] = int32(i)
		}

		base := arraylist.Int32Of(want...)
		imm := arraylist.NewImmutableInt32(want...)
		sync := arraylist.NewSynchronizedInt32From(arraylist.Int32Of(want...))

		variants := map[string]par.Segmenter[int32]{"base": base, "immutable": imm, "synchronized": sync}
		for name, seg := range variants {
			for _, n := range []int{1, 2, 7, size + 1} {
				got := segmentsMultiset(seg, n)
				// Normalize nil vs empty for the size==0 case.
				if len(got) == 0 && len(want) == 0 {
					continue
				}
				if !slices.Equal(got, want) {
					t.Fatalf("%s size=%d n=%d: multiset = %v, want %v", name, size, n, got, want)
				}
			}
		}
	}
}

// TestSegmentsCountMatchesLenInt32 checks the segment count is min(n, len)
// clamped to >= 1 — no empty trailing padding, no over-splitting past len.
func TestSegmentsCountMatchesLenInt32(t *testing.T) {
	l := arraylist.Int32Of(1, 2, 3, 4, 5) // len 5
	cases := []struct{ n, want int }{
		{-1, 1}, {0, 1}, {1, 1}, {3, 3}, {5, 5}, {6, 5}, {100, 5},
	}
	for _, c := range cases {
		if got := len(l.Segments(c.n)); got != c.want {
			t.Errorf("Segments(%d): %d segments, want %d", c.n, got, c.want)
		}
	}
}

// TestSegmentsAreReRunnableInt32 proves a returned view can be iterated more than
// once (par may re-run a View across terminals).
func TestSegmentsAreReRunnableInt32(t *testing.T) {
	l := arraylist.Int32Of(10, 20, 30, 40, 50, 60)
	segs := l.Segments(3)
	for i, s := range segs {
		first := slices.Collect(s)
		again := slices.Collect(s)
		if !slices.Equal(first, again) {
			t.Fatalf("segment %d not re-runnable: %v then %v", i, first, again)
		}
	}
}

// TestSegmentsFeedParFromInt32 is the payoff: an ArrayList plugs into par.From
// and a parallel Count matches the sequential answer — the interface is used for
// real, not merely asserted.
func TestSegmentsFeedParFromInt32(t *testing.T) {
	l := arraylist.NewInt32()
	for i := int32(0); i < 10_000; i++ {
		l.Add(i)
	}
	got, err := par.From[int32](l, par.Workers(8)).Count(context.Background(), func(v int32) bool {
		return v%2 == 0
	})
	if err != nil {
		t.Fatalf("par.From(list).Count: %v", err)
	}
	if got != 5000 {
		t.Fatalf("parallel even-count = %d, want 5000", got)
	}
}
