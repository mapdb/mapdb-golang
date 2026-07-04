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

// Numeric is the constraint for numeric reductions (Sum): every built-in integer
// and floating-point kind. Defined locally so par stays free of collection deps.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// Sum adds every element, folding per-segment and combining the partials. A free
// function because the Numeric constraint cannot be attached to a View[T] method.
// Floating-point summation is order-dependent, so the result may differ slightly
// from a sequential sum across worker counts.
func Sum[T Numeric](ctx context.Context, v View[T]) (T, error) {
	partials, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (T, error) {
		var s T
		for x := range seg {
			if cancelled(cctx) {
				var z T
				return z, cctx.Err()
			}
			s += x
		}
		return s, nil
	})
	var total T
	if err != nil {
		return total, err
	}
	for _, p := range partials {
		total += p
	}
	return total, nil
}

// MinFunc returns the least element under less (less(a, b) reports a < b) and
// whether the source was non-empty. Ties resolve to the earliest by segment
// order. A free function to keep the comparator signature explicit.
func MinFunc[T any](ctx context.Context, v View[T], less func(a, b T) bool) (T, bool, error) {
	return extremum(ctx, v, func(candidate, best T) bool { return less(candidate, best) })
}

// MaxFunc returns the greatest element under less (less(a, b) reports a < b) and
// whether the source was non-empty. Ties resolve to the earliest by segment order.
func MaxFunc[T any](ctx context.Context, v View[T], less func(a, b T) bool) (T, bool, error) {
	return extremum(ctx, v, func(candidate, best T) bool { return less(best, candidate) })
}

// extremum folds each segment to its best element (per better) then reduces the
// per-segment bests. better(candidate, best) reports that candidate should
// replace best; earliest-by-segment-order wins ties because replacement requires
// strictly-better, never equal.
func extremum[T any](ctx context.Context, v View[T], better func(candidate, best T) bool) (T, bool, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (opt[T], error) {
		var best opt[T]
		for x := range seg {
			if cancelled(cctx) {
				return opt[T]{}, cctx.Err()
			}
			if !best.ok || better(x, best.v) {
				best = opt[T]{x, true}
			}
		}
		return best, nil
	})
	var zero T
	if err != nil {
		return zero, false, err
	}
	var best opt[T]
	for _, p := range parts {
		if p.ok && (!best.ok || better(p.v, best.v)) {
			best = p
		}
	}
	return best.v, best.ok, nil
}
