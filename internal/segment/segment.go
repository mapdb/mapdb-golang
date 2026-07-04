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
	return SplitIndex(len(xs), n, func(i int) T { return xs[i] })
}

// SplitIndex is the index-addressed form of Split for sources that expose their
// elements by position rather than as a slice — a computed interval, a heap
// array, any type with a Len and an O(1) at(i). It cuts the index space [0,total)
// into the same balanced, contiguous, non-overlapping ranges Split produces, and
// each view yields at(j) for j across its range. at is called only with indices
// in [0,total), so an at that panics out of range is safe here.
//
// Split(xs, n) is exactly SplitIndex(len(xs), n, index-into-xs); routing both
// through one core keeps every family's segmentation identical by construction.
func SplitIndex[T any](total, n int, at func(int) T) []iter.Seq[T] {
	ranges := SplitRanges(total, n)
	segs := make([]iter.Seq[T], len(ranges))
	for i, r := range ranges {
		lo, hi := r[0], r[1] // per-iteration copies; each closure gets its own range
		segs[i] = func(yield func(T) bool) {
			for j := lo; j < hi; j++ {
				if !yield(at(j)) {
					return
				}
			}
		}
	}
	return segs
}

// SplitRanges returns the balanced, contiguous, non-overlapping [lo,hi) index
// ranges that Split and SplitIndex partition [0,total) into: min(n,total) ranges
// (or a single empty [0,0) range when total == 0 or n < 1), with the total%n
// remainder spread one extra index each across the leading ranges. It is the
// range core both splitters delegate to.
//
// Exposed for sources whose elements are NOT one-per-index, where SplitIndex does
// not fit — e.g. a bitset dividing its backing word array, each word range then
// expanding to the set bit positions it contains. Returned as [2]int{lo, hi}
// pairs; the ranges tile [0,total) with no gap or overlap.
func SplitRanges(total, n int) [][2]int {
	if n > total {
		n = total
	}
	if n < 1 {
		n = 1
	}
	ranges := make([][2]int, n)
	chunk := total / n
	remainder := total % n
	lo := 0
	for i := range ranges {
		hi := lo + chunk
		if i < remainder {
			hi++
		}
		ranges[i] = [2]int{lo, hi}
		lo = hi
	}
	return ranges
}
