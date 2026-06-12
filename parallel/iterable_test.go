// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package parallel

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func TestParallel_LenIsEmpty(t *testing.T) {
	if AsParallel([]int{}).Len() != 0 {
		t.Fatal("expected empty")
	}
	p := AsParallel([]int{1, 2, 3})
	if p.Len() == 0 {
		t.Fatal("expected non-empty")
	}
	if p.Len() != 3 {
		t.Fatalf("expected len 3, got %d", p.Len())
	}
}

func TestParallel_ForEach(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	var sum atomic.Int64
	AsParallelWith(data, 1000, 8).ForEach(func(v int) { sum.Add(int64(v)) })
	expected := int64(n) * int64(n-1) / 2
	if sum.Load() != expected {
		t.Fatalf("expected %d, got %d", expected, sum.Load())
	}
}

func TestParallel_SelectRejectPreserveOrder(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	p := AsParallelWith(data, 1000, 8)

	evens := p.Select(func(v int) bool { return v%2 == 0 })
	var wantEvens []int
	for _, v := range data {
		if v%2 == 0 {
			wantEvens = append(wantEvens, v)
		}
	}
	if !slices.Equal(evens, wantEvens) {
		t.Fatal("Select did not preserve order or content")
	}

	odds := p.Reject(func(v int) bool { return v%2 == 0 })
	var wantOdds []int
	for _, v := range data {
		if v%2 != 0 {
			wantOdds = append(wantOdds, v)
		}
	}
	if !slices.Equal(odds, wantOdds) {
		t.Fatal("Reject did not preserve order or content")
	}
}

func TestParallel_Collect(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	p := AsParallelWith(data, 1000, 8)
	got := ParallelCollect(p, func(v int) int { return v * v })
	for i, v := range got {
		if v != i*i {
			t.Fatalf("index %d: expected %d, got %d", i, i*i, v)
		}
	}
	if len(got) != n {
		t.Fatalf("expected %d, got %d", n, len(got))
	}
}

func TestParallel_CountAnyAllNoneDetect(t *testing.T) {
	n := 1000
	data := makeRange(n)
	p := AsParallelWith(data, 1, 16)

	if c := p.Count(func(v int) bool { return v >= 500 }); c != 500 {
		t.Fatalf("Count: expected 500, got %d", c)
	}
	if !p.AnySatisfy(func(v int) bool { return v == 999 }) {
		t.Fatal("AnySatisfy: expected true")
	}
	if p.AnySatisfy(func(v int) bool { return v == 1000 }) {
		t.Fatal("AnySatisfy: expected false")
	}
	if !p.AllSatisfy(func(v int) bool { return v >= 0 }) {
		t.Fatal("AllSatisfy: expected true")
	}
	if p.AllSatisfy(func(v int) bool { return v < 999 }) {
		t.Fatal("AllSatisfy: expected false")
	}
	if !p.NoneSatisfy(func(v int) bool { return v < 0 }) {
		t.Fatal("NoneSatisfy: expected true")
	}
	if p.NoneSatisfy(func(v int) bool { return v == 42 }) {
		t.Fatal("NoneSatisfy: expected false")
	}

	if v, ok := p.Detect(func(v int) bool { return v == 42 }); !ok || v != 42 {
		t.Fatalf("Detect: expected 42, got %d ok=%v", v, ok)
	}
	if _, ok := p.Detect(func(v int) bool { return v == 1000 }); ok {
		t.Fatal("Detect: expected not found")
	}
}

func TestParallel_DetectSequentialPath(t *testing.T) {
	// Small input: stays sequential (below default min fork size).
	p := AsParallel([]int{1, 2, 3, 4, 5})
	if v, ok := p.Detect(func(v int) bool { return v > 3 }); !ok || v != 4 {
		t.Fatalf("expected 4, got %d ok=%v", v, ok)
	}
	if _, ok := p.Detect(func(v int) bool { return v > 10 }); ok {
		t.Fatal("expected not found")
	}
}

func TestParallel_Sum(t *testing.T) {
	n := 100_000
	data := makeRange(n)
	p := AsParallelWith(data, 1000, 8)
	if s := ParallelSum(p); s != n*(n-1)/2 {
		t.Fatalf("expected %d, got %d", n*(n-1)/2, s)
	}

	floats := make([]float64, 50_000)
	for i := range floats {
		floats[i] = 1.0
	}
	if s := ParallelSum(AsParallelWith(floats, 1000, 8)); s != 50_000.0 {
		t.Fatalf("expected 50000, got %f", s)
	}
}

func TestParallel_Empty(t *testing.T) {
	p := AsParallel([]int{})
	if p.Count(func(int) bool { return true }) != 0 {
		t.Fatal("expected count 0")
	}
	if p.AnySatisfy(func(int) bool { return true }) {
		t.Fatal("expected AnySatisfy false")
	}
	if !p.AllSatisfy(func(int) bool { return false }) {
		t.Fatal("expected AllSatisfy vacuously true")
	}
	if _, ok := p.Detect(func(int) bool { return true }); ok {
		t.Fatal("expected Detect not found")
	}
	if ParallelSum(p) != 0 {
		t.Fatal("expected sum 0")
	}
}

// TestParallel_StressConcurrency hammers every parallel operation from many
// goroutines at once over a large dataset, forcing the batch fork path. Run
// under -race this exercises the goroutine/atomic/mutex synchronization.
func TestParallel_StressConcurrency(t *testing.T) {
	n := 100_000
	data := makeRange(n)
	p := AsParallelWith(data, 1, 16)

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)
	var failures atomic.Int64
	for w := 0; w < workers; w++ {
		go func(seed int) {
			defer wg.Done()
			switch seed % 6 {
			case 0:
				var cnt atomic.Int64
				p.ForEach(func(int) { cnt.Add(1) })
				if cnt.Load() != int64(n) {
					failures.Add(1)
				}
			case 1:
				if len(p.Select(func(v int) bool { return v%2 == 0 })) != n/2 {
					failures.Add(1)
				}
			case 2:
				got := ParallelCollect(p, func(v int) int { return v + 1 })
				if len(got) != n || got[0] != 1 || got[n-1] != n {
					failures.Add(1)
				}
			case 3:
				if p.Count(func(v int) bool { return v%3 == 0 }) != (n+2)/3 {
					failures.Add(1)
				}
			case 4:
				if !p.AnySatisfy(func(v int) bool { return v == n-1 }) {
					failures.Add(1)
				}
				if v, ok := p.Detect(func(v int) bool { return v == seed*1000 }); !ok || v != seed*1000 {
					failures.Add(1)
				}
			case 5:
				if ParallelSum(p) != n*(n-1)/2 {
					failures.Add(1)
				}
			}
		}(w)
	}
	wg.Wait()
	if f := failures.Load(); f != 0 {
		t.Fatalf("%d concurrent operations produced wrong results", f)
	}
}
