
package stream

import "iter"

// Reduce performs a left fold over the sequence.
func Reduce[V any](seq iter.Seq[V], initial V, accumulator func(V, V) V) V {
	result := initial
	for v := range seq {
		result = accumulator(result, v)
	}
	return result
}

// Count returns the number of elements in the sequence.
func Count[V any](seq iter.Seq[V]) int {
	n := 0
	for range seq {
		n++
	}
	return n
}

// Sum returns the sum of elements (for numeric types).
func Sum[V interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}](seq iter.Seq[V]) V {
	var sum V
	for v := range seq {
		sum += v
	}
	return sum
}

// Min returns the minimum element, or zero and false if empty.
func Min[V interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}](seq iter.Seq[V]) (V, bool) {
	first := true
	var min V
	for v := range seq {
		if first || v < min {
			min = v
			first = false
		}
	}
	return min, !first
}

// Max returns the maximum element, or zero and false if empty.
func Max[V interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64
}](seq iter.Seq[V]) (V, bool) {
	first := true
	var max V
	for v := range seq {
		if first || v > max {
			max = v
			first = false
		}
	}
	return max, !first
}
