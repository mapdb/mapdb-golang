
package hashmap

import (
	"iter"
)

// ImmutableCharInt32HashMap is an immutable view of a CharInt32HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableCharInt32HashMap struct {
	delegate *CharInt32HashMap
}

// NewImmutableCharInt32HashMap creates an immutable map from key-value pairs.
func NewImmutableCharInt32HashMap(pairs ...struct {
	Key   uint16
	Value int32
}) *ImmutableCharInt32HashMap {
	m := NewCharInt32HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableCharInt32HashMap{delegate: m}
}

// ImmutableCharInt32HashMapFrom creates an immutable copy of a mutable map.
func ImmutableCharInt32HashMapFrom(m *CharInt32HashMap) *ImmutableCharInt32HashMap {
	copy := NewCharInt32HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k uint16, v int32) {
		copy.Put(k, v)
	})
	return &ImmutableCharInt32HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableCharInt32HashMap) Get(key uint16) (int32, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableCharInt32HashMap) GetOrDefault(key uint16, defaultValue int32) int32 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableCharInt32HashMap) ContainsKey(key uint16) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableCharInt32HashMap) ContainsValue(value int32) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableCharInt32HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableCharInt32HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableCharInt32HashMap) All() iter.Seq2[uint16, int32] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableCharInt32HashMap) Keys() iter.Seq[uint16] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableCharInt32HashMap) Values() iter.Seq[int32] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableCharInt32HashMap) ForEach(f func(uint16, int32)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableCharInt32HashMap) Select(predicate func(uint16, int32) bool) *ImmutableCharInt32HashMap {
	return &ImmutableCharInt32HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableCharInt32HashMap) Reject(predicate func(uint16, int32) bool) *ImmutableCharInt32HashMap {
	return &ImmutableCharInt32HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableCharInt32HashMap) AnySatisfy(predicate func(uint16, int32) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableCharInt32HashMap) AllSatisfy(predicate func(uint16, int32) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableCharInt32HashMap) NoneSatisfy(predicate func(uint16, int32) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableCharInt32HashMap) KeysToSlice() []uint16 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableCharInt32HashMap) ValuesToSlice() []int32 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableCharInt32HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableCharInt32HashMap) Equals(other *ImmutableCharInt32HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableCharInt32HashMap) ToMutable() *CharInt32HashMap {
	copy := NewCharInt32HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k uint16, v int32) {
		copy.Put(k, v)
	})
	return copy
}
