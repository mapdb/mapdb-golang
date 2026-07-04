// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import (
	"cmp"
	"iter"
)

// These are the type-changing and constrained operations that a Go method cannot
// express (methods may not introduce type parameters or constraints). Each
// returns a Seq so a chain resumes immediately: seq.Map(s, f).Filter(...).

// Map applies f to each element, producing a Seq of the results. Lazy, O(1)
// memory. ⟨EC: collect⟩
func Map[V, W any](s Seq[V], f func(V) W) Seq[W] {
	return func(yield func(W) bool) {
		for v := range s {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// MapWhere applies f only to elements satisfying pred, a fused filter+map. Lazy,
// O(1) memory. ⟨EC: collectIf⟩
func MapWhere[V, W any](s Seq[V], pred func(V) bool, f func(V) W) Seq[W] {
	return func(yield func(W) bool) {
		for v := range s {
			if pred(v) && !yield(f(v)) {
				return
			}
		}
	}
}

// FlatMap maps each element to a sub-sequence and concatenates them. Lazy, O(1)
// memory beyond each sub-sequence. ⟨EC: flatCollect⟩
func FlatMap[V, W any](s Seq[V], f func(V) iter.Seq[W]) Seq[W] {
	return func(yield func(W) bool) {
		for v := range s {
			for w := range f(v) {
				if !yield(w) {
					return
				}
			}
		}
	}
}

// Distinct yields the elements of s in first-seen order, skipping later
// duplicates. Lazy (streams and short-circuits) but O(distinct) memory: it holds
// a set of the values seen so far.
func Distinct[V comparable](s Seq[V]) Seq[V] {
	return func(yield func(V) bool) {
		seen := make(map[V]struct{})
		for v := range s {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// DistinctBy yields elements whose key(v) is seen for the first time, in
// encounter order. Lazy, O(distinct-keys) memory.
func DistinctBy[V any, K comparable](s Seq[V], key func(V) K) Seq[V] {
	return func(yield func(V) bool) {
		seen := make(map[K]struct{})
		for v := range s {
			k := key(v)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// Fold is the general reduction: it threads an accumulator of any type A through
// the sequence left-to-right, returning acc(acc(…acc(initial, e0)…), en-1).
// Returns initial for an empty sequence. Eager, O(n) time. ⟨EC: injectInto⟩
func Fold[V, A any](s Seq[V], initial A, acc func(A, V) A) A {
	res := initial
	for v := range s {
		res = acc(res, v)
	}
	return res
}

// Scan is the lazy running fold: it yields initial, then each successive
// accumulator acc(state, e). So Scan(Of(1,2,3), 0, +) yields 0,1,3,6. Lazy,
// O(1) memory.
func Scan[V, A any](s Seq[V], initial A, acc func(A, V) A) Seq[A] {
	return func(yield func(A) bool) {
		state := initial
		if !yield(state) {
			return
		}
		for v := range s {
			state = acc(state, v)
			if !yield(state) {
				return
			}
		}
	}
}

// Partition splits s into the elements matching pred and the rest, preserving
// order within each. Eager: it consumes s once and returns two re-runnable Seqs.
// O(n) memory.
func Partition[V any](s Seq[V], pred func(V) bool) (matching, rest Seq[V]) {
	var yes, no []V
	for v := range s {
		if pred(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return FromSlice(yes), FromSlice(no)
}

// Sum returns the sum of the elements, or the zero value for an empty sequence.
// Eager, O(n). ⟨EC: sumOf⟩
func Sum[V Numeric](s Seq[V]) V {
	var total V
	for v := range s {
		total += v
	}
	return total
}

// Min returns the smallest element and true, or the zero value and false if the
// sequence is empty. Eager, O(n). For a custom order use the MinFunc method.
func Min[V cmp.Ordered](s Seq[V]) (V, bool) {
	var best V
	found := false
	for v := range s {
		if !found || v < best {
			best, found = v, true
		}
	}
	return best, found
}

// Max returns the largest element and true, or the zero value and false if the
// sequence is empty. Eager, O(n). For a custom order use the MaxFunc method.
func Max[V cmp.Ordered](s Seq[V]) (V, bool) {
	var best V
	found := false
	for v := range s {
		if !found || v > best {
			best, found = v, true
		}
	}
	return best, found
}
