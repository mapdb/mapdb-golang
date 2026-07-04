// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par_test

import (
	"context"
	"iter"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mapdb/mapdb-golang/par"
)

// pairsSeg is a hand-written Segmenter2 over parallel key/value slices, split into
// n balanced contiguous chunks — the pair analog of the slice segmenter used to
// exercise From without any collection dependency.
type pairsSeg struct{ ks, vs []int }

func (p pairsSeg) Segments2(n int) []iter.Seq2[int, int] {
	total := len(p.ks)
	if n > total {
		n = total
	}
	if n < 1 {
		n = 1
	}
	out := make([]iter.Seq2[int, int], n)
	chunk, rem, lo := total/n, total%n, 0
	for i := range out {
		hi := lo + chunk
		if i < rem {
			hi++
		}
		l, h := lo, hi
		out[i] = func(yield func(int, int) bool) {
			for j := l; j < h; j++ {
				if !yield(p.ks[j], p.vs[j]) {
					return
				}
			}
		}
		lo = hi
	}
	return out
}

var _ par.Segmenter2[int, int] = pairsSeg{}

func makePairs(n int) pairsSeg {
	ks := make([]int, n)
	vs := make([]int, n)
	for i := range ks {
		ks[i] = i
		vs[i] = i * 10
	}
	return pairsSeg{ks, vs}
}

// TestFrom2ForEachVisitsEveryPair checks From2 + View2.ForEach fans out over the
// pair segments and calls f once per (k, v) — verified via a concurrent-safe
// collector, since ForEach is unordered/concurrent.
func TestFrom2ForEachVisitsEveryPair(t *testing.T) {
	const n = 5000
	var mu sync.Mutex
	seen := make(map[int]int, n)
	err := par.From2[int, int](makePairs(n), par.Workers(8)).ForEach(context.Background(), func(k, v int) {
		mu.Lock()
		seen[k] = v
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("ForEach: %v", err)
	}
	if len(seen) != n {
		t.Fatalf("saw %d distinct keys, want %d", len(seen), n)
	}
	for k, v := range seen {
		if v != k*10 {
			t.Fatalf("pair (%d,%d) mismatched, want value %d", k, v, k*10)
		}
	}
}

// TestFrom2Count counts pairs matching a (k, v) predicate across segments.
func TestFrom2Count(t *testing.T) {
	got, err := par.From2[int, int](makePairs(1000), par.Workers(4)).
		Count(context.Background(), func(k, v int) bool { return k%2 == 0 })
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 500 {
		t.Fatalf("Count = %d, want 500", got)
	}
}

// TestFold2SumsValues aggregates over pairs; the merge combines per-segment sums.
func TestFold2SumsValues(t *testing.T) {
	const n = 10_000
	var want int64
	for i := 0; i < n; i++ {
		want += int64(i * 10)
	}
	got, err := par.Fold2(context.Background(), par.From2[int, int](makePairs(n), par.Workers(8)),
		func() int64 { return 0 },
		func(acc int64, k, v int) int64 { return acc + int64(v) },
		func(a, b int64) int64 { return a + b },
	)
	if err != nil {
		t.Fatalf("Fold2: %v", err)
	}
	if got != want {
		t.Fatalf("Fold2 sum = %d, want %d", got, want)
	}
}

// TestFrom2Cancel: a cancelled context surfaces as an error, not a wrong result.
func TestFrom2Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls atomic.Int64
	err := par.From2[int, int](makePairs(10_000), par.Workers(8)).ForEach(ctx, func(k, v int) {
		calls.Add(1)
	})
	if err == nil {
		t.Fatal("want context error from a pre-cancelled ForEach, got nil")
	}
}
