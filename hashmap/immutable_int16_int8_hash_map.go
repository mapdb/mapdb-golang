
package hashmap

import (
	"iter"
)

// ImmutableInt16Int8HashMap is an immutable view of a Int16Int8HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt16Int8HashMap struct {
	delegate *Int16Int8HashMap
}

// NewImmutableInt16Int8HashMap creates an immutable map from key-value pairs.
func NewImmutableInt16Int8HashMap(pairs ...struct {
	Key   int16
	Value int8
}) *ImmutableInt16Int8HashMap {
	m := NewInt16Int8HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt16Int8HashMap{delegate: m}
}

// ImmutableInt16Int8HashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt16Int8HashMapFrom(m *Int16Int8HashMap) *ImmutableInt16Int8HashMap {
	copy := NewInt16Int8HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int16, v int8) {
		copy.Put(k, v)
	})
	return &ImmutableInt16Int8HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt16Int8HashMap) Get(key int16) (int8, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt16Int8HashMap) GetOrDefault(key int16, defaultValue int8) int8 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt16Int8HashMap) ContainsKey(key int16) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt16Int8HashMap) ContainsValue(value int8) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt16Int8HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt16Int8HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt16Int8HashMap) All() iter.Seq2[int16, int8] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt16Int8HashMap) Keys() iter.Seq[int16] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt16Int8HashMap) Values() iter.Seq[int8] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt16Int8HashMap) ForEach(f func(int16, int8)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt16Int8HashMap) Select(predicate func(int16, int8) bool) *ImmutableInt16Int8HashMap {
	return &ImmutableInt16Int8HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt16Int8HashMap) Reject(predicate func(int16, int8) bool) *ImmutableInt16Int8HashMap {
	return &ImmutableInt16Int8HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt16Int8HashMap) AnySatisfy(predicate func(int16, int8) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt16Int8HashMap) AllSatisfy(predicate func(int16, int8) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt16Int8HashMap) NoneSatisfy(predicate func(int16, int8) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt16Int8HashMap) KeysToSlice() []int16 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt16Int8HashMap) ValuesToSlice() []int8 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt16Int8HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt16Int8HashMap) Equals(other *ImmutableInt16Int8HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt16Int8HashMap) ToMutable() *Int16Int8HashMap {
	copy := NewInt16Int8HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int16, v int8) {
		copy.Put(k, v)
	})
	return copy
}
