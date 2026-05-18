
package hashmap

import (
	"iter"
)

// ImmutableInt16ObjectHashMap is an immutable view of a Int16ObjectHashMap.
type ImmutableInt16ObjectHashMap[V any] struct {
	delegate *Int16ObjectHashMap[V]
}

// NewImmutableInt16ObjectHashMapFrom creates an immutable int16-object map by copying entries from a mutable map.
func NewImmutableInt16ObjectHashMapFrom[V any](m *Int16ObjectHashMap[V]) *ImmutableInt16ObjectHashMap[V] {
	copy := NewInt16ObjectHashMapWithCapacity[V](m.Size() * 2)
	m.ForEach(func(k int16, v V) {
		copy.Put(k, v)
	})
	return &ImmutableInt16ObjectHashMap[V]{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt16ObjectHashMap[V]) Get(key int16) (V, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt16ObjectHashMap[V]) GetOrDefault(key int16, defaultValue V) V {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt16ObjectHashMap[V]) ContainsKey(key int16) bool {
	return m.delegate.ContainsKey(key)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt16ObjectHashMap[V]) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt16ObjectHashMap[V]) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt16ObjectHashMap[V]) All() iter.Seq2[int16, V] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt16ObjectHashMap[V]) Keys() iter.Seq[int16] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt16ObjectHashMap[V]) Values() iter.Seq[V] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt16ObjectHashMap[V]) ForEach(f func(int16, V)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries satisfying the predicate.
func (m *ImmutableInt16ObjectHashMap[V]) Select(predicate func(int16, V) bool) *ImmutableInt16ObjectHashMap[V] {
	return &ImmutableInt16ObjectHashMap[V]{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries not satisfying the predicate.
func (m *ImmutableInt16ObjectHashMap[V]) Reject(predicate func(int16, V) bool) *ImmutableInt16ObjectHashMap[V] {
	return &ImmutableInt16ObjectHashMap[V]{delegate: m.delegate.Reject(predicate)}
}

// String returns a string representation.
func (m *ImmutableInt16ObjectHashMap[V]) String() string {
	return m.delegate.String()
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt16ObjectHashMap[V]) ToMutable() *Int16ObjectHashMap[V] {
	copy := NewInt16ObjectHashMapWithCapacity[V](m.Size() * 2)
	m.ForEach(func(k int16, v V) {
		copy.Put(k, v)
	})
	return copy
}
