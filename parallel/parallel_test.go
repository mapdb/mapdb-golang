// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package parallel

import (
	"slices"
	"sync/atomic"
	"testing"
)

// ── ForEach ───────────────────────────────────────────────────────────

func TestForEach_Empty(t *testing.T) {
	ForEach([]int{}, func(int) { t.Fatal("should not be called") })
}

func TestForEach_Small_Sequential(t *testing.T) {
	var sum atomic.Int64
	data := []int{1, 2, 3, 4, 5}
	ForEach(data, func(v int) { sum.Add(int64(v)) })
	if sum.Load() != 15 {
		t.Fatalf("expected 15, got %d", sum.Load())
	}
}

func TestForEach_Large_Parallel(t *testing.T) {
	n := 100_000
	data := makeRange(n)
	var sum atomic.Int64
	ForEachWith(data, func(v int) { sum.Add(int64(v)) }, 1000, 8)
	expected := int64(n) * int64(n-1) / 2
	if sum.Load() != expected {
		t.Fatalf("expected %d, got %d", expected, sum.Load())
	}
}

// ── Select ────────────────────────────────────────────────────────────

func TestSelect_Empty(t *testing.T) {
	result := Select([]int{}, func(int) bool { return true })
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestSelect_Small_Sequential(t *testing.T) {
	data := []int{1, 2, 3, 4, 5, 6}
	evens := Select(data, func(v int) bool { return v%2 == 0 })
	if !slices.Equal(evens, []int{2, 4, 6}) {
		t.Fatalf("expected [2,4,6], got %v", evens)
	}
}

func TestSelect_Large_PreservesOrder(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	evens := SelectWith(data, func(v int) bool { return v%2 == 0 }, 1000, 8)
	if len(evens) != n/2 {
		t.Fatalf("expected %d evens, got %d", n/2, len(evens))
	}
	// Verify order is preserved
	for i := 1; i < len(evens); i++ {
		if evens[i] <= evens[i-1] {
			t.Fatalf("order broken at index %d: %d <= %d", i, evens[i], evens[i-1])
		}
	}
	// Verify all are even
	for _, v := range evens {
		if v%2 != 0 {
			t.Fatalf("non-even value: %d", v)
		}
	}
}

func TestSelect_AllMatch(t *testing.T) {
	data := makeRange(20_000)
	result := SelectWith(data, func(int) bool { return true }, 1000, 4)
	if !slices.Equal(result, data) {
		t.Fatal("expected all elements")
	}
}

func TestSelect_NoneMatch(t *testing.T) {
	data := makeRange(20_000)
	result := SelectWith(data, func(int) bool { return false }, 1000, 4)
	if result != nil {
		t.Fatalf("expected nil, got %d elements", len(result))
	}
}

// ── Reject ────────────────────────────────────────────────────────────

func TestReject_Large(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	odds := RejectWith(data, func(v int) bool { return v%2 == 0 }, 1000, 8)
	if len(odds) != n/2 {
		t.Fatalf("expected %d odds, got %d", n/2, len(odds))
	}
	for _, v := range odds {
		if v%2 == 0 {
			t.Fatalf("even value in reject result: %d", v)
		}
	}
}

// ── Collect ───────────────────────────────────────────────────────────

func TestCollect_Empty(t *testing.T) {
	result := Collect([]int{}, func(v int) string { return "" })
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestCollect_Small_Sequential(t *testing.T) {
	data := []int{1, 2, 3}
	doubled := Collect(data, func(v int) int { return v * 2 })
	if !slices.Equal(doubled, []int{2, 4, 6}) {
		t.Fatalf("expected [2,4,6], got %v", doubled)
	}
}

func TestCollect_Large_PreservesOrder(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	result := CollectWith(data, func(v int) int { return v * 10 }, 1000, 8)
	if len(result) != n {
		t.Fatalf("expected %d, got %d", n, len(result))
	}
	for i, v := range result {
		if v != i*10 {
			t.Fatalf("index %d: expected %d, got %d", i, i*10, v)
		}
	}
}

func TestCollect_TypeChange(t *testing.T) {
	data := makeRange(20_000)
	strs := CollectWith(data, func(v int) bool { return v%2 == 0 }, 1000, 4)
	if len(strs) != 20_000 {
		t.Fatalf("expected 20000, got %d", len(strs))
	}
	if strs[0] != true || strs[1] != false {
		t.Fatalf("expected true/false, got %v/%v", strs[0], strs[1])
	}
}

// ── Count ─────────────────────────────────────────────────────────────

func TestCount_Empty(t *testing.T) {
	if c := Count([]int{}, func(int) bool { return true }); c != 0 {
		t.Fatalf("expected 0, got %d", c)
	}
}

func TestCount_Small(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	c := Count(data, func(v int) bool { return v > 3 })
	if c != 2 {
		t.Fatalf("expected 2, got %d", c)
	}
}

func TestCount_Large(t *testing.T) {
	n := 100_000
	data := makeRange(n)
	c := CountWith(data, func(v int) bool { return v%3 == 0 }, 1000, 8)
	expected := 0
	for _, v := range data {
		if v%3 == 0 {
			expected++
		}
	}
	if c != expected {
		t.Fatalf("expected %d, got %d", expected, c)
	}
}

// ── AnySatisfy / AllSatisfy ───────────────────────────────────────────

func TestAnySatisfy_Empty(t *testing.T) {
	if AnySatisfy([]int{}, func(int) bool { return true }) {
		t.Fatal("expected false for empty")
	}
}

func TestAnySatisfy_Found(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	if !AnySatisfyWith(data, func(v int) bool { return v == n-1 }, 1000, 8) {
		t.Fatal("expected true")
	}
}

func TestAnySatisfy_NotFound(t *testing.T) {
	n := 50_000
	data := makeRange(n)
	if AnySatisfyWith(data, func(v int) bool { return v == n }, 1000, 8) {
		t.Fatal("expected false")
	}
}

func TestAllSatisfy(t *testing.T) {
	data := makeRange(50_000)
	if !AllSatisfy(data, func(v int) bool { return v >= 0 }) {
		t.Fatal("expected all >= 0")
	}
	if AllSatisfy(data, func(v int) bool { return v < 100 }) {
		t.Fatal("expected not all < 100")
	}
}

// ── Sum ───────────────────────────────────────────────────────────────

func TestSum_Empty(t *testing.T) {
	if s := Sum([]int{}); s != 0 {
		t.Fatalf("expected 0, got %d", s)
	}
}

func TestSum_Small(t *testing.T) {
	if s := Sum([]int{1, 2, 3, 4, 5}); s != 15 {
		t.Fatalf("expected 15, got %d", s)
	}
}

func TestSum_Large(t *testing.T) {
	n := 100_000
	data := makeRange(n)
	s := SumWith(data, 1000, 8)
	expected := n * (n - 1) / 2
	if s != expected {
		t.Fatalf("expected %d, got %d", expected, s)
	}
}

func TestSum_Float64(t *testing.T) {
	data := make([]float64, 50_000)
	for i := range data {
		data[i] = 1.0
	}
	s := SumWith(data, 1000, 8)
	if s != 50_000.0 {
		t.Fatalf("expected 50000, got %f", s)
	}
}

// ── splitBatches ──────────────────────────────────────────────────────

func TestSplitBatches(t *testing.T) {
	batches := splitBatches(10, 3)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	// 10 / 3 = 3 with remainder 1, so first batch gets 4
	if batches[0] != (batch{0, 4}) {
		t.Fatalf("batch 0: expected {0,4}, got %v", batches[0])
	}
	if batches[1] != (batch{4, 7}) {
		t.Fatalf("batch 1: expected {4,7}, got %v", batches[1])
	}
	if batches[2] != (batch{7, 10}) {
		t.Fatalf("batch 2: expected {7,10}, got %v", batches[2])
	}

	// More tasks than elements
	batches = splitBatches(3, 10)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches (capped), got %d", len(batches))
	}
}

// ── helpers ───────────────────────────────────────────────────────────

func makeRange(n int) []int {
	data := make([]int, n)
	for i := range data {
		data[i] = i
	}
	return data
}
