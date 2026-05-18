
package multimap

import (
	"fmt"
	"strings"
)

// Int8CharSetMultimap is a set multimap from int8 keys to uint16 values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type Int8CharSetMultimap struct {
	data map[int8][]uint16
	size int
}

// NewInt8CharSetMultimap creates a new empty Int8CharSetMultimap.
func NewInt8CharSetMultimap() *Int8CharSetMultimap {
	return &Int8CharSetMultimap{
		data: make(map[int8][]uint16),
		size: 0,
	}
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *Int8CharSetMultimap) Put(key int8, value uint16) {
	for _, existing := range m.data[key] {
		if existing == value {
			return
		}
	}
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Int8CharSetMultimap) Get(key int8) []uint16 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Int8CharSetMultimap) GetAll(key int8) []uint16 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]uint16, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Int8CharSetMultimap) RemoveAll(key int8) []uint16 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *Int8CharSetMultimap) ContainsKey(key int8) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Int8CharSetMultimap) ContainsKeyValue(key int8, value uint16) bool {
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
func (m *Int8CharSetMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Int8CharSetMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Int8CharSetMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Int8CharSetMultimap) Clear() {
	m.data = make(map[int8][]uint16)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Int8CharSetMultimap) ForEach(f func(int8, uint16)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Int8CharSetMultimap) ForEachKeyValues(f func(int8, []uint16)) {
	for key, vals := range m.data {
		copied := make([]uint16, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Int8CharSetMultimap) Keys() []int8 {
	result := make([]int8, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Int8CharSetMultimap) Values() []uint16 {
	result := make([]uint16, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Int8CharSetMultimap) Select(predicate func(int8, uint16) bool) *Int8CharSetMultimap {
	result := NewInt8CharSetMultimap()
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
func (m *Int8CharSetMultimap) Reject(predicate func(int8, uint16) bool) *Int8CharSetMultimap {
	result := NewInt8CharSetMultimap()
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
func (m *Int8CharSetMultimap) String() string {
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
func (m *Int8CharSetMultimap) Equals(other *Int8CharSetMultimap) bool {
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
func (m *Int8CharSetMultimap) KeysToSlice() []int8 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Int8CharSetMultimap) ValuesToSlice() []uint16 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Int8CharSetMultimap) WithKeyValue(key int8, value uint16) *Int8CharSetMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Int8CharSetMultimap) WithoutKey(key int8) *Int8CharSetMultimap {
	m.RemoveAll(key)
	return m
}
