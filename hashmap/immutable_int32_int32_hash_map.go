
package hashmap

import (
	"iter"
)

// ImmutableInt32Int32HashMap is an immutable view of a Int32Int32HashMap.
// It exposes only read operations. Any attempt to modify requires
// creating a mutable copy first via ToMutable().
type ImmutableInt32Int32HashMap struct {
	delegate *Int32Int32HashMap
}

// NewImmutableInt32Int32HashMap creates an immutable map from key-value pairs.
func NewImmutableInt32Int32HashMap(pairs ...struct {
	Key   int32
	Value int32
}) *ImmutableInt32Int32HashMap {
	m := NewInt32Int32HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return &ImmutableInt32Int32HashMap{delegate: m}
}

// ImmutableInt32Int32HashMapFrom creates an immutable copy of a mutable map.
func ImmutableInt32Int32HashMapFrom(m *Int32Int32HashMap) *ImmutableInt32Int32HashMap {
	copy := NewInt32Int32HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int32, v int32) {
		copy.Put(k, v)
	})
	return &ImmutableInt32Int32HashMap{delegate: copy}
}

// Get returns the value for the given key and true if found.
func (m *ImmutableInt32Int32HashMap) Get(key int32) (int32, bool) {
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *ImmutableInt32Int32HashMap) GetOrDefault(key int32, defaultValue int32) int32 {
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *ImmutableInt32Int32HashMap) ContainsKey(key int32) bool {
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *ImmutableInt32Int32HashMap) ContainsValue(value int32) bool {
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *ImmutableInt32Int32HashMap) Size() int {
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *ImmutableInt32Int32HashMap) IsEmpty() bool {
	return m.delegate.IsEmpty()
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ImmutableInt32Int32HashMap) All() iter.Seq2[int32, int32] {
	return m.delegate.All()
}

// Keys returns an iter.Seq that yields all keys.
func (m *ImmutableInt32Int32HashMap) Keys() iter.Seq[int32] {
	return m.delegate.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *ImmutableInt32Int32HashMap) Values() iter.Seq[int32] {
	return m.delegate.Values()
}

// ForEach calls the given function for each key-value pair.
func (m *ImmutableInt32Int32HashMap) ForEach(f func(int32, int32)) {
	m.delegate.ForEach(f)
}

// Select returns a new immutable map with entries that satisfy the predicate.
func (m *ImmutableInt32Int32HashMap) Select(predicate func(int32, int32) bool) *ImmutableInt32Int32HashMap {
	return &ImmutableInt32Int32HashMap{delegate: m.delegate.Select(predicate)}
}

// Reject returns a new immutable map with entries that do not satisfy the predicate.
func (m *ImmutableInt32Int32HashMap) Reject(predicate func(int32, int32) bool) *ImmutableInt32Int32HashMap {
	return &ImmutableInt32Int32HashMap{delegate: m.delegate.Reject(predicate)}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *ImmutableInt32Int32HashMap) AnySatisfy(predicate func(int32, int32) bool) bool {
	return m.delegate.AnySatisfy(predicate)
}

// AllSatisfy returns true if all entries satisfy the predicate.
func (m *ImmutableInt32Int32HashMap) AllSatisfy(predicate func(int32, int32) bool) bool {
	return m.delegate.AllSatisfy(predicate)
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *ImmutableInt32Int32HashMap) NoneSatisfy(predicate func(int32, int32) bool) bool {
	return m.delegate.NoneSatisfy(predicate)
}

// KeysToSlice returns all keys as a slice.
func (m *ImmutableInt32Int32HashMap) KeysToSlice() []int32 {
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns all values as a slice.
func (m *ImmutableInt32Int32HashMap) ValuesToSlice() []int32 {
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *ImmutableInt32Int32HashMap) String() string {
	return m.delegate.String()
}

// Equals returns true if the other immutable map has the same entries.
func (m *ImmutableInt32Int32HashMap) Equals(other *ImmutableInt32Int32HashMap) bool {
	return m.delegate.Equals(other.delegate)
}

// ToMutable returns a mutable copy of this map.
func (m *ImmutableInt32Int32HashMap) ToMutable() *Int32Int32HashMap {
	copy := NewInt32Int32HashMapWithCapacity(m.Size() * 2)
	m.ForEach(func(k int32, v int32) {
		copy.Put(k, v)
	})
	return copy
}
