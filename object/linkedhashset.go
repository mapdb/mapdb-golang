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

// LinkedHashSet is a generic insertion-ordered set backed by a Go map
// and a doubly-linked list. Iteration follows insertion order.
// It implements MutableSet[T].
type LinkedHashSet[T comparable] struct {
	m    map[T]*lhsEntry[T]
	head *lhsEntry[T]
	tail *lhsEntry[T]
}

type lhsEntry[T comparable] struct {
	value      T
	prev, next *lhsEntry[T]
}

// NewLinkedHashSet creates an empty LinkedHashSet.
func NewLinkedHashSet[T comparable]() *LinkedHashSet[T] {
	return &LinkedHashSet[T]{m: make(map[T]*lhsEntry[T])}
}

// NewLinkedHashSetFrom creates a LinkedHashSet from existing elements.
func NewLinkedHashSetFrom[T comparable](values ...T) *LinkedHashSet[T] {
	s := NewLinkedHashSet[T]()
	for _, v := range values {
		s.Add(v)
	}
	return s
}

// ── Sized ─────────────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) Size() int     { return len(s.m) }
func (s *LinkedHashSet[T]) IsEmpty() bool { return len(s.m) == 0 }

// ── Iterable ──────────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := s.head; e != nil; e = e.next {
			if !yield(e.value) {
				return
			}
		}
	}
}

func (s *LinkedHashSet[T]) ForEach(f func(T)) {
	for e := s.head; e != nil; e = e.next {
		f(e.value)
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) Contains(value T) bool {
	_, ok := s.m[value]
	return ok
}

func (s *LinkedHashSet[T]) AnySatisfy(predicate func(T) bool) bool {
	for e := s.head; e != nil; e = e.next {
		if predicate(e.value) {
			return true
		}
	}
	return false
}

func (s *LinkedHashSet[T]) AllSatisfy(predicate func(T) bool) bool {
	for e := s.head; e != nil; e = e.next {
		if !predicate(e.value) {
			return false
		}
	}
	return true
}

func (s *LinkedHashSet[T]) NoneSatisfy(predicate func(T) bool) bool {
	for e := s.head; e != nil; e = e.next {
		if predicate(e.value) {
			return false
		}
	}
	return true
}

// ── Convertible ───────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) ToSlice() []T {
	result := make([]T, 0, len(s.m))
	for e := s.head; e != nil; e = e.next {
		result = append(result, e.value)
	}
	return result
}

// ── MutableSet ────────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) Add(value T) bool {
	if _, ok := s.m[value]; ok {
		return false
	}
	e := &lhsEntry[T]{value: value, prev: s.tail}
	if s.tail != nil {
		s.tail.next = e
	} else {
		s.head = e
	}
	s.tail = e
	s.m[value] = e
	return true
}

func (s *LinkedHashSet[T]) Remove(value T) bool {
	e, ok := s.m[value]
	if !ok {
		return false
	}
	s.unlink(e)
	delete(s.m, value)
	return true
}

func (s *LinkedHashSet[T]) Clear() {
	clear(s.m)
	s.head = nil
	s.tail = nil
}

func (s *LinkedHashSet[T]) unlink(e *lhsEntry[T]) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		s.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		s.tail = e.prev
	}
}

// ── Set operations ────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) Union(other *LinkedHashSet[T]) *LinkedHashSet[T] {
	result := NewLinkedHashSet[T]()
	for e := s.head; e != nil; e = e.next {
		result.Add(e.value)
	}
	for e := other.head; e != nil; e = e.next {
		result.Add(e.value)
	}
	return result
}

func (s *LinkedHashSet[T]) Intersect(other *LinkedHashSet[T]) *LinkedHashSet[T] {
	result := NewLinkedHashSet[T]()
	for e := s.head; e != nil; e = e.next {
		if other.Contains(e.value) {
			result.Add(e.value)
		}
	}
	return result
}

func (s *LinkedHashSet[T]) Difference(other *LinkedHashSet[T]) *LinkedHashSet[T] {
	result := NewLinkedHashSet[T]()
	for e := s.head; e != nil; e = e.next {
		if !other.Contains(e.value) {
			result.Add(e.value)
		}
	}
	return result
}

func (s *LinkedHashSet[T]) SymmetricDifference(other *LinkedHashSet[T]) *LinkedHashSet[T] {
	result := NewLinkedHashSet[T]()
	for e := s.head; e != nil; e = e.next {
		if !other.Contains(e.value) {
			result.Add(e.value)
		}
	}
	for e := other.head; e != nil; e = e.next {
		if !s.Contains(e.value) {
			result.Add(e.value)
		}
	}
	return result
}

// ── Functional operations ─────────────────────────────────────────────

func (s *LinkedHashSet[T]) Select(predicate func(T) bool) *LinkedHashSet[T] {
	result := NewLinkedHashSet[T]()
	for e := s.head; e != nil; e = e.next {
		if predicate(e.value) {
			result.Add(e.value)
		}
	}
	return result
}

func (s *LinkedHashSet[T]) Reject(predicate func(T) bool) *LinkedHashSet[T] {
	result := NewLinkedHashSet[T]()
	for e := s.head; e != nil; e = e.next {
		if !predicate(e.value) {
			result.Add(e.value)
		}
	}
	return result
}

func (s *LinkedHashSet[T]) Detect(predicate func(T) bool) (T, bool) {
	for e := s.head; e != nil; e = e.next {
		if predicate(e.value) {
			return e.value, true
		}
	}
	var zero T
	return zero, false
}

func (s *LinkedHashSet[T]) Count(predicate func(T) bool) int {
	n := 0
	for e := s.head; e != nil; e = e.next {
		if predicate(e.value) {
			n++
		}
	}
	return n
}

// ── Stringer ──────────────────────────────────────────────────────────

func (s *LinkedHashSet[T]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for e := s.head; e != nil; e = e.next {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", e.value)
		first = false
	}
	b.WriteByte('}')
	return b.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableSet[int] = (*LinkedHashSet[int])(nil)
var _ MutableSet[string] = (*LinkedHashSet[string])(nil)
