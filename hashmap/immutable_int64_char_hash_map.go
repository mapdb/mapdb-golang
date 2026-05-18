
package hashmap

import (
	"iter"
)

// ImmutableInt64CharHashMap is an immutable view of a Int64CharHashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt64CharHashMap struct {
	delegate *Int64CharHashMap
}

// NewImmutableInt64CharHashMap creates an immutable map from key-value pairs.
func NewImmutableInt64CharHashMap(pairs ...struct {
	Key   int64
	Value uint16
}) *ImmutableInt64CharHashMap {
	m := NewInt64CharHashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt64CharHashMap{delegate: m}
}

// ImmutableInt64CharHashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt64CharHashMapFrom(m *Int64CharHashMap) *ImmutableInt64CharHashMap {
	copy := NewInt64CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int64, v uint16) {
		copy.Put(k, v)
	})
	return &ImmutableInt64CharHashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt64CharHashMap) Get(key int64) (uint16, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt64CharHashMap) GetOrDefault(key int64, defaultValue uint16) uint16 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt64CharHashMap) ContainsKey(key int64) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt64CharHashMap) ContainsValue(value uint16) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt64CharHashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt64CharHashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt64CharHashMap) All() iter.Seq2[int64, uint16] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt64CharHashMap) Keys() iter.Seq[int64] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt64CharHashMap) Values() iter.Seq[uint16] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt64CharHashMap) ForEach(f func(int64, uint16)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt64CharHashMap) Select(predicate func(int64, uint16) bool) *ImmutableInt64CharHashMap {
	return &ImmutableInt64CharHashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt64CharHashMap) Reject(predicate func(int64, uint16) bool) *ImmutableInt64CharHashMap {
	return &ImmutableInt64CharHashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt64CharHashMap) AnySatisfy(predicate func(int64, uint16) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt64CharHashMap) AllSatisfy(predicate func(int64, uint16) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt64CharHashMap) NoneSatisfy(predicate func(int64, uint16) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt64CharHashMap) KeysToSlice() []int64 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt64CharHashMap) ValuesToSlice() []uint16 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt64CharHashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt64CharHashMap) Equals(other *ImmutableInt64CharHashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt64CharHashMap) ToMutable() *Int64CharHashMap {
	copy := NewInt64CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int64, v uint16) {
		copy.Put(k, v)
	})
	return copy
}
