// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package bitset_test

import (
	"context"
	"iter"
	"slices"
	"testing"

	"github.com/mapdb/mapdb-golang/bitset"
	"github.com/mapdb/mapdb-golang/par"
)

// Both bitset variants are par.Segmenter[int] via the generated Segments (the
// load-bearing par.From on-ramp).
var (
	_ par.Segmenter[int] = (*bitset.BitSet)(nil)
	_ par.Segmenter[int] = (*bitset.SynchronizedBitSet)(nil)
)

func concat(segs []iter.Seq[int]) []int {
	var out []int
	for _, s := range segs {
		out = append(out, slices.Collect(s)...)
	}
	return out
}

// TestSegmentsCoverAllBits is the load-bearing case: bits spread across MANY
// words (so the word-range split actually cuts across words), with clustered,
// sparse, and boundary (word-edge) positions. concat(Segments(n)) must equal the
// ascending set-bit order — the same as ToSlice/All — with no bit dropped,
// duplicated, or misattributed at a word boundary. ∀ n including n > word count.
func TestSegmentsCoverAllBits(t *testing.T) {
	// Positions chosen to straddle word boundaries (63/64, 127/128) and leave
	// whole words empty (a sparse high word at 1000).
	bitsSet := []int{0, 1, 5, 63, 64, 65, 127, 128, 200, 201, 202, 511, 512, 1000}
	b := bitset.NewBitSet()
	for _, p := range bitsSet {
		b.Set(p)
	}
	want := slices.Clone(bitsSet)
	slices.Sort(want)

	// Sanity: ToSlice (trusted oracle) and All agree.
	if got := b.ToSlice(); !slices.Equal(got, want) {
		t.Fatalf("ToSlice = %v, want %v", got, want)
	}
	if got := slices.Collect(b.All()); !slices.Equal(got, want) {
		t.Fatalf("All = %v, want %v", got, want)
	}

	for _, n := range []int{1, 2, 3, 7, 17, 100} {
		got := concat(b.Segments(n))
		if !slices.Equal(got, want) {
			t.Fatalf("n=%d: concat = %v, want %v", n, got, want)
		}
	}
}

// TestSegmentsEmptyAndAllZeroWords: an empty bitset and a bitset whose words
// exist but hold no set bits both yield no elements without panicking, and
// Segments returns at least one (empty) view.
func TestSegmentsEmptyAndAllZeroWords(t *testing.T) {
	empty := bitset.NewBitSet()
	if got := concat(empty.Segments(4)); len(got) != 0 {
		t.Fatalf("empty concat = %v, want none", got)
	}
	if len(empty.Segments(4)) != 1 { // total=0 words → one empty view
		t.Fatalf("empty Segments(4) count = %d, want 1", len(empty.Segments(4)))
	}

	// Allocate words (via a high Set then Clear) so len(words) > 0 but no bits set.
	z := bitset.NewBitSet()
	z.Set(500)
	z.Clear(500)
	if got := concat(z.Segments(3)); len(got) != 0 {
		t.Fatalf("all-zero-words concat = %v, want none", got)
	}
}

// TestSegmentsSyncMatches: the synchronized variant (splits a materialized
// ascending snapshot) covers the same bits.
func TestSegmentsSyncMatches(t *testing.T) {
	sb := bitset.NewSynchronizedBitSet()
	for _, p := range []int{3, 64, 65, 300, 301} {
		sb.Set(p)
	}
	want := []int{3, 64, 65, 300, 301}
	for _, n := range []int{1, 2, 7} {
		got := concat(sb.Segments(n))
		if !slices.Equal(got, want) {
			t.Fatalf("sync n=%d: concat = %v, want %v", n, got, want)
		}
	}
}

// TestSegmentsFeedParFrom: a dense bitset drives a real parallel reduction,
// covering every set bit exactly once across word ranges.
func TestSegmentsFeedParFrom(t *testing.T) {
	b := bitset.NewBitSet()
	var want int64
	for i := 0; i < 10_000; i++ {
		if i%3 == 0 {
			b.Set(i)
			want += int64(i)
		}
	}
	got, err := par.Fold(context.Background(), par.From[int](b, par.Workers(8)),
		func() int64 { return 0 },
		func(acc int64, v int) int64 { return acc + int64(v) },
		func(a, b int64) int64 { return a + b },
	)
	if err != nil {
		t.Fatalf("par.Fold over bitset: %v", err)
	}
	if got != want {
		t.Fatalf("parallel sum of set bits = %d, want %d", got, want)
	}
}
