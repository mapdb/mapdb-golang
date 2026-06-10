
package multimap

import (
	"fmt"
	"math"
	"strings"
)

// Float64Float64ListMultimap is a list multimap from float64 keys to float64 values.
// Each key maps to a slice of values, preserving insertion order per key.
type Float64Float64ListMultimap struct {
	data map[uint64][]float64
	keys map[uint64]float64
	size int
}

// NewFloat64Float64ListMultimap creates a new empty Float64Float64ListMultimap.
func NewFloat64Float64ListMultimap() *Float64Float64ListMultimap {
	return &Float64Float64ListMultimap{
		data: make(map[uint64][]float64),
		keys: make(map[uint64]float64),
		size: 0,
	}
}

// Put adds a value to the list for the given key.
func (m *Float64Float64ListMultimap) Put(key float64, value float64) {
	kb := math.Float64bits(key)
	m.data[kb] = append(m.data[kb], value)
	m.keys[kb] = key
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Float64Float64ListMultimap) Get(key float64) []float64 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Float64Float64ListMultimap) GetAll(key float64) []float64 {
	vals := m.data[math.Float64bits(key)]
	if vals == nil {
		return nil
	}
	result := make([]float64, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Float64Float64ListMultimap) RemoveAll(key float64) []float64 {
	kb := math.Float64bits(key)
	vals, ok := m.data[kb]
	if !ok {
		return nil
	}
	delete(m.data, kb)
	delete(m.keys, kb)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *Float64Float64ListMultimap) ContainsKey(key float64) bool {
	_, ok := m.data[math.Float64bits(key)]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Float64Float64ListMultimap) ContainsKeyValue(key float64, value float64) bool {
	vals, ok := m.data[math.Float64bits(key)]
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
func (m *Float64Float64ListMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Float64Float64ListMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Float64Float64ListMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Float64Float64ListMultimap) Clear() {
	m.data = make(map[uint64][]float64)
	m.keys = make(map[uint64]float64)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Float64Float64ListMultimap) ForEach(f func(float64, float64)) {
	for kb, vals := range m.data {
		key := m.keys[kb]
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Float64Float64ListMultimap) ForEachKeyValues(f func(float64, []float64)) {
	for kb, vals := range m.data {
		key := m.keys[kb]
		copied := make([]float64, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Float64Float64ListMultimap) Keys() []float64 {
	result := make([]float64, 0, len(m.data))
	for _, key := range m.keys {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Float64Float64ListMultimap) Values() []float64 {
	result := make([]float64, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Float64Float64ListMultimap) Select(predicate func(float64, float64) bool) *Float64Float64ListMultimap {
	result := NewFloat64Float64ListMultimap()
	for kb, vals := range m.data {
		key := m.keys[kb]
		for _, val := range vals {
			if predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// Reject returns a new multimap containing only key-value pairs that do not satisfy the predicate.
func (m *Float64Float64ListMultimap) Reject(predicate func(float64, float64) bool) *Float64Float64ListMultimap {
	result := NewFloat64Float64ListMultimap()
	for kb, vals := range m.data {
		key := m.keys[kb]
		for _, val := range vals {
			if !predicate(key, val) {
				result.Put(key, val)
			}
		}
	}
	return result
}

// String returns a string representation of the multimap.
func (m *Float64Float64ListMultimap) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for kb, vals := range m.data {
		key := m.keys[kb]
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
func (m *Float64Float64ListMultimap) Equals(other *Float64Float64ListMultimap) bool {
	if m.size != other.size {
		return false
	}
	if len(m.data) != len(other.data) {
		return false
	}
	for kb, vals := range m.data {
		otherVals, ok := other.data[kb]
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
func (m *Float64Float64ListMultimap) KeysToSlice() []float64 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Float64Float64ListMultimap) ValuesToSlice() []float64 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Float64Float64ListMultimap) WithKeyValue(key float64, value float64) *Float64Float64ListMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Float64Float64ListMultimap) WithoutKey(key float64) *Float64Float64ListMultimap {
	m.RemoveAll(key)
	return m
}
