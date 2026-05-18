
package hashmap

import (
	"iter"
)

// ImmutableObjectInt32HashMap is an immutable view of an ObjectInt32HashMap.
type ImmutableObjectInt32HashMap[K comparable] struct {
	delegate *ObjectInt32HashMap[K]
}

// NewImmutableObjectInt32HashMap creates an immutable object-int32 map by copying entries from a mutable map.
func NewImmutableObjectInt32HashMapFrom[K comparable](m *ObjectInt32HashMap[K]) *ImmutableObjectInt32HashMap[K] {
	copy := NewObjectInt32HashMapWithCapacity[K](m.Size() * 2)
	m.ForEach(func(k K, v int32) {
		copy.Put(k, v)
	})
	return &ImmutableObjectInt32HashMap[K]{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableObjectInt32HashMap[K]) Get(key K) (int32, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableObjectInt32HashMap[K]) GetOrDefault(key K, defaultValue int32) int32 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableObjectInt32HashMap[K]) ContainsKey(key K) bool {
	return m.delegate.ContainsKey(key)
}

// Size returns the number of key-value pairs.
func (m *ImmutableObjectInt32HashMap[K]) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableObjectInt32HashMap[K]) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableObjectInt32HashMap[K]) All() iter.Seq2[K, int32] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableObjectInt32HashMap[K]) Keys() iter.Seq[K] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableObjectInt32HashMap[K]) Values() iter.Seq[int32] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableObjectInt32HashMap[K]) ForEach(f func(K, int32)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries satisfying the predicate.
func (m *ImmutableObjectInt32HashMap[K]) Select(predicate func(K, int32) bool) *ImmutableObjectInt32HashMap[K] {
	return &ImmutableObjectInt32HashMap[K]{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries not satisfying the predicate.
func (m *ImmutableObjectInt32HashMap[K]) Reject(predicate func(K, int32) bool) *ImmutableObjectInt32HashMap[K] {
	return &ImmutableObjectInt32HashMap[K]{delegate: m.delegate.Reject(predicate)}
}

// String returns a string representation.
func (m *ImmutableObjectInt32HashMap[K]) String() string {
	return m.delegate.String()
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableObjectInt32HashMap[K]) ToMutable() *ObjectInt32HashMap[K] {
	copy := NewObjectInt32HashMapWithCapacity[K](m.Size() * 2)
	m.ForEach(func(k K, v int32) {
		copy.Put(k, v)
	})
	return copy
}
