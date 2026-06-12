package stream

import "iter"

// Any returns true if any element satisfies the predicate.
func Any[V any](seq iter.Seq[V], predicate func(V) bool) bool {
	for v := range seq {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All returns true if all elements satisfy the predicate.
func All[V any](seq iter.Seq[V], predicate func(V) bool) bool {
	for v := range seq {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// None returns true if no element satisfies the predicate.
func None[V any](seq iter.Seq[V], predicate func(V) bool) bool {
	return !Any(seq, predicate)
}

// Contains returns true if the sequence contains the value.
func Contains[V comparable](seq iter.Seq[V], value V) bool {
	for v := range seq {
		if v == value {
			return true
		}
	}
	return false
}

// First returns the first element, or zero value and false if empty.
func First[V any](seq iter.Seq[V]) (V, bool) {
	for v := range seq {
		return v, true
	}
	var zero V
	return zero, false
}

// Last returns the last element, or zero value and false if empty.
func Last[V any](seq iter.Seq[V]) (V, bool) {
	var last V
	found := false
	for v := range seq {
		last = v
		found = true
	}
	return last, found
}
