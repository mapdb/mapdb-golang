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
//
// Float elements (T with underlying type float32/float64) use bit-pattern
// identity, not ==: NaN is counted as one element, NaN payloads are distinct
// elements, and -0.0 and +0.0 are distinct elements; see keyIndex.
type HashBag[T comparable] struct {
	counts keyIndex[T, int]
	size   int // total count including multiplicities
}

// NewHashBag creates an empty HashBag.
func NewHashBag[T comparable]() *HashBag[T] {
	return &HashBag[T]{counts: newKeyIndex[T, int](0)}
}

// NewHashBagFrom creates a HashBag from existing elements.
func NewHashBagFrom[T comparable](values ...T) *HashBag[T] {
	b := &HashBag[T]{counts: newKeyIndex[T, int](len(values))}
	for _, v := range values {
		b.Add(v)
	}
	return b
}

// ── Sized ─────────────────────────────────────────────────────────────

// Len returns the total number of elements (counting duplicates).
// Use b.Len() == 0 to test for emptiness.
func (b *HashBag[T]) Len() int { return b.size }

// ── Bag ───────────────────────────────────────────────────────────────

func (b *HashBag[T]) OccurrencesOf(value T) int {
	n, _ := b.counts.get(value)
	return n
}
func (b *HashBag[T]) SizeDistinct() int { return b.counts.len() }

// ── Iterable ──────────────────────────────────────────────────────────

// All yields each element once per occurrence.
//
// Float keys: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for v, count := range b.counts.all() {
			for i := 0; i < count; i++ {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// ForEach: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) ForEach(f func(T)) {
	for v, count := range b.counts.all() {
		for i := 0; i < count; i++ {
			f(v)
		}
	}
}

// ForEachWithOccurrences calls f once per distinct value with its count.
//
// Float keys: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) ForEachWithOccurrences(f func(T, int)) {
	for v, count := range b.counts.all() {
		f(v, count)
	}
}

// ── Searchable ────────────────────────────────────────────────────────

func (b *HashBag[T]) Contains(value T) bool { return b.OccurrencesOf(value) > 0 }

// AnySatisfy: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) AnySatisfy(predicate func(T) bool) bool {
	for v := range b.counts.all() {
		if predicate(v) {
			return true
		}
	}
	return false
}

// AllSatisfy: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) AllSatisfy(predicate func(T) bool) bool {
	for v := range b.counts.all() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// NoneSatisfy: the HashBag must not be mutated during this iteration when
// its element type is float32 or float64 (see the package documentation,
// section "Float keys").
func (b *HashBag[T]) NoneSatisfy(predicate func(T) bool) bool {
	for v := range b.counts.all() {
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
	for v, count := range b.counts.all() {
		for i := 0; i < count; i++ {
			result = append(result, v)
		}
	}
	return result
}

// ── MutableBag ────────────────────────────────────────────────────────

func (b *HashBag[T]) Add(value T) {
	n, _ := b.counts.get(value)
	b.counts.put(value, n+1)
	b.size++
}

// AddOccurrences adds multiple occurrences of value.
func (b *HashBag[T]) AddOccurrences(value T, occurrences int) {
	if occurrences <= 0 {
		return
	}
	n, _ := b.counts.get(value)
	b.counts.put(value, n+occurrences)
	b.size += occurrences
}

// Remove removes one occurrence of value. Returns true if it was present.
func (b *HashBag[T]) Remove(value T) bool {
	n, _ := b.counts.get(value)
	if n <= 0 {
		return false
	}
	b.size--
	if n == 1 {
		b.counts.remove(value)
	} else {
		b.counts.put(value, n-1)
	}
	return true
}

func (b *HashBag[T]) Clear() {
	b.counts.clear()
	b.size = 0
}

// ── Top/Bottom occurrences ────────────────────────────────────────────

// TopOccurrences returns the n most frequent values as (value, count) pairs,
// sorted by descending count. A negative n panics.
func (b *HashBag[T]) TopOccurrences(n int) []ValueCount[T] {
	if n < 0 {
		panic("mapdb: TopOccurrences: negative count")
	}
	pairs := b.toValueCounts()
	slices.SortFunc(pairs, func(a, c ValueCount[T]) int { return c.Count - a.Count })
	if n > len(pairs) {
		n = len(pairs)
	}
	return pairs[:n]
}

// BottomOccurrences returns the n least frequent values. A negative n panics.
func (b *HashBag[T]) BottomOccurrences(n int) []ValueCount[T] {
	if n < 0 {
		panic("mapdb: BottomOccurrences: negative count")
	}
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
	pairs := make([]ValueCount[T], 0, b.counts.len())
	for v, c := range b.counts.all() {
		pairs = append(pairs, ValueCount[T]{Value: v, Count: c})
	}
	return pairs
}

// ── Functional operations ─────────────────────────────────────────────

// Select: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) Select(predicate func(T) bool) *HashBag[T] {
	result := NewHashBag[T]()
	for v, count := range b.counts.all() {
		if predicate(v) {
			result.counts.put(v, count)
			result.size += count
		}
	}
	return result
}

// Reject: the HashBag must not be mutated during this iteration when its
// element type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBag[T]) Reject(predicate func(T) bool) *HashBag[T] {
	result := NewHashBag[T]()
	for v, count := range b.counts.all() {
		if !predicate(v) {
			result.counts.put(v, count)
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
	for v, count := range b.counts.all() {
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
