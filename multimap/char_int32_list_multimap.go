
package multimap

import (
	"fmt"
	"strings"
)

// CharInt32ListMultimap is a list multimap from uint16 keys to int32 values.
// Each key maps to a slice of values, preserving insertion order per key.
type CharInt32ListMultimap struct {
	data map[uint16][]int32
	size int
}

// NewCharInt32ListMultimap creates a new empty CharInt32ListMultimap.
func NewCharInt32ListMultimap() *CharInt32ListMultimap {
	return &CharInt32ListMultimap{
		data: make(map[uint16][]int32),
		size: 0,
	}
}

// Put adds a value to the list for the given key.
func (m *CharInt32ListMultimap) Put(key uint16, value int32) {
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *CharInt32ListMultimap) Get(key uint16) []int32 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *CharInt32ListMultimap) GetAll(key uint16) []int32 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]int32, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *CharInt32ListMultimap) RemoveAll(key uint16) []int32 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *CharInt32ListMultimap) ContainsKey(key uint16) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *CharInt32ListMultimap) ContainsKeyValue(key uint16, value int32) bool {
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
func (m *CharInt32ListMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *CharInt32ListMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *CharInt32ListMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *CharInt32ListMultimap) Clear() {
	m.data = make(map[uint16][]int32)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *CharInt32ListMultimap) ForEach(f func(uint16, int32)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *CharInt32ListMultimap) ForEachKeyValues(f func(uint16, []int32)) {
	for key, vals := range m.data {
		copied := make([]int32, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *CharInt32ListMultimap) Keys() []uint16 {
	result := make([]uint16, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *CharInt32ListMultimap) Values() []int32 {
	result := make([]int32, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *CharInt32ListMultimap) Select(predicate func(uint16, int32) bool) *CharInt32ListMultimap {
	result := NewCharInt32ListMultimap()
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
func (m *CharInt32ListMultimap) Reject(predicate func(uint16, int32) bool) *CharInt32ListMultimap {
	result := NewCharInt32ListMultimap()
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
func (m *CharInt32ListMultimap) String() string {
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
func (m *CharInt32ListMultimap) Equals(other *CharInt32ListMultimap) bool {
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
func (m *CharInt32ListMultimap) KeysToSlice() []uint16 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *CharInt32ListMultimap) ValuesToSlice() []int32 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *CharInt32ListMultimap) WithKeyValue(key uint16, value int32) *CharInt32ListMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *CharInt32ListMultimap) WithoutKey(key uint16) *CharInt32ListMultimap {
	m.RemoveAll(key)
	return m
}
