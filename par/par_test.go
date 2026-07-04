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
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mapdb/mapdb-golang/internal/segment"
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

func TestWorkerGoexitReportedNotSilent(t *testing.T) {
	// A callback that exits via runtime.Goexit (e.g. t.FailNow inside a parallel
	// op) must surface as an error, not a silently-dropped segment result.
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := Map(context.Background(), v, func(x int) int {
		if x == 5000 {
			runtime.Goexit()
		}
		return x
	})
	if !errors.Is(err, errGoexit) {
		t.Fatalf("err = %v, want errGoexit", err)
	}
	if got != nil {
		t.Fatalf("result should be discarded on error, got len %d", len(got))
	}
}

// sliceSegmenter is a hand-written Segmenter for exercising From. It splits a
// backing slice into contiguous ranges, exactly like FromSlice but through the
// public interface.
type sliceSegmenter struct{ xs []int }

func (s sliceSegmenter) Segments(n int) []iter.Seq[int] {
	return segment.Split(s.xs, n)
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

func TestAnyAllNone(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	ctx := context.Background()

	if got, err := v.Any(ctx, func(x int) bool { return x == 9999 }); err != nil || !got {
		t.Fatalf("Any(present) = (%v, %v), want (true, nil)", got, err)
	}
	if got, err := v.Any(ctx, func(x int) bool { return x < 0 }); err != nil || got {
		t.Fatalf("Any(absent) = (%v, %v), want (false, nil)", got, err)
	}
	if got, err := v.All(ctx, func(x int) bool { return x >= 0 }); err != nil || !got {
		t.Fatalf("All(true) = (%v, %v), want (true, nil)", got, err)
	}
	if got, err := v.All(ctx, func(x int) bool { return x < 9999 }); err != nil || got {
		t.Fatalf("All(one fails) = (%v, %v), want (false, nil)", got, err)
	}
	if got, err := v.None(ctx, func(x int) bool { return x < 0 }); err != nil || !got {
		t.Fatalf("None(true) = (%v, %v), want (true, nil)", got, err)
	}
	if got, err := v.None(ctx, func(x int) bool { return x == 0 }); err != nil || got {
		t.Fatalf("None(has zero) = (%v, %v), want (false, nil)", got, err)
	}
}

func TestAnyAllNoneEmpty(t *testing.T) {
	v := FromSlice([]int{}, Workers(8), MinPerWorker(1))
	ctx := context.Background()
	if got, _ := v.Any(ctx, func(int) bool { return true }); got {
		t.Fatal("Any(empty) = true, want false")
	}
	if got, _ := v.All(ctx, func(int) bool { return false }); !got {
		t.Fatal("All(empty) = false, want true (vacuous)")
	}
	if got, _ := v.None(ctx, func(int) bool { return true }); !got {
		t.Fatal("None(empty) = false, want true (vacuous)")
	}
}

func TestAnyShortCircuits(t *testing.T) {
	// Match on the very first element of segment 0; with a shared found flag the
	// other segments must stop early, so far fewer than n predicate calls happen.
	const n = 1_000_000
	v := FromSlice(iotaSlice(n), Workers(8), MinPerWorker(1))
	var calls int64
	got, err := v.Any(context.Background(), func(x int) bool {
		atomic.AddInt64(&calls, 1)
		return x == 0
	})
	if err != nil || !got {
		t.Fatalf("Any = (%v, %v), want (true, nil)", got, err)
	}
	if c := atomic.LoadInt64(&calls); c >= n {
		t.Fatalf("Any did not short-circuit: %d predicate calls of %d", c, n)
	}
}

func TestFind(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	ctx := context.Background()
	// first-by-segment-order: lowest matching value overall.
	if got, ok, err := v.Find(ctx, func(x int) bool { return x >= 100 && x%7 == 0 }); err != nil || !ok || got != 105 {
		t.Fatalf("Find = (%d, %v, %v), want (105, true, nil)", got, ok, err)
	}
	if got, ok, err := v.Find(ctx, func(x int) bool { return x < 0 }); err != nil || ok || got != 0 {
		t.Fatalf("Find(absent) = (%d, %v, %v), want (0, false, nil)", got, ok, err)
	}
}

func TestSum(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := Sum(context.Background(), v)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if want := 10_000 * 9_999 / 2; got != want {
		t.Fatalf("Sum = %d, want %d", got, want)
	}
}

func TestMinMaxFunc(t *testing.T) {
	// Shuffle-free but non-trivial: values 0..n-1 laid out so min/max aren't at ends.
	xs := iotaSlice(10_000)
	xs[0], xs[9999] = 9999, 0 // move max to front, min to back
	v := FromSlice(xs, Workers(8), MinPerWorker(1))
	ctx := context.Background()
	less := func(a, b int) bool { return a < b }

	if got, ok, err := MinFunc(ctx, v, less); err != nil || !ok || got != 0 {
		t.Fatalf("MinFunc = (%d, %v, %v), want (0, true, nil)", got, ok, err)
	}
	if got, ok, err := MaxFunc(ctx, v, less); err != nil || !ok || got != 9999 {
		t.Fatalf("MaxFunc = (%d, %v, %v), want (9999, true, nil)", got, ok, err)
	}

	empty := FromSlice([]int{}, Workers(8), MinPerWorker(1))
	if _, ok, err := MinFunc(ctx, empty, less); err != nil || ok {
		t.Fatalf("MinFunc(empty) ok = %v, want false", ok)
	}
}

func TestSearchReduceCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	if _, err := v.Any(ctx, func(int) bool { return true }); !errors.Is(err, context.Canceled) {
		t.Fatalf("Any(cancelled) err = %v, want context.Canceled", err)
	}
	if _, _, err := v.Find(ctx, func(int) bool { return true }); !errors.Is(err, context.Canceled) {
		t.Fatalf("Find(cancelled) err = %v, want context.Canceled", err)
	}
	if _, err := Sum(ctx, v); !errors.Is(err, context.Canceled) {
		t.Fatalf("Sum(cancelled) err = %v, want context.Canceled", err)
	}
}

