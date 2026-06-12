package stream

import (
	"iter"
	"slices"

	"github.com/mapdb/mapdb-golang/object"
)

// GroupBy groups elements by a key function, returning an
// object.HashMultimap so callers get a first-class library-native type
// instead of a raw Go map. Size/SizeDistinct/All iteration/ForEach
// work as on any HashMultimap.
//
// An earlier version returned map[K][]V; that shape forced callers out
// of the library's functional API for the grouped result. The rename
// is source-breaking for that handful of callers — use GroupByToMap
// below if you need the raw map shape.
func GroupBy[V any, K comparable](seq iter.Seq[V], keyFunc func(V) K) *object.HashMultimap[K, V] {
	result := object.NewHashMultimap[K, V]()
	for v := range seq {
		result.Put(keyFunc(v), v)
	}
	return result
}

// GroupByToMap is the escape hatch for callers that genuinely want a
// bare map[K][]V (e.g. for marshalling) rather than a HashMultimap.
func GroupByToMap[V any, K comparable](seq iter.Seq[V], keyFunc func(V) K) map[K][]V {
	result := make(map[K][]V)
	for v := range seq {
		key := keyFunc(v)
		result[key] = append(result[key], v)
	}
	return result
}

// Partition splits elements into two sequences: the elements matching
// the predicate and the elements that do not.
//
// The source seq is consumed exactly once and the predicate is called
// exactly once per element. The result is materialised eagerly — the
// returned seqs are backed by two slices and are therefore re-runnable
// even when the input is a single-shot seq (e.g. one backed by a
// channel, a generator with state, or any pull-based source).
//
// This matches Eclipse Collections Java's PartitionIterable contract,
// where the selected/rejected sides are already materialised by the
// time partition() returns. An earlier implementation re-ran the
// source seq twice under two separate Filters; that silently broke
// single-shot seqs and doubled the predicate-call count on re-runnable
// seqs.
func Partition[V any](seq iter.Seq[V], predicate func(V) bool) (matching iter.Seq[V], notMatching iter.Seq[V]) {
	var yes, no []V
	for v := range seq {
		if predicate(v) {
			yes = append(yes, v)
		} else {
			no = append(no, v)
		}
	}
	return slices.Values(yes), slices.Values(no)
}

// Distinct returns a sequence with duplicate elements removed.
func Distinct[V comparable](seq iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		seen := make(map[V]struct{})
		for v := range seq {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Chunk breaks the sequence into sub-slices of at most n elements.
func Chunk[V any](seq iter.Seq[V], n int) iter.Seq[[]V] {
	return func(yield func([]V) bool) {
		chunk := make([]V, 0, n)
		for v := range seq {
			chunk = append(chunk, v)
			if len(chunk) == n {
				if !yield(chunk) {
					return
				}
				chunk = make([]V, 0, n)
			}
		}
		if len(chunk) > 0 {
			if !yield(chunk) {
				return
			}
		}
	}
}

// Chain concatenates multiple sequences into one.
func Chain[V any](seqs ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, seq := range seqs {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}
