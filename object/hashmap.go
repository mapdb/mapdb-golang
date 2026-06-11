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

// HashMap is a generic unordered map backed by a Go builtin map.
// It implements MutableMap[K, V].
type HashMap[K comparable, V any] struct {
	m map[K]V
}

// NewHashMap creates an empty HashMap.
func NewHashMap[K comparable, V any]() *HashMap[K, V] {
	return &HashMap[K, V]{m: make(map[K]V)}
}

// NewHashMapWithCapacity creates a HashMap with pre-allocated capacity.
func NewHashMapWithCapacity[K comparable, V any](capacity int) *HashMap[K, V] {
	return &HashMap[K, V]{m: make(map[K]V, capacity)}
}

// ── MapIterable ───────────────────────────────────────────────────────

func (h *HashMap[K, V]) Get(key K) (V, bool) {
	v, ok := h.m[key]
	return v, ok
}

// GetOrDefault returns the value for key, or defaultValue if not found.
func (h *HashMap[K, V]) GetOrDefault(key K, defaultValue V) V {
	if v, ok := h.m[key]; ok {
		return v
	}
	return defaultValue
}

func (h *HashMap[K, V]) ContainsKey(key K) bool {
	_, ok := h.m[key]
	return ok
}

func (h *HashMap[K, V]) Size() int { return len(h.m) }

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (h *HashMap[K, V]) Len() int      { return h.Size() }
func (h *HashMap[K, V]) IsEmpty() bool { return len(h.m) == 0 }

func (h *HashMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range h.m {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (h *HashMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range h.m {
			if !yield(k) {
				return
			}
		}
	}
}

func (h *HashMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range h.m {
			if !yield(v) {
				return
			}
		}
	}
}

func (h *HashMap[K, V]) ForEach(f func(K, V)) {
	for k, v := range h.m {
		f(k, v)
	}
}

func (h *HashMap[K, V]) AnySatisfy(predicate func(K, V) bool) bool {
	for k, v := range h.m {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

func (h *HashMap[K, V]) AllSatisfy(predicate func(K, V) bool) bool {
	for k, v := range h.m {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

func (h *HashMap[K, V]) NoneSatisfy(predicate func(K, V) bool) bool {
	for k, v := range h.m {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// ── MutableMap ────────────────────────────────────────────────────────

func (h *HashMap[K, V]) Put(key K, value V) (V, bool) {
	if h.m == nil {
		h.m = make(map[K]V)
	}
	old, existed := h.m[key]
	h.m[key] = value
	return old, existed
}

func (h *HashMap[K, V]) Remove(key K) (V, bool) {
	old, existed := h.m[key]
	if existed {
		delete(h.m, key)
	}
	return old, existed
}

func (h *HashMap[K, V]) Clear() {
	clear(h.m)
}

// ── Functional operations ─────────────────────────────────────────────

// Select returns a new HashMap with entries satisfying the predicate.
func (h *HashMap[K, V]) Select(predicate func(K, V) bool) *HashMap[K, V] {
	result := NewHashMap[K, V]()
	for k, v := range h.m {
		if predicate(k, v) {
			result.m[k] = v
		}
	}
	return result
}

// Reject returns a new HashMap with entries NOT satisfying the predicate.
func (h *HashMap[K, V]) Reject(predicate func(K, V) bool) *HashMap[K, V] {
	result := NewHashMap[K, V]()
	for k, v := range h.m {
		if !predicate(k, v) {
			result.m[k] = v
		}
	}
	return result
}

// Detect returns the first entry satisfying the predicate (iteration order is undefined).
func (h *HashMap[K, V]) Detect(predicate func(K, V) bool) (K, V, bool) {
	for k, v := range h.m {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk K
	var zv V
	return zk, zv, false
}

// Count returns the number of entries satisfying the predicate.
func (h *HashMap[K, V]) Count(predicate func(K, V) bool) int {
	n := 0
	for k, v := range h.m {
		if predicate(k, v) {
			n++
		}
	}
	return n
}

// KeysToSlice returns all keys as a slice.
func (h *HashMap[K, V]) KeysToSlice() []K {
	result := make([]K, 0, len(h.m))
	for k := range h.m {
		result = append(result, k)
	}
	return result
}

// ValuesToSlice returns all values as a slice.
func (h *HashMap[K, V]) ValuesToSlice() []V {
	result := make([]V, 0, len(h.m))
	for _, v := range h.m {
		result = append(result, v)
	}
	return result
}

// ── Stringer ──────────────────────────────────────────────────────────

func (h *HashMap[K, V]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for k, v := range h.m {
		if !first {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%v: %v", k, v)
		first = false
	}
	b.WriteByte('}')
	return b.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableMap[string, int] = (*HashMap[string, int])(nil)
var _ MutableMap[int, string] = (*HashMap[int, string])(nil)
