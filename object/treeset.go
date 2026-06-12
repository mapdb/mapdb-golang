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

// TreeSet is a sorted set backed by a red-black tree with a pluggable Comparator.
// Elements are maintained in the order defined by the comparator.
type TreeSet[T any] struct {
	tree TreeMap[T, struct{}]
}

// NewTreeSet creates an empty TreeSet using the given comparator.
func NewTreeSet[T any](cmp Comparator[T]) *TreeSet[T] {
	return &TreeSet[T]{tree: *NewTreeMap[T, struct{}](cmp)}
}

func (s *TreeSet[T]) Add(value T) bool {
	_, existed := s.tree.Put(value, struct{}{})
	return !existed
}

func (s *TreeSet[T]) Remove(value T) bool {
	_, existed := s.tree.Remove(value)
	return existed
}

func (s *TreeSet[T]) Contains(value T) bool { return s.tree.ContainsKey(value) }

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *TreeSet[T]) Len() int { return s.tree.Len() }
func (s *TreeSet[T]) Clear()   { s.tree.Clear() }

// Min returns the smallest element.
func (s *TreeSet[T]) Min() (T, bool) {
	k, _, ok := s.tree.Min()
	return k, ok
}

// Max returns the largest element.
func (s *TreeSet[T]) Max() (T, bool) {
	k, _, ok := s.tree.Max()
	return k, ok
}

// All returns an iterator over elements in sorted order.
func (s *TreeSet[T]) All() iter.Seq[T] { return s.tree.Keys() }

func (s *TreeSet[T]) ForEach(f func(T)) {
	s.tree.ForEach(func(k T, _ struct{}) { f(k) })
}

func (s *TreeSet[T]) ToSlice() []T {
	result := make([]T, 0, s.tree.Len())
	s.tree.ForEach(func(k T, _ struct{}) { result = append(result, k) })
	return result
}

func (s *TreeSet[T]) Select(predicate func(T) bool) *TreeSet[T] {
	result := NewTreeSet(s.tree.cmp)
	s.ForEach(func(v T) {
		if predicate(v) {
			result.Add(v)
		}
	})
	return result
}

func (s *TreeSet[T]) Reject(predicate func(T) bool) *TreeSet[T] {
	result := NewTreeSet(s.tree.cmp)
	s.ForEach(func(v T) {
		if !predicate(v) {
			result.Add(v)
		}
	})
	return result
}

func (s *TreeSet[T]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	s.ForEach(func(v T) {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v", v)
		first = false
	})
	b.WriteByte('}')
	return b.String()
}
