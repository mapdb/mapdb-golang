// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package seq

// Terminal methods consume the sequence. Ones that scan the whole input are O(n)
// time; the short-circuiting ones (First, Find, Any, All, None) stop at the
// first decisive element, so they terminate on infinite sources.

// ForEach calls f on each element in order. O(n) time, O(1) memory.
func (s Seq[T]) ForEach(f func(T)) {
	for v := range s {
		f(v)
	}
}

// Into pours every element of s into dst via dst.Add and returns dst, bridging a
// lazy Seq to any collection with the Adder capability (Add(value) bool) — a list,
// set, bag, or any user type with that method. It is the bulk-load counterpart to
// GroupByInto. A free function, not a method, because the sink type cannot be a
// Seq method's type parameter.
//
// The bool that Add returns — insertion info for sets, always true for lists and
// bags — is deliberately IGNORED: pouring duplicates into a set must not truncate
// the pipeline. A stop-capable sink, if ever needed, gets its own protocol (its
// false meaning "stop"); the two meanings never share the Add method (11 §4).
//
// Eager, O(n) over a finite s; an infinite s never returns (bound it with Take
// first).
func Into[T any, A interface{ Add(T) bool }](s Seq[T], dst A) A {
	for v := range s {
		dst.Add(v)
	}
	return dst
}

// ToSlice materializes all elements into a new slice, in order. Eager, O(n).
func (s Seq[T]) ToSlice() []T {
	var out []T
	for v := range s {
		out = append(out, v)
	}
	return out
}

// Count returns the number of elements. O(n) time; does not terminate on an
// infinite source.
func (s Seq[T]) Count() int {
	n := 0
	for range s {
		n++
	}
	return n
}

// CountFunc returns the number of elements that satisfy pred. O(n) time.
func (s Seq[T]) CountFunc(pred func(T) bool) int {
	n := 0
	for v := range s {
		if pred(v) {
			n++
		}
	}
	return n
}

// First returns the first element and true, or the zero value and false if the
// sequence is empty. Short-circuits after one element.
func (s Seq[T]) First() (T, bool) {
	for v := range s {
		return v, true
	}
	var zero T
	return zero, false
}

// Last returns the final element and true, or the zero value and false if the
// sequence is empty. O(n); does not terminate on an infinite source.
func (s Seq[T]) Last() (T, bool) {
	var last T
	found := false
	for v := range s {
		last, found = v, true
	}
	return last, found
}

// Find returns the first element satisfying pred and true, or the zero value and
// false if none does. Short-circuits at the first match. ⟨EC: detect⟩
func (s Seq[T]) Find(pred func(T) bool) (T, bool) {
	for v := range s {
		if pred(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Any reports whether at least one element satisfies pred. Short-circuits at the
// first match; false for an empty sequence. ⟨EC: anySatisfy⟩
func (s Seq[T]) Any(pred func(T) bool) bool {
	for v := range s {
		if pred(v) {
			return true
		}
	}
	return false
}

// All reports whether every element satisfies pred. Short-circuits at the first
// failure; true for an empty sequence (vacuous truth). ⟨EC: allSatisfy⟩
func (s Seq[T]) All(pred func(T) bool) bool {
	for v := range s {
		if !pred(v) {
			return false
		}
	}
	return true
}

// None reports whether no element satisfies pred. Short-circuits at the first
// match; true for an empty sequence. ⟨EC: noneSatisfy⟩
func (s Seq[T]) None(pred func(T) bool) bool {
	for v := range s {
		if pred(v) {
			return false
		}
	}
	return true
}

// Reduce combines the elements left-to-right starting from initial:
// op(op(op(initial, e0), e1), …). Returns initial for an empty sequence. O(n).
// For a result of a different type than the element, use the Fold free function.
// ⟨EC: injectInto⟩
func (s Seq[T]) Reduce(initial T, op func(T, T) T) T {
	acc := initial
	for v := range s {
		acc = op(acc, v)
	}
	return acc
}

// MinFunc returns the minimum element under cmp (cmp(a,b) < 0 means a orders
// before b) and true, or the zero value and false if the sequence is empty.
// O(n).
func (s Seq[T]) MinFunc(cmp func(a, b T) int) (T, bool) {
	var best T
	found := false
	for v := range s {
		if !found || cmp(v, best) < 0 {
			best, found = v, true
		}
	}
	return best, found
}

// MaxFunc returns the maximum element under cmp (cmp(a,b) > 0 means a orders
// after b) and true, or the zero value and false if the sequence is empty. O(n).
func (s Seq[T]) MaxFunc(cmp func(a, b T) int) (T, bool) {
	var best T
	found := false
	for v := range s {
		if !found || cmp(v, best) > 0 {
			best, found = v, true
		}
	}
	return best, found
}
