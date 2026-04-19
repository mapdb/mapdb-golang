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

// HashBag is a generic multiset (bag) backed by map[T]int.
// It implements MutableBag[T].
type HashBag[T comparable] struct {
	counts map[T]int
	size   int // total count including multiplicities
}

// NewHashBag creates an empty HashBag.
func NewHashBag[T comparable]() *HashBag[T] {
	return &HashBag[T]{counts: make(map[T]int)}
}

// NewHashBagFrom creates a HashBag from existing elements.
func NewHashBagFrom[T comparable](values ...T) *HashBag[T] {
	b := &HashBag[T]{counts: make(map[T]int, len(values))}
	for _, v := range values {
		b.counts[v]++
		b.size++
	}
	return b
}

// ── Sized ─────────────────────────────────────────────────────────────

func (b *HashBag[T]) Size() int     { return b.size }
func (b *HashBag[T]) IsEmpty() bool { return b.size == 0 }

// ── Bag ───────────────────────────────────────────────────────────────

func (b *HashBag[T]) OccurrencesOf(value T) int { return b.counts[value] }
func (b *HashBag[T]) SizeDistinct() int         { return len(b.counts) }

// ── Iterable ──────────────────────────────────────────────────────────

// All yields each element once per occurrence.
func (b *HashBag[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v, count := range b.counts {
			for i := 0; i < count; i++ {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func (b *HashBag[T]) ForEach(f func(T)) {
	for v, count := range b.counts {
		for i := 0; i < count; i++ {
			f(v)
		}
	}
}

// ForEachWithOccurrences calls f once per distinct value with its count.
func (b *HashBag[T]) ForEachWithOccurrences(f func(T, int)) {
	for v, count := range b.counts {
		f(v, count)
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (b *HashBag[T]) Contains(value T) bool { return b.counts[value] > 0 }

func (b *HashBag[T]) AnySatisfy(predicate func(T) bool) bool {
	for v := range b.counts {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (b *HashBag[T]) AllSatisfy(predicate func(T) bool) bool {
	for v := range b.counts {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (b *HashBag[T]) NoneSatisfy(predicate func(T) bool) bool {
	for v := range b.counts {
		if predicate(v) {
			return false
		}
	}
	return true
}

// ── Convertible ───────────────────────────────────────────────────────

// ToSlice returns all elements (with multiplicities) as a slice.
func (b *HashBag[T]) ToSlice() []T {
	result := make([]T, 0, b.size)
	for v, count := range b.counts {
		for i := 0; i < count; i++ {
			result = append(result, v)
		}
	}
	return result
}

// ── MutableBag ────────────────────────────────────────────────────────

func (b *HashBag[T]) Add(value T) {
	b.counts[value]++
	b.size++
}

// AddOccurrences adds multiple occurrences of value.
func (b *HashBag[T]) AddOccurrences(value T, occurrences int) {
	if occurrences <= 0 {
		return
	}
	b.counts[value] += occurrences
	b.size += occurrences
}

// Remove removes one occurrence of value. Returns true if it was present.
func (b *HashBag[T]) Remove(value T) bool {
	if b.counts[value] <= 0 {
		return false
	}
	b.counts[value]--
	b.size--
	if b.counts[value] == 0 {
		delete(b.counts, value)
	}
	return true
}

func (b *HashBag[T]) Clear() {
	clear(b.counts)
	b.size = 0
}

// ── Top/Bottom occurrences ────────────────────────────────────────────

// TopOccurrences returns the n most frequent values as (value, count) pairs,
// sorted by descending count.
func (b *HashBag[T]) TopOccurrences(n int) []ValueCount[T] {
	pairs := b.toValueCounts()
	slices.SortFunc(pairs, func(a, c ValueCount[T]) int { return c.Count - a.Count })
	if n > len(pairs) {
		n = len(pairs)
	}
	return pairs[:n]
}

// BottomOccurrences returns the n least frequent values.
func (b *HashBag[T]) BottomOccurrences(n int) []ValueCount[T] {
	pairs := b.toValueCounts()
	slices.SortFunc(pairs, func(a, c ValueCount[T]) int { return a.Count - c.Count })
	if n > len(pairs) {
		n = len(pairs)
	}
	return pairs[:n]
}

// ValueCount pairs a value with its occurrence count.
type ValueCount[T comparable] struct {
	Value T
	Count int
}

func (b *HashBag[T]) toValueCounts() []ValueCount[T] {
	pairs := make([]ValueCount[T], 0, len(b.counts))
	for v, c := range b.counts {
		pairs = append(pairs, ValueCount[T]{Value: v, Count: c})
	}
	return pairs
}

// ── Functional operations ─────────────────────────────────────────────

func (b *HashBag[T]) Select(predicate func(T) bool) *HashBag[T] {
	result := NewHashBag[T]()
	for v, count := range b.counts {
		if predicate(v) {
			result.counts[v] = count
			result.size += count
		}
	}
	return result
}

func (b *HashBag[T]) Reject(predicate func(T) bool) *HashBag[T] {
	result := NewHashBag[T]()
	for v, count := range b.counts {
		if !predicate(v) {
			result.counts[v] = count
			result.size += count
		}
	}
	return result
}

// ── Stringer ──────────────────────────────────────────────────────────

func (b *HashBag[T]) String() string {
	var s strings.Builder
	s.WriteByte('{')
	first := true
	for v, count := range b.counts {
		if !first {
			s.WriteString(", ")
		}
		fmt.Fprintf(&s, "%v×%d", v, count)
		first = false
	}
	s.WriteByte('}')
	return s.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableBag[int] = (*HashBag[int])(nil)
var _ MutableBag[string] = (*HashBag[string])(nil)
