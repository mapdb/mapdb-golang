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

// HashBiMap is a generic bidirectional map backed by two Go builtin maps.
// Both keys and values must be unique (bijection). When a Put inserts a
// value that already exists under a different key, the old key is removed.
//
// It implements MutableBiMap[K, V].
type HashBiMap[K, V comparable] struct {
	forward map[K]V
	inverse map[V]K
}

// NewHashBiMap creates an empty HashBiMap.
func NewHashBiMap[K, V comparable]() *HashBiMap[K, V] {
	return &HashBiMap[K, V]{
		forward: make(map[K]V),
		inverse: make(map[V]K),
	}
}

// NewHashBiMapWithCapacity creates a HashBiMap with pre-allocated capacity.
func NewHashBiMapWithCapacity[K, V comparable](capacity int) *HashBiMap[K, V] {
	return &HashBiMap[K, V]{
		forward: make(map[K]V, capacity),
		inverse: make(map[V]K, capacity),
	}
}

// ── MapIterable ───────────────────────────────────────────────────────

func (b *HashBiMap[K, V]) Get(key K) (V, bool) {
	v, ok := b.forward[key]
	return v, ok
}

func (b *HashBiMap[K, V]) ContainsKey(key K) bool {
	_, ok := b.forward[key]
	return ok
}

func (b *HashBiMap[K, V]) Size() int { return len(b.forward) }

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (b *HashBiMap[K, V]) Len() int      { return b.Size() }
func (b *HashBiMap[K, V]) IsEmpty() bool { return len(b.forward) == 0 }

func (b *HashBiMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range b.forward {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (b *HashBiMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range b.forward {
			if !yield(k) {
				return
			}
		}
	}
}

func (b *HashBiMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range b.forward {
			if !yield(v) {
				return
			}
		}
	}
}

func (b *HashBiMap[K, V]) ForEach(f func(K, V)) {
	for k, v := range b.forward {
		f(k, v)
	}
}

func (b *HashBiMap[K, V]) AnySatisfy(predicate func(K, V) bool) bool {
	for k, v := range b.forward {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

func (b *HashBiMap[K, V]) AllSatisfy(predicate func(K, V) bool) bool {
	for k, v := range b.forward {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

func (b *HashBiMap[K, V]) NoneSatisfy(predicate func(K, V) bool) bool {
	for k, v := range b.forward {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// ── BiMap ─────────────────────────────────────────────────────────────

// GetInverse returns the key for the given value (reverse lookup).
func (b *HashBiMap[K, V]) GetInverse(value V) (K, bool) {
	k, ok := b.inverse[value]
	return k, ok
}

// ContainsValue returns true if the value exists in the map.
func (b *HashBiMap[K, V]) ContainsValue(value V) bool {
	_, ok := b.inverse[value]
	return ok
}

// ── MutableBiMap ──────────────────────────────────────────────────────

// Put adds a key-value pair. If the value already exists under a different
// key, that old key is removed to maintain the bijection invariant.
// Returns the old value for the key if it existed.
func (b *HashBiMap[K, V]) Put(key K, value V) (V, bool) {
	if b.forward == nil {
		b.forward = make(map[K]V)
		b.inverse = make(map[V]K)
	}
	// If this value already maps to a different key, remove that key
	if existingKey, ok := b.inverse[value]; ok {
		if existingKey != key {
			delete(b.forward, existingKey)
			delete(b.inverse, value)
		}
	}

	// If this key already maps to a different value, remove old inverse
	oldValue, existed := b.forward[key]
	if existed {
		delete(b.inverse, oldValue)
	}

	b.forward[key] = value
	b.inverse[value] = key
	return oldValue, existed
}

// ForcePut is an alias for Put — the bijection invariant is always enforced.
func (b *HashBiMap[K, V]) ForcePut(key K, value V) (V, bool) {
	return b.Put(key, value)
}

func (b *HashBiMap[K, V]) Remove(key K) (V, bool) {
	v, ok := b.forward[key]
	if ok {
		delete(b.forward, key)
		delete(b.inverse, v)
	}
	return v, ok
}

// RemoveInverse removes the entry with the given value (reverse removal).
func (b *HashBiMap[K, V]) RemoveInverse(value V) (K, bool) {
	k, ok := b.inverse[value]
	if ok {
		delete(b.inverse, value)
		delete(b.forward, k)
	}
	return k, ok
}

func (b *HashBiMap[K, V]) Clear() {
	clear(b.forward)
	clear(b.inverse)
}

// ── View ──────────────────────────────────────────────────────────────

// Inverse returns a new HashBiMap with keys and values swapped.
// This is a snapshot copy, not a live view.
func (b *HashBiMap[K, V]) Inverse() *HashBiMap[V, K] {
	inv := NewHashBiMapWithCapacity[V, K](len(b.forward))
	for k, v := range b.forward {
		inv.forward[v] = k
		inv.inverse[k] = v
	}
	return inv
}

// ── Convenience ───────────────────────────────────────────────────────

// KeysToSlice returns all keys as a slice.
func (b *HashBiMap[K, V]) KeysToSlice() []K {
	result := make([]K, 0, len(b.forward))
	for k := range b.forward {
		result = append(result, k)
	}
	return result
}

// ValuesToSlice returns all values as a slice.
func (b *HashBiMap[K, V]) ValuesToSlice() []V {
	result := make([]V, 0, len(b.forward))
	for _, v := range b.forward {
		result = append(result, v)
	}
	return result
}

// ── Stringer ──────────────────────────────────────────────────────────

func (b *HashBiMap[K, V]) String() string {
	var s strings.Builder
	s.WriteString("{BiMap: ")
	first := true
	for k, v := range b.forward {
		if !first {
			s.WriteString(", ")
		}
		fmt.Fprintf(&s, "%v↔%v", k, v)
		first = false
	}
	s.WriteByte('}')
	return s.String()
}

// ── Interface compliance ──────────────────────────────────────────────

var _ MutableBiMap[string, int] = (*HashBiMap[string, int])(nil)
var _ MutableBiMap[int, string] = (*HashBiMap[int, string])(nil)
