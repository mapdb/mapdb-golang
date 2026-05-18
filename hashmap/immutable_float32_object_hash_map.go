
package hashmap

import (
	"iter"
)

// ImmutableFloat32ObjectHashMap is an immutable view of a Float32ObjectHashMap.
type ImmutableFloat32ObjectHashMap[V any] struct {
	delegate *Float32ObjectHashMap[V]
}

// NewImmutableFloat32ObjectHashMapFrom creates an immutable float32-object map by copying entries from a mutable map.
func NewImmutableFloat32ObjectHashMapFrom[V any](m *Float32ObjectHashMap[V]) *ImmutableFloat32ObjectHashMap[V] {
	copy := NewFloat32ObjectHashMapWithCapacity[V](m.Size() * 2)
	m.ForEach(func(k float32, v V) {
		copy.Put(k, v)
	})
	return &ImmutableFloat32ObjectHashMap[V]{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableFloat32ObjectHashMap[V]) Get(key float32) (V, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableFloat32ObjectHashMap[V]) GetOrDefault(key float32, defaultValue V) V {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableFloat32ObjectHashMap[V]) ContainsKey(key float32) bool {
	return m.delegate.ContainsKey(key)
}

// Size returns the number of key-value pairs.
func (m *ImmutableFloat32ObjectHashMap[V]) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableFloat32ObjectHashMap[V]) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableFloat32ObjectHashMap[V]) All() iter.Seq2[float32, V] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableFloat32ObjectHashMap[V]) Keys() iter.Seq[float32] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableFloat32ObjectHashMap[V]) Values() iter.Seq[V] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableFloat32ObjectHashMap[V]) ForEach(f func(float32, V)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries satisfying the predicate.
func (m *ImmutableFloat32ObjectHashMap[V]) Select(predicate func(float32, V) bool) *ImmutableFloat32ObjectHashMap[V] {
	return &ImmutableFloat32ObjectHashMap[V]{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries not satisfying the predicate.
func (m *ImmutableFloat32ObjectHashMap[V]) Reject(predicate func(float32, V) bool) *ImmutableFloat32ObjectHashMap[V] {
	return &ImmutableFloat32ObjectHashMap[V]{delegate: m.delegate.Reject(predicate)}
}

// String returns a string representation.
func (m *ImmutableFloat32ObjectHashMap[V]) String() string {
	return m.delegate.String()
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableFloat32ObjectHashMap[V]) ToMutable() *Float32ObjectHashMap[V] {
	copy := NewFloat32ObjectHashMapWithCapacity[V](m.Size() * 2)
	m.ForEach(func(k float32, v V) {
		copy.Put(k, v)
	})
	return copy
}
