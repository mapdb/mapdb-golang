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

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *ArrayStack[T]) Len() int { return len(s.items) }

// ── Stack ─────────────────────────────────────────────────────────────

// Peek returns the top element without removing it. The boolean is false when
// the stack is empty.
func (s *ArrayStack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// PeekAt returns the element at distance from the top (0 = top).
// It panics if index is out of range.
func (s *ArrayStack[T]) PeekAt(index int) T {
	if index < 0 || index >= len(s.items) {
		panic(fmt.Sprintf("object.ArrayStack: index out of range [%d] with length %d", index, s.Len()))
	}
	return s.items[len(s.items)-1-index]
}

// ── MutableStack ──────────────────────────────────────────────────────

func (s *ArrayStack[T]) Push(value T) {
	s.items = append(s.items, value)
}

// Pop removes and returns the top element. The boolean is false when the stack
// is empty.
func (s *ArrayStack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return top, true
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
