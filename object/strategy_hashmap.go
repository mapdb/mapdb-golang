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

type shmEntry[K any, V any] struct {
	key      K
	value    V
	occupied bool
}

// HashMapWithStrategy is an open-addressing hash map that uses a pluggable
// HashingStrategy for key identity. This allows case-insensitive maps,
// maps keyed by extracted fields, etc.
type HashMapWithStrategy[K any, V any] struct {
	entries  []shmEntry[K, V]
	size     int
	strategy HashingStrategy[K]
}

// NewHashMapWithStrategy creates an empty map using the given hashing strategy.
func NewHashMapWithStrategy[K any, V any](strategy HashingStrategy[K]) *HashMapWithStrategy[K, V] {
	return NewHashMapWithStrategyCapacity[K, V](strategy, strategySetDefaultCapacity)
}

// NewHashMapWithStrategyCapacity creates an empty map with pre-allocated capacity.
func NewHashMapWithStrategyCapacity[K any, V any](strategy HashingStrategy[K], capacity int) *HashMapWithStrategy[K, V] {
	cap := strategyNextPow2(capacity)
	return &HashMapWithStrategy[K, V]{
		entries:  make([]shmEntry[K, V], cap),
		size:     0,
		strategy: strategy,
	}
}

func (m *HashMapWithStrategy[K, V]) Put(key K, value V) (V, bool) {
	if m.needsResize() {
		m.resize()
	}
	mask := len(m.entries) - 1
	idx := int(m.strategy.HashCode(key)) & mask
	for {
		e := &m.entries[idx]
		if !e.occupied {
			e.key = key
			e.value = value
			e.occupied = true
			m.size++
			var zero V
			return zero, false
		}
		if m.strategy.Equals(e.key, key) {
			old := e.value
			e.value = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

func (m *HashMapWithStrategy[K, V]) Get(key K) (V, bool) {
	if m.size == 0 {
		var zero V
		return zero, false
	}
	mask := len(m.entries) - 1
	idx := int(m.strategy.HashCode(key)) & mask
	for {
		e := &m.entries[idx]
		if !e.occupied {
			var zero V
			return zero, false
		}
		if m.strategy.Equals(e.key, key) {
			return e.value, true
		}
		idx = (idx + 1) & mask
	}
}

func (m *HashMapWithStrategy[K, V]) GetOrDefault(key K, defaultValue V) V {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

func (m *HashMapWithStrategy[K, V]) Remove(key K) (V, bool) {
	if m.size == 0 {
		var zero V
		return zero, false
	}
	mask := len(m.entries) - 1
	idx := int(m.strategy.HashCode(key)) & mask
	for {
		e := &m.entries[idx]
		if !e.occupied {
			var zero V
			return zero, false
		}
		if m.strategy.Equals(e.key, key) {
			old := e.value
			m.entries[idx] = shmEntry[K, V]{}
			m.size--
			m.rehashFrom(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

func (m *HashMapWithStrategy[K, V]) ContainsKey(key K) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *HashMapWithStrategy[K, V]) Size() int { return m.size }

// Len returns the number of elements. It is an alias for Size, matching
// Go convention (sort.Interface, container/list, bytes.Buffer).
func (m *HashMapWithStrategy[K, V]) Len() int      { return m.Size() }
func (m *HashMapWithStrategy[K, V]) IsEmpty() bool { return m.size == 0 }

func (m *HashMapWithStrategy[K, V]) Clear() {
	for i := range m.entries {
		m.entries[i] = shmEntry[K, V]{}
	}
	m.size = 0
}

func (m *HashMapWithStrategy[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].key, m.entries[i].value) {
					return
				}
			}
		}
	}
}

func (m *HashMapWithStrategy[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].key) {
					return
				}
			}
		}
	}
}

func (m *HashMapWithStrategy[K, V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].value) {
					return
				}
			}
		}
	}
}

func (m *HashMapWithStrategy[K, V]) ForEach(f func(K, V)) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key, m.entries[i].value)
		}
	}
}

func (m *HashMapWithStrategy[K, V]) Select(predicate func(K, V) bool) *HashMapWithStrategy[K, V] {
	result := NewHashMapWithStrategy[K, V](m.strategy)
	m.ForEach(func(k K, v V) {
		if predicate(k, v) {
			result.Put(k, v)
		}
	})
	return result
}

func (m *HashMapWithStrategy[K, V]) Reject(predicate func(K, V) bool) *HashMapWithStrategy[K, V] {
	result := NewHashMapWithStrategy[K, V](m.strategy)
	m.ForEach(func(k K, v V) {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	})
	return result
}

func (m *HashMapWithStrategy[K, V]) KeysToSlice() []K {
	result := make([]K, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].key)
		}
	}
	return result
}

func (m *HashMapWithStrategy[K, V]) ValuesToSlice() []V {
	result := make([]V, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].value)
		}
	}
	return result
}

func (m *HashMapWithStrategy[K, V]) String() string {
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for i := range m.entries {
		if m.entries[i].occupied {
			if !first {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "%v: %v", m.entries[i].key, m.entries[i].value)
			first = false
		}
	}
	b.WriteByte('}')
	return b.String()
}

// ── internal ──────────────────────────────────────────────────────────

func (m *HashMapWithStrategy[K, V]) needsResize() bool {
	return (m.size+1)*4 >= len(m.entries)*3
}

func (m *HashMapWithStrategy[K, V]) resize() {
	old := m.entries
	newCap := len(old) * 2
	if newCap == 0 {
		newCap = strategySetDefaultCapacity
	}
	m.entries = make([]shmEntry[K, V], newCap)
	m.size = 0
	for i := range old {
		if old[i].occupied {
			m.Put(old[i].key, old[i].value)
		}
	}
}

func (m *HashMapWithStrategy[K, V]) rehashFrom(deleted, mask int) {
	c := len(m.entries)
	idx := (deleted + 1) & mask
	for m.entries[idx].occupied {
		ideal := int(m.strategy.HashCode(m.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			m.entries[deleted] = m.entries[idx]
			m.entries[idx] = shmEntry[K, V]{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}
