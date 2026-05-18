
package hashmap

import (
	"iter"
)

// ImmutableInt8Int16HashMap is an immutable view of a Int8Int16HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt8Int16HashMap struct {
	delegate *Int8Int16HashMap
}

// NewImmutableInt8Int16HashMap creates an immutable map from key-value pairs.
func NewImmutableInt8Int16HashMap(pairs ...struct {
	Key   int8
	Value int16
}) *ImmutableInt8Int16HashMap {
	m := NewInt8Int16HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt8Int16HashMap{delegate: m}
}

// ImmutableInt8Int16HashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt8Int16HashMapFrom(m *Int8Int16HashMap) *ImmutableInt8Int16HashMap {
	copy := NewInt8Int16HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int8, v int16) {
		copy.Put(k, v)
	})
	return &ImmutableInt8Int16HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt8Int16HashMap) Get(key int8) (int16, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt8Int16HashMap) GetOrDefault(key int8, defaultValue int16) int16 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt8Int16HashMap) ContainsKey(key int8) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt8Int16HashMap) ContainsValue(value int16) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt8Int16HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt8Int16HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt8Int16HashMap) All() iter.Seq2[int8, int16] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt8Int16HashMap) Keys() iter.Seq[int8] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt8Int16HashMap) Values() iter.Seq[int16] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt8Int16HashMap) ForEach(f func(int8, int16)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt8Int16HashMap) Select(predicate func(int8, int16) bool) *ImmutableInt8Int16HashMap {
	return &ImmutableInt8Int16HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt8Int16HashMap) Reject(predicate func(int8, int16) bool) *ImmutableInt8Int16HashMap {
	return &ImmutableInt8Int16HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt8Int16HashMap) AnySatisfy(predicate func(int8, int16) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt8Int16HashMap) AllSatisfy(predicate func(int8, int16) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt8Int16HashMap) NoneSatisfy(predicate func(int8, int16) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt8Int16HashMap) KeysToSlice() []int8 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt8Int16HashMap) ValuesToSlice() []int16 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt8Int16HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt8Int16HashMap) Equals(other *ImmutableInt8Int16HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt8Int16HashMap) ToMutable() *Int8Int16HashMap {
	copy := NewInt8Int16HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int8, v int16) {
		copy.Put(k, v)
	})
	return copy
}
