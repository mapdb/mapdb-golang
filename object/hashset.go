// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"fmt"
	"iter"
	"strings"
)

// HashSet is a generic unordered set backed by a Go map.
// It implements MutableSet[T].
//
// Float elements (T with underlying type float32/float64) use bit-pattern
// identity, not ==: NaN is found and not duplicated, NaN payloads are
// distinct elements, and -0.0 and +0.0 are distinct elements; see keyIndex.
type HashSet[T comparable] struct {
	m keyIndex[T, struct{}]
}

// NewHashSet creates an empty HashSet.
func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{m: newKeyIndex[T, struct{}](0)}
}

// NewHashSetFrom creates a HashSet from existing elements.
func NewHashSetFrom[T comparable](values ...T) *HashSet[T] {
	s := &HashSet[T]{m: newKeyIndex[T, struct{}](len(values))}
	for _, v := range values {
		s.m.put(v, struct{}{})
	}
	return s
}

// ── Sized ─────────────────────────────────────────────────────────────

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *HashSet[T]) Len() int { return s.m.len() }

// ── Iterable ──────────────────────────────────────────────────────────

func (s *HashSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s.m.all() {
			if !yield(v) {
				return
			}
		}
	}
}

func (s *HashSet[T]) ForEach(f func(T)) {
	for v := range s.m.all() {
		f(v)
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (s *HashSet[T]) Contains(value T) bool {
	return s.m.contains(value)
}

func (s *HashSet[T]) AnySatisfy(predicate func(T) bool) bool {
	for v := range s.m.all() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *HashSet[T]) AllSatisfy(predicate func(T) bool) bool {
	for v := range s.m.all() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *HashSet[T]) NoneSatisfy(predicate func(T) bool) bool {
	for v := range s.m.all() {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ── Convertible ───────────────────────────────────────────────────────

func (s *HashSet[T]) ToSlice() []T {
	result := make([]T, 0, s.m.len())
	for v := range s.m.all() {
		result = append(result, v)
	}
	return result
}

// ── MutableSet ────────────────────────────────────────────────────────

func (s *HashSet[T]) Add(value T) bool {
	if s.m.contains(value) {
		return false // keep the stored element (matters for == -equal composites)
	}
	s.m.put(value, struct{}{})
	return true
}

func (s *HashSet[T]) Remove(value T) bool {
	_, existed := s.m.remove(value)
	return existed
}

func (s *HashSet[T]) Clear() {
	s.m.clear()
}

// ── Set operations ────────────────────────────────────────────────────

// Union returns a new set containing all elements from both sets.
func (s *HashSet[T]) Union(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m.all() {
		result.m.put(v, struct{}{})
	}
	for v := range other.m.all() {
		result.m.put(v, struct{}{})
	}
	return result
}

// Intersect returns a new set containing only elements present in both sets.
func (s *HashSet[T]) Intersect(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	smaller, larger := s, other
	if smaller.m.len() > larger.m.len() {
		smaller, larger = larger, smaller
	}
	for v := range smaller.m.all() {
		if larger.m.contains(v) {
			result.m.put(v, struct{}{})
		}
	}
	return result
}

// Difference returns a new set containing elements in s but not in other.
func (s *HashSet[T]) Difference(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m.all() {
		if !other.m.contains(v) {
			result.m.put(v, struct{}{})
		}
	}
	return result
}

// SymmetricDifference returns elements in either set but not both.
func (s *HashSet[T]) SymmetricDifference(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m.all() {
		if !other.m.contains(v) {
			result.m.put(v, struct{}{})
		}
	}
	for v := range other.m.all() {
		if !s.m.contains(v) {
			result.m.put(v, struct{}{})
		}
	}
	return result
}

// ── Functional operations ─────────────────────────────────────────────

func (s *HashSet[T]) Select(predicate func(T) bool) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m.all() {
		if predicate(v) {
			result.m.put(v, struct{}{})
		}
	}
	return result
}

func (s *HashSet[T]) Reject(predicate func(T) bool) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m.all() {
		if !predicate(v) {
			result.m.put(v, struct{}{})
		}
	}
	return result
}

func (s *HashSet[T]) Detect(predicate func(T) bool) (T, bool) {
	for v := range s.m.all() {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func (s *HashSet[T]) Count(predicate func(T) bool) int {
	n := 0
	for v := range s.m.all() {
		if predicate(v) {
			n++
		}
	}
	return n
}

// ── Stringer ──────────────────────────────────────────────────────────

func (s *HashSet[T]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for v := range s.m.all() {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", v)
		first = false
	}
	b.WriteByte('}')
	return b.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableSet[int] = (*HashSet[int])(nil)
var _ MutableSet[string] = (*HashSet[string])(nil)
