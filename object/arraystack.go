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

// ArrayStack is a generic LIFO stack backed by a slice.
// It implements MutableStack[T].
type ArrayStack[T comparable] struct {
	items []T
}

// NewArrayStack creates an empty ArrayStack.
func NewArrayStack[T comparable]() *ArrayStack[T] {
	return &ArrayStack[T]{}
}

// NewArrayStackFrom creates an ArrayStack with initial elements (last element is top).
func NewArrayStackFrom[T comparable](values ...T) *ArrayStack[T] {
	cp := make([]T, len(values))
	copy(cp, values)
	return &ArrayStack[T]{items: cp}
}

// ── Sized ─────────────────────────────────────────────────────────────

func (s *ArrayStack[T]) Size() int { return len(s.items) }

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (s *ArrayStack[T]) Len() int      { return s.Size() }
func (s *ArrayStack[T]) IsEmpty() bool { return len(s.items) == 0 }

// ── Stack ─────────────────────────────────────────────────────────────

func (s *ArrayStack[T]) Peek() (T, error) {
	if len(s.items) == 0 {
		var zero T
		return zero, fmt.Errorf("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

// PeekAt returns the element at distance from the top (0 = top).
func (s *ArrayStack[T]) PeekAt(index int) (T, error) {
	if index < 0 || index >= len(s.items) {
		var zero T
		return zero, fmt.Errorf("index out of bounds: %d (size %d)", index, len(s.items))
	}
	return s.items[len(s.items)-1-index], nil
}

// ── MutableStack ──────────────────────────────────────────────────────

func (s *ArrayStack[T]) Push(value T) {
	s.items = append(s.items, value)
}

func (s *ArrayStack[T]) Pop() (T, error) {
	if len(s.items) == 0 {
		var zero T
		return zero, fmt.Errorf("stack is empty")
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, nil
}

func (s *ArrayStack[T]) Clear() {
	s.items = s.items[:0]
}

// ── Iterable (top to bottom) ──────────────────────────────────────────

func (s *ArrayStack[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(s.items) - 1; i >= 0; i-- {
			if !yield(s.items[i]) {
				return
			}
		}
	}
}

func (s *ArrayStack[T]) ForEach(f func(T)) {
	for i := len(s.items) - 1; i >= 0; i-- {
		f(s.items[i])
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (s *ArrayStack[T]) Contains(value T) bool {
	for _, v := range s.items {
		if v == value {
			return true
		}
	}
	return false
}

func (s *ArrayStack[T]) AnySatisfy(predicate func(T) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *ArrayStack[T]) AllSatisfy(predicate func(T) bool) bool {
	for _, v := range s.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *ArrayStack[T]) NoneSatisfy(predicate func(T) bool) bool {
	for _, v := range s.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ── Convertible ───────────────────────────────────────────────────────

// ToSlice returns elements from top to bottom.
func (s *ArrayStack[T]) ToSlice() []T {
	result := make([]T, len(s.items))
	for i, j := len(s.items)-1, 0; i >= 0; i, j = i-1, j+1 {
		result[j] = s.items[i]
	}
	return result
}

// ── Stringer ──────────────────────────────────────────────────────────

func (s *ArrayStack[T]) String() string {
	var b strings.Builder
	b.WriteString("[top: ")
	for i := len(s.items) - 1; i >= 0; i-- {
		if i < len(s.items)-1 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", s.items[i])
	}
	b.WriteByte(']')
	return b.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableStack[int] = (*ArrayStack[int])(nil)
var _ MutableStack[string] = (*ArrayStack[string])(nil)
