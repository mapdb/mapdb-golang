
package multimap

import (
	"fmt"
	"math"
	"strings"
)

// Int16Float64SetMultimap is a set multimap from int16 keys to float64 values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type Int16Float64SetMultimap struct {
	data map[int16][]float64
	size int
}

// NewInt16Float64SetMultimap creates a new empty Int16Float64SetMultimap.
func NewInt16Float64SetMultimap() *Int16Float64SetMultimap {
	return &Int16Float64SetMultimap{
		data: make(map[int16][]float64),
		size: 0,
	}
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *Int16Float64SetMultimap) Put(key int16, value float64) {
	for _, existing := range m.data[key] {
		if math.Float64bits(existing) == math.Float64bits(value) {
			return
		}
	}
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Int16Float64SetMultimap) Get(key int16) []float64 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Int16Float64SetMultimap) GetAll(key int16) []float64 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]float64, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Int16Float64SetMultimap) RemoveAll(key int16) []float64 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *Int16Float64SetMultimap) ContainsKey(key int16) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Int16Float64SetMultimap) ContainsKeyValue(key int16, value float64) bool {
	vals, ok := m.data[key]
	if !ok {
		return false
	}
	for _, v := range vals {
		if math.Float64bits(v) == math.Float64bits(value) {
			return true
		}
	}
	return false
}

// KeysCount returns the number of distinct keys.
func (m *Int16Float64SetMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Int16Float64SetMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Int16Float64SetMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Int16Float64SetMultimap) Clear() {
	m.data = make(map[int16][]float64)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Int16Float64SetMultimap) ForEach(f func(int16, float64)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Int16Float64SetMultimap) ForEachKeyValues(f func(int16, []float64)) {
	for key, vals := range m.data {
		copied := make([]float64, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Int16Float64SetMultimap) Keys() []int16 {
	result := make([]int16, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Int16Float64SetMultimap) Values() []float64 {
	result := make([]float64, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Int16Float64SetMultimap) Select(predicate func(int16, float64) bool) *Int16Float64SetMultimap {
	result := NewInt16Float64SetMultimap()
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
func (m *Int16Float64SetMultimap) Reject(predicate func(int16, float64) bool) *Int16Float64SetMultimap {
	result := NewInt16Float64SetMultimap()
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
func (m *Int16Float64SetMultimap) String() string {
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
func (m *Int16Float64SetMultimap) Equals(other *Int16Float64SetMultimap) bool {
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
			if !(math.Float64bits(val) == math.Float64bits(otherVals[i])) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all distinct keys as a slice.
func (m *Int16Float64SetMultimap) KeysToSlice() []int16 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Int16Float64SetMultimap) ValuesToSlice() []float64 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Int16Float64SetMultimap) WithKeyValue(key int16, value float64) *Int16Float64SetMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Int16Float64SetMultimap) WithoutKey(key int16) *Int16Float64SetMultimap {
	m.RemoveAll(key)
	return m
}
