// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
)

// iota returns 0..n-1 as a slice.
func iotaSlice(n int) []int {
	xs := make([]int, n)
	for i := range xs {
		xs[i] = i
	}
	return xs
}

func TestForEachVisitsEachElementOnce(t *testing.T) {
	const n = 10_000
	// Force real fan-out: many workers, tiny per-worker floor.
	v := FromSlice(iotaSlice(n), Workers(8), MinPerWorker(1))
	var seen [n]int32
	err := v.ForEach(context.Background(), func(x int) {
		atomic.AddInt32(&seen[x], 1)
	})
	if err != nil {
		t.Fatalf("ForEach: %v", err)
	}
	for i, c := range seen {
		if c != 1 {
			t.Fatalf("element %d visited %d times, want 1", i, c)
		}
	}
}

func TestCount(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := v.Count(context.Background(), func(x int) bool { return x%2 == 0 })
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 5000 {
		t.Fatalf("Count = %d, want 5000", got)
	}
}

func TestFilterPreservesOrder(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := v.Filter(context.Background(), func(x int) bool { return x%3 == 0 })
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if !sort.IntsAreSorted(got) {
		t.Fatalf("Filter result not in source order")
	}
	if len(got) != (10_000+2)/3 {
		t.Fatalf("Filter len = %d, want %d", len(got), (10_000+2)/3)
	}
	for i, x := range got {
		if x != i*3 {
			t.Fatalf("got[%d] = %d, want %d", i, x, i*3)
		}
	}
}

func TestMapPreservesOrder(t *testing.T) {
	v := FromSlice(iotaSlice(5000), Workers(8), MinPerWorker(1))
	got, err := Map(context.Background(), v, func(x int) string { return fmt.Sprintf("v%d", x) })
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(got) != 5000 {
		t.Fatalf("Map len = %d, want 5000", len(got))
	}
	for i, s := range got {
		if want := fmt.Sprintf("v%d", i); s != want {
			t.Fatalf("got[%d] = %q, want %q", i, s, want)
		}
	}
}

func TestReduceSum(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := v.Reduce(context.Background(), 0, func(a, b int) int { return a + b })
	if err != nil {
		t.Fatalf("Reduce: %v", err)
	}
	want := 10_000 * (10_000 - 1) / 2
	if got != want {
		t.Fatalf("Reduce = %d, want %d", got, want)
	}
}

func TestFoldConcatPreservesOrder(t *testing.T) {
	// Fold each segment into its own slice, merge left-to-right → source order.
	v := FromSlice(iotaSlice(3000), Workers(8), MinPerWorker(1))
	got, err := Fold(context.Background(), v,
		func() []int { return nil },
		func(a []int, x int) []int { return append(a, x) },
		func(a, b []int) []int { return append(a, b...) },
	)
	if err != nil {
		t.Fatalf("Fold: %v", err)
	}
	if len(got) != 3000 {
		t.Fatalf("Fold len = %d, want 3000", len(got))
	}
	for i, x := range got {
		if x != i {
			t.Fatalf("got[%d] = %d, want %d", i, x, i)
		}
	}
}

func TestReRunnable(t *testing.T) {
	v := FromSlice(iotaSlice(1000), Workers(4), MinPerWorker(1))
	for run := 0; run < 3; run++ {
		got, err := v.Count(context.Background(), func(x int) bool { return true })
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if got != 1000 {
			t.Fatalf("run %d: Count = %d, want 1000", run, got)
		}
	}
}

func TestEmpty(t *testing.T) {
	v := FromSlice([]int{}, Workers(8), MinPerWorker(1))
	if got, err := v.Count(context.Background(), func(int) bool { return true }); err != nil || got != 0 {
		t.Fatalf("Count empty = (%d, %v), want (0, nil)", got, err)
	}
	if got, err := v.Filter(context.Background(), func(int) bool { return true }); err != nil || got != nil {
		t.Fatalf("Filter empty = (%v, %v), want (nil, nil)", got, err)
	}
	if err := v.ForEach(context.Background(), func(int) { t.Fatal("f called on empty") }); err != nil {
		t.Fatalf("ForEach empty: %v", err)
	}
}