func TestForEachErrSuccess(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	var seen [10_000]int32
	err := v.ForEachErr(context.Background(), func(_ context.Context, x int) error {
		atomic.AddInt32(&seen[x], 1)
		return nil
	})
	if err != nil {
		t.Fatalf("ForEachErr: %v", err)
	}
	for i, c := range seen {
		if c != 1 {
			t.Fatalf("element %d visited %d times, want 1", i, c)
		}
	}
}

func TestForEachErrPropagatesAndCancels(t *testing.T) {
	sentinel := errors.New("sentinel")
	v := FromSlice(iotaSlice(1_000_000), Workers(8), MinPerWorker(1))
	var calls int64
	err := v.ForEachErr(context.Background(), func(_ context.Context, x int) error {
		atomic.AddInt64(&calls, 1)
		if x == 0 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
	if c := atomic.LoadInt64(&calls); c >= 1_000_000 {
		t.Fatalf("first error did not cancel siblings: %d calls", c)
	}
}

func TestMapErr(t *testing.T) {
	v := FromSlice(iotaSlice(5000), Workers(8), MinPerWorker(1))
	got, err := MapErr(context.Background(), v, func(_ context.Context, x int) (int, error) {
		return x * 2, nil
	})
	if err != nil {
		t.Fatalf("MapErr: %v", err)
	}
	for i, x := range got {
		if x != i*2 {
			t.Fatalf("got[%d] = %d, want %d", i, x, i*2)
		}
	}

	boom := errors.New("boom")
	res, err := MapErr(context.Background(), v, func(_ context.Context, x int) (int, error) {
		if x == 2500 {
			return 0, boom
		}
		return x, nil
	})
	if !errors.Is(err, boom) || res != nil {
		t.Fatalf("MapErr(fail) = (%v, %v), want (nil, boom)", res, err)
	}
}

func TestFilterErr(t *testing.T) {
	v := FromSlice(iotaSlice(9000), Workers(8), MinPerWorker(1))
	got, err := v.FilterErr(context.Background(), func(_ context.Context, x int) (bool, error) {
		return x%3 == 0, nil
	})
	if err != nil {
		t.Fatalf("FilterErr: %v", err)
	}
	if !sort.IntsAreSorted(got) || len(got) != 3000 {
		t.Fatalf("FilterErr result wrong: sorted=%v len=%d", sort.IntsAreSorted(got), len(got))
	}
	for i, x := range got {
		if x != i*3 {
			t.Fatalf("got[%d] = %d, want %d", i, x, i*3)
		}
	}
}

// TestErrCallbackObservesCancellation is the whole point of the ctx-aware twins:
// when one worker errors, a sibling blocked inside its callback must be released
// via the ctx it was handed — otherwise the op would hang. The test completing
// (not timing out) proves the callback's ctx becomes done.
func TestErrCallbackObservesCancellation(t *testing.T) {
	sentinel := errors.New("trigger")
	v := FromSlice(iotaSlice(100_000), Workers(4), MinPerWorker(1))
	err := v.ForEachErr(context.Background(), func(cctx context.Context, x int) error {
		if x == 0 {
			return sentinel // segment 0 trips immediately, cancelling the rest
		}
		// Every other element blocks until cancellation reaches its ctx.
		<-cctx.Done()
		return cctx.Err()
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel (first error wins before cancel)", err)
	}
}

func TestErrOpsPanicContained(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	defer func() {
		if _, ok := recover().(*PanicError); !ok {
			t.Fatal("want *PanicError from a panicking …Err callback")
		}
	}()
	_, _ = MapErr(context.Background(), v, func(_ context.Context, x int) (int, error) {
		if x == 5000 {
			panic("boom")
		}
		return x, nil
	})
	t.Fatal("unreachable")
}

func TestCountBy(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := CountBy(context.Background(), v, func(x int) int { return x % 10 })
	if err != nil {
		t.Fatalf("CountBy: %v", err)
	}
	if len(got) != 10 {
		t.Fatalf("CountBy keys = %d, want 10", len(got))
	}
	for k, c := range got {
		if c != 1000 {
			t.Fatalf("bucket %d = %d, want 1000", k, c)
		}
	}
}

func TestAggregateBySum(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	got, err := AggregateBy(context.Background(), v,
		func(x int) int { return x % 4 },    // key
		func() int { return 0 },             // newAcc
		func(a, x int) int { return a + x }, // acc within segment
		func(a, b int) int { return a + b }, // merge across segments (associative)
	)
	if err != nil {
		t.Fatalf("AggregateBy: %v", err)
	}
	// Independent oracle: sum of all x with x%4 == k.
	want := map[int]int{}
	for x := 0; x < 10_000; x++ {
		want[x%4] += x
	}
	if len(got) != len(want) {
		t.Fatalf("keys = %d, want %d", len(got), len(want))
	}
	for k, w := range want {
		if got[k] != w {
			t.Fatalf("bucket %d = %d, want %d", k, got[k], w)
		}
	}
}

func TestAggregateByEmptyAndCancel(t *testing.T) {
	empty := FromSlice([]int{}, Workers(8), MinPerWorker(1))
	got, err := CountBy(context.Background(), empty, func(x int) int { return x })
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("CountBy(empty) = (%v, %v), want (empty non-nil, nil)", got, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	if _, err := CountBy(ctx, v, func(x int) int { return x }); !errors.Is(err, context.Canceled) {
		t.Fatalf("CountBy(cancelled) err = %v, want context.Canceled", err)
	}
	if _, err := AggregateBy(ctx, v, func(x int) int { return x },
		func() int { return 0 }, func(a, x int) int { return a + x }, func(a, b int) int { return a + b },
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("AggregateBy(cancelled) err = %v, want context.Canceled", err)
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

// ── chunk-pump (FromSeq) ───────────────────────────────────────────────────

// seqOfSlice is a re-runnable iter.Seq over xs (test helper).
func seqOfSlice(xs []int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, x := range xs {
			if !yield(x) {
				return
			}
		}
	}
}

// countFrom is an UNBOUNDED iter.Seq: start, start+1, … forever. Terminals that
// don't terminate over it will hang the test (that's the point).
func countFrom(start int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := start; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func TestFromSeqCountSum(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(10_000)), Workers(4), MinPerWorker(64))
	ctx := context.Background()
	if got, err := v.Count(ctx, func(x int) bool { return x%2 == 0 }); err != nil || got != 5000 {
		t.Fatalf("Count = (%d, %v), want (5000, nil)", got, err)
	}
	if got, err := Sum(ctx, v); err != nil || got != 10_000*9_999/2 {
		t.Fatalf("Sum = (%d, %v)", got, err)
	}
}

func TestFromSeqMapUnorderedButComplete(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(3000)), Workers(4), MinPerWorker(32))
	got, err := Map(context.Background(), v, func(x int) int { return x * 2 })
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if len(got) != 3000 {
		t.Fatalf("Map len = %d, want 3000", len(got))
	}
	sort.Ints(got) // chunk-pump results are unordered; the SET must be complete
	for i, x := range got {
		if x != i*2 {
			t.Fatalf("after sort got[%d] = %d, want %d", i, x, i*2)
		}
	}
}

func TestFromSeqCountBy(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(10_000)), Workers(4), MinPerWorker(64))
	got, err := CountBy(context.Background(), v, func(x int) int { return x % 10 })
	if err != nil {
		t.Fatalf("CountBy: %v", err)
	}
	for k, c := range got {
		if c != 1000 {
			t.Fatalf("bucket %d = %d, want 1000", k, c)
		}
	}
}

