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

// HashMultimap is a generic unordered multimap: each key maps to a list of
// values, with duplicates preserved in insertion order. It mirrors Eclipse
// Collections Java's HashBagMultimap / FastListMultimap contract, but the
// Go API uses a list-shaped value collection so V does not need to be
// comparable.
//
// K must be `comparable` (hashable under Go's ==). V is any type; if you
// need value-based operations (Contains, RemoveKeyValue) build a
// HashMultimap[K, V] and use HashMultimap.RemoveMatching with an equality
// function, or switch to a ByField HashingStrategy-backed store.
type HashMultimap[K comparable, V any] struct {
	m map[K][]V
	// totalSize caches the total value count so Size() is O(1).
	totalSize int
}

// NewHashMultimap creates an empty HashMultimap.
func NewHashMultimap[K comparable, V any]() *HashMultimap[K, V] {
	return &HashMultimap[K, V]{m: make(map[K][]V)}
}

// NewHashMultimapWithCapacity pre-allocates space for approximately
// `keyCapacity` distinct keys.
func NewHashMultimapWithCapacity[K comparable, V any](keyCapacity int) *HashMultimap[K, V] {
	return &HashMultimap[K, V]{m: make(map[K][]V, keyCapacity)}
}

// Put appends v to the list at key k.
func (h *HashMultimap[K, V]) Put(k K, v V) {
	if h.m == nil {
		h.m = make(map[K][]V)
	}
	h.m[k] = append(h.m[k], v)
	h.totalSize++
}

// PutAll appends every value in values to the list at key k.
func (h *HashMultimap[K, V]) PutAll(k K, values ...V) {
	if h.m == nil {
		h.m = make(map[K][]V)
	}
	h.m[k] = append(h.m[k], values...)
	h.totalSize += len(values)
}

// Get returns a defensive copy of the values associated with k in insertion
// order, or nil if k has no values.
func (h *HashMultimap[K, V]) Get(k K) []V {
	return h.GetCopy(k)
}

// GetCopy returns a defensive copy of the values for k.
func (h *HashMultimap[K, V]) GetCopy(k K) []V {
	src := h.m[k]
	if src == nil {
		return nil
	}
	out := make([]V, len(src))
	copy(out, src)
	return out
}

// ContainsKey returns true if any values are stored under k.
func (h *HashMultimap[K, V]) ContainsKey(k K) bool {
	_, ok := h.m[k]
	return ok
}

// RemoveKey removes all values for k and returns the removed slice, or
// nil if k was not present.
func (h *HashMultimap[K, V]) RemoveKey(k K) []V {
	vs, ok := h.m[k]
	if !ok {
		return nil
	}
	delete(h.m, k)
	h.totalSize -= len(vs)
	return vs
}

// RemoveMatching removes every value v at key k for which eq(v, target)
// is true, compacting the remaining values in place. Returns the number
// of values removed. Use this when V is not comparable but the caller
// has an equivalence predicate.
func (h *HashMultimap[K, V]) RemoveMatching(k K, target V, eq func(V, V) bool) int {
	vs, ok := h.m[k]
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
		delete(h.m, k)
	} else {
		h.m[k] = out
	}
	h.totalSize -= removed
	return removed
}

// Len returns the total number of values stored across all keys.
// Use h.Len() == 0 to test for emptiness.
func (h *HashMultimap[K, V]) Len() int { return h.totalSize }

// SizeDistinct returns the number of distinct keys.
func (h *HashMultimap[K, V]) SizeDistinct() int { return len(h.m) }

// Clear removes all entries.
func (h *HashMultimap[K, V]) Clear() {
	h.m = make(map[K][]V)
	h.totalSize = 0
}

// All yields every (key, value) pair. The iteration order within one key
// is insertion order; across keys it is the Go map's randomised order.
func (h *HashMultimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, vs := range h.m {
			for _, v := range vs {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Keys yields each distinct key once.
func (h *HashMultimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range h.m {
			if !yield(k) {
				return
			}
		}
	}
}

// Values yields every value across all keys.
func (h *HashMultimap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, vs := range h.m {
			for _, v := range vs {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// ForEachKeyMultiValues invokes f once per key with a defensive copy of the
// values at that key.
func (h *HashMultimap[K, V]) ForEachKeyMultiValues(f func(K, []V)) {
	for k, vs := range h.m {
		cp := make([]V, len(vs))
		copy(cp, vs)
		f(k, cp)
	}
}

// ForEachKey invokes f once per distinct key.
func (h *HashMultimap[K, V]) ForEachKey(f func(K)) {
	for k := range h.m {
		f(k)
	}
}

// ForEach invokes f for every (key, value) pair.
func (h *HashMultimap[K, V]) ForEach(f func(K, V)) {
	for k, vs := range h.m {
		for _, v := range vs {
			f(k, v)
		}
	}
}

// ToMap returns a defensive copy as a plain Go map of slices. Callers
// receive a shallow copy of both the map and each value slice.
func (h *HashMultimap[K, V]) ToMap() map[K][]V {
	out := make(map[K][]V, len(h.m))
	for k, vs := range h.m {
		cp := make([]V, len(vs))
		copy(cp, vs)
		out[k] = cp
	}
	return out
}

// String returns a readable representation. Format example:
//
//	{alice=[1, 2], bob=[3]}
func (h *HashMultimap[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, vs := range h.m {
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
	}
	sb.WriteString("}")
	return sb.String()
}
