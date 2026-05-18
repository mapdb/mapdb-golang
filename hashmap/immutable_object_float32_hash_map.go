
package hashmap

import (
	"iter"
)

// ImmutableObjectFloat32HashMap is an immutable view of an ObjectFloat32HashMap.
type ImmutableObjectFloat32HashMap[K comparable] struct {
	delegate *ObjectFloat32HashMap[K]
}

// NewImmutableObjectFloat32HashMap creates an immutable object-float32 map by copying entries from a mutable map.
func NewImmutableObjectFloat32HashMapFrom[K comparable](m *ObjectFloat32HashMap[K]) *ImmutableObjectFloat32HashMap[K] {
	copy := NewObjectFloat32HashMapWithCapacity[K](m.Size() * 2)
	m.ForEach(func(k K, v float32) {
		copy.Put(k, v)
	})
	return &ImmutableObjectFloat32HashMap[K]{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableObjectFloat32HashMap[K]) Get(key K) (float32, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableObjectFloat32HashMap[K]) GetOrDefault(key K, defaultValue float32) float32 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableObjectFloat32HashMap[K]) ContainsKey(key K) bool {
	return m.delegate.ContainsKey(key)
}

// Size returns the number of key-value pairs.
func (m *ImmutableObjectFloat32HashMap[K]) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableObjectFloat32HashMap[K]) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableObjectFloat32HashMap[K]) All() iter.Seq2[K, float32] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableObjectFloat32HashMap[K]) Keys() iter.Seq[K] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableObjectFloat32HashMap[K]) Values() iter.Seq[float32] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableObjectFloat32HashMap[K]) ForEach(f func(K, float32)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries satisfying the predicate.
func (m *ImmutableObjectFloat32HashMap[K]) Select(predicate func(K, float32) bool) *ImmutableObjectFloat32HashMap[K] {
	return &ImmutableObjectFloat32HashMap[K]{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries not satisfying the predicate.
func (m *ImmutableObjectFloat32HashMap[K]) Reject(predicate func(K, float32) bool) *ImmutableObjectFloat32HashMap[K] {
	return &ImmutableObjectFloat32HashMap[K]{delegate: m.delegate.Reject(predicate)}
}

// String returns a string representation.
func (m *ImmutableObjectFloat32HashMap[K]) String() string {
	return m.delegate.String()
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableObjectFloat32HashMap[K]) ToMutable() *ObjectFloat32HashMap[K] {
	copy := NewObjectFloat32HashMapWithCapacity[K](m.Size() * 2)
	m.ForEach(func(k K, v float32) {
		copy.Put(k, v)
	})
	return copy
}
