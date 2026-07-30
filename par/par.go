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
	"runtime/debug"
	"sync"
)

// errGoexit is reported for a worker that terminated via runtime.Goexit (e.g. a
// callback calling t.FailNow / testify require) rather than returning. Without
// this the worker's segment would silently contribute a zero result.
var errGoexit = errors.New("par: worker exited abnormally (runtime.Goexit)")

// Segmenter is the one capability the parallel design rests on: a source that
// can cut itself into k ≤ n independently iterable, re-runnable, non-overlapping
// views that together cover it. k may be less than n (small or hard-to-split
// sources may return 1). Views are live and their boundaries are fixed from the
// source's state at the Segments call: mutating the source any time after
// Segments returns and before the returned views are exhausted or discarded is
// undefined behavior.
type Segmenter[T any] interface {
	Segments(n int) []iter.Seq[T]
}

// PanicError wraps a value recovered from a panicking worker callback. The
// executor re-raises it (via panic) on the caller's goroutine, preserving the
// original value and the worker's stack.
type PanicError struct {
	Value any    // the value passed to panic in the worker
	Stack []byte // debug.Stack() captured at the recover site
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("par: panic in parallel worker: %v\n\n%s", e.Value, e.Stack)
}

// Unwrap exposes the panic value as an error when it is one, so errors.As/Is can
// reach through a re-raised PanicError. Returns nil when the value is not an error.
func (e *PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

// defaultMinPerWorker is the per-worker element floor below which an input runs
// sequentially. 1024 sits at the trivial-callback crossover measured by
// BenchmarkCountCrossover (§8): on a 32-core host, ~750–1024 elements/worker is
// where a worker's ~1µs spawn/schedule overhead stops dominating a trivial
// callback. It is a floor, not a target — callers with expensive callbacks lower
// it via MinPerWorker (their crossover is far smaller).
const defaultMinPerWorker = 1024

// Option configures a View's execution. Options are applied at construction.
type Option func(*config)

type config struct {
	workers      int
	minPerWorker int
}

// Workers sets the maximum number of goroutines (and thus segments) a terminal
// fans out to. Values < 1 are ignored. Default: runtime.GOMAXPROCS(0).
func Workers(n int) Option {
	return func(c *config) {
		if n >= 1 {
			c.workers = n
		}
	}
}

// MinPerWorker sets the minimum elements per worker; a sized source with fewer
// than this many elements per potential worker uses fewer workers, and one below
// the floor runs sequentially. Values < 1 are ignored. Default 1024 (measured;
// see defaultMinPerWorker) — lower it for expensive per-element callbacks.
func MinPerWorker(n int) Option {
	return func(c *config) {
		if n >= 1 {
			c.minPerWorker = n
		}
	}
}

func newConfig(opts []Option) config {
	c := config{workers: runtime.GOMAXPROCS(0), minPerWorker: defaultMinPerWorker}
	if c.workers < 1 {
		c.workers = 1
	}
	for _, o := range opts {
		o(&c)
	}
	return c
}

// segmentCount decides how many segments to request for a source of the given
// element count (size < 0 means unknown). It never exceeds workers, and for a
// known size it also caps by the per-worker floor so tiny inputs stay sequential.
func (c config) segmentCount(size int) int {
	w := c.workers
	if w < 1 {
		w = 1
	}
	if size < 0 {
		return w // unknown size: trust the worker budget
	}
	if size == 0 {
		return 0
	}
	mpw := c.minPerWorker
	if mpw < 1 {
		mpw = 1
	}
	maxByLoad := (size + mpw - 1) / mpw // ceil(size / mpw)
	if maxByLoad < w {
		w = maxByLoad
	}
	if w < 1 {
		w = 1
	}
	return w
}

// View is a reusable handle over a source plus its execution config. It performs
// no work and starts no goroutines until a terminal op runs. The zero View is not
// usable; construct one with FromSlice, From, or FromSeq.
//
// A View is either segment-based (segment != nil — a splittable source, terminals
// preserve segment order) or a chunk-pump (pump != nil — an unsplittable single-
// shot seq, terminals are unordered; see FromSeq). Exactly one is set.
type View[T any] struct {
	// segment produces up to n balanced, re-runnable, non-overlapping segments.
	segment func(n int) []iter.Seq[T]
	// pump is the single-shot source for the chunk-pump execution model (§6).
	pump iter.Seq[T]
	// pumpCtx is the ctx-aware chunk-pump source: it is handed the engine's
	// internal context so a blocking receive can select on cancellation and
	// short-circuit teardown (FromSeqCtx). Exactly one of pump/pumpCtx is set on a
	// chunk-pump view.
	pumpCtx func(ctx context.Context) iter.Seq[T]
	// size is the element count if known, else -1 (drives the sequential fallback).
	size int
	cfg  config
}

// From builds a View over any Segmenter. The source's size is treated as unknown,
// so the MinPerWorker sequential fallback does not apply — a Segmenter that wants
// it should return a single segment for small inputs itself.
func From[T any](src Segmenter[T], opts ...Option) View[T] {
	return View[T]{
		segment: src.Segments,
		size:    -1,
		cfg:     newConfig(opts),
	}
}

// ── internal executor ─────────────────────────────────────────────────────

// work is a per-unit task (a segment, or a chunk of the pump). It must honor ctx
// cancellation while iterating (select on ctx.Done()); the executor relies on it
// to stop a running unit when a sibling fails, cancels, or panics.
type work[T, R any] func(ctx context.Context, seg iter.Seq[T]) (R, error)

// run dispatches to the segment or chunk-pump engine and collects per-unit
// results. Segment results are in segment order; chunk-pump results are in
// completion order (unordered). Used by the non-short-circuiting terminals.
func run[T, R any](ctx context.Context, v View[T], w work[T, R]) ([]R, error) {
	return runEarly(ctx, v, w, nil)
}

// runEarly is run with an optional early-stop predicate for short-circuiting
// terminals (Any/Find). earlyDone is consulted only by the chunk-pump puller, so
// it can stop draining an unbounded source once the terminal has its answer; the
// segment engine ignores it (segments are finite and its workers short-circuit
// via their own polled flag).
func runEarly[T, R any](ctx context.Context, v View[T], w work[T, R], earlyDone func() bool) ([]R, error) {
	if v.pump != nil || v.pumpCtx != nil {
		return runChunks(ctx, v, w, earlyDone)
	}
	return runSegments(ctx, v, w)
}

// runSegments splits v, runs work on each segment across a bounded pool, and
// returns the per-segment results in segment order. The first error or panic
// cancels the siblings; a panic is re-raised (wrapped in *PanicError) on the
// caller's goroutine after all workers drain.
func runSegments[T, R any](ctx context.Context, v View[T], w work[T, R]) ([]R, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := v.cfg.segmentCount(v.size)
	if n == 0 {
		return nil, nil
	}
	segs := v.segment(n)
	if len(segs) == 0 {
		return nil, nil
	}
	if len(segs) == 1 {
		// Sequential fast path: no goroutine, but the same panic wrapping so
		// callers see uniform *PanicError semantics regardless of worker count.
		r, pv, err := runOne(ctx, segs[0], w)
		if pv != nil {
			panic(pv)
		}
		if err != nil {
			return nil, err
		}
		return []R{r}, nil
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]R, len(segs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	var panicVal *PanicError
	fail := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
			cancel()
		}
		mu.Unlock()
	}
	failPanic := func(rec any, stack []byte) {
		mu.Lock()
		if panicVal == nil {
			panicVal = &PanicError{Value: rec, Stack: stack}
			cancel()
		}
		mu.Unlock()
	}

	wg.Add(len(segs))
	for i, s := range segs {
		go func(i int, s iter.Seq[T]) {
			defer wg.Done()
			normal := false // distinguishes a natural return from runtime.Goexit
			defer func() {
				if rec := recover(); rec != nil {
					failPanic(rec, debug.Stack())
					return
				}
				if !normal {
					fail(errGoexit)
				}
			}()
			r, err := w(cctx, s)
			normal = true // w returned; a Goexit inside w skips this line
			if err != nil {
				fail(err)
				return
			}
			results[i] = r
		}(i, s)
	}
	wg.Wait()

	if panicVal != nil {
		panic(panicVal)
	}
	if firstErr != nil {
		return nil, firstErr
	}
	// No post-hoc ctx.Err() check: cancellation is cooperative — a worker that
	// observes it while iterating reports it via firstErr. An operation that runs
	// to completion before cancellation lands returns its result (errgroup-style,
	// the prior art doc.go cites), and both worker-count paths agree.
	return results, nil
}

