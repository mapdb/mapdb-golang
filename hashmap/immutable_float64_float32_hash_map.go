
package hashmap

import (
	"iter"
)

// ImmutableFloat64Float32HashMap is an immutable view of a Float64Float32HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableFloat64Float32HashMap struct {
	delegate *Float64Float32HashMap
}

// NewImmutableFloat64Float32HashMap creates an immutable map from key-value pairs.
func NewImmutableFloat64Float32HashMap(pairs ...struct {
	Key   float64
	Value float32
}) *ImmutableFloat64Float32HashMap {
	m := NewFloat64Float32HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableFloat64Float32HashMap{delegate: m}
}

// ImmutableFloat64Float32HashMapFrom creates an immutable copy of a mutable map.
func ImmutableFloat64Float32HashMapFrom(m *Float64Float32HashMap) *ImmutableFloat64Float32HashMap {
	copy := NewFloat64Float32HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k float64, v float32) {
		copy.Put(k, v)
	})
	return &ImmutableFloat64Float32HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableFloat64Float32HashMap) Get(key float64) (float32, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableFloat64Float32HashMap) GetOrDefault(key float64, defaultValue float32) float32 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableFloat64Float32HashMap) ContainsKey(key float64) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableFloat64Float32HashMap) ContainsValue(value float32) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableFloat64Float32HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableFloat64Float32HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableFloat64Float32HashMap) All() iter.Seq2[float64, float32] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableFloat64Float32HashMap) Keys() iter.Seq[float64] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableFloat64Float32HashMap) Values() iter.Seq[float32] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableFloat64Float32HashMap) ForEach(f func(float64, float32)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableFloat64Float32HashMap) Select(predicate func(float64, float32) bool) *ImmutableFloat64Float32HashMap {
	return &ImmutableFloat64Float32HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableFloat64Float32HashMap) Reject(predicate func(float64, float32) bool) *ImmutableFloat64Float32HashMap {
	return &ImmutableFloat64Float32HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableFloat64Float32HashMap) AnySatisfy(predicate func(float64, float32) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableFloat64Float32HashMap) AllSatisfy(predicate func(float64, float32) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableFloat64Float32HashMap) NoneSatisfy(predicate func(float64, float32) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableFloat64Float32HashMap) KeysToSlice() []float64 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableFloat64Float32HashMap) ValuesToSlice() []float32 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableFloat64Float32HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableFloat64Float32HashMap) Equals(other *ImmutableFloat64Float32HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableFloat64Float32HashMap) ToMutable() *Float64Float32HashMap {
	copy := NewFloat64Float32HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k float64, v float32) {
		copy.Put(k, v)
	})
	return copy
}
