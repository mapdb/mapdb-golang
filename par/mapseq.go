// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"iter"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// MapSeq is the STREAMING parallel map: it applies f across the source in
// parallel and returns results through an [iter.Seq] as each worker produces
// them, rather than materializing a []R like [Map]. It returns two values:
//
//   - the result seq — UNORDERED (completion order, not source order) and
//     SINGLE-SHOT (iterating it drives the workers once; a second iteration
//     yields nothing).
//   - a join func() error — the JOIN POINT. Call it after the seq is exhausted
//     (or abandoned). It waits for the workers to finish, returns ctx.Err() if
//     the operation was cancelled, and RE-PANICS (wrapped in [*PanicError]) any
//     panic a worker's f raised. A bare seq return has nowhere to surface a
//     worker panic or a cancellation, which is why the join point exists.
//
// Nothing runs until the seq is first iterated (lazy, like every View terminal);
// if the seq is never iterated, join() returns nil. Abandoning the seq early
// (breaking the range) cancels the workers and drains them before the range
// returns, so no goroutine leaks — but results already computed are discarded.
//
// A free function, not a method, because Go methods cannot introduce R. f is
// infallible (func(T) R); the only reportable error is cancellation. Works over
// both engines: a segment source streams one goroutine per segment; a chunk-pump
// source ([FromSeq]) streams its worker pool. Either way concurrency ≤ Workers.
func MapSeq[T, R any](ctx context.Context, v View[T], f func(T) R) (iter.Seq[R], func() error) {
	if ctx == nil {
		ctx = context.Background()
	}

	workers := v.cfg.workers
	if workers < 1 {
		workers = 1
	}

	cctx, cancel := context.WithCancel(ctx)
	results := make(chan R, 2*workers)
	allDone := make(chan struct{}) // closed once every producer goroutine has exited

	var (
		mu       sync.Mutex
		firstErr error
		panicVal *PanicError
		engaged  atomic.Bool // true once the engine has been started
	)
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

	// emit sends one result, or reports that the op is winding down (cancelled or
	// the consumer abandoned the seq) so the caller stops promptly.
	emit := func(r R) bool {
		select {
		case results <- r:
			return true
		case <-cctx.Done():
			return false
		}
	}

	// finalize runs once, after every producer has exited, to close the result
	// stream and the join barrier. It LATCHES external cancellation here — at
	// engine completion — rather than in join(): sampling ctx.Err() when join() is
	// eventually called would falsely report Canceled if the caller cancels the
	// parent ctx after a fully-successful stream but before calling join. (Our own
	// teardown cancels cctx, not ctx, so a clean run/abandon latches nil.)
	finalize := func() {
		mu.Lock()
		if firstErr == nil {
			if err := ctx.Err(); err != nil {
				firstErr = err
			}
		}
		mu.Unlock()
		close(results)
		close(allDone)
	}

	start := func() {
		engaged.Store(true)
		if cctx.Err() != nil {
			// Pre-cancelled: no work; finalize latches ctx.Err() and closes so the
			// seq yields nothing and join reports the cancellation.
			finalize()
			return
		}

		var wg sync.WaitGroup // producers whose completion means no more sends

		// segWork runs one segment (or one chunk) as a seq, emitting f(x) per
		// element, with the same panic + runtime.Goexit containment as runSegments:
		// a callback that panics or Goexits must be reported, not silently dropped.
		segWork := func(s iter.Seq[T]) {
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
			for x := range s {
				if cancelled(cctx) {
					normal = true // cooperative stop, not a callback failure
					return
				}
				if !emit(f(x)) {
					normal = true
					return
				}
			}
			normal = true
		}

		if v.pump != nil || v.pumpCtx != nil {
			// Chunk-pump streaming: one puller drains the source into chunks; the
			// worker pool maps each chunk's elements and streams them.
			source := v.pump
			if v.pumpCtx != nil {
				source = v.pumpCtx(cctx)
			}
			chunkSize := v.cfg.minPerWorker
			if chunkSize < 1 {
				chunkSize = 1
			}
			chunks := make(chan []T, 2*workers)

			var pwg sync.WaitGroup
			pwg.Add(1)
			go func() {
				defer pwg.Done()
				defer close(chunks)
				// Contain a panic from the SOURCE seq itself (not just f): the segment
				// engine catches source panics inside its guarded worker, so the pump
				// puller must too, or a source panic crashes the process instead of
				// surfacing at join. Recover runs before close(chunks)/pwg.Done so the
				// panic is latched and cctx cancelled before workers drain out.
				//
				// The normal sentinel matches the chunk worker below: a source that
				// exits via runtime.Goexit unwinds this goroutine with recover() nil
				// and close(chunks) still running, so join would report success over a
				// truncated stream.
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
				send := func() bool {
					select {
					case chunks <- buf:
						buf = make([]T, 0, chunkSize)
						return true
					case <-cctx.Done():
						return false
					}
				}
				for x := range source {
					if cancelled(cctx) {
						normal = true // cancellation is a natural stop, not a Goexit
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
					send()
				}
				normal = true // source drained; a Goexit inside it skips this line
			}()

			wg.Add(workers)
			for i := 0; i < workers; i++ {
				go func() {
					defer wg.Done()
					for {
						var chunk []T
						select {
						case c, ok := <-chunks:
							if !ok {
								return
							}
							chunk = c
						case <-cctx.Done():
							return
						}
						segWork(sliceSeq(chunk))
						if cancelled(cctx) {
							return
						}
					}
				}()
			}

			go func() {
				wg.Wait()  // all workers done → no more sends
				pwg.Wait() // and the puller has exited (leak-free)
				finalize()
			}()
			return
		}

		// Segment streaming: one goroutine per balanced segment.
		n := v.cfg.segmentCount(v.size)
		segs := v.segment(n)
		wg.Add(len(segs))
		for _, s := range segs {
			go func(s iter.Seq[T]) {
				defer wg.Done()
				segWork(s)
			}(s)
		}
		go func() {
			wg.Wait()
			finalize()
		}()
	}

	var startOnce sync.Once
	seq := func(yield func(R) bool) {
		first := false
		startOnce.Do(func() { first = true; start() })
		if !first {
			return // single-shot: the source was already consumed
		}
		// A consumer that panics mid-range unwinds past us; cancel on the way out so
		// producers tear down instead of blocking forever on a full channel. On a
		// normal or abandoned exit this is a harmless no-op (already cancelled/drained).
		defer cancel()
		for r := range results {
			if !yield(r) {
				cancel()            // consumer abandoned; stop producing
				for range results { // drain in-flight sends until the closer closes
				}
				return
			}
		}
	}

	join := func() error {
		if !engaged.Load() {
			return nil // seq never iterated: no work started, nothing to report
		}
		<-allDone // workers finished; firstErr/panicVal are stable (happens-before)
		if panicVal != nil {
			panic(panicVal)
		}
		// firstErr already reflects any worker error AND (latched by finalize at
		// engine completion) external cancellation — so a parent cancel that lands
		// after the stream finished is not misreported.
		return firstErr
	}

	return seq, join
}
