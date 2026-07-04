// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"errors"
	"runtime"
	"sort"
	"sync/atomic"
	"testing"
	"time"
)

// drain collects every value the seq yields into a slice.
func drain[R any](seq func(func(R) bool)) []R {
	var out []R
	for r := range seq {
		out = append(out, r)
	}
	return out
}

func TestMapSeqStreamsAllSegment(t *testing.T) {
	const n = 20_000
	v := FromSlice(iotaSlice(n), Workers(8), MinPerWorker(1))
	seq, join := MapSeq(context.Background(), v, func(x int) int { return x * 2 })
	got := drain(seq)
	if err := join(); err != nil {
		t.Fatalf("join: %v", err)
	}
	// Unordered: compare as sets. Every element mapped exactly once.
	if len(got) != n {
		t.Fatalf("len(got) = %d, want %d", len(got), n)
	}
	sort.Ints(got)
	for i := 0; i < n; i++ {
		if got[i] != i*2 {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], i*2)
		}
	}
}

func TestMapSeqStreamsAllChunkPump(t *testing.T) {
	const n = 20_000
	v := FromSeq(seqOfSlice(iotaSlice(n)), Workers(8), MinPerWorker(128))
	seq, join := MapSeq(context.Background(), v, func(x int) int { return x + 1 })
	got := drain(seq)
	if err := join(); err != nil {
		t.Fatalf("join: %v", err)
	}
	if len(got) != n {
		t.Fatalf("len(got) = %d, want %d", len(got), n)
	}
	sort.Ints(got)
	for i := 0; i < n; i++ {
		if got[i] != i+1 {
			t.Fatalf("got[%d] = %d, want %d", i, got[i], i+1)
		}
	}
}

func TestMapSeqSingleShot(t *testing.T) {
	v := FromSlice(iotaSlice(1000), Workers(4), MinPerWorker(1))
	seq, join := MapSeq(context.Background(), v, func(x int) int { return x })
	first := drain(seq)
	second := drain(seq) // single-shot: nothing left
	if err := join(); err != nil {
		t.Fatalf("join: %v", err)
	}
	if len(first) != 1000 {
		t.Fatalf("first pass len = %d, want 1000", len(first))
	}
	if len(second) != 0 {
		t.Fatalf("second pass len = %d, want 0 (single-shot)", len(second))
	}
}

func TestMapSeqNeverIteratedJoinsNil(t *testing.T) {
	v := FromSlice(iotaSlice(1000), Workers(4), MinPerWorker(1))
	_, join := MapSeq(context.Background(), v, func(x int) int { return x })
	if err := join(); err != nil {
		t.Fatalf("join on never-iterated seq = %v, want nil", err)
	}
}

// TestMapSeqAbandonNoLeak breaks out of the range early: the workers must be
// cancelled and drained before the range returns, leaving no goroutines behind.
func TestMapSeqAbandonNoLeak(t *testing.T) {
	base := runtime.NumGoroutine()
	// Unbounded chunk-pump source: without proper teardown its puller/workers leak.
	v := FromSeq(countFrom(0), Workers(8), MinPerWorker(64))
	seq, join := MapSeq(context.Background(), v, func(x int) int { return x })

	taken := 0
	for range seq {
		taken++
		if taken == 100 {
			break // abandon
		}
	}
	if taken != 100 {
		t.Fatalf("took %d before break, want 100", taken)
	}
	// Abandoning is not an error.
	if err := join(); err != nil {
		t.Fatalf("join after abandon = %v, want nil", err)
	}
	// All producer goroutines have exited (join waited on allDone). Allow a brief
	// settle for the scheduler and assert we are back near baseline.
	assertGoroutinesSettle(t, base)
}

func TestMapSeqPanicReraisedAtJoin(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	seq, join := MapSeq(context.Background(), v, func(x int) int {
		if x == 5000 {
			panic("boom")
		}
		return x
	})
	// Iteration itself must not panic — it ends when the stream closes.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("iteration panicked (%v); panic should surface at join", r)
			}
		}()
		drain(seq)
	}()
	// join re-raises the contained panic, wrapped in *PanicError.
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("join did not re-raise the worker panic")
		}
		pe, ok := r.(*PanicError)
		if !ok {
			t.Fatalf("re-raised %T, want *PanicError", r)
		}
		if pe.Value != "boom" {
			t.Fatalf("panic value = %v, want boom", pe.Value)
		}
	}()
	join()
}

