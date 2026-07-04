// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"iter"
)

// This file adds the key/value (pair) parallel layer on top of the single-value
// engine. A Segmenter2 splits a map-like source into iter.Seq2 segments; From2
// adapts those into the single-value engine by carrying each (key, value) as a
// pair, so runSegments — with all its panic/cancel/order guarantees — drives them
// unchanged. No second executor exists; View2 and the pair terminals are thin
// projections over View[pair].

// Segmenter2 is the pair analog of Segmenter: a key/value source that can cut
// itself into k ≤ n independently iterable, re-runnable, non-overlapping Seq2
// views that together cover it. The same live-view rule applies — the boundaries
// are fixed at the Segments2 call, so mutating the source before the returned
// views are exhausted or discarded is undefined behavior. Ordered sources
// (a sorted map) return ordered segments.
type Segmenter2[K, V any] interface {
	Segments2(n int) []iter.Seq2[K, V]
}

// pair carries a key/value element through the single-value engine. It is
// unexported: callers see only the (K, V) callbacks of View2's terminals, never
// the wrapper.
type pair[K, V any] struct {
	k K
	v V
}

// seg2Adapter presents a Segmenter2 as a Segmenter[pair] by wrapping each Seq2
// segment's (k, v) yields into pair values. The wrapping is lazy per segment, so
// segments stay re-runnable and no pairs are materialized ahead of iteration.
type seg2Adapter[K, V any] struct{ src Segmenter2[K, V] }

func (a seg2Adapter[K, V]) Segments(n int) []iter.Seq[pair[K, V]] {
	s2 := a.src.Segments2(n)
	segs := make([]iter.Seq[pair[K, V]], len(s2))
	for i := range s2 {
		seg := s2[i] // capture this segment for the closure
		segs[i] = func(yield func(pair[K, V]) bool) {
			for k, v := range seg {
				if !yield(pair[K, V]{k, v}) {
					return
				}
			}
		}
	}
	return segs
}

// View2 is the pair analog of View: a reusable handle over a key/value source and
// its execution config, performing no work until a terminal runs. Construct one
// with From2. Its terminals mirror View's but take (key, value) callbacks.
type View2[K, V any] struct {
	inner View[pair[K, V]]
}

// From2 builds a View2 over a Segmenter2's (key, value) pairs. Options (Workers,
// MinPerWorker) apply exactly as for From. The source size is treated as unknown,
// like From — a Segmenter2 that wants a sequential fallback for small inputs
// returns a single segment itself.
func From2[K, V any](src Segmenter2[K, V], opts ...Option) View2[K, V] {
	return View2[K, V]{inner: From[pair[K, V]](seg2Adapter[K, V]{src}, opts...)}
}

// ForEach applies f to every (key, value) pair, fanning out across segments.
// Unordered and concurrent, with the same contract as View.ForEach: f must be
// safe for concurrent invocation; returns ctx.Err() if cancelled; re-panics (as
// *PanicError) if f panics.
func (v View2[K, V]) ForEach(ctx context.Context, f func(K, V)) error {
	return v.inner.ForEach(ctx, func(p pair[K, V]) { f(p.k, p.v) })
}

// Count returns the number of pairs satisfying pred, summed across segments.
func (v View2[K, V]) Count(ctx context.Context, pred func(K, V) bool) (int, error) {
	return v.inner.Count(ctx, func(p pair[K, V]) bool { return pred(p.k, p.v) })
}

// Fold2 combines the pairs into an accumulator A: each segment folds its own
// pairs from newAcc() via acc, then the per-segment accumulators combine with
// merge (which must be associative). It is the pair analog of Fold and the
// general pair reducer. Free function, not a method, because A is an extra type
// parameter.
func Fold2[K, V, A any](
	ctx context.Context,
	v View2[K, V],
	newAcc func() A,
	acc func(A, K, V) A,
	merge func(A, A) A,
) (A, error) {
	return Fold(ctx, v.inner, newAcc, func(a A, p pair[K, V]) A { return acc(a, p.k, p.v) }, merge)
}
