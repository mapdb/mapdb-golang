
package hashmap

import (
	"iter"
)

// ImmutableInt32Float64HashMap is an immutable view of a Int32Float64HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt32Float64HashMap struct {
	delegate *Int32Float64HashMap
}

// NewImmutableInt32Float64HashMap creates an immutable map from key-value pairs.
func NewImmutableInt32Float64HashMap(pairs ...struct {
	Key   int32
	Value float64
}) *ImmutableInt32Float64HashMap {
	m := NewInt32Float64HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt32Float64HashMap{delegate: m}
}

// ImmutableInt32Float64HashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt32Float64HashMapFrom(m *Int32Float64HashMap) *ImmutableInt32Float64HashMap {
	copy := NewInt32Float64HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int32, v float64) {
		copy.Put(k, v)
	})
	return &ImmutableInt32Float64HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt32Float64HashMap) Get(key int32) (float64, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt32Float64HashMap) GetOrDefault(key int32, defaultValue float64) float64 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt32Float64HashMap) ContainsKey(key int32) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt32Float64HashMap) ContainsValue(value float64) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt32Float64HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt32Float64HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt32Float64HashMap) All() iter.Seq2[int32, float64] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt32Float64HashMap) Keys() iter.Seq[int32] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt32Float64HashMap) Values() iter.Seq[float64] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt32Float64HashMap) ForEach(f func(int32, float64)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt32Float64HashMap) Select(predicate func(int32, float64) bool) *ImmutableInt32Float64HashMap {
	return &ImmutableInt32Float64HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt32Float64HashMap) Reject(predicate func(int32, float64) bool) *ImmutableInt32Float64HashMap {
	return &ImmutableInt32Float64HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt32Float64HashMap) AnySatisfy(predicate func(int32, float64) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt32Float64HashMap) AllSatisfy(predicate func(int32, float64) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt32Float64HashMap) NoneSatisfy(predicate func(int32, float64) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt32Float64HashMap) KeysToSlice() []int32 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt32Float64HashMap) ValuesToSlice() []float64 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt32Float64HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt32Float64HashMap) Equals(other *ImmutableInt32Float64HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt32Float64HashMap) ToMutable() *Int32Float64HashMap {
	copy := NewInt32Float64HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int32, v float64) {
		copy.Put(k, v)
	})
	return copy
}
