// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package par

import (
	"context"
	"iter"
	"sync/atomic"
)

// opt is an optional value: a segment result that may or may not hold an element.
type opt[T any] struct {
	v  T
	ok bool
}

// Any reports whether any element satisfies pred. It short-circuits: once one
// segment finds a match, the others stop pulling at their next element boundary
// (they poll a shared flag). The honest limit is the design's stated one — an
// in-flight callback that is mid-execution when the match is found still runs to
// completion, because a plain func(T) bool has no channel to hear the stop on.
func (v View[T]) Any(ctx context.Context, pred func(T) bool) (bool, error) {
	return anyMatch(ctx, v, pred)
}

// All reports whether every element satisfies pred (vacuously true when empty).
// Short-circuits on the first counterexample.
func (v View[T]) All(ctx context.Context, pred func(T) bool) (bool, error) {
	any, err := anyMatch(ctx, v, func(x T) bool { return !pred(x) })
	if err != nil {
		return false, err
	}
	return !any, nil
}

// None reports whether no element satisfies pred (vacuously true when empty).
// Short-circuits on the first match.
func (v View[T]) None(ctx context.Context, pred func(T) bool) (bool, error) {
	any, err := anyMatch(ctx, v, pred)
	if err != nil {
		return false, err
	}
	return !any, nil
}

// anyMatch is the shared short-circuiting existential scan. Workers poll a shared
// atomic.Bool so a match in one segment halts the others at their next element.
func anyMatch[T any](ctx context.Context, v View[T], pred func(T) bool) (bool, error) {
	var found atomic.Bool
	_, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (struct{}, error) {
		for x := range seg {
			if found.Load() {
				return struct{}{}, nil
			}
			if cancelled(cctx) {
				return struct{}{}, cctx.Err()
			}
			if pred(x) {
				found.Store(true)
				return struct{}{}, nil
			}
		}
		return struct{}{}, nil
	})
	if err != nil {
		return false, err
	}
	return found.Load(), nil
}

// Find returns the first element (in segment order) satisfying pred, and whether
// one was found. Each segment scans only until its own first match; the earliest
// segment's match wins. Later segments are not cancelled by an earlier match
// (that would risk cancelling a still-searching earlier segment and losing the
// true first) — cross-segment short-circuit is a later refinement.
func (v View[T]) Find(ctx context.Context, pred func(T) bool) (T, bool, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (opt[T], error) {
		for x := range seg {
			if cancelled(cctx) {
				return opt[T]{}, cctx.Err()
			}
			if pred(x) {
				return opt[T]{x, true}, nil
			}
		}
		return opt[T]{}, nil
	})
	var zero T
	if err != nil {
		return zero, false, err
	}
	for _, p := range parts {
		if p.ok {
			return p.v, true, nil
		}
	}
	return zero, false, nil
}
