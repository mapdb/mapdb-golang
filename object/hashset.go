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
type HashSet[T comparable] struct {
	m map[T]struct{}
}

// NewHashSet creates an empty HashSet.
func NewHashSet[T comparable]() *HashSet[T] {
	return &HashSet[T]{m: make(map[T]struct{})}
}

// NewHashSetFrom creates a HashSet from existing elements.
func NewHashSetFrom[T comparable](values ...T) *HashSet[T] {
	s := &HashSet[T]{m: make(map[T]struct{}, len(values))}
	for _, v := range values {
		s.m[v] = struct{}{}
	}
	return s
}

// ── Sized ─────────────────────────────────────────────────────────────

func (s *HashSet[T]) Size() int     { return len(s.m) }
func (s *HashSet[T]) IsEmpty() bool { return len(s.m) == 0 }

// ── Iterable ──────────────────────────────────────────────────────────

func (s *HashSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s.m {
			if !yield(v) {
				return
			}
		}
	}
}

func (s *HashSet[T]) ForEach(f func(T)) {
	for v := range s.m {
		f(v)
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (s *HashSet[T]) Contains(value T) bool {
	_, ok := s.m[value]
	return ok
}

func (s *HashSet[T]) AnySatisfy(predicate func(T) bool) bool {
	for v := range s.m {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *HashSet[T]) AllSatisfy(predicate func(T) bool) bool {
	for v := range s.m {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *HashSet[T]) NoneSatisfy(predicate func(T) bool) bool {
	for v := range s.m {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ── Convertible ───────────────────────────────────────────────────────

func (s *HashSet[T]) ToSlice() []T {
	result := make([]T, 0, len(s.m))
	for v := range s.m {
		result = append(result, v)
	}
	return result
}

// ── MutableSet ────────────────────────────────────────────────────────

func (s *HashSet[T]) Add(value T) bool {
	if _, ok := s.m[value]; ok {
		return false
	}
	s.m[value] = struct{}{}
	return true
}

func (s *HashSet[T]) Remove(value T) bool {
	if _, ok := s.m[value]; !ok {
		return false
	}
	delete(s.m, value)
	return true
}

func (s *HashSet[T]) Clear() {
	clear(s.m)
}

// ── Set operations ────────────────────────────────────────────────────

// Union returns a new set containing all elements from both sets.
func (s *HashSet[T]) Union(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m {
		result.m[v] = struct{}{}
	}
	for v := range other.m {
		result.m[v] = struct{}{}
	}
	return result
}

// Intersect returns a new set containing only elements present in both sets.
func (s *HashSet[T]) Intersect(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	smaller, larger := s, other
	if len(smaller.m) > len(larger.m) {
		smaller, larger = larger, smaller
	}
	for v := range smaller.m {
		if _, ok := larger.m[v]; ok {
			result.m[v] = struct{}{}
		}
	}
	return result
}

// Difference returns a new set containing elements in s but not in other.
func (s *HashSet[T]) Difference(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m {
		if _, ok := other.m[v]; !ok {
			result.m[v] = struct{}{}
		}
	}
	return result
}

// SymmetricDifference returns elements in either set but not both.
func (s *HashSet[T]) SymmetricDifference(other *HashSet[T]) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m {
		if _, ok := other.m[v]; !ok {
			result.m[v] = struct{}{}
		}
	}
	for v := range other.m {
		if _, ok := s.m[v]; !ok {
			result.m[v] = struct{}{}
		}
	}
	return result
}

// ── Functional operations ─────────────────────────────────────────────

func (s *HashSet[T]) Select(predicate func(T) bool) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m {
		if predicate(v) {
			result.m[v] = struct{}{}
		}
	}
	return result
}

func (s *HashSet[T]) Reject(predicate func(T) bool) *HashSet[T] {
	result := NewHashSet[T]()
	for v := range s.m {
		if !predicate(v) {
			result.m[v] = struct{}{}
		}
	}
	return result
}

func (s *HashSet[T]) Detect(predicate func(T) bool) (T, bool) {
	for v := range s.m {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

func (s *HashSet[T]) Count(predicate func(T) bool) int {
	n := 0
	for v := range s.m {
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
	for v := range s.m {
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
