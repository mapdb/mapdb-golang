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

// Map applies f to every element in parallel and returns the results in source
// (segment) order. A free function, not a method, because Go methods cannot
// introduce the new type parameter R.
func Map[T, R any](ctx context.Context, v View[T], f func(T) R) ([]R, error) {
	parts, err := run(ctx, v, func(cctx context.Context, seg iter.Seq[T]) ([]R, error) {
		var out []R
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			out = append(out, f(x))
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return concat(parts), nil
}

// Fold is the associative-merge reduction with a per-segment accumulator. Each
// segment starts a fresh accumulator via newAcc, folds its elements with acc,
// and the segment accumulators are combined left-to-right with merge. Only merge
// must be associative and newAcc()'s result neutral for it — acc itself may be
// order-dependent within a segment. This is the escape hatch from Reduce's
// stronger associativity requirement.
func Fold[T, A any](ctx context.Context, v View[T], newAcc func() A, acc func(A, T) A, merge func(A, A) A) (A, error) {
	partials, err := run(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (A, error) {
		a := newAcc()
		for x := range seg {
			if cancelled(cctx) {
				var zero A
				return zero, cctx.Err()
			}
			a = acc(a, x)
		}
		return a, nil
	})
	if err != nil {
		var zero A
		return zero, err
	}
	result := newAcc()
	for _, p := range partials {
		result = merge(result, p)
	}
	return result, nil
}
