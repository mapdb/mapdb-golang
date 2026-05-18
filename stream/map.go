
package stream

import "iter"

// Map transforms each element using the given function.
func Map[V, W any](seq iter.Seq[V], transform func(V) W) iter.Seq[W] {
	return func(yield func(W) bool) {
		for v := range seq {
			if !yield(transform(v)) {
				return
			}
		}
	}
}

// Map2 transforms each key-value pair using the given function.
func Map2[K1, V1, K2, V2 any](seq iter.Seq2[K1, V1], transform func(K1, V1) (K2, V2)) iter.Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		for k, v := range seq {
			k2, v2 := transform(k, v)
			if !yield(k2, v2) {
				return
			}
		}
	}
}

// FlatMap transforms each element into a sequence and concatenates the results.
func FlatMap[V, W any](seq iter.Seq[V], transform func(V) iter.Seq[W]) iter.Seq[W] {
	return func(yield func(W) bool) {
		for v := range seq {
			for w := range transform(v) {
				if !yield(w) {
					return
				}
			}
		}
	}
}

// Zip combines two sequences into a sequence of pairs.
func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		nextB, stopB := iter.Pull(b)
		defer stopB()
		for va := range a {
			vb, ok := nextB()
			if !ok {
				return
			}
			if !yield(va, vb) {
				return
			}
		}
	}
}

// Enumerate adds indices to elements: (0, first), (1, second), ...
func Enumerate[V any](seq iter.Seq[V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0
		for v := range seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}
