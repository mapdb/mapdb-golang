
package multimap

import (
	"fmt"
	"iter"
	"strings"
)

// Multimap is a generic one-to-many map: each key maps to a slice of values.
// Uses Go generics — works with any comparable key and any value type.
//
// This is the Go equivalent of Eclipse Collections' Multimap / ArrayListMultimap.
type Multimap[K comparable, V any] struct {
	data map[K][]V
	size int // total values across all keys
}

// NewMultimap creates a new empty Multimap.
func NewMultimap[K comparable, V any]() *Multimap[K, V] {
	return &Multimap[K, V]{data: make(map[K][]V)}
}

// Put adds a value for the key. Always appends (does not check for duplicates).
func (m *Multimap[K, V]) Put(key K, value V) {
	m.data[key] = append(m.data[key], value)
	m.size++
}

// PutAll adds multiple values for the key.
func (m *Multimap[K, V]) PutAll(key K, values ...V) {
	m.data[key] = append(m.data[key], values...)
	m.size += len(values)
}

// Get returns a copy of all values for the key. Returns nil if key not present.
func (m *Multimap[K, V]) Get(key K) []V {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]V, len(vals))
	copy(result, vals)
	return result
}

// ContainsKey returns true if the key has at least one value.
func (m *Multimap[K, V]) ContainsKey(key K) bool {
	vals, ok := m.data[key]
	return ok && len(vals) > 0
}

// RemoveAll removes all values for the key. Returns the removed values.
func (m *Multimap[K, V]) RemoveAll(key K) []V {
	vals := m.data[key]
	if vals != nil {
		m.size -= len(vals)
		delete(m.data, key)
	}
	return vals
}

// Size returns total number of values across all keys.
func (m *Multimap[K, V]) Size() int { return m.size }

// SizeDistinct returns the number of distinct keys.
func (m *Multimap[K, V]) SizeDistinct() int { return len(m.data) }

// IsEmpty returns true if there are no entries.
func (m *Multimap[K, V]) IsEmpty() bool { return m.size == 0 }

// Clear removes all entries.
func (m *Multimap[K, V]) Clear() { m.data = make(map[K][]V); m.size = 0 }

// Keys returns an iter.Seq of distinct keys.
func (m *Multimap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range m.data {
			if !yield(k) {
				return
			}
		}
	}
}

// All returns an iter.Seq2 that yields each (key, value) pair.
// Keys with multiple values appear multiple times.
func (m *Multimap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, vals := range m.data {
			for _, v := range vals {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// ForEachKey calls f for each distinct key with a copy of its values.
func (m *Multimap[K, V]) ForEachKey(f func(K, []V)) {
	for k, vals := range m.data {
		copied := make([]V, len(vals))
		copy(copied, vals)
		f(k, copied)
	}
}

// ForEach calls f for each (key, value) pair.
func (m *Multimap[K, V]) ForEach(f func(K, V)) {
	for k, vals := range m.data {
		for _, v := range vals {
			f(k, v)
		}
	}
}

// String returns a string representation.
func (m *Multimap[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, vals := range m.data {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, vals)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}
