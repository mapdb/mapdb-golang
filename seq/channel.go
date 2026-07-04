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
//
// # Cancellation is cooperative — the crucial caveat
//
// These adapters run a source Seq on a goroutine with a loop of the shape
// "for v := range s { select send-or-cancel }". Cancellation (ctx or consumer
// break) can only be observed AT the select — i.e. between elements, or while the
// producer is blocked trying to hand an element off. It CANNOT preempt the source
// itself: if s is blocked INSIDE its own body waiting for the next element before
// it calls yield again — the classic case being FromChannel over an open, idle
// channel — the producer goroutine stays parked there and neither ctx nor a
// consumer break can reach it until s finally yields or returns. This is inherent
// to the pull-based iter.Seq model; a generic adapter has no preemption hook into
// s. The escape hatch is a cancellation-aware source: use FromChannelCtx (not
// FromChannel) for channel/IO-backed inputs, so cancellation is honored even while
// the source is idle.

// FromChannel adopts a receive channel as a Seq: it yields each value until ch is
// closed. Single-shot — ranging drains ch, and a second range sees only what is
// left. Breaking the range stops reading but never closes ch (the sender owns
// that). O(1) memory.
//
// Not cancellation-aware: a range blocked on an open, idle ch waits until a value
// arrives or ch closes — nothing else can wake it. When feeding this into
// ToChannel/Buffered under a Context, use FromChannelCtx instead so cancellation
// is honored while idle (see the package caveat above).
func FromChannel[T any](ch <-chan T) Seq[T] {
	return func(yield func(T) bool) {
		for v := range ch {
			if !yield(v) {
				return
			}
		}
	}
}

// FromChannelCtx is the cancellation-aware FromChannel: it yields values until ch
// closes, the consumer breaks, or ctx is cancelled — whichever comes first, even
// while ch is idle. This is the source to use for channel/IO-backed inputs that
// flow into ToChannel/Buffered, because it lets cancellation actually tear the
// producer down (plain FromChannel cannot — see the package caveat). Single-shot,
// O(1) memory. ctx must be non-nil.
func FromChannelCtx[T any](ctx context.Context, ch <-chan T) Seq[T] {
	return func(yield func(T) bool) {
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

// ToChannel runs s on a new goroutine and returns a buffered channel (capacity
// buf) of its elements; the channel is closed when s is exhausted or ctx is
// cancelled. The caller MUST either drain the channel to completion or cancel ctx
// — otherwise, if s outproduces a consumer that walks away, the goroutine blocks
// forever on send and leaks. O(buf) memory plus one goroutine.
//
// Cancellation is cooperative (see the package caveat): cancelling ctx stops the
// producer and closes the channel promptly only while the producer is at its send
// or s cooperatively returns; if s is blocked inside its own body before the next
// yield (e.g. plain FromChannel on an idle channel), the goroutine stays parked
// there until s unblocks. Use a cancellation-aware source (FromChannelCtx) for
// such inputs to make cancel-then-close reliable.
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
// The producer is torn down when the consumer breaks the range or ctx is cancelled
// — with the cooperative-cancellation caveat (see package doc): teardown is prompt
// while the producer is at its channel send or s returns after yield reports stop,
// but a source blocked inside its own body before the next yield (plain
// FromChannel on an idle channel) parks the goroutine until it unblocks. Use
// FromChannelCtx for such sources. Re-runnable exactly when s is (each range spawns
// a fresh producer); single-shot when s is. O(n) memory plus one goroutine per
// active range.
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
