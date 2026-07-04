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
// sources may return 1). Views are live: mutating the source while segment seqs
// are being consumed is undefined behavior.
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

// defaultMinPerWorker is the provisional per-worker element floor below which an
// input runs sequentially. 13-parallel-design.md §8 replaces this guess with a
// measured crossover point.
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
// the floor runs sequentially. Values < 1 are ignored. Default provisional (§8).
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

// View is a reusable handle over a splittable source plus its execution config.
// It performs no work and starts no goroutines until a terminal op runs. The
// zero View is not usable; construct one with FromSlice or From.
type View[T any] struct {
	// segment produces up to n balanced, re-runnable, non-overlapping segments.
	segment func(n int) []iter.Seq[T]
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

// work is a per-segment task. It must honor ctx cancellation while iterating
// (select on ctx.Done()); the executor relies on it to stop a running segment
// when a sibling fails, cancels, or panics.
type work[T, R any] func(ctx context.Context, seg iter.Seq[T]) (R, error)

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
