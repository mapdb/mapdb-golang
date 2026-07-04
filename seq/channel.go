// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import "context"

// This is the ONLY file in seq that touches goroutines and channels (design
// principle P7). Channels ARE Go's blocking queue, so these adapters bridge to
// them rather than reinventing a queue type. Every function here either takes a
// context.Context (non-nil; use context.Background() if you have none) or is torn
// down by the consumer breaking the range.

// FromChannel adopts a receive channel as a Seq: it yields each value until ch is
// closed. Single-shot — ranging drains ch, and a second range sees only what is
// left. Breaking the range stops reading but never closes ch (the sender owns
// that). O(1) memory.
func FromChannel[T any](ch <-chan T) Seq[T] {
	return func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	}
}

// ToChannel runs s on a new goroutine and returns a buffered channel (capacity
// buf) of its elements; the channel is closed when s is exhausted or ctx is
// cancelled. The caller MUST either drain the channel to completion or cancel ctx
// — otherwise, if s outproduces a consumer that walks away, the goroutine blocks
// forever on send and leaks. O(buf) memory plus one goroutine.
func ToChannel[T any](ctx context.Context, s Seq[T], buf int) <-chan T {
	out := make(chan T, buf)
	go func() {
		defer close(out)
		for v := range s {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Buffered overlaps production with consumption: each range starts a producer
// goroutine that runs s into an n-deep channel while the consumer pulls from it,
// so an IO-backed or CPU-heavy source computes its next elements ahead of demand.
// The producer is torn down when the consumer breaks the range or ctx is
// cancelled — no leak. Re-runnable exactly when s is (each range spawns a fresh
// producer); single-shot when s is. O(n) memory plus one goroutine per active
// range.
func Buffered[T any](ctx context.Context, s Seq[T], n int) Seq[T] {
	return func(yield func(T) bool) {
		ch := make(chan T, n)
		done := make(chan struct{})
		go func() {
			defer close(ch)
			for v := range s {
				select {
				case ch <- v:
				case <-done: // consumer stopped
					return
				case <-ctx.Done():
					return
				}
			}
		}()
		// Signal the producer to stop on any exit (consumer break or ctx done),
		// so it never blocks forever on a send into an abandoned channel.
		defer close(done)
		for {
			select {
			case v, ok := <-ch:
				if !ok {
					return // producer finished (s exhausted or ctx cancelled)
				}
				if !yield(v) {
					return // consumer stopped; defer close(done) tears the producer down
				}
			case <-ctx.Done():
				return
			}
		}
	}
}
