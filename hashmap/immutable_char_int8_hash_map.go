
package hashmap

import (
	"iter"
)

// ImmutableCharInt8HashMap is an immutable view of a CharInt8HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableCharInt8HashMap struct {
	delegate *CharInt8HashMap
}

// NewImmutableCharInt8HashMap creates an immutable map from key-value pairs.
func NewImmutableCharInt8HashMap(pairs ...struct {
	Key   uint16
	Value int8
}) *ImmutableCharInt8HashMap {
	m := NewCharInt8HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableCharInt8HashMap{delegate: m}
}

// ImmutableCharInt8HashMapFrom creates an immutable copy of a mutable map.
func ImmutableCharInt8HashMapFrom(m *CharInt8HashMap) *ImmutableCharInt8HashMap {
	copy := NewCharInt8HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k uint16, v int8) {
		copy.Put(k, v)
	})
	return &ImmutableCharInt8HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableCharInt8HashMap) Get(key uint16) (int8, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableCharInt8HashMap) GetOrDefault(key uint16, defaultValue int8) int8 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableCharInt8HashMap) ContainsKey(key uint16) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableCharInt8HashMap) ContainsValue(value int8) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableCharInt8HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableCharInt8HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableCharInt8HashMap) All() iter.Seq2[uint16, int8] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableCharInt8HashMap) Keys() iter.Seq[uint16] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableCharInt8HashMap) Values() iter.Seq[int8] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableCharInt8HashMap) ForEach(f func(uint16, int8)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableCharInt8HashMap) Select(predicate func(uint16, int8) bool) *ImmutableCharInt8HashMap {
	return &ImmutableCharInt8HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableCharInt8HashMap) Reject(predicate func(uint16, int8) bool) *ImmutableCharInt8HashMap {
	return &ImmutableCharInt8HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableCharInt8HashMap) AnySatisfy(predicate func(uint16, int8) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableCharInt8HashMap) AllSatisfy(predicate func(uint16, int8) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableCharInt8HashMap) NoneSatisfy(predicate func(uint16, int8) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableCharInt8HashMap) KeysToSlice() []uint16 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableCharInt8HashMap) ValuesToSlice() []int8 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableCharInt8HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableCharInt8HashMap) Equals(other *ImmutableCharInt8HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableCharInt8HashMap) ToMutable() *CharInt8HashMap {
	copy := NewCharInt8HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k uint16, v int8) {
		copy.Put(k, v)
	})
	return copy
}
