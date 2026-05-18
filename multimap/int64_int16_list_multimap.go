
package multimap

import (
	"fmt"
	"strings"
)

// Int64Int16ListMultimap is a list multimap from int64 keys to int16 values.
// Each key maps to a slice of values, preserving insertion order per key.
type Int64Int16ListMultimap struct {
	data map[int64][]int16
	size int
}

// NewInt64Int16ListMultimap creates a new empty Int64Int16ListMultimap.
func NewInt64Int16ListMultimap() *Int64Int16ListMultimap {
	return &Int64Int16ListMultimap{
		data: make(map[int64][]int16),
		size: 0,
	}
}

// Put adds a value to the list for the given key.
func (m *Int64Int16ListMultimap) Put(key int64, value int16) {
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Int64Int16ListMultimap) Get(key int64) []int16 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Int64Int16ListMultimap) GetAll(key int64) []int16 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]int16, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Int64Int16ListMultimap) RemoveAll(key int64) []int16 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *Int64Int16ListMultimap) ContainsKey(key int64) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Int64Int16ListMultimap) ContainsKeyValue(key int64, value int16) bool {
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
func (m *Int64Int16ListMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Int64Int16ListMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Int64Int16ListMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Int64Int16ListMultimap) Clear() {
	m.data = make(map[int64][]int16)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Int64Int16ListMultimap) ForEach(f func(int64, int16)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Int64Int16ListMultimap) ForEachKeyValues(f func(int64, []int16)) {
	for key, vals := range m.data {
		copied := make([]int16, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Int64Int16ListMultimap) Keys() []int64 {
	result := make([]int64, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Int64Int16ListMultimap) Values() []int16 {
	result := make([]int16, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Int64Int16ListMultimap) Select(predicate func(int64, int16) bool) *Int64Int16ListMultimap {
	result := NewInt64Int16ListMultimap()
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
func (m *Int64Int16ListMultimap) Reject(predicate func(int64, int16) bool) *Int64Int16ListMultimap {
	result := NewInt64Int16ListMultimap()
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
func (m *Int64Int16ListMultimap) String() string {
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
func (m *Int64Int16ListMultimap) Equals(other *Int64Int16ListMultimap) bool {
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
func (m *Int64Int16ListMultimap) KeysToSlice() []int64 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Int64Int16ListMultimap) ValuesToSlice() []int16 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Int64Int16ListMultimap) WithKeyValue(key int64, value int16) *Int64Int16ListMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Int64Int16ListMultimap) WithoutKey(key int64) *Int64Int16ListMultimap {
	m.RemoveAll(key)
	return m
}
