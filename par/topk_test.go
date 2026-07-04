// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"
	"testing"
)

func less(a, b int) bool { return a < b }

// scrambled returns a deterministic permutation of 0..n-1 (fixed seed), so the top
// values are spread across segments and the merge phase must actually select —
// yet the expected result stays pinned (the set {0..n-1} is unchanged by shuffling).
func scrambled(n int) []int {
	xs := make([]int, n)
	for i := range xs {
		xs[i] = i
	}
	r := rand.New(rand.NewSource(42))
	r.Shuffle(n, func(i, j int) { xs[i], xs[j] = xs[j], xs[i] })
	return xs
}

func TestTopKBasic(t *testing.T) {
	const n = 50_000
	// Force real fan-out across segments so the per-segment/merge split is exercised.
	v := FromSlice(scrambled(n), Workers(8), MinPerWorker(1))
	got, err := TopK(context.Background(), v, 10, less)
	if err != nil {
		t.Fatalf("TopK: %v", err)
	}
	want := []int{n - 1, n - 2, n - 3, n - 4, n - 5, n - 6, n - 7, n - 8, n - 9, n - 10}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %d, want %d (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestTopKChunkPump(t *testing.T) {
	const n = 20_000
	// Chunk-pump source: unordered, single-shot — TopK is order-agnostic so it fits.
	v := FromSeq(seqOfSlice(scrambled(n)), Workers(8), MinPerWorker(256))
	got, err := TopK(context.Background(), v, 5, less)
	if err != nil {
		t.Fatalf("TopK: %v", err)
	}
	want := []int{n - 1, n - 2, n - 3, n - 4, n - 5}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %d, want %d (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestTopKFewerThanK(t *testing.T) {
	v := FromSlice([]int{3, 1, 2}, Workers(4), MinPerWorker(1))
	got, err := TopK(context.Background(), v, 10, less)
	if err != nil {
		t.Fatalf("TopK: %v", err)
	}
	want := []int{3, 2, 1}
	if len(got) != 3 {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestTopKNonPositiveKAndEmpty(t *testing.T) {
	ctx := context.Background()
	for _, k := range []int{0, -1} {
		got, err := TopK(ctx, FromSlice(iotaSlice(100), MinPerWorker(1)), k, less)
		if err != nil {
			t.Fatalf("TopK k=%d: %v", k, err)
		}
		if got == nil || len(got) != 0 {
			t.Fatalf("TopK k=%d = %v, want non-nil empty", k, got)
		}
	}
	// Empty source with a positive k also yields a non-nil empty slice.
	got, err := TopK(ctx, FromSlice([]int{}, MinPerWorker(1)), 5, less)
	if err != nil {
		t.Fatalf("TopK empty: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("TopK empty = %v, want non-nil empty", got)
	}
}

func TestTopKTiesReturnCorrectCount(t *testing.T) {
	// All-equal source: every element ties. The SET is unspecified but the COUNT
	// and values are pinned (all 7s), and it must not panic under fan-out.
	xs := make([]int, 10_000)
	for i := range xs {
		xs[i] = 7
	}
	v := FromSlice(xs, Workers(8), MinPerWorker(1))
	got, err := TopK(context.Background(), v, 4, less)
	if err != nil {
		t.Fatalf("TopK: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("len(got) = %d, want 4", len(got))
	}
	for i, x := range got {
		if x != 7 {
			t.Fatalf("got[%d] = %d, want 7", i, x)
		}
	}
}

func TestTopKMaxHeapSemanticsViaReversedLess(t *testing.T) {
	// Passing a reversed less selects the k SMALLEST, greatest-first under that order
	// (i.e. ascending true values): a smoke test that less drives selection, not a
	// hardcoded direction.
	const n = 1000
	v := FromSlice(scrambled(n), Workers(4), MinPerWorker(1))
	got, err := TopK(context.Background(), v, 3, func(a, b int) bool { return a > b })
	if err != nil {
		t.Fatalf("TopK: %v", err)
	}
	want := []int{0, 1, 2}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestTopKCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancelled
	_, err := TopK(ctx, FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1)), 10, less)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

// cancellingLess is a valid comparator that cancels the ctx once it has been
// consulted `after` times — the deterministic mid-run cancel pattern (a worker
// observes cancellation between elements and returns cctx.Err()).
func cancellingLess(cancel context.CancelFunc, after int32) func(a, b int) bool {
	var calls int32
	return func(a, b int) bool {
		if atomic.AddInt32(&calls, 1) == after {
			cancel()
		}
		return a < b
	}
}

func TestTopKCancelMidRunSegment(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	v := FromSlice(iotaSlice(200_000), Workers(8), MinPerWorker(1))
	_, err := TopK(ctx, v, 16, cancellingLess(cancel, 20))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("segment mid-run err = %v, want context.Canceled", err)
	}
}

func TestTopKCancelMidRunChunkPump(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	// countFrom is unbounded — only cancellation stops it, proving the chunk-pump
	// surfaces external cancel (and does not hang).
	v := FromSeq(countFrom(0), Workers(8), MinPerWorker(64))
	_, err := TopK(ctx, v, 16, cancellingLess(cancel, 20))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("chunk-pump mid-run err = %v, want context.Canceled", err)
	}
}

// TestTopKBoundaryK pins k at 1, len-1 and len over unsorted data with duplicates
// at the boundary — the extraction/off-by-one cases the large happy-path tests
// don't isolate. Values and order are asserted (the k-th boundary here never
// splits a tie group, so the result is fully determined).
func TestTopKBoundaryK(t *testing.T) {
	src := []int{5, 1, 5, 3, 2, 4, 1} // len 7; dups at max (5,5) and min (1,1)
	cases := []struct {
		k    int
		want []int
	}{
		{1, []int{5}},
		{6, []int{5, 5, 4, 3, 2, 1}},    // len-1: drops one (tied) 1
		{7, []int{5, 5, 4, 3, 2, 1, 1}}, // len: full descending sort
		{8, []int{5, 5, 4, 3, 2, 1, 1}}, // k>len: all, no over-pop
	}
	for _, c := range cases {
		v := FromSlice(append([]int(nil), src...), Workers(4), MinPerWorker(1))
		got, err := TopK(context.Background(), v, c.k, less)
		if err != nil {
			t.Fatalf("k=%d: %v", c.k, err)
		}
		if len(got) != len(c.want) {
			t.Fatalf("k=%d: got %v, want %v", c.k, got, c.want)
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Fatalf("k=%d: got %v, want %v", c.k, got, c.want)
			}
		}
	}
}
