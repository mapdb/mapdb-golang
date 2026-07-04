// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import "iter"

// Seq2 is the key-value mirror of Seq: a lazy, chainable sequence of pairs over
// iter.Seq2[K, V]. It ranges directly (for k, v := range s) and carries the small
// method set below; type-changing operations (Map2, MapKeys, …) are free
// functions returning Seq2 or Seq so a pipeline keeps flowing. It is the currency
// for map-shaped data: FromMap, collection map views (treemap.HeadMap, a
// multimap's All), and Enumerate all produce a Seq2.
type Seq2[K, V any] iter.Seq2[K, V]

// Std2 releases s as a standard library iter.Seq2[K, V]. Zero-cost.
func (s Seq2[K, V]) Std2() iter.Seq2[K, V] { return iter.Seq2[K, V](s) }

// Filter returns the pairs of s for which pred holds. Lazy, O(1) memory,
// preserves re-runnability.
func (s Seq2[K, V]) Filter(pred func(K, V) bool) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range s {
			if pred(k, v) && !yield(k, v) {
				return
			}
		}
	}
}

// Keys returns the key half of s as a Seq[K], in the same order. Lazy, O(1)
// memory.
func (s Seq2[K, V]) Keys() Seq[K] {
	return func(yield func(K) bool) {
		for k := range s {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns the value half of s as a Seq[V], in the same order. Lazy, O(1)
// memory.
func (s Seq2[K, V]) Values() Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Swap returns s with keys and values exchanged. Lazy, O(1) memory.
func (s Seq2[K, V]) Swap() Seq2[V, K] {
	return func(yield func(V, K) bool) {
		for k, v := range s {
			if !yield(v, k) {
				return
			}
		}
	}
}

// ForEach calls f on each pair in order. O(n) time, O(1) memory.
func (s Seq2[K, V]) ForEach(f func(K, V)) {
	for k, v := range s {
		f(k, v)
	}
}

// Enumerate pairs each element of s with its zero-based position, yielding
// (0, e0), (1, e1), …. Lazy, O(1) memory. ⟨EC: zipWithIndex⟩
func (s Seq[T]) Enumerate() Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for v := range s {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}

// FromMap returns a Seq2 over the entries of m. Map iteration order is
// unspecified and may vary between runs; the result is re-runnable but not
// order-stable. Lazy in the sense that iteration drives it, O(1) extra memory.
func FromMap[K comparable, V any](m map[K]V) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Map2 transforms each pair into a new pair of possibly different types. Lazy,
// O(1) memory.
func Map2[K, V, K2, V2 any](s Seq2[K, V], f func(K, V) (K2, V2)) Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range s {
			k2, v2 := f(k, v)
			if !yield(k2, v2) {
				return
			}
		}
	}
}

// MapKeys rewrites each key with f, leaving values untouched. Lazy, O(1) memory.
func MapKeys[K, V, K2 any](s Seq2[K, V], f func(K, V) K2) Seq2[K2, V] {
	return func(yield func(K2, V) bool) {
		for k, v := range s {
			if !yield(f(k, v), v) {
				return
			}
		}
	}
}

// MapValues rewrites each value with f, leaving keys untouched. Lazy, O(1)
// memory. ⟨EC: collectValues⟩
func MapValues[K, V, V2 any](s Seq2[K, V], f func(K, V) V2) Seq2[K, V2] {
	return func(yield func(K, V2) bool) {
		for k, v := range s {
			if !yield(k, f(k, v)) {
				return
			}
		}
	}
}

// GroupBy2 regroups a pair sequence by a derived key: it maps each (k, v) to a
// new key g(k, v) and collects the values under it. Eager, O(n) memory; returns
// a re-runnable map. For the common "group a flat Seq[V] by key(v)" case use
// GroupBy in group.go.
func GroupBy2[K, V, G comparable](s Seq2[K, V], g func(K, V) G) map[G][]V {
	out := make(map[G][]V)
	for k, v := range s {
		gk := g(k, v)
		out[gk] = append(out[gk], v)
	}
	return out
}
