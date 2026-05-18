
package multimap

import (
	"fmt"
	"strings"
)

// Float64CharListMultimap is a list multimap from float64 keys to uint16 values.
// Each key maps to a slice of values, preserving insertion order per key.
type Float64CharListMultimap struct {
	data map[float64][]uint16
	size int
}

// NewFloat64CharListMultimap creates a new empty Float64CharListMultimap.
func NewFloat64CharListMultimap() *Float64CharListMultimap {
	return &Float64CharListMultimap{
		data: make(map[float64][]uint16),
		size: 0,
	}
}

// Put adds a value to the list for the given key.
func (m *Float64CharListMultimap) Put(key float64, value uint16) {
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Float64CharListMultimap) Get(key float64) []uint16 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Float64CharListMultimap) GetAll(key float64) []uint16 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]uint16, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Float64CharListMultimap) RemoveAll(key float64) []uint16 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *Float64CharListMultimap) ContainsKey(key float64) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Float64CharListMultimap) ContainsKeyValue(key float64, value uint16) bool {
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
func (m *Float64CharListMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Float64CharListMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Float64CharListMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Float64CharListMultimap) Clear() {
	m.data = make(map[float64][]uint16)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Float64CharListMultimap) ForEach(f func(float64, uint16)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Float64CharListMultimap) ForEachKeyValues(f func(float64, []uint16)) {
	for key, vals := range m.data {
		copied := make([]uint16, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Float64CharListMultimap) Keys() []float64 {
	result := make([]float64, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Float64CharListMultimap) Values() []uint16 {
	result := make([]uint16, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Float64CharListMultimap) Select(predicate func(float64, uint16) bool) *Float64CharListMultimap {
	result := NewFloat64CharListMultimap()
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
func (m *Float64CharListMultimap) Reject(predicate func(float64, uint16) bool) *Float64CharListMultimap {
	result := NewFloat64CharListMultimap()
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
func (m *Float64CharListMultimap) String() string {
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
func (m *Float64CharListMultimap) Equals(other *Float64CharListMultimap) bool {
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
func (m *Float64CharListMultimap) KeysToSlice() []float64 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Float64CharListMultimap) ValuesToSlice() []uint16 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Float64CharListMultimap) WithKeyValue(key float64, value uint16) *Float64CharListMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Float64CharListMultimap) WithoutKey(key float64) *Float64CharListMultimap {
	m.RemoveAll(key)
	return m
}
