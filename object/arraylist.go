// Copyright (c) 2026 Jan Kotek.
// Derived from Eclipse Collections (Copyright (c) Goldman Sachs and others).
// Licensed under the Eclipse Public License v1.0 and Eclipse Distribution License v1.0.
// See LICENSE-EPL-1.0.txt and LICENSE-EDL-1.0.txt.
// USE AT YOUR OWN RISK — THIS SOFTWARE IS PROVIDED WITHOUT WARRANTY OF ANY KIND.

package object

import (
	"fmt"
	"iter"
	"slices"
	"strings"
)

// ArrayList is a generic ordered list backed by a Go slice.
// It implements MutableList[T].
type ArrayList[T comparable] struct {
	items []T
}

// NewArrayList creates an empty ArrayList.
func NewArrayList[T comparable]() *ArrayList[T] {
	return &ArrayList[T]{}
}

// NewArrayListFrom creates an ArrayList from existing elements.
func NewArrayListFrom[T comparable](values ...T) *ArrayList[T] {
	cp := make([]T, len(values))
	copy(cp, values)
	return &ArrayList[T]{items: cp}
}

// NewArrayListWithCapacity creates an ArrayList with pre-allocated capacity.
func NewArrayListWithCapacity[T comparable](capacity int) *ArrayList[T] {
	return &ArrayList[T]{items: make([]T, 0, capacity)}
}

// ── Sized ─────────────────────────────────────────────────────────────

// Len returns the number of elements. Use a.Len() == 0 to test for emptiness.
func (a *ArrayList[T]) Len() int { return len(a.items) }

// ── Iterable ──────────────────────────────────────────────────────────

func (a *ArrayList[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range a.items {
			if !yield(v) {
				return
			}
		}
	}
}

func (a *ArrayList[T]) ForEach(f func(T)) {
	for _, v := range a.items {
		f(v)
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (a *ArrayList[T]) Contains(value T) bool {
	for _, v := range a.items {
		if v == value {
			return true
		}
	}
	return false
}

func (a *ArrayList[T]) AnySatisfy(predicate func(T) bool) bool {
	for _, v := range a.items {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (a *ArrayList[T]) AllSatisfy(predicate func(T) bool) bool {
	for _, v := range a.items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (a *ArrayList[T]) NoneSatisfy(predicate func(T) bool) bool {
	for _, v := range a.items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ── Convertible ───────────────────────────────────────────────────────

func (a *ArrayList[T]) ToSlice() []T {
	cp := make([]T, len(a.items))
	copy(cp, a.items)
	return cp
}

// ── List ──────────────────────────────────────────────────────────────

func (a *ArrayList[T]) Get(index int) T {
	if index < 0 || index >= len(a.items) {
		panic(fmt.Sprintf("object.ArrayList: index out of range [%d] with length %d", index, a.Len()))
	}
	return a.items[index]
}

func (a *ArrayList[T]) IndexOf(value T) int {
	for i, v := range a.items {
		if v == value {
			return i
		}
	}
	return -1
}

// ── MutableList ───────────────────────────────────────────────────────

func (a *ArrayList[T]) Add(value T) {
	a.items = append(a.items, value)
}

func (a *ArrayList[T]) Set(index int, value T) T {
	if index < 0 || index >= len(a.items) {
		panic(fmt.Sprintf("object.ArrayList: index out of range [%d] with length %d", index, a.Len()))
	}
	old := a.items[index]
	a.items[index] = value
	return old
}

func (a *ArrayList[T]) Clear() {
	a.items = a.items[:0]
}

// ── Functional operations ─────────────────────────────────────────────

// Select returns a new ArrayList containing only elements that satisfy the predicate.
func (a *ArrayList[T]) Select(predicate func(T) bool) *ArrayList[T] {
	result := NewArrayList[T]()
	for _, v := range a.items {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Reject returns a new ArrayList containing only elements that do NOT satisfy the predicate.
func (a *ArrayList[T]) Reject(predicate func(T) bool) *ArrayList[T] {
	result := NewArrayList[T]()
	for _, v := range a.items {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// Detect returns the first element satisfying the predicate, or (zero, false).
func (a *ArrayList[T]) Detect(predicate func(T) bool) (T, bool) {
	for _, v := range a.items {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// Count returns the number of elements satisfying the predicate.
func (a *ArrayList[T]) Count(predicate func(T) bool) int {
	n := 0
	for _, v := range a.items {
		if predicate(v) {
			n++
		}
	}
	return n
}

// InjectInto folds the collection from left with an initial value.
func (a *ArrayList[T]) InjectInto(initial any, f func(any, T) any) any {
	acc := initial
	for _, v := range a.items {
		acc = f(acc, v)
	}
	return acc
}

// Sort sorts the list in place using the provided comparison function.
// Returns the list for chaining.
func (a *ArrayList[T]) Sort(less func(a, b T) bool) *ArrayList[T] {
	slices.SortFunc(a.items, func(x, y T) int {
		if less(x, y) {
			return -1
		}
		if less(y, x) {
			return 1
		}
		return 0
	})
	return a
}

// Reversed returns a new ArrayList with elements in reverse order.
func (a *ArrayList[T]) Reversed() *ArrayList[T] {
	result := NewArrayListWithCapacity[T](len(a.items))
	for i := len(a.items) - 1; i >= 0; i-- {
		result.Add(a.items[i])
	}
	return result
}

// Distinct returns a new ArrayList with duplicate elements removed (first occurrence kept).
func (a *ArrayList[T]) Distinct() *ArrayList[T] {
	seen := make(map[T]struct{}, len(a.items))
	result := NewArrayList[T]()
	for _, v := range a.items {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result.Add(v)
		}
	}
	return result
}

// Remove removes the first occurrence of value. Returns true if found.
func (a *ArrayList[T]) Remove(value T) bool {
	for i, v := range a.items {
		if v == value {
			a.items = append(a.items[:i], a.items[i+1:]...)
			return true
		}
	}
	return false
}

// ── Stringer ──────────────────────────────────────────────────────────

func (a *ArrayList[T]) String() string {
	var b strings.Builder
	b.WriteByte('[')
	for i, v := range a.items {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", v)
	}
	b.WriteByte(']')
	return b.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableList[int] = (*ArrayList[int])(nil)
var _ MutableList[string] = (*ArrayList[string])(nil)
