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

// The …Err ops are the errgroup-shaped twins: the callback is fallible and
// receives the operation's context. The first callback error wins — it cancels
// the internal context, so sibling workers stop and, crucially, a callback
// currently blocked can observe cancellation on the ctx it was handed (this is
// why the twins pass ctx to the callback where the plain terminals do not). A
// callback panic is still contained and re-raised as *PanicError, exactly as for
// the infallible terminals.

// ForEachErr applies f to every element, fanning out across segments. It returns
// the first error any callback produced (siblings are cancelled), or ctx.Err() on
// cancellation, or nil. Like ForEach it is unordered, so f must be safe for
// concurrent invocation.
func (v View[T]) ForEachErr(ctx context.Context, f func(context.Context, T) error) error {
	_, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (struct{}, error) {
		for x := range seg {
			if cancelled(cctx) {
				return struct{}{}, cctx.Err()
			}
			if err := f(cctx, x); err != nil {
				return struct{}{}, err
			}
		}
		return struct{}{}, nil
	})
	return err
}

// FilterErr returns the elements for which pred reports true, in source (segment)
// order. The first pred error aborts the whole operation (siblings cancelled) and
// the partial results are discarded.
func (v View[T]) FilterErr(ctx context.Context, pred func(context.Context, T) (bool, error)) ([]T, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) ([]T, error) {
		var out []T
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			ok, err := pred(cctx, x)
			if err != nil {
				return nil, err
			}
			if ok {
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

// MapErr applies f to every element in parallel and returns the results in source
// (segment) order. The first f error aborts the whole operation (siblings
// cancelled) and no partial slice is returned. A free function, not a method,
// because Go methods cannot introduce the new type parameter R.
func MapErr[T, R any](ctx context.Context, v View[T], f func(context.Context, T) (R, error)) ([]R, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) ([]R, error) {
		var out []R
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			r, err := f(cctx, x)
			if err != nil {
				return nil, err
			}
			out = append(out, r)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return concat(parts), nil
}
