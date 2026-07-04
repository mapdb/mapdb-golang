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

const strategySetDefaultCapacity = 16

type shsEntry[T any] struct {
	value    T
	occupied bool
}

// HashSetWithStrategy is an open-addressing hash set that uses a pluggable
// HashingStrategy for identity. This allows case-insensitive sets,
// sets keyed by extracted fields, etc.
//
// Go's built-in map requires the comparable constraint and uses built-in
// equality, so this uses a custom hash table instead.
type HashSetWithStrategy[T any] struct {
	entries  []shsEntry[T]
	size     int
	strategy HashingStrategy[T]
}

// NewHashSetWithStrategy creates an empty set using the given hashing strategy.
func NewHashSetWithStrategy[T any](strategy HashingStrategy[T]) *HashSetWithStrategy[T] {
	return NewHashSetWithStrategyCapacity(strategy, strategySetDefaultCapacity)
}

// NewHashSetWithStrategyCapacity creates an empty set with pre-allocated capacity.
func NewHashSetWithStrategyCapacity[T any](strategy HashingStrategy[T], capacity int) *HashSetWithStrategy[T] {
	cap := strategyNextPow2(capacity)
	return &HashSetWithStrategy[T]{
		entries:  make([]shsEntry[T], cap),
		size:     0,
		strategy: strategy,
	}
}

func (s *HashSetWithStrategy[T]) Add(value T) bool {
	if s.needsResize() {
		s.resize()
	}
	mask := len(s.entries) - 1
	idx := int(s.strategy.HashCode(value)) & mask
	for {
		e := &s.entries[idx]
		if !e.occupied {
			e.value = value
			e.occupied = true
			s.size++
			return true
		}
		if s.strategy.Equals(e.value, value) {
			return false
		}
		idx = (idx + 1) & mask
	}
}

func (s *HashSetWithStrategy[T]) Remove(value T) bool {
	if s.size == 0 {
		return false
	}
	mask := len(s.entries) - 1
	idx := int(s.strategy.HashCode(value)) & mask
	for {
		e := &s.entries[idx]
		if !e.occupied {
			return false
		}
		if s.strategy.Equals(e.value, value) {
			s.entries[idx] = shsEntry[T]{}
			s.size--
			s.rehashFrom(idx, mask)
			return true
		}
		idx = (idx + 1) & mask
	}
}

func (s *HashSetWithStrategy[T]) Contains(value T) bool {
	if s.size == 0 {
		return false
	}
	mask := len(s.entries) - 1
	idx := int(s.strategy.HashCode(value)) & mask
	for {
		e := &s.entries[idx]
		if !e.occupied {
			return false
		}
		if s.strategy.Equals(e.value, value) {
			return true
		}
		idx = (idx + 1) & mask
	}
}

// Len returns the number of elements. Use s.Len() == 0 to test for emptiness.
func (s *HashSetWithStrategy[T]) Len() int { return s.size }

func (s *HashSetWithStrategy[T]) Clear() {
	for i := range s.entries {
		s.entries[i] = shsEntry[T]{}
	}
	s.size = 0
}

func (s *HashSetWithStrategy[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range s.entries {
			if s.entries[i].occupied {
				if !yield(s.entries[i].value) {
					return
				}
			}
		}
	}
}

func (s *HashSetWithStrategy[T]) ForEach(f func(T)) {
	for i := range s.entries {
		if s.entries[i].occupied {
			f(s.entries[i].value)
		}
	}
}

func (s *HashSetWithStrategy[T]) ToSlice() []T {
	result := make([]T, 0, s.size)
	for i := range s.entries {
		if s.entries[i].occupied {
			result = append(result, s.entries[i].value)
		}
	}
	return result
}

// AnySatisfy returns true if any element satisfies the predicate.
func (s *HashSetWithStrategy[T]) AnySatisfy(predicate func(T) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all elements satisfy the predicate (vacuously true
// when empty).
func (s *HashSetWithStrategy[T]) AllSatisfy(predicate func(T) bool) bool {
	for v := range s.All() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no element satisfies the predicate (vacuously true
// when empty).
func (s *HashSetWithStrategy[T]) NoneSatisfy(predicate func(T) bool) bool {
	for v := range s.All() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *HashSetWithStrategy[T]) Select(predicate func(T) bool) *HashSetWithStrategy[T] {
	result := NewHashSetWithStrategy(s.strategy)
	s.ForEach(func(v T) {
		if predicate(v) {
			result.Add(v)
		}
	})
	return result
}

func (s *HashSetWithStrategy[T]) Reject(predicate func(T) bool) *HashSetWithStrategy[T] {
	result := NewHashSetWithStrategy(s.strategy)
	s.ForEach(func(v T) {
		if !predicate(v) {
			result.Add(v)
		}
	})
	return result
}

func (s *HashSetWithStrategy[T]) Union(other *HashSetWithStrategy[T]) *HashSetWithStrategy[T] {
	result := NewHashSetWithStrategyCapacity(s.strategy, (s.size+other.size)*2)
	s.ForEach(func(v T) { result.Add(v) })
	other.ForEach(func(v T) { result.Add(v) })
	return result
}

func (s *HashSetWithStrategy[T]) Intersect(other *HashSetWithStrategy[T]) *HashSetWithStrategy[T] {
	result := NewHashSetWithStrategy(s.strategy)
	s.ForEach(func(v T) {
		if other.Contains(v) {
			result.Add(v)
		}
	})
	return result
}

func (s *HashSetWithStrategy[T]) Difference(other *HashSetWithStrategy[T]) *HashSetWithStrategy[T] {
	result := NewHashSetWithStrategy(s.strategy)
	s.ForEach(func(v T) {
		if !other.Contains(v) {
			result.Add(v)
		}
	})
	return result
}

func (s *HashSetWithStrategy[T]) String() string {
	if s.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteByte('{')
	first := true
	for i := range s.entries {
		if s.entries[i].occupied {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", s.entries[i].value)
			first = false
		}
	}
	sb.WriteByte('}')
	return sb.String()
}

// ── internal ──────────────────────────────────────────────────────────

func (s *HashSetWithStrategy[T]) needsResize() bool {
	return (s.size+1)*4 >= len(s.entries)*3
}

func (s *HashSetWithStrategy[T]) resize() {
	old := s.entries
	newCap := len(old) * 2
	if newCap == 0 {
		newCap = strategySetDefaultCapacity
	}
	s.entries = make([]shsEntry[T], newCap)
	s.size = 0
	for i := range old {
		if old[i].occupied {
			s.Add(old[i].value)
		}
	}
}

func (s *HashSetWithStrategy[T]) rehashFrom(deleted, mask int) {
	c := len(s.entries)
	idx := (deleted + 1) & mask
	for s.entries[idx].occupied {
		ideal := int(s.strategy.HashCode(s.entries[idx].value)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			s.entries[deleted] = s.entries[idx]
			s.entries[idx] = shsEntry[T]{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}

func strategyNextPow2(n int) int {
	if n <= 0 {
		return strategySetDefaultCapacity
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32 // no-op on 32-bit platforms (Go shifts are width-defined), required on 64-bit
	n++
	return n
}

// Strategy sets are the exemplar of the comparable→any relaxation (11 §4): a
// non-comparable element type ([]int) satisfies MutableSet only because the
// hierarchy no longer requires comparable — this assert would not compile before
// the relaxation. Identity comes from the strategy, not ==.
var (
	_ MutableSet[int]   = (*HashSetWithStrategy[int])(nil)
	_ MutableSet[[]int] = (*HashSetWithStrategy[[]int])(nil)
)
