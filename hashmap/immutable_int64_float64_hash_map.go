
package hashmap

import (
	"iter"
)

// ImmutableInt64Float64HashMap is an immutable view of a Int64Float64HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt64Float64HashMap struct {
	delegate *Int64Float64HashMap
}

// NewImmutableInt64Float64HashMap creates an immutable map from key-value pairs.
func NewImmutableInt64Float64HashMap(pairs ...struct {
	Key   int64
	Value float64
}) *ImmutableInt64Float64HashMap {
	m := NewInt64Float64HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt64Float64HashMap{delegate: m}
}

// ImmutableInt64Float64HashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt64Float64HashMapFrom(m *Int64Float64HashMap) *ImmutableInt64Float64HashMap {
	copy := NewInt64Float64HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int64, v float64) {
		copy.Put(k, v)
	})
	return &ImmutableInt64Float64HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt64Float64HashMap) Get(key int64) (float64, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt64Float64HashMap) GetOrDefault(key int64, defaultValue float64) float64 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt64Float64HashMap) ContainsKey(key int64) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt64Float64HashMap) ContainsValue(value float64) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt64Float64HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt64Float64HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt64Float64HashMap) All() iter.Seq2[int64, float64] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt64Float64HashMap) Keys() iter.Seq[int64] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt64Float64HashMap) Values() iter.Seq[float64] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt64Float64HashMap) ForEach(f func(int64, float64)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt64Float64HashMap) Select(predicate func(int64, float64) bool) *ImmutableInt64Float64HashMap {
	return &ImmutableInt64Float64HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt64Float64HashMap) Reject(predicate func(int64, float64) bool) *ImmutableInt64Float64HashMap {
	return &ImmutableInt64Float64HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt64Float64HashMap) AnySatisfy(predicate func(int64, float64) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt64Float64HashMap) AllSatisfy(predicate func(int64, float64) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt64Float64HashMap) NoneSatisfy(predicate func(int64, float64) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt64Float64HashMap) KeysToSlice() []int64 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt64Float64HashMap) ValuesToSlice() []float64 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt64Float64HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt64Float64HashMap) Equals(other *ImmutableInt64Float64HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt64Float64HashMap) ToMutable() *Int64Float64HashMap {
	copy := NewInt64Float64HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int64, v float64) {
		copy.Put(k, v)
	})
	return copy
}