func TestSequentialFallbackIsCorrect(t *testing.T) {
	// MinPerWorker larger than the input → a single segment (sequential path).
	v := FromSlice(iotaSlice(100), Workers(8), MinPerWorker(1_000_000))
	got, err := v.Reduce(context.Background(), 0, func(a, b int) int { return a + b })
	if err != nil {
		t.Fatalf("Reduce: %v", err)
	}
	if want := 100 * 99 / 2; got != want {
		t.Fatalf("Reduce = %d, want %d", got, want)
	}
}

func TestPreCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	called := int32(0)
	err := v.ForEach(ctx, func(int) { atomic.AddInt32(&called, 1) })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("f called %d times on pre-cancelled ctx, want 0", called)
	}
}

func TestCancellationMidRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	v := FromSlice(iotaSlice(100_000), Workers(8), MinPerWorker(1))
	var count int32
	err := v.ForEach(ctx, func(int) {
		// Cancel partway through; siblings and this worker should wind down.
		if atomic.AddInt32(&count, 1) == 10 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if got := atomic.LoadInt32(&count); got >= 100_000 {
		t.Fatalf("cancellation did not short-circuit: visited all %d", got)
	}
}

func TestPanicContainment(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	var pe *PanicError
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected re-raised panic")
			}
			var ok bool
			pe, ok = r.(*PanicError)
			if !ok {
				t.Fatalf("recovered %T, want *PanicError", r)
			}
		}()
		_ = v.ForEach(context.Background(), func(x int) {
			if x == 5000 {
				panic("boom")
			}
		})
	}()
	if pe.Value != "boom" {
		t.Fatalf("PanicError.Value = %v, want boom", pe.Value)
	}
	if len(pe.Stack) == 0 {
		t.Fatal("PanicError.Stack empty")
	}
}

func TestPanicContainmentSingleSegment(t *testing.T) {
	// Sequential path must wrap panics identically.
	v := FromSlice(iotaSlice(100), Workers(1), MinPerWorker(1_000_000))
	defer func() {
		r := recover()
		if _, ok := r.(*PanicError); !ok {
			t.Fatalf("recovered %T, want *PanicError", r)
		}
	}()
	_ = v.ForEach(context.Background(), func(x int) {
		if x == 50 {
			panic(errors.New("kaboom"))
		}
	})
	t.Fatal("unreachable")
}

func TestPanicErrorUnwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	pe := &PanicError{Value: sentinel}
	if !errors.Is(pe, sentinel) {
		t.Fatal("errors.Is could not reach the wrapped error value")
	}
	if got := (&PanicError{Value: "not an error"}).Unwrap(); got != nil {
		t.Fatalf("Unwrap of non-error = %v, want nil", got)
	}
}

// sliceSegmenter is a hand-written Segmenter for exercising From. It splits a
// backing slice into contiguous ranges, exactly like FromSlice but through the
// public interface.
type sliceSegmenter struct{ xs []int }

func (s sliceSegmenter) Segments(n int) []iter.Seq[int] {
	return sliceSegments(s.xs, n)
}

func TestFromSegmenter(t *testing.T) {
	v := From[int](sliceSegmenter{iotaSlice(4000)}, Workers(8))
	got, err := v.Count(context.Background(), func(x int) bool { return x%2 == 0 })
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if got != 2000 {
		t.Fatalf("Count = %d, want 2000", got)
	}
	// Re-runnable through the interface too.
	sum, err := v.Reduce(context.Background(), 0, func(a, b int) int { return a + b })
	if err != nil {
		t.Fatalf("Reduce: %v", err)
	}
	if want := 4000 * 3999 / 2; sum != want {
		t.Fatalf("Reduce = %d, want %d", sum, want)
	}
}

func TestConcurrentAccumulationRaceFree(t *testing.T) {
	// A mutex-guarded sink under -race proves the engine adds no data race of
	// its own; the user is responsible for the callback's own safety.
	v := FromSlice(iotaSlice(50_000), Workers(8), MinPerWorker(1))
	var mu sync.Mutex
	sum := 0
	err := v.ForEach(context.Background(), func(x int) {
		mu.Lock()
		sum += x
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("ForEach: %v", err)
	}
	if want := 50_000 * 49_999 / 2; sum != want {
		t.Fatalf("sum = %d, want %d", sum, want)
	}
}
