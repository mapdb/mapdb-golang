
package hashmap

import (
	"iter"
)

// ImmutableFloat64CharHashMap is an immutable view of a Float64CharHashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableFloat64CharHashMap struct {
	delegate *Float64CharHashMap
}

// NewImmutableFloat64CharHashMap creates an immutable map from key-value pairs.
func NewImmutableFloat64CharHashMap(pairs ...struct {
	Key   float64
	Value uint16
}) *ImmutableFloat64CharHashMap {
	m := NewFloat64CharHashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableFloat64CharHashMap{delegate: m}
}

// ImmutableFloat64CharHashMapFrom creates an immutable copy of a mutable map.
func ImmutableFloat64CharHashMapFrom(m *Float64CharHashMap) *ImmutableFloat64CharHashMap {
	copy := NewFloat64CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k float64, v uint16) {
		copy.Put(k, v)
	})
	return &ImmutableFloat64CharHashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableFloat64CharHashMap) Get(key float64) (uint16, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableFloat64CharHashMap) GetOrDefault(key float64, defaultValue uint16) uint16 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableFloat64CharHashMap) ContainsKey(key float64) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableFloat64CharHashMap) ContainsValue(value uint16) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableFloat64CharHashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableFloat64CharHashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableFloat64CharHashMap) All() iter.Seq2[float64, uint16] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableFloat64CharHashMap) Keys() iter.Seq[float64] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableFloat64CharHashMap) Values() iter.Seq[uint16] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableFloat64CharHashMap) ForEach(f func(float64, uint16)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableFloat64CharHashMap) Select(predicate func(float64, uint16) bool) *ImmutableFloat64CharHashMap {
	return &ImmutableFloat64CharHashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableFloat64CharHashMap) Reject(predicate func(float64, uint16) bool) *ImmutableFloat64CharHashMap {
	return &ImmutableFloat64CharHashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableFloat64CharHashMap) AnySatisfy(predicate func(float64, uint16) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableFloat64CharHashMap) AllSatisfy(predicate func(float64, uint16) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableFloat64CharHashMap) NoneSatisfy(predicate func(float64, uint16) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableFloat64CharHashMap) KeysToSlice() []float64 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableFloat64CharHashMap) ValuesToSlice() []uint16 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableFloat64CharHashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableFloat64CharHashMap) Equals(other *ImmutableFloat64CharHashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableFloat64CharHashMap) ToMutable() *Float64CharHashMap {
	copy := NewFloat64CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k float64, v uint16) {
		copy.Put(k, v)
	})
	return copy
}