func TestFromSeqReRunnableOverSlice(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(5000)), Workers(4), MinPerWorker(64))
	for run := 0; run < 3; run++ {
		got, err := v.Count(context.Background(), func(int) bool { return true })
		if err != nil || got != 5000 {
			t.Fatalf("run %d: Count = (%d, %v), want (5000, nil)", run, got, err)
		}
	}
}

func TestFromSeqEmpty(t *testing.T) {
	v := FromSeq(seqOfSlice(nil), Workers(4), MinPerWorker(64))
	if got, err := v.Count(context.Background(), func(int) bool { return true }); err != nil || got != 0 {
		t.Fatalf("Count(empty) = (%d, %v), want (0, nil)", got, err)
	}
}

// The critical property: short-circuit terminals must TERMINATE over an unbounded
// source by stopping the puller. If the puller ignored the early-done flag these
// would hang.
func TestFromSeqAnyTerminatesOverUnbounded(t *testing.T) {
	v := FromSeq(countFrom(0), Workers(4), MinPerWorker(16))
	got, err := v.Any(context.Background(), func(x int) bool { return x == 5 })
	if err != nil || !got {
		t.Fatalf("Any = (%v, %v), want (true, nil)", got, err)
	}
}

func TestFromSeqFindTerminatesOverUnbounded(t *testing.T) {
	v := FromSeq(countFrom(0), Workers(4), MinPerWorker(16))
	got, ok, err := v.Find(context.Background(), func(x int) bool { return x == 7 })
	if err != nil || !ok || !(got == 7 || got >= 0) {
		t.Fatalf("Find = (%d, %v, %v), want a match", got, ok, err)
	}
}

