
package hashmap

import (
	"iter"
)

// ImmutableFloat32CharHashMap is an immutable view of a Float32CharHashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableFloat32CharHashMap struct {
	delegate *Float32CharHashMap
}

// NewImmutableFloat32CharHashMap creates an immutable map from key-value pairs.
func NewImmutableFloat32CharHashMap(pairs ...struct {
	Key   float32
	Value uint16
}) *ImmutableFloat32CharHashMap {
	m := NewFloat32CharHashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableFloat32CharHashMap{delegate: m}
}

// ImmutableFloat32CharHashMapFrom creates an immutable copy of a mutable map.
func ImmutableFloat32CharHashMapFrom(m *Float32CharHashMap) *ImmutableFloat32CharHashMap {
	copy := NewFloat32CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k float32, v uint16) {
		copy.Put(k, v)
	})
	return &ImmutableFloat32CharHashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableFloat32CharHashMap) Get(key float32) (uint16, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableFloat32CharHashMap) GetOrDefault(key float32, defaultValue uint16) uint16 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableFloat32CharHashMap) ContainsKey(key float32) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableFloat32CharHashMap) ContainsValue(value uint16) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableFloat32CharHashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableFloat32CharHashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableFloat32CharHashMap) All() iter.Seq2[float32, uint16] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableFloat32CharHashMap) Keys() iter.Seq[float32] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableFloat32CharHashMap) Values() iter.Seq[uint16] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableFloat32CharHashMap) ForEach(f func(float32, uint16)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableFloat32CharHashMap) Select(predicate func(float32, uint16) bool) *ImmutableFloat32CharHashMap {
	return &ImmutableFloat32CharHashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableFloat32CharHashMap) Reject(predicate func(float32, uint16) bool) *ImmutableFloat32CharHashMap {
	return &ImmutableFloat32CharHashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableFloat32CharHashMap) AnySatisfy(predicate func(float32, uint16) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableFloat32CharHashMap) AllSatisfy(predicate func(float32, uint16) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableFloat32CharHashMap) NoneSatisfy(predicate func(float32, uint16) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableFloat32CharHashMap) KeysToSlice() []float32 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableFloat32CharHashMap) ValuesToSlice() []uint16 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableFloat32CharHashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableFloat32CharHashMap) Equals(other *ImmutableFloat32CharHashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableFloat32CharHashMap) ToMutable() *Float32CharHashMap {
	copy := NewFloat32CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k float32, v uint16) {
		copy.Put(k, v)
	})
	return copy
}
