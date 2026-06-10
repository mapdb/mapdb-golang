
package multimap

import (
	"math"
	"fmt"
	"strings"
)

// Float32CharSetMultimap is a set multimap from float32 keys to uint16 values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type Float32CharSetMultimap struct {
	data map[uint32][]uint16
	keys map[uint32]float32
	size int
}

// NewFloat32CharSetMultimap creates a new empty Float32CharSetMultimap.
func NewFloat32CharSetMultimap() *Float32CharSetMultimap {
	return &Float32CharSetMultimap{
		data: make(map[uint32][]uint16),
		keys: make(map[uint32]float32),
		size: 0,
	}
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *Float32CharSetMultimap) Put(key float32, value uint16) {
	kb := math.Float32bits(key)
	for _, existing := range m.data[kb] {
		if existing == value {
			return
		}
	}
	m.data[kb] = append(m.data[kb], value)
	m.keys[kb] = key
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *Float32CharSetMultimap) Get(key float32) []uint16 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *Float32CharSetMultimap) GetAll(key float32) []uint16 {
	vals := m.data[math.Float32bits(key)]
	if vals == nil {
		return nil
	}
	result := make([]uint16, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *Float32CharSetMultimap) RemoveAll(key float32) []uint16 {
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
func (m *Float32CharSetMultimap) ContainsKey(key float32) bool {
	_, ok := m.data[math.Float32bits(key)]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *Float32CharSetMultimap) ContainsKeyValue(key float32, value uint16) bool {
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
func (m *Float32CharSetMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *Float32CharSetMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *Float32CharSetMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *Float32CharSetMultimap) Clear() {
	m.data = make(map[uint32][]uint16)
	m.keys = make(map[uint32]float32)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *Float32CharSetMultimap) ForEach(f func(float32, uint16)) {
	for kb, vals := range m.data {
		key := m.keys[kb]
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *Float32CharSetMultimap) ForEachKeyValues(f func(float32, []uint16)) {
	for kb, vals := range m.data {
		key := m.keys[kb]
		copied := make([]uint16, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *Float32CharSetMultimap) Keys() []float32 {
	result := make([]float32, 0, len(m.data))
	for _, key := range m.keys {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *Float32CharSetMultimap) Values() []uint16 {
	result := make([]uint16, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *Float32CharSetMultimap) Select(predicate func(float32, uint16) bool) *Float32CharSetMultimap {
	result := NewFloat32CharSetMultimap()
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
func (m *Float32CharSetMultimap) Reject(predicate func(float32, uint16) bool) *Float32CharSetMultimap {
	result := NewFloat32CharSetMultimap()
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
func (m *Float32CharSetMultimap) String() string {
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
func (m *Float32CharSetMultimap) Equals(other *Float32CharSetMultimap) bool {
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
func (m *Float32CharSetMultimap) KeysToSlice() []float32 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *Float32CharSetMultimap) ValuesToSlice() []uint16 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *Float32CharSetMultimap) WithKeyValue(key float32, value uint16) *Float32CharSetMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *Float32CharSetMultimap) WithoutKey(key float32) *Float32CharSetMultimap {
	m.RemoveAll(key)
	return m
}
