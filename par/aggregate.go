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

// These keyed reducers materialize a plain map[K]… rather than a collection
// family, so par stays free of the collection import cycle. Each worker builds a
// private map over its segment; the merge phase folds the per-segment maps in
// segment order. GroupBy into an *object.HashMultimap is the family-returning
// variant, gated on the Phase-2 par↔collection direction.

// CountBy tallies elements by key: the returned map holds, for each key produced
// by key, how many elements mapped to it. Free function because K is a fresh type
// parameter. Returns an empty (non-nil) map for an empty source.
func CountBy[T any, K comparable](ctx context.Context, v View[T], key func(T) K) (map[K]int, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (map[K]int, error) {
		m := make(map[K]int)
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			m[key(x)]++
		}
		return m, nil
	})
	if err != nil {
		return nil, err
	}
	out := make(map[K]int)
	for _, p := range parts {
		for k, c := range p {
			out[k] += c
		}
	}
	return out, nil
}

// AggregateBy folds elements into a per-key accumulator. For each element, key
// selects a bucket; within a segment the bucket is folded with acc (which may be
// order-dependent). Per-segment buckets are then combined by merge, which MUST be
// associative — the result for a key is the left-to-right merge of the segments'
// buckets in segment order. newAcc creates a fresh accumulator the first time a
// key appears in a segment; it need not be a merge identity. Returns an empty
// (non-nil) map for an empty source.
func AggregateBy[T any, K comparable, A any](
	ctx context.Context,
	v View[T],
	key func(T) K,
	newAcc func() A,
	acc func(A, T) A,
	merge func(A, A) A,
) (map[K]A, error) {
	parts, err := runSegments(ctx, v, func(cctx context.Context, seg iter.Seq[T]) (map[K]A, error) {
		m := make(map[K]A)
		for x := range seg {
			if cancelled(cctx) {
				return nil, cctx.Err()
			}
			k := key(x)
			a, ok := m[k]
			if !ok {
				a = newAcc()
			}
			m[k] = acc(a, x)
		}
		return m, nil
	})
	if err != nil {
		return nil, err
	}
	out := make(map[K]A)
	for _, p := range parts {
		for k, a := range p {
			if cur, ok := out[k]; ok {
				out[k] = merge(cur, a)
			} else {
				out[k] = a
			}
		}
	}
	return out, nil
}
