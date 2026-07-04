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

// ForEach applies f to every element, fanning out across segments. Unordered:
// calls within a segment follow that segment's order, but calls across segments
// interleave arbitrarily, so f must be safe for concurrent invocation. Returns
// ctx.Err() if cancelled before completion; re-panics (as *PanicError) if f panics.
func (v View[T]) ForEach(ctx context.Context, f func(T)) error {
	_, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (struct{}, error) {
		for x := range seg {
			if cancelled(cctx) {
				return struct{}{}, cctx.Err()
			}
			f(x)
		}
		return struct{}{}, nil
	})
	return err
}

// Count returns the number of elements satisfying pred, summed across segments.
func (v View[T]) Count(ctx context.Context, pred func(T) bool) (int, error) {
	counts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (int, error) {
		c := 0
		for x := range seg {
			if cancelled(cctx) {
				return 0, cctx.Err()
			}
			if pred(x) {
				c++
			}
		}
		return c, nil
	})
	if err != nil {
		return 0, err
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	return total, nil
}

// Filter returns the elements satisfying pred, in source (segment) order. O(n)
// in the number of matches.
func (v View[T]) Filter(ctx context.Context, pred func(T) bool) ([]T, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) ([]T, error) {
		var out []T
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			if pred(x) {
				out = append(out, x)
			}
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return concat(parts), nil
}

// Reduce folds every element into a single value. identity MUST be neutral for
// op and op MUST be associative — each segment folds independently from identity
// and the segment results are combined left-to-right with op. This contract is
// documented, not checked. For a non-associative element step, use Fold, whose
// per-segment accumulator only needs merge to be associative.
func (v View[T]) Reduce(ctx context.Context, identity T, op func(acc, x T) T) (T, error) {
	partials, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (T, error) {
		acc := identity
		for x := range seg {
			if cancelled(cctx) {
				return identity, cctx.Err()
			}
			acc = op(acc, x)
		}
		return acc, nil
	})
	if err != nil {
		return identity, err
	}
	result := identity
	for _, p := range partials {
		result = op(result, p)
	}
	return result, nil
}

// concat flattens per-segment slices into one, preserving segment order.
func concat[T any](parts [][]T) []T {
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	if total == 0 {
		return nil
	}
	out := make([]T, 0, total)
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}
