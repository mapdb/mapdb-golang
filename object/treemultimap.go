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

// TreeMultimap is a sorted multimap: each key maps to a list of values,
// with duplicates preserved in insertion order. Keys are iterated in the
// order defined by the Comparator passed to the constructor. Mirrors
// Eclipse Collections Java's TreeBagMultimap / SortedSetMultimap
// contract, but uses a list-shaped value collection so V need not be
// comparable.
//
// Backed by the hand-written TreeMap[K, []V]; all TreeMap-side ops are
// O(log N) in the number of distinct keys.
type TreeMultimap[K any, V any] struct {
	tm        *TreeMap[K, []V]
	totalSize int
}

// NewTreeMultimap creates an empty TreeMultimap ordered by cmp.
func NewTreeMultimap[K any, V any](cmp Comparator[K]) *TreeMultimap[K, V] {
	return &TreeMultimap[K, V]{tm: NewTreeMap[K, []V](cmp)}
}

// Put appends v to the list at key k.
func (t *TreeMultimap[K, V]) Put(k K, v V) {
	if existing, ok := t.tm.Get(k); ok {
		t.tm.Put(k, append(existing, v))
	} else {
		t.tm.Put(k, []V{v})
	}
	t.totalSize++
}

// PutAll appends every value in values to the list at key k.
func (t *TreeMultimap[K, V]) PutAll(k K, values ...V) {
	if len(values) == 0 {
		return
	}
	if existing, ok := t.tm.Get(k); ok {
		t.tm.Put(k, append(existing, values...))
	} else {
		cp := make([]V, len(values))
		copy(cp, values)
		t.tm.Put(k, cp)
	}
	t.totalSize += len(values)
}

// Get returns a defensive copy of the values associated with k.
func (t *TreeMultimap[K, V]) Get(k K) []V {
	return t.GetCopy(k)
}

// GetCopy returns a defensive copy of the values for k.
func (t *TreeMultimap[K, V]) GetCopy(k K) []V {
	v, _ := t.tm.Get(k)
	if v == nil {
		return nil
	}
	out := make([]V, len(v))
	copy(out, v)
	return out
}

// ContainsKey returns true if any values are stored under k.
func (t *TreeMultimap[K, V]) ContainsKey(k K) bool {
	return t.tm.ContainsKey(k)
}

// RemoveKey removes all values for k and returns them, or nil if absent.
func (t *TreeMultimap[K, V]) RemoveKey(k K) []V {
	vs, ok := t.tm.Remove(k)
	if !ok {
		return nil
	}
	t.totalSize -= len(vs)
	return vs
}

// RemoveMatching removes every value v at key k where eq(v, target) is true.
// Returns the number of values removed.
func (t *TreeMultimap[K, V]) RemoveMatching(k K, target V, eq func(V, V) bool) int {
	vs, ok := t.tm.Get(k)
	if !ok {
		return 0
	}
	out := vs[:0]
	removed := 0
	for _, v := range vs {
		if eq(v, target) {
			removed++
			continue
		}
		out = append(out, v)
	}
	if removed == 0 {
		return 0
	}
	if len(out) == 0 {
		t.tm.Remove(k)
	} else {
		t.tm.Put(k, out)
	}
	t.totalSize -= removed
	return removed
}

// Size returns the total number of values.
func (t *TreeMultimap[K, V]) Size() int { return t.totalSize }

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (t *TreeMultimap[K, V]) Len() int { return t.Size() }

// SizeDistinct returns the number of distinct keys.
func (t *TreeMultimap[K, V]) SizeDistinct() int { return t.tm.Size() }

// IsEmpty reports whether the multimap is empty.
func (t *TreeMultimap[K, V]) IsEmpty() bool { return t.totalSize == 0 }

// Clear removes all entries (retains the comparator).
func (t *TreeMultimap[K, V]) Clear() {
	t.tm.Clear()
	t.totalSize = 0
}

// All yields every (key, value) pair in key-sorted order; within one
// key the order is insertion order.
func (t *TreeMultimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, vs := range t.tm.All() {
			for _, v := range vs {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Keys yields each distinct key once in sorted order.
func (t *TreeMultimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range t.tm.All() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values yields every value across all keys in key-sorted / insertion order.
func (t *TreeMultimap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, vs := range t.tm.All() {
			for _, v := range vs {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// ForEachKeyMultiValues iterates keys in sorted order, passing a defensive
// copy of each key's values.
func (t *TreeMultimap[K, V]) ForEachKeyMultiValues(f func(K, []V)) {
	t.tm.ForEach(func(k K, vs []V) {
		cp := make([]V, len(vs))
		copy(cp, vs)
		f(k, cp)
	})
}

// ForEachKey iterates keys in sorted order.
func (t *TreeMultimap[K, V]) ForEachKey(f func(K)) {
	t.tm.ForEach(func(k K, _ []V) { f(k) })
}

// ForEach invokes f for every (key, value) pair in sorted/insertion order.
func (t *TreeMultimap[K, V]) ForEach(f func(K, V)) {
	t.tm.ForEach(func(k K, vs []V) {
		for _, v := range vs {
			f(k, v)
		}
	})
}

// String renders as {k1=[v1, v2], k2=[v3]} in key-sorted order.
func (t *TreeMultimap[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	t.tm.ForEach(func(k K, vs []V) {
		if !first {
			sb.WriteString(", ")
		}
		first = false
		fmt.Fprintf(&sb, "%v=[", k)
		for i, v := range vs {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", v)
		}
		sb.WriteString("]")
	})
	sb.WriteString("}")
	return sb.String()
}
