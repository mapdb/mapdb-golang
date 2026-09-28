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
//
// The returned builtin map uses Go == key identity, not the bit-pattern
// identity the collection spec requires for float keys (algorithms.md "NaN
// must hash and compare by bit pattern"). For a key type whose underlying
// type is float32 or float64 this loses information: a pair keyed -0.0 and a
// pair keyed +0.0 collapse into one entry (the later value wins), and every
// NaN-keyed pair becomes its own entry that no lookup can reach (NaN != NaN;
// such entries are visible only by ranging over the map). When keys may be
// floats, collect into an object.HashMap (object.NewHashMap[K, V]() and Put),
// which keeps float keys by bit pattern.
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
