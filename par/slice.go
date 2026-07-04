// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import "iter"

// FromSlice builds a View over xs. The slice is split into contiguous index
// ranges (O(1) per segment); the slice is not copied, so mutating it while a
// terminal runs is undefined behavior — the same live-view rule as any Segmenter.
// The slice's length is known, so small inputs fall back to sequential execution
// below MinPerWorker.
func FromSlice[T any](xs []T, opts ...Option) View[T] {
	return View[T]{
		segment: func(n int) []iter.Seq[T] { return sliceSegments(xs, n) },
		size:    len(xs),
		cfg:     newConfig(opts),
	}
}

// sliceSegments cuts xs into up to n balanced contiguous segments. Each segment
// is a re-runnable iter.Seq over a fixed index range, so the View is reusable.
func sliceSegments[T any](xs []T, n int) []iter.Seq[T] {
	batches := splitBatches(len(xs), n)
	segs := make([]iter.Seq[T], len(batches))
	for i, b := range batches {
		lo, hi := b.lo, b.hi
		segs[i] = func(yield func(T) bool) {
			for j := lo; j < hi; j++ {
				if !yield(xs[j]) {
					return
				}
			}
		}
	}
	return segs
}

type batch struct{ lo, hi int }

// splitBatches partitions [0,n) into taskCount balanced contiguous ranges,
// distributing the remainder across the first ranges. Mirrors parallel.splitBatches.
func splitBatches(n, taskCount int) []batch {
	if taskCount > n {
		taskCount = n
	}
	if taskCount < 1 {
		taskCount = 1
	}
	batches := make([]batch, taskCount)
	chunkSize := n / taskCount
	remainder := n % taskCount
	lo := 0
	for i := range batches {
		hi := lo + chunkSize
		if i < remainder {
			hi++
		}
		batches[i] = batch{lo, hi}
		lo = hi
	}
	return batches
}
