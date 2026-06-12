package stream

import "iter"

// ToSlice collects all elements from the sequence into a slice.
func ToSlice[V any](seq iter.Seq[V]) []V {
	var result []V
	for v := range seq {
		result = append(result, v)
	}
	return result
}

// ToMap collects key-value pairs from an iter.Seq2 into a Go map.
func ToMap[K comparable, V any](seq iter.Seq2[K, V]) map[K]V {
	result := make(map[K]V)
	for k, v := range seq {
		result[k] = v
	}
	return result
}

// ForEach calls the function for each element.
func ForEach[V any](seq iter.Seq[V], f func(V)) {
	for v := range seq {
		f(v)
	}
}

// ForEach2 calls the function for each key-value pair.
func ForEach2[K, V any](seq iter.Seq2[K, V], f func(K, V)) {
	for k, v := range seq {
		f(k, v)
	}
}
