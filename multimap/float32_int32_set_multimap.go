
package multimap

import (
	"fmt"
	"strings"
)

// Float32Int32SetMultimap is a set multimap from float32 keys to int32 values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type Float32Int32SetMultimap struct {
	data map[float32][]int32
	size int
}

// NewFloat32Int32SetMultimap creates a new empty Float32Int32SetMultimap.
func NewFloat32Int32SetMultimap() *Float32Int32SetMultimap {
	return &Float32Int32SetMultimap{
		data: make(map[float32][]int32),
		size: 0,
	}
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *Float32Int32SetMultimap) Put(key float32, value int32) {
	for _, existing := range m.data[key] {
		if existing == value {
			return
		}
	}
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Float32Int32SetMultimap) Get(key float32) []int32 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Float32Int32SetMultimap) GetAll(key float32) []int32 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]int32, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Float32Int32SetMultimap) RemoveAll(key float32) []int32 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *Float32Int32SetMultimap) ContainsKey(key float32) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Float32Int32SetMultimap) ContainsKeyValue(key float32, value int32) bool {
	vals, ok := m.data[key]
	if !ok {
		return false
	}
	for _, v := range vals {
		if v == value {
			return true
		}
	}
	return false
}

// KeysCount returns the number of distinct keys.
func (m *Float32Int32SetMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Float32Int32SetMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Float32Int32SetMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Float32Int32SetMultimap) Clear() {
	m.data = make(map[float32][]int32)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Float32Int32SetMultimap) ForEach(f func(float32, int32)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Float32Int32SetMultimap) ForEachKeyValues(f func(float32, []int32)) {
	for key, vals := range m.data {
		copied := make([]int32, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Float32Int32SetMultimap) Keys() []float32 {
	result := make([]float32, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Float32Int32SetMultimap) Values() []int32 {
	result := make([]int32, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Float32Int32SetMultimap) Select(predicate func(float32, int32) bool) *Float32Int32SetMultimap {
	result := NewFloat32Int32SetMultimap()
	for key, vals := range m.data {
		for _, val := range vals {
			if predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// Reject returns a new multimap containing only key-value pairs that do not satisfy the predicate.
func (m *Float32Int32SetMultimap) Reject(predicate func(float32, int32) bool) *Float32Int32SetMultimap {
	result := NewFloat32Int32SetMultimap()
	for key, vals := range m.data {
		for _, val := range vals {
			if !predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// String returns a string representation of the multimap.
func (m *Float32Int32SetMultimap) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for key, vals := range m.data {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v=[", key)
		for i, val := range vals {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v", val)
		}
		sb.WriteString("]")
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other multimap has the same key-value pairs in the same order per key.
func (m *Float32Int32SetMultimap) Equals(other *Float32Int32SetMultimap) bool {
	if m.size != other.size {
		return false
	}
	if len(m.data) != len(other.data) {
		return false
	}
	for key, vals := range m.data {
		otherVals, ok := other.data[key]
		if !ok || len(vals) != len(otherVals) {
			return false
		}
		for i, val := range vals {
			if !(val == otherVals[i]) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all distinct keys as a slice.
func (m *Float32Int32SetMultimap) KeysToSlice() []float32 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Float32Int32SetMultimap) ValuesToSlice() []int32 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Float32Int32SetMultimap) WithKeyValue(key float32, value int32) *Float32Int32SetMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Float32Int32SetMultimap) WithoutKey(key float32) *Float32Int32SetMultimap {
	m.RemoveAll(key)
	return m
}
