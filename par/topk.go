// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"container/heap"
	"context"
	"iter"
)

// TopK returns the k elements that rank highest under less (less(a, b) reports
// a < b), in descending order (greatest first). Each segment keeps a bounded
// min-heap of size k as it iterates — O(n log k) time, O(workers·k) memory, no
// full sort — and a merge phase selects the global top-k from the per-segment
// survivors. Returns an empty slice for k <= 0 or an empty source; a source with
// fewer than k elements yields all of them, still sorted.
//
// A free function (not a View method) to keep the comparator explicit, matching
// [MinFunc]/[MaxFunc]. less takes a bool rather than an object.Comparator so par
// stays free of collection-family imports.
//
// Ties at the k-th boundary are resolved arbitrarily: which of several equal
// elements makes the cut can vary with segmentation and worker count. This is
// inherent to parallel bounded selection — do not rely on tie ordering.
func TopK[T any](ctx context.Context, v View[T], k int, less func(a, b T) bool) ([]T, error) {
	if k <= 0 {
		return []T{}, nil
	}
	parts, err := run(ctx, v, func(cctx context.Context, seg iter.Seq[T]) ([]T, error) {
		h := &topkHeap[T]{less: less}
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			h.offer(x, k)
		}
		return h.data, nil
	})
	if err != nil {
		return nil, err
	}
	// Merge phase: fold every per-segment survivor set (each ≤ k) into one bounded
	// heap, then drain it greatest-first. Merge input is ≤ workers·k, independent
	// of source size.
	merged := &topkHeap[T]{less: less}
	for _, p := range parts {
		for _, x := range p {
			merged.offer(x, k)
		}
	}
	out := make([]T, merged.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(merged).(T) // pop the current minimum → fills the tail
	}
	return out, nil
}

// topkHeap is a min-heap on less (root = the smallest retained element), so once
// it holds k elements, offering a larger one evicts the current minimum and only
// the k largest survive. Mirrors seq.TopK's heap but keyed on a less-bool.
type topkHeap[T any] struct {
	data []T
	less func(a, b T) bool
}

func (h *topkHeap[T]) Len() int           { return len(h.data) }
func (h *topkHeap[T]) Less(i, j int) bool { return h.less(h.data[i], h.data[j]) }
func (h *topkHeap[T]) Swap(i, j int)      { h.data[i], h.data[j] = h.data[j], h.data[i] }
func (h *topkHeap[T]) Push(x any)         { h.data = append(h.data, x.(T)) }
func (h *topkHeap[T]) Pop() any {
	old := h.data
	n := len(old)
	v := old[n-1]
	h.data = old[:n-1]
	return v
}

// offer admits x into the top-k: pushed directly while below capacity, otherwise
// it replaces the current minimum only if it is strictly greater.
func (h *topkHeap[T]) offer(x T, k int) {
	if h.Len() < k {
		heap.Push(h, x)
	} else if h.less(h.data[0], x) {
		h.data[0] = x
		heap.Fix(h, 0)
	}
}
