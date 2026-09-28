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
//
// Float keys and float values (underlying type float32/float64) use
// bit-pattern identity, not ==: NaN is found and not duplicated, NaN payloads
// are distinct, and -0.0 and +0.0 are distinct; see keyIndex.
type HashBiMap[K, V comparable] struct {
	forward keyIndex[K, V]
	inverse keyIndex[V, K]
}

// NewHashBiMap creates an empty HashBiMap.
func NewHashBiMap[K, V comparable]() *HashBiMap[K, V] {
	return &HashBiMap[K, V]{
		forward: newKeyIndex[K, V](0),
		inverse: newKeyIndex[V, K](0),
	}
}

// NewHashBiMapWithCapacity creates a HashBiMap with pre-allocated capacity.
func NewHashBiMapWithCapacity[K, V comparable](capacity int) *HashBiMap[K, V] {
	return &HashBiMap[K, V]{
		forward: newKeyIndex[K, V](capacity),
		inverse: newKeyIndex[V, K](capacity),
	}
}

// ── MapIterable ───────────────────────────────────────────────────────

func (b *HashBiMap[K, V]) Get(key K) (V, bool) {
	return b.forward.get(key)
}

func (b *HashBiMap[K, V]) ContainsKey(key K) bool {
	return b.forward.contains(key)
}

// Len returns the number of entries. Use b.Len() == 0 to test for emptiness.
func (b *HashBiMap[K, V]) Len() int { return b.forward.len() }

// All: the HashBiMap must not be mutated during this iteration when its key
// type is float32 or float64 (see the package documentation, section "Float
// keys").
func (b *HashBiMap[K, V]) All() iter.Seq2[K, V] {
	return b.forward.all()
}

// Keys: the HashBiMap must not be mutated during this iteration when its key
// type is float32 or float64 (see the package documentation, section "Float
// keys").
func (b *HashBiMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range b.forward.all() {
			if !yield(k) {
				return
			}
		}
	}
}

// Values: the HashBiMap must not be mutated during this iteration when its
// key type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBiMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range b.forward.all() {
			if !yield(v) {
				return
			}
		}
	}
}

// ForEach: the HashBiMap must not be mutated during this iteration when its
// key type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBiMap[K, V]) ForEach(f func(K, V)) {
	for k, v := range b.forward.all() {
		f(k, v)
	}
}

// AnySatisfy: the HashBiMap must not be mutated during this iteration when
// its key type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBiMap[K, V]) AnySatisfy(predicate func(K, V) bool) bool {
	for k, v := range b.forward.all() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy: the HashBiMap must not be mutated during this iteration when
// its key type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBiMap[K, V]) AllSatisfy(predicate func(K, V) bool) bool {
	for k, v := range b.forward.all() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy: the HashBiMap must not be mutated during this iteration when
// its key type is float32 or float64 (see the package documentation, section
// "Float keys").
func (b *HashBiMap[K, V]) NoneSatisfy(predicate func(K, V) bool) bool {
	for k, v := range b.forward.all() {
		if predicate(k, v) {
			return false
		}
	}
	return true
}

// ── BiMap ─────────────────────────────────────────────────────────────

// GetInverse returns the key for the given value (reverse lookup).
func (b *HashBiMap[K, V]) GetInverse(value V) (K, bool) {
	return b.inverse.get(value)
}

// ContainsValue returns true if the value exists in the map.
func (b *HashBiMap[K, V]) ContainsValue(value V) bool {
	return b.inverse.contains(value)
}

// ── MutableBiMap ──────────────────────────────────────────────────────

// Put adds a key-value pair. If the value already exists under a different
// key, that old key is removed to maintain the bijection invariant.
// Returns the old value for the key if it existed.
func (b *HashBiMap[K, V]) Put(key K, value V) (V, bool) {
	// If this value already maps to a different key, remove that key
	if existingKey, ok := b.inverse.get(value); ok {
		if !sameKey(existingKey, key) {
			b.forward.remove(existingKey)
			b.inverse.remove(value)
		}
	}

	// If this key already maps to a different value, remove old inverse
	oldValue, existed := b.forward.get(key)
	if existed {
		b.inverse.remove(oldValue)
	}

	b.forward.put(key, value)
	b.inverse.put(value, key)
	return oldValue, existed
}

// ForcePut is an alias for Put — the bijection invariant is always enforced.
func (b *HashBiMap[K, V]) ForcePut(key K, value V) (V, bool) {
	return b.Put(key, value)
}

func (b *HashBiMap[K, V]) Remove(key K) (V, bool) {
	v, ok := b.forward.remove(key)
	if ok {
		b.inverse.remove(v)
	}
	return v, ok
}

// RemoveInverse removes the entry with the given value (reverse removal).
func (b *HashBiMap[K, V]) RemoveInverse(value V) (K, bool) {
	k, ok := b.inverse.remove(value)
	if ok {
		b.forward.remove(k)
	}
	return k, ok
}

func (b *HashBiMap[K, V]) Clear() {
	b.forward.clear()
	b.inverse.clear()
}

// ── View ──────────────────────────────────────────────────────────────

// Inverse returns a new HashBiMap with keys and values swapped.
// This is a snapshot copy, not a live view.
func (b *HashBiMap[K, V]) Inverse() *HashBiMap[V, K] {
	inv := NewHashBiMapWithCapacity[V, K](b.forward.len())
	for k, v := range b.forward.all() {
		inv.forward.put(v, k)
		inv.inverse.put(k, v)
	}
	return inv
}

// ── Convenience ───────────────────────────────────────────────────────

// KeysToSlice returns all keys as a slice.
func (b *HashBiMap[K, V]) KeysToSlice() []K {
	result := make([]K, 0, b.forward.len())
	for k := range b.forward.all() {
		result = append(result, k)
	}
	return result
}

// ValuesToSlice returns all values as a slice.
func (b *HashBiMap[K, V]) ValuesToSlice() []V {
	result := make([]V, 0, b.forward.len())
	for _, v := range b.forward.all() {
		result = append(result, v)
	}
	return result
}

// ── Stringer ──────────────────────────────────────────────────────────

func (b *HashBiMap[K, V]) String() string {
	var s strings.Builder
	s.WriteString("{BiMap: ")
	first := true
	for k, v := range b.forward.all() {
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
