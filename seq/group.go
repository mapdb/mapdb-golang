// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

import (
	"container/heap"
	"fmt"
	"iter"

	"github.com/mapdb/mapdb-golang/object"
)

// This file holds the pair-producing and grouping free functions: the ops whose
// result is a Seq2, a map keyed by a derived key, or a collection sink. They are
// free functions because each introduces a type parameter or a constraint that a
// Go method cannot express.

// Zip pairs a and b positionally: (a0, b0), (a1, b1), …, stopping when either
// runs dry. It ranges a and Pulls b (one iter.Pull, torn down when the range
// ends), so only b pays the Pull cost. Lazy; re-runnable when a and b are.
func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B] {
	return func(yield func(A, B) bool) {
		next, stop := iter.Pull(iter.Seq[B](b))
		defer stop()
		for av := range a {
			bv, ok := next()
			if !ok {
				return
			}
			if !yield(av, bv) {
				return
			}
		}
	}
}

// Pairwise yields each consecutive overlapping pair of s: (e0, e1), (e1, e2), ….
// A sequence of fewer than two elements yields nothing. Lazy, O(1) memory.
func Pairwise[V any](s Seq[V]) Seq2[V, V] {
	return func(yield func(V, V) bool) {
		var prev V
		has := false
		for v := range s {
			if has && !yield(prev, v) {
				return
			}
			prev, has = v, true
		}
	}
}

// Window yields sliding windows of exactly n consecutive elements, advancing one
// element at a time: [e0..e(n-1)], [e1..en], …. Sequences shorter than n yield
// nothing. Each window is a fresh copy, safe to retain. Lazy in the source,
// O(n) memory. Panics if n <= 0.
func Window[V any](s Seq[V], n int) iter.Seq[[]V] {
	if n <= 0 {
		panic("seq: Window size must be positive")
	}
	return func(yield func([]V) bool) {
		buf := make([]V, 0, n)
		for v := range s {
			if len(buf) == n {
				buf = append(buf[:0], buf[1:]...) // drop the oldest
			}
			buf = append(buf, v)
			if len(buf) == n {
				win := make([]V, n)
				copy(win, buf)
				if !yield(win) {
					return
				}
			}
		}
	}
}

// GroupBy collects the elements of s into a HashMultimap keyed by key(v),
// preserving encounter order within each group. Eager, O(n) memory. ⟨EC: groupBy⟩
func GroupBy[V any, K comparable](s Seq[V], key func(V) K) *object.HashMultimap[K, V] {
	m := object.NewHashMultimap[K, V]()
	for v := range s {
		m.Put(key(v), v)
	}
	return m
}

// GroupByInto is GroupBy with a caller-supplied sink: any type with Put(K, V),
// including the primitive multimaps, so they are first-class targets rather than
// forced through object.HashMultimap. Returns dst for chaining. Eager, O(n).
func GroupByInto[V any, K comparable, M interface{ Put(K, V) }](s Seq[V], key func(V) K, dst M) M {
	for v := range s {
		dst.Put(key(v), v)
	}
	return dst
}

// CountBy tallies how many elements fall under each key. Eager, O(distinct-keys)
// memory. ⟨EC: countBy⟩
func CountBy[V any, K comparable](s Seq[V], key func(V) K) map[K]int {
	out := make(map[K]int)
	for v := range s {
		out[key(v)]++
	}
	return out
}

// SumBy sums val(v) under each key(v). Eager, O(distinct-keys) memory.
// ⟨EC: sumByInt/…⟩
func SumBy[V any, K comparable, N Numeric](s Seq[V], key func(V) K, val func(V) N) map[K]N {
	out := make(map[K]N)
	for v := range s {
		out[key(v)] += val(v)
	}
	return out
}

// AggregateBy is the general reduce-by-key: for each element it selects a group
// key(v), lazily creates that group's accumulator with newAcc on first sight,
// and folds the element in with acc. Eager, O(distinct-keys) memory.
// ⟨EC: aggregateBy⟩
func AggregateBy[V any, K comparable, A any](s Seq[V], key func(V) K, newAcc func() A, acc func(A, V) A) map[K]A {
	out := make(map[K]A)
	for v := range s {
		k := key(v)
		cur, ok := out[k]
		if !ok {
			cur = newAcc()
		}
		out[k] = acc(cur, v)
	}
	return out
}

// Average returns the arithmetic mean of the elements as a float64 and true, or
// 0 and false for an empty sequence. Eager, O(n). ⟨EC: averageOf⟩
func Average[V Numeric](s Seq[V]) (float64, bool) {
	var total float64
	n := 0
	for v := range s {
		total += float64(v)
		n++
	}
	if n == 0 {
		return 0, false
	}
	return total / float64(n), true
}

// TopK returns the k elements that rank highest under cmp, in descending order
// (greatest first). It keeps a bounded min-heap of size k, so it runs in
// O(n log k) time and O(k) memory — no full sort. Returns nil for k <= 0; a
// shorter sequence yields all its elements, still sorted. ⟨EC: topOccurrences⟩
func TopK[V any](s Seq[V], k int, cmp object.Comparator[V]) []V {
	if k <= 0 {
		return nil
	}
	h := &topkHeap[V]{cmp: cmp}
	for v := range s {
		if h.Len() < k {
			heap.Push(h, v)
		} else if cmp(v, h.data[0]) > 0 {
			h.data[0] = v
			heap.Fix(h, 0)
		}
	}
	out := make([]V, h.Len())
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(h).(V)
	}
	return out
}

// topkHeap is a min-heap on cmp (root = smallest retained), so pushing past
// capacity evicts the current minimum and only the k largest survive.
type topkHeap[V any] struct {
	data []V
	cmp  object.Comparator[V]
}

func (h *topkHeap[V]) Len() int           { return len(h.data) }
func (h *topkHeap[V]) Less(i, j int) bool { return h.cmp(h.data[i], h.data[j]) < 0 }
func (h *topkHeap[V]) Swap(i, j int)      { h.data[i], h.data[j] = h.data[j], h.data[i] }
func (h *topkHeap[V]) Push(x any)         { h.data = append(h.data, x.(V)) }
func (h *topkHeap[V]) Pop() any {
	old := h.data
	n := len(old)
	v := old[n-1]
	h.data = old[:n-1]
	return v
}

// DuplicatePolicy tells ToMap what to do when a key repeats.
type DuplicatePolicy int

const (
	// ErrorOnDuplicate makes ToMap return an error on the first repeated key.
	ErrorOnDuplicate DuplicatePolicy = iota
	// KeepFirst keeps the value first seen for a key and ignores later ones.
	KeepFirst
	// KeepLast lets a later value overwrite an earlier one for the same key.
	KeepLast
)

// ToMap collects a pair sequence into a map, resolving repeated keys per policy:
// ErrorOnDuplicate returns an error at the first collision; KeepFirst keeps the
// earliest value; KeepLast keeps the latest. Eager, O(distinct-keys) memory.
func ToMap[K comparable, V any](s Seq2[K, V], policy DuplicatePolicy) (map[K]V, error) {
	out := make(map[K]V)
	for k, v := range s {
		if _, exists := out[k]; exists {
			switch policy {
			case ErrorOnDuplicate:
				return nil, fmt.Errorf("seq: ToMap: duplicate key %v", k)
			case KeepFirst:
				continue
			case KeepLast:
				// fall through to overwrite below
			}
		}
		out[k] = v
	}
	return out, nil
}
