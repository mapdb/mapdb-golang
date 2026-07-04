// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

// Package segment holds the balanced slice-splitting helper shared by the
// parallel layer (par.FromSlice) and every slice-backed collection's Segments
// method, so a single definition backs the Segmenter capability rather than one
// stamped into each generated file. A single source of truth also makes the
// contract structural: par.FromSlice(xs) and par.From(listOver(xs)) split
// identically because they call the same function.
package segment

import "iter"

// Split cuts xs into up to n balanced, contiguous, non-overlapping index-range
// views whose concatenation covers every element of xs exactly once — the
// Segmenter contract par.From rests on. It returns k = min(n, len(xs)) views,
// or a single (possibly empty) view when xs is empty or n <= 1. The remainder
// (len(xs) mod n) is spread one extra element each across the first views, so
// segment lengths differ by at most one.
//
// Each view is a re-runnable iter.Seq over a fixed range of xs and holds no copy
// (O(1) memory per view). The views are live over xs: mutating xs while a view
// is being consumed is undefined behavior — the same rule as any Segmenter.
func Split[T any](xs []T, n int) []iter.Seq[T] {
	total := len(xs)
	if n > total {
		n = total
	}
	if n < 1 {
		n = 1
	}
	segs := make([]iter.Seq[T], n)
	chunk := total / n
	remainder := total % n
	lo := 0
	for i := range segs {
		hi := lo + chunk
		if i < remainder {
			hi++
		}
		l, h := lo, hi // snapshot the range for this view's closure
		segs[i] = func(yield func(T) bool) {
			for j := l; j < h; j++ {
				if !yield(xs[j]) {
					return
				}
			}
		}
		lo = hi
	}
	return segs
}