// runOne runs a single segment inline, converting a panic into a *PanicError so
// the single-segment path matches the multi-segment path's contract. (A
// runtime.Goexit inside w unwinds the caller's own goroutine — loud, not the
// silent data loss the multi-segment path guards against, so no flag is needed.)
func runOne[T, R any](ctx context.Context, seg iter.Seq[T], w work[T, R]) (r R, pv *PanicError, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			pv = &PanicError{Value: rec, Stack: debug.Stack()}
		}
	}()
	r, err = w(ctx, seg)
	return
}

// cancelled reports whether ctx is done, without blocking. Terminals call it in
// their per-element loop so a running segment stops promptly on sibling failure.
func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// runChunks is the chunk-pump engine (§6): one puller goroutine drains the
// single-shot pump seq into MinPerWorker-sized []T chunks over a bounded channel;
// Workers goroutines consume chunks and run w on each (as a seq over the chunk).
// Per-chunk results accumulate in COMPLETION ORDER (unordered). Backpressure is
// the channel depth (2×workers).
//
// Semantics vs the segment engine (documented on FromSeq): results are unordered,
// and external cancellation IS reported via a final ctx.Err() check — chunk-pump
// workers stop by ceasing to receive rather than by returning an error, so
// nothing else would surface it.
//
// Cancellation reaches the source only cooperatively: the puller can stop the
// source at a yield boundary, and — for a ctx-aware source (pumpCtx / FromSeqCtx)
// bound to cctx — a blocking receive can select on cctx and unblock. A short-
// circuit (earlyDone) cancels cctx so it tears such a source down; the resulting
// context.Canceled artifact is discounted at the join (short-circuit still
// succeeds). A plain pump that blocks between yields is only stopped once it
// yields again — see FromSeq.
//
// Panic/error containment matches runSegments: first panic wins (re-raised as
// *PanicError after all goroutines drain), first real error cancels the rest, and
// a worker that exits via runtime.Goexit is reported as errGoexit rather than
// silently dropping its chunk.
func runChunks[T, R any](ctx context.Context, v View[T], w work[T, R], earlyDone func() bool) ([]R, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	workers := v.cfg.workers
	if workers < 1 {
		workers = 1
	}
	chunkSize := v.cfg.minPerWorker
	if chunkSize < 1 {
		chunkSize = 1
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	source := v.pump
	if v.pumpCtx != nil {
		source = v.pumpCtx(cctx) // bind a ctx-aware source to the engine ctx
	}

	chunks := make(chan []T, 2*workers)

	var mu sync.Mutex
	var results []R
	var firstErr error
	var panicVal *PanicError
	fail := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
			cancel()
		}
		mu.Unlock()
	}
	failPanic := func(rec any, stack []byte) {
		mu.Lock()
		if panicVal == nil {
			panicVal = &PanicError{Value: rec, Stack: stack}
			cancel()
		}
		mu.Unlock()
	}

	// Puller: drain the source into chunks. Stops on cancellation or early-done,
	// closes the channel on exit so workers terminate.
	var pwg sync.WaitGroup
	pwg.Add(1)
	go func() {
		defer pwg.Done()
		defer close(chunks)
		// Contain a panic from the SOURCE seq itself. The segment engine catches
		// source panics inside its guarded worker (runSegments); the pump puller
		// runs the source outside any worker, so without this a source panic would
		// crash the process instead of surfacing as *PanicError. Recover runs before
		// close(chunks)/pwg.Done so the panic is latched and cctx cancelled first.
		//
		// The normal sentinel is the same one the segment and chunk workers use: a
		// source that exits via runtime.Goexit unwinds this goroutine silently —
		// recover() returns nil and close(chunks) still runs — so without it the
		// terminal would report a clean EOF over a truncated stream.
		normal := false
		defer func() {
			if rec := recover(); rec != nil {
				failPanic(rec, debug.Stack())
				return
			}
			if !normal {
				fail(errGoexit)
			}
		}()
		buf := make([]T, 0, chunkSize)
		send := func() bool { // returns false if the op is winding down
			select {
			case chunks <- buf:
				buf = make([]T, 0, chunkSize)
				return true
			case <-cctx.Done():
				return false
			}
		}
		for x := range source {
			if cancelled(cctx) || (earlyDone != nil && earlyDone()) {
				normal = true // winding down is a natural stop, not a Goexit
				return
			}
			buf = append(buf, x)
			if len(buf) == chunkSize {
				if !send() {
					normal = true
					return
				}
			}
		}
		if len(buf) > 0 {
			send() // final partial chunk; result ignored (we close regardless)
		}
		normal = true // source drained; a Goexit inside it skips this line
	}()

	// Workers: consume chunks, run w on each, accumulate results in completion order.
	var wwg sync.WaitGroup
	wwg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wwg.Done()
			for {
				var chunk []T
				select {
				case c, ok := <-chunks:
					if !ok {
						return // channel closed and drained
					}
					chunk = c
				case <-cctx.Done():
					return
				}
				// Process one chunk with the same panic + Goexit containment as the
				// segment worker: a callback that exits via runtime.Goexit must be
				// reported (errGoexit), not silently drop the chunk's result.
				normal := false
				func() {
					defer func() {
						if rec := recover(); rec != nil {
							failPanic(rec, debug.Stack())
							return
						}
						if !normal {
							fail(errGoexit)
						}
					}()
					r, err := w(cctx, sliceSeq(chunk))
					normal = true
					if err != nil {
						fail(err)
						return
					}
					mu.Lock()
					results = append(results, r)
					mu.Unlock()
				}()
				// Short-circuit: cancel cctx so a ctx-aware source and the siblings
				// tear down promptly (the join discounts the resulting cancel).
				if earlyDone != nil && earlyDone() {
					cancel()
					return
				}
			}
		}()
	}

	wwg.Wait()
	pwg.Wait()

	if panicVal != nil {
		panic(panicVal)
	}
	if firstErr != nil {
		// A short-circuit cancels cctx, so a sibling mid-chunk may report
		// context.Canceled; that is an artifact of our own teardown, not a real
		// failure, so discount it when the terminal actually short-circuited.
		// Real callback errors (and external cancel, caught below) still surface.
		shortCircuited := earlyDone != nil && earlyDone()
		if !(shortCircuited && errors.Is(firstErr, context.Canceled)) {
			return nil, firstErr
		}
	}
	// Chunk-pump workers stop by not receiving, not by erroring, so external
	// cancellation is surfaced here. earlyDone does not cancel the CALLER ctx, so a
	// short-circuit success leaves it live and is not misreported.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// sliceSeq is a re-runnable iter.Seq over xs (a materialized chunk).
func sliceSeq[T any](xs []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, x := range xs {
			if !yield(x) {
				return
			}
		}
	}
}
