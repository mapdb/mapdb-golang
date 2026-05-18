
package hashmap

import (
	"iter"
)

// ImmutableInt16CharHashMap is an immutable view of a Int16CharHashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt16CharHashMap struct {
	delegate *Int16CharHashMap
}

// NewImmutableInt16CharHashMap creates an immutable map from key-value pairs.
func NewImmutableInt16CharHashMap(pairs ...struct {
	Key   int16
	Value uint16
}) *ImmutableInt16CharHashMap {
	m := NewInt16CharHashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt16CharHashMap{delegate: m}
}

// ImmutableInt16CharHashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt16CharHashMapFrom(m *Int16CharHashMap) *ImmutableInt16CharHashMap {
	copy := NewInt16CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int16, v uint16) {
		copy.Put(k, v)
	})
	return &ImmutableInt16CharHashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt16CharHashMap) Get(key int16) (uint16, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt16CharHashMap) GetOrDefault(key int16, defaultValue uint16) uint16 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt16CharHashMap) ContainsKey(key int16) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt16CharHashMap) ContainsValue(value uint16) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt16CharHashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt16CharHashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt16CharHashMap) All() iter.Seq2[int16, uint16] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt16CharHashMap) Keys() iter.Seq[int16] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt16CharHashMap) Values() iter.Seq[uint16] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt16CharHashMap) ForEach(f func(int16, uint16)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt16CharHashMap) Select(predicate func(int16, uint16) bool) *ImmutableInt16CharHashMap {
	return &ImmutableInt16CharHashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt16CharHashMap) Reject(predicate func(int16, uint16) bool) *ImmutableInt16CharHashMap {
	return &ImmutableInt16CharHashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt16CharHashMap) AnySatisfy(predicate func(int16, uint16) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt16CharHashMap) AllSatisfy(predicate func(int16, uint16) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt16CharHashMap) NoneSatisfy(predicate func(int16, uint16) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt16CharHashMap) KeysToSlice() []int16 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt16CharHashMap) ValuesToSlice() []uint16 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt16CharHashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt16CharHashMap) Equals(other *ImmutableInt16CharHashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt16CharHashMap) ToMutable() *Int16CharHashMap {
	copy := NewInt16CharHashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int16, v uint16) {
		copy.Put(k, v)
	})
	return copy
}