func TestMapSeqGoexitReportedAtJoin(t *testing.T) {
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	seq, join := MapSeq(context.Background(), v, func(x int) int {
		if x == 5000 {
			runtime.Goexit()
		}
		return x
	})
	drain(seq)
	if err := join(); !errors.Is(err, errGoexit) {
		t.Fatalf("join = %v, want errGoexit", err)
	}
}

func TestMapSeqPreCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	v := FromSlice(iotaSlice(10_000), Workers(8), MinPerWorker(1))
	seq, join := MapSeq(ctx, v, func(x int) int { return x })
	got := drain(seq)
	if len(got) != 0 {
		t.Fatalf("pre-cancelled yielded %d results, want 0", len(got))
	}
	if err := join(); !errors.Is(err, context.Canceled) {
		t.Fatalf("join = %v, want context.Canceled", err)
	}
}

func TestMapSeqMidRunCancel(t *testing.T) {
	base := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // vet: guarantee cancel on every path (idempotent with the mid-loop call)
	v := FromSeq(countFrom(0), Workers(8), MinPerWorker(64))
	seq, join := MapSeq(ctx, v, func(x int) int { return x })

	var seen int32
	for range seq {
		if atomic.AddInt32(&seen, 1) == 200 {
			cancel() // external cancel mid-stream
		}
	}
	if err := join(); !errors.Is(err, context.Canceled) {
		t.Fatalf("join = %v, want context.Canceled", err)
	}
	assertGoroutinesSettle(t, base)
}

// TestMapSeqJoinIgnoresParentCancelAfterCompletion is the regression for the
// codex-found bug: join() must not report Canceled when the parent ctx is
// cancelled AFTER a fully-successful stream. Cancellation is latched at engine
// completion, not sampled when join() happens to be called.
func TestMapSeqJoinIgnoresParentCancelAfterCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	v := FromSlice(iotaSlice(1000), Workers(4), MinPerWorker(1))
	seq, join := MapSeq(ctx, v, func(x int) int { return x })
	got := drain(seq)
	if len(got) != 1000 {
		t.Fatalf("len(got) = %d, want 1000", len(got))
	}
	cancel() // parent cancelled only after the stream fully completed
	if err := join(); err != nil {
		t.Fatalf("join after completed stream then parent cancel = %v, want nil", err)
	}
}

// panickingSeq yields lo..hi-1 then panics — a SOURCE (not f) panic.
func panickingSeq(hi int) func(func(int) bool) {
	return func(yield func(int) bool) {
		for i := 0; i < hi; i++ {
			if !yield(i) {
				return
			}
		}
		panic("source boom")
	}
}

// TestMapSeqSourcePanicContainedAtJoin: a panic in the chunk-pump SOURCE seq must
// be contained by the puller and re-raised at join, not crash the process.
func TestMapSeqSourcePanicContainedAtJoin(t *testing.T) {
	v := FromSeq(panickingSeq(500), Workers(4), MinPerWorker(64))
	seq, join := MapSeq(context.Background(), v, func(x int) int { return x })
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("iteration panicked (%v); source panic should surface at join", r)
			}
		}()
		drain(seq)
	}()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("join did not re-raise the source panic")
		}
		if pe, ok := r.(*PanicError); !ok || pe.Value != "source boom" {
			t.Fatalf("re-raised %v, want *PanicError{source boom}", r)
		}
	}()
	join()
}

// TestChunkPumpSourcePanicContained checks the same containment on the
// MATERIALIZING chunk-pump engine (runChunks): a source panic surfaces as a
// re-raised *PanicError from the terminal, not a process crash.
func TestChunkPumpSourcePanicContained(t *testing.T) {
	v := FromSeq(panickingSeq(500), Workers(4), MinPerWorker(64))
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Count did not re-raise the source panic")
		}
		if pe, ok := r.(*PanicError); !ok || pe.Value != "source boom" {
			t.Fatalf("re-raised %v, want *PanicError{source boom}", r)
		}
	}()
	_, _ = v.Count(context.Background(), func(int) bool { return true })
}

// assertGoroutinesSettle waits (briefly, with retries) for the live goroutine
// count to return near a baseline, tolerating scheduler slack.
func assertGoroutinesSettle(t *testing.T, base int) {
	t.Helper()
	for i := 0; i < 50; i++ {
		if runtime.NumGoroutine() <= base+2 {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("goroutines did not settle: have %d, baseline %d", runtime.NumGoroutine(), base)
}
