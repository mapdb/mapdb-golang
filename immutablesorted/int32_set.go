// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK -- THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package immutablesorted

import "github.com/mapdb/mapdb-golang/rangev"

// ImmutableInt32SortedSet is a compact immutable sorted set backed by a single
// packed ascending int32 array, queried by binary search. The element analogue
// of ImmutableInt32Int32SortedMap.
type ImmutableInt32SortedSet struct {
	elems []int32
}

// FromSortedInt32 builds a set from a strictly-ascending element slice (COPIED
// -- a snapshot independent of the caller's slice).
//
// It PANICS if the elements are out-of-order or contain a duplicate. Empty and
// single-element input are valid.
func FromSortedInt32(elements []int32) *ImmutableInt32SortedSet {
	assertStrictlyAscendingInt32(elements)
	e := make([]int32, len(elements))
	copy(e, elements)
	return &ImmutableInt32SortedSet{elems: e}
}

// Size returns the number of elements.
func (s *ImmutableInt32SortedSet) Size() int { return len(s.elems) }

// IsEmpty reports whether the set has no elements.
func (s *ImmutableInt32SortedSet) IsEmpty() bool { return len(s.elems) == 0 }

// Contains reports whether elem is present.
func (s *ImmutableInt32SortedSet) Contains(elem int32) bool {
	_, ok := searchInt32(s.elems, elem)
	return ok
}

// First returns the minimum element and true, or (0, false) if empty.
func (s *ImmutableInt32SortedSet) First() (int32, bool) {
	if len(s.elems) == 0 {
		return 0, false
	}
	return s.elems[0], true
}

// Last returns the maximum element and true, or (0, false) if empty.
func (s *ImmutableInt32SortedSet) Last() (int32, bool) {
	if len(s.elems) == 0 {
		return 0, false
	}
	return s.elems[len(s.elems)-1], true
}

// Floor returns the greatest element <= k and true, or (0, false).
func (s *ImmutableInt32SortedSet) Floor(k int32) (int32, bool) {
	i, ok := searchInt32(s.elems, k)
	if ok {
		return s.elems[i], true
	}
	if i == 0 {
		return 0, false
	}
	return s.elems[i-1], true
}

// Ceiling returns the least element >= k and true, or (0, false).
func (s *ImmutableInt32SortedSet) Ceiling(k int32) (int32, bool) {
	i, _ := searchInt32(s.elems, k)
	if i >= len(s.elems) {
		return 0, false
	}
	return s.elems[i], true
}

// Lower returns the greatest element < k (strict) and true, or (0, false).
func (s *ImmutableInt32SortedSet) Lower(k int32) (int32, bool) {
	i, _ := searchInt32(s.elems, k)
	if i == 0 {
		return 0, false
	}
	return s.elems[i-1], true
}

// Higher returns the least element > k (strict) and true, or (0, false).
func (s *ImmutableInt32SortedSet) Higher(k int32) (int32, bool) {
	i, ok := searchInt32(s.elems, k)
	if ok {
		i++
	}
	if i >= len(s.elems) {
		return 0, false
	}
	return s.elems[i], true
}

// Rank returns the number of elements strictly less than elem (lower-bound
// index, in 0..=Size()). Defined for present and absent elements.
func (s *ImmutableInt32SortedSet) Rank(elem int32) int {
	i, _ := searchInt32(s.elems, elem)
	return i
}

// Select returns the i-th smallest element (0-based) and true, or (0, false) if
// i < 0 or i >= Size().
func (s *ImmutableInt32SortedSet) Select(i int) (int32, bool) {
	if i < 0 || i >= len(s.elems) {
		return 0, false
	}
	return s.elems[i], true
}

// Elements returns the elements in ascending order (a fresh snapshot copy).
func (s *ImmutableInt32SortedSet) Elements() []int32 {
	out := make([]int32, len(s.elems))
	copy(out, s.elems)
	return out
}

// DescendingElements returns all elements, descending.
func (s *ImmutableInt32SortedSet) DescendingElements() []int32 {
	n := len(s.elems)
	out := make([]int32, n)
	for i, e := range s.elems {
		out[n-1-i] = e
	}
	return out
}

// RangeElements returns the elements in range, ascending. Bracketed by two
// binary searches from the range's cut semantics (overflow-safe at the signed
// extremes). A fresh snapshot copy.
func (s *ImmutableInt32SortedSet) RangeElements(r rangev.Int32Range) []int32 {
	lo, hi := r.Bracket(s.elems)
	out := make([]int32, hi-lo)
	copy(out, s.elems[lo:hi])
	return out
}

// DescendingRangeElements returns the elements in range, descending.
func (s *ImmutableInt32SortedSet) DescendingRangeElements(r rangev.Int32Range) []int32 {
	lo, hi := r.Bracket(s.elems)
	n := hi - lo
	out := make([]int32, n)
	for i := 0; i < n; i++ {
		out[n-1-i] = s.elems[lo+i]
	}
	return out
}
