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

// Floor returns the largest element <= value, or zero and false.
func (s *TreeSet[T]) Floor(value T) (T, bool) {
	k, _, ok := s.tree.Floor(value)
	return k, ok
}

// Ceiling returns the smallest element >= value, or zero and false.
func (s *TreeSet[T]) Ceiling(value T) (T, bool) {
	k, _, ok := s.tree.Ceiling(value)
	return k, ok
}

// Lower returns the largest element strictly < value, or zero and false.
func (s *TreeSet[T]) Lower(value T) (T, bool) {
	k, _, ok := s.tree.Lower(value)
	return k, ok
}

// Higher returns the smallest element strictly > value, or zero and false.
func (s *TreeSet[T]) Higher(value T) (T, bool) {
	k, _, ok := s.tree.Higher(value)
	return k, ok
}

// First is an alias of Min.
func (s *TreeSet[T]) First() (T, bool) { return s.Min() }

// Last is an alias of Max.
func (s *TreeSet[T]) Last() (T, bool) { return s.Max() }

// PollFirst removes and returns the smallest element, or zero and false if
// empty. Does not trap on an empty set.
func (s *TreeSet[T]) PollFirst() (T, bool) {
	k, _, ok := s.tree.PollFirstEntry()
	return k, ok
}

// PollLast removes and returns the largest element, or zero and false if empty.
// Does not trap on an empty set.
func (s *TreeSet[T]) PollLast() (T, bool) {
	k, _, ok := s.tree.PollLastEntry()
	return k, ok
}

// DescendingSet returns an iterator over elements in descending order.
func (s *TreeSet[T]) DescendingSet() iter.Seq[T] { return s.tree.DescendingKeys() }

// SubSetCopy returns a new independent TreeSet of the elements in [from, to)
// under this set's comparator. It is a materialized snapshot, not a live view:
// mutating it never affects the original and vice versa. The snapshot PRESERVES
// the source set's comparator, so a reverse/custom/float total-order set keeps
// its ordering semantics in the slice (it never resets to natural order).
func (s *TreeSet[T]) SubSetCopy(from, to T) *TreeSet[T] {
	out := NewTreeSet(s.tree.cmp)
	sub := s.tree.SubMapCopy(from, to)
	for k := range sub.Keys() {
		out.Add(k)
	}
	return out
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

// SelectWhere returns a new TreeSet with elements satisfying the predicate.
//
// Named SelectWhere (not Select) so the bare Select name is reserved for the
// order-statistic Select (i-th smallest by 0-based rank), per
// spec/features/rank-select.md.
func (s *TreeSet[T]) SelectWhere(predicate func(T) bool) *TreeSet[T] {
	result := NewTreeSet(s.tree.cmp)
	s.ForEach(func(v T) {
		if predicate(v) {
			result.Add(v)
		}
	})
	return result
}

// Rank returns the number of elements strictly less than value under the set's
// comparator — the 0-based lower-bound index value occupies (if present) or
// would occupy (if absent). Result is in 0..=Len(). Pure query.
func (s *TreeSet[T]) Rank(value T) int { return s.tree.Rank(value) }

// Select returns the i-th smallest element (0-based) and true, or the zero value
// and false if i >= Len() or i < 0. Out-of-range and negative indices return
// absence and do not trap. This is the order-statistic select; the predicate
// filtering convenience is SelectWhere.
func (s *TreeSet[T]) Select(i int) (T, bool) { return s.tree.SelectKey(i) }

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
