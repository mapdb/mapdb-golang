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
//
// Float keys (underlying type float32/float64) use bit-pattern identity, not
// ==: NaN is found and not duplicated, NaN payloads are distinct, and -0.0 and
// +0.0 are distinct; see keyIndex. ToMap cannot preserve that (it returns a
// builtin map).
type HashMultimap[K comparable, V any] struct {
	m keyIndex[K, []V]
	// totalSize caches the total value count so Size() is O(1).
	totalSize int
}

// NewHashMultimap creates an empty HashMultimap.
func NewHashMultimap[K comparable, V any]() *HashMultimap[K, V] {
	return &HashMultimap[K, V]{m: newKeyIndex[K, []V](0)}
}

// NewHashMultimapWithCapacity pre-allocates space for approximately
// `keyCapacity` distinct keys.
func NewHashMultimapWithCapacity[K comparable, V any](keyCapacity int) *HashMultimap[K, V] {
	return &HashMultimap[K, V]{m: newKeyIndex[K, []V](keyCapacity)}
}

// Put appends v to the list at key k.
func (h *HashMultimap[K, V]) Put(k K, v V) {
	vs, _ := h.m.get(k)
	h.m.put(k, append(vs, v))
	h.totalSize++
}

// PutAll appends every value in values to the list at key k.
// An empty values list leaves the multimap unchanged.
func (h *HashMultimap[K, V]) PutAll(k K, values ...V) {
	if len(values) == 0 {
		return
	}
	vs, _ := h.m.get(k)
	h.m.put(k, append(vs, values...))
	h.totalSize += len(values)
}

// Get returns a defensive copy of the values associated with k in insertion
// order, or nil if k has no values.
func (h *HashMultimap[K, V]) Get(k K) []V {
	return h.GetCopy(k)
}

// GetCopy returns a defensive copy of the values for k.
func (h *HashMultimap[K, V]) GetCopy(k K) []V {
	src, _ := h.m.get(k)
	if src == nil {
		return nil
	}
	out := make([]V, len(src))
	copy(out, src)
	return out
}

// ContainsKey returns true if any values are stored under k.
func (h *HashMultimap[K, V]) ContainsKey(k K) bool {
	return h.m.contains(k)
}

// RemoveKey removes all values for k and returns the removed slice, or
// nil if k was not present.
func (h *HashMultimap[K, V]) RemoveKey(k K) []V {
	vs, ok := h.m.remove(k)
	if !ok {
		return nil
	}
	h.totalSize -= len(vs)
	return vs
}

// RemoveMatching removes every value v at key k for which eq(v, target)
// is true, compacting the remaining values in place. Returns the number
// of values removed. Use this when V is not comparable but the caller
// has an equivalence predicate.
func (h *HashMultimap[K, V]) RemoveMatching(k K, target V, eq func(V, V) bool) int {
	vs, ok := h.m.get(k)
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
	clear(vs[len(out):])
	if len(out) == 0 {
		h.m.remove(k)
	} else {
		h.m.put(k, out)
	}
	h.totalSize -= removed
	return removed
}

// Len returns the total number of values stored across all keys.
// Use h.Len() == 0 to test for emptiness.
func (h *HashMultimap[K, V]) Len() int { return h.totalSize }

// SizeDistinct returns the number of distinct keys.
func (h *HashMultimap[K, V]) SizeDistinct() int { return h.m.len() }

// Clear removes all entries.
func (h *HashMultimap[K, V]) Clear() {
	h.m = newKeyIndex[K, []V](0)
	h.totalSize = 0
}

// All yields every (key, value) pair. The iteration order within one key
// is insertion order; across keys it is unspecified (the Go map's
// randomised order, or table order for float keys).
//
// Float keys: the HashMultimap must not be mutated during this iteration
// when its key type is float32 or float64 (see the package documentation,
// section "Float keys").
func (h *HashMultimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, vs := range h.m.all() {
			for _, v := range vs {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// Keys yields each distinct key once.
//
// Float keys: the HashMultimap must not be mutated during this iteration
// when its key type is float32 or float64 (see the package documentation,
// section "Float keys").
func (h *HashMultimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range h.m.all() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values yields every value across all keys.
//
// Float keys: the HashMultimap must not be mutated during this iteration
// when its key type is float32 or float64 (see the package documentation,
// section "Float keys").
func (h *HashMultimap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, vs := range h.m.all() {
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
//
// Float keys: the HashMultimap must not be mutated during this iteration
// when its key type is float32 or float64 (see the package documentation,
// section "Float keys").
func (h *HashMultimap[K, V]) ForEachKeyMultiValues(f func(K, []V)) {
	for k, vs := range h.m.all() {
		cp := make([]V, len(vs))
		copy(cp, vs)
		f(k, cp)
	}
}

// ForEachKey invokes f once per distinct key.
//
// Float keys: the HashMultimap must not be mutated during this iteration
// when its key type is float32 or float64 (see the package documentation,
// section "Float keys").
func (h *HashMultimap[K, V]) ForEachKey(f func(K)) {
	for k := range h.m.all() {
		f(k)
	}
}

// ForEach invokes f for every (key, value) pair.
//
// Float keys: the HashMultimap must not be mutated during this iteration
// when its key type is float32 or float64 (see the package documentation,
// section "Float keys").
func (h *HashMultimap[K, V]) ForEach(f func(K, V)) {
	for k, vs := range h.m.all() {
		for _, v := range vs {
			f(k, v)
		}
	}
}

// ToMap returns a defensive copy as a plain Go map of slices. Callers
// receive a shallow copy of both the map and each value slice.
//
// The returned builtin map uses Go == key identity, not the bit-pattern
// identity of the multimap itself. For a float32 or float64 key type this
// loses information: -0.0 and +0.0 merge into one map key (only one of the
// two value lists survives, which one is unspecified), and every NaN key
// becomes a map entry that no lookup can reach (NaN != NaN; such entries
// are visible only by ranging over the map).
func (h *HashMultimap[K, V]) ToMap() map[K][]V {
	out := make(map[K][]V, h.m.len())
	for k, vs := range h.m.all() {
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
	for k, vs := range h.m.all() {
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
