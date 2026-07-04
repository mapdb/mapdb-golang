// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package treemap_test

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/mapdb/mapdb-golang/par"
	"github.com/mapdb/mapdb-golang/treemap"
)

// A sorted map is a par.Segmenter2[K,V] via the rank-pruned Segments2 (the
// load-bearing par.From2 on-ramp, delivering ORDERED key/value segments).
var _ par.Segmenter2[int32, int32] = (*treemap.Int32Int32)(nil)

type kv struct{ k, v int32 }

// buildMap returns a map {i: i*10 | 0<=i<size} with keys inserted in the given
// order, so the tree SHAPE varies while the sorted contents are identical.
func buildMap(size int, order string) *treemap.Int32Int32 {
	keys := make([]int32, size)
	for i := range keys {
		keys[i] = int32(i)
	}
	switch order {
	case "desc":
		for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
			keys[i], keys[j] = keys[j], keys[i]
		}
	case "shuf":
		r := rand.New(rand.NewSource(int64(size) * 2654435761))
		r.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	}
	m := treemap.NewInt32Int32()
	for _, k := range keys {
		m.Put(k, k*10)
	}
	return m
}

func allPairs(m *treemap.Int32Int32) []kv {
	var out []kv
	for k, v := range m.All() {
		out = append(out, kv{k, v})
	}
	return out
}

// TestSegments2ConcatEqualsAllInOrder is the load-bearing differential test: for
// every tree shape, size, and split count, concatenating Segments2(n) must equal
// All() EXACTLY — same (key, value) pairs, ascending by key, no gap, overlap,
// drop, duplicate, or misattributed boundary, and each value paired with its
// correct key. All() is the trusted in-order oracle.
func TestSegments2ConcatEqualsAllInOrder(t *testing.T) {
	for _, order := range []string{"asc", "desc", "shuf"} {
		for _, size := range []int{0, 1, 2, 3, 7, 8, 31, 100, 257} {
			m := buildMap(size, order)
			want := allPairs(m)
			for _, n := range []int{1, 2, 3, 7, 8, size + 1, 1000} {
				var got []kv
				for _, seg := range m.Segments2(n) {
					for k, v := range seg {
						got = append(got, kv{k, v})
					}
				}
				if size == 0 {
					if len(got) != 0 {
						t.Fatalf("%s size=0 n=%d: got %v, want empty", order, n, got)
					}
					continue
				}
				if len(got) != len(want) {
					t.Fatalf("%s size=%d n=%d: got %d pairs, want %d", order, size, n, len(got), len(want))
				}
				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("%s size=%d n=%d pair %d: got %v, want %v", order, size, n, i, got[i], want[i])
					}
				}
			}
		}
	}
}

// TestSegments2CountMatchesLen: segment count is min(n, Len) clamped to >= 1.
func TestSegments2CountMatchesLen(t *testing.T) {
	m := buildMap(5, "asc")
	cases := []struct{ n, want int }{
		{-1, 1}, {0, 1}, {1, 1}, {3, 3}, {5, 5}, {6, 5}, {100, 5},
	}
	for _, c := range cases {
		if got := len(m.Segments2(c.n)); got != c.want {
			t.Errorf("Segments2(%d): %d segments, want %d", c.n, got, c.want)
		}
	}
	if got := len(treemap.NewInt32Int32().Segments2(4)); got != 1 {
		t.Errorf("empty Segments2(4): %d segments, want 1", got)
	}
}

// TestFrom2OverTreemap is the payoff: a sorted map plugs into par.From2 and drives
// real parallel pair terminals — ForEach visits every (k,v), and Fold2 aggregates
// keys and values with a correct total.
func TestFrom2OverTreemap(t *testing.T) {
	const n = 10_000
	m := treemap.NewInt32Int32()
	var wantKeys, wantVals int64
	for i := int32(0); i < n; i++ {
		m.Put(i, i*2)
		wantKeys += int64(i)
		wantVals += int64(i * 2)
	}

	// ForEach: collect concurrently and confirm every pair with its own value.
	var mu sync.Mutex
	seen := make(map[int32]int32, n)
	if err := par.From2[int32, int32](m, par.Workers(8)).ForEach(context.Background(), func(k, v int32) {
		mu.Lock()
		seen[k] = v
		mu.Unlock()
	}); err != nil {
		t.Fatalf("ForEach: %v", err)
	}
	if len(seen) != n {
		t.Fatalf("ForEach saw %d keys, want %d", len(seen), n)
	}
	for k, v := range seen {
		if v != k*2 {
			t.Fatalf("pair (%d,%d) wrong, want value %d", k, v, k*2)
		}
	}

	// Fold2: sum keys and values in one pass.
	type sums struct{ k, v int64 }
	got, err := par.Fold2(context.Background(), par.From2[int32, int32](m, par.Workers(8)),
		func() sums { return sums{} },
		func(a sums, k, v int32) sums { return sums{a.k + int64(k), a.v + int64(v)} },
		func(a, b sums) sums { return sums{a.k + b.k, a.v + b.v} },
	)
	if err != nil {
		t.Fatalf("Fold2: %v", err)
	}
	if got.k != wantKeys || got.v != wantVals {
		t.Fatalf("Fold2 = {k:%d v:%d}, want {k:%d v:%d}", got.k, got.v, wantKeys, wantVals)
	}
}
