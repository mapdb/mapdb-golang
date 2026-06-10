
package multimap

import (
	"math"
	"fmt"
	"strings"
)

// Float32Int64ListMultimap is a list multimap from float32 keys to int64 values.
// Each key maps to a slice of values, preserving insertion order per key.
type Float32Int64ListMultimap struct {
	data map[uint32][]int64
	keys map[uint32]float32
	size int
}

// NewFloat32Int64ListMultimap creates a new empty Float32Int64ListMultimap.
func NewFloat32Int64ListMultimap() *Float32Int64ListMultimap {
	return &Float32Int64ListMultimap{
		data: make(map[uint32][]int64),
		keys: make(map[uint32]float32),
		size: 0,
	}
}

// Put adds a value to the list for the given key.
func (m *Float32Int64ListMultimap) Put(key float32, value int64) {
	kb := math.Float32bits(key)
	m.data[kb] = append(m.data[kb], value)
	m.keys[kb] = key
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Float32Int64ListMultimap) Get(key float32) []int64 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Float32Int64ListMultimap) GetAll(key float32) []int64 {
	vals := m.data[math.Float32bits(key)]
	if vals == nil {
		return nil
	}
	result := make([]int64, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Float32Int64ListMultimap) RemoveAll(key float32) []int64 {
	kb := math.Float32bits(key)
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
func (m *Float32Int64ListMultimap) ContainsKey(key float32) bool {
	_, ok := m.data[math.Float32bits(key)]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Float32Int64ListMultimap) ContainsKeyValue(key float32, value int64) bool {
	vals, ok := m.data[math.Float32bits(key)]
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
func (m *Float32Int64ListMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Float32Int64ListMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Float32Int64ListMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Float32Int64ListMultimap) Clear() {
	m.data = make(map[uint32][]int64)
	m.keys = make(map[uint32]float32)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Float32Int64ListMultimap) ForEach(f func(float32, int64)) {
	for kb, vals := range m.data {
		key := m.keys[kb]
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Float32Int64ListMultimap) ForEachKeyValues(f func(float32, []int64)) {
	for kb, vals := range m.data {
		key := m.keys[kb]
		copied := make([]int64, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Float32Int64ListMultimap) Keys() []float32 {
	result := make([]float32, 0, len(m.data))
	for _, key := range m.keys {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Float32Int64ListMultimap) Values() []int64 {
	result := make([]int64, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Float32Int64ListMultimap) Select(predicate func(float32, int64) bool) *Float32Int64ListMultimap {
	result := NewFloat32Int64ListMultimap()
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
func (m *Float32Int64ListMultimap) Reject(predicate func(float32, int64) bool) *Float32Int64ListMultimap {
	result := NewFloat32Int64ListMultimap()
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
func (m *Float32Int64ListMultimap) String() string {
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
func (m *Float32Int64ListMultimap) Equals(other *Float32Int64ListMultimap) bool {
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
			if !(val == otherVals[i]) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all distinct keys as a slice.
func (m *Float32Int64ListMultimap) KeysToSlice() []float32 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Float32Int64ListMultimap) ValuesToSlice() []int64 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Float32Int64ListMultimap) WithKeyValue(key float32, value int64) *Float32Int64ListMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Float32Int64ListMultimap) WithoutKey(key float32) *Float32Int64ListMultimap {
	m.RemoveAll(key)
	return m
}