func TestFromSeqAllNoneFinite(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(1000)), Workers(4), MinPerWorker(16))
	ctx := context.Background()
	if got, err := v.All(ctx, func(x int) bool { return x >= 0 }); err != nil || !got {
		t.Fatalf("All = (%v, %v), want (true, nil)", got, err)
	}
	if got, err := v.None(ctx, func(x int) bool { return x < 0 }); err != nil || !got {
		t.Fatalf("None = (%v, %v), want (true, nil)", got, err)
	}
}

// Error-based cancellation must also stop the puller (via the internal ctx),
// terminating a fallible op over an unbounded source.
func TestFromSeqForEachErrTerminatesOverUnbounded(t *testing.T) {
	sentinel := errors.New("sentinel")
	v := FromSeq(countFrom(0), Workers(4), MinPerWorker(16))
	err := v.ForEachErr(context.Background(), func(_ context.Context, x int) error {
		if x >= 50 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
}

// External cancellation must terminate AND be reported (chunk-pump workers stop by
// ceasing to receive, so runChunks' final ctx.Err() check is what surfaces it).
func TestFromSeqExternalCancelTerminatesAndReports(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	v := FromSeq(countFrom(0), Workers(4), MinPerWorker(16))
	var n int64
	err := v.ForEach(ctx, func(int) {
		if atomic.AddInt64(&n, 1) == 100 {
			cancel()
		}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestFromSeqPreCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := FromSeq(seqOfSlice(iotaSlice(10_000)), Workers(4), MinPerWorker(16))
	called := int32(0)
	got, err := v.Count(ctx, func(int) bool { atomic.AddInt32(&called, 1); return true })
	if !errors.Is(err, context.Canceled) || got != 0 {
		t.Fatalf("Count(pre-cancel) = (%d, %v), want (0, Canceled)", got, err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("pred called %d times on pre-cancel, want 0", called)
	}
}

func TestFromSeqPanicContained(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(10_000)), Workers(4), MinPerWorker(16))
	defer func() {
		if _, ok := recover().(*PanicError); !ok {
			t.Fatalf("want *PanicError from a panicking chunk-pump callback")
		}
	}()
	_ = v.ForEach(context.Background(), func(x int) {
		if x == 5000 {
			panic("boom")
		}
	})
	t.Fatal("unreachable")
}

// ── chunk-pump review fixes: Goexit report + blocking-source teardown ───────

// chanSeqCtx is a ctx-aware channel source: its receive selects on ctx.Done(), so
// the engine can tear it down even while it is blocked waiting for a value.
func chanSeqCtx(ctx context.Context, ch <-chan int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return
				}
				if !yield(v) {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

func TestFromSeqWorkerGoexitReported(t *testing.T) {
	v := FromSeq(seqOfSlice(iotaSlice(10_000)), Workers(4), MinPerWorker(16))
	got, err := Map(context.Background(), v, func(x int) int {
		if x == 5000 {
			runtime.Goexit()
		}
		return x
	})
	if !errors.Is(err, errGoexit) {
		t.Fatalf("err = %v, want errGoexit (Goexit must not be silently dropped)", err)
	}
	if got != nil {
		t.Fatalf("result should be discarded on error, got len %d", len(got))
	}
}

// External cancellation must tear down a BLOCKING source (would hang if the
// source's receive did not observe the engine ctx).
func TestFromSeqCtxExternalCancelBlockingSource(t *testing.T) {
	ch := make(chan int) // never sends, never closes → the source blocks
	ctx, cancel := context.WithCancel(context.Background())
	v := FromSeqCtx(func(c context.Context) iter.Seq[int] { return chanSeqCtx(c, ch) },
		Workers(4), MinPerWorker(16))
	go cancel()
	err := v.ForEach(ctx, func(int) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

// Short-circuit over a ctx-aware source that yields one match then goes quiet must
// terminate: earlyDone cancels the engine ctx, unblocking the idle source. This
// hangs without the fix.
func TestFromSeqCtxAnyShortCircuitBlockingSource(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42 // one matching value; channel then stays open and quiet (blocks)
	v := FromSeqCtx(func(c context.Context) iter.Seq[int] { return chanSeqCtx(c, ch) },
		Workers(2), MinPerWorker(1))
	got, err := v.Any(context.Background(), func(x int) bool { return x == 42 })
	if err != nil || !got {
		t.Fatalf("Any = (%v, %v), want (true, nil)", got, err)
	}
}

// A ctx-aware source that eventually closes drives a normal (uncancelled) run.
func TestFromSeqCtxNormalCompletion(t *testing.T) {
	ch := make(chan int, 100)
	for i := 0; i < 100; i++ {
		ch <- i
	}
	close(ch)
	v := FromSeqCtx(func(c context.Context) iter.Seq[int] { return chanSeqCtx(c, ch) },
		Workers(4), MinPerWorker(8))
	got, err := v.Count(context.Background(), func(x int) bool { return x%2 == 0 })
	if err != nil || got != 50 {
		t.Fatalf("Count = (%d, %v), want (50, nil)", got, err)
	}
}
