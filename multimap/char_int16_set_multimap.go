
package multimap

import (
	"fmt"
	"strings"
)

// CharInt16SetMultimap is a set multimap from uint16 keys to int16 values.
// Each key maps to a set of unique values (duplicates on Put are silently dropped).
type CharInt16SetMultimap struct {
	data map[uint16][]int16
	size int
}

// NewCharInt16SetMultimap creates a new empty CharInt16SetMultimap.
func NewCharInt16SetMultimap() *CharInt16SetMultimap {
	return &CharInt16SetMultimap{
		data: make(map[uint16][]int16),
		size: 0,
	}
}

// Put adds a value to the set for the given key. Idempotent: a duplicate
// value for the same key is silently dropped.
func (m *CharInt16SetMultimap) Put(key uint16, value int16) {
	for _, existing := range m.data[key] {
		if existing == value {
			return
		}
	}
	m.data[key] = append(m.data[key], value)
	m.size++
}

// Get returns a copy of the values for the given key. Returns nil if the key is absent.
func (m *CharInt16SetMultimap) Get(key uint16) []int16 {
	return m.GetAll(key)
}

// GetAll returns a copy of the values for the given key.
func (m *CharInt16SetMultimap) GetAll(key uint16) []int16 {
	vals := m.data[key]
	if vals == nil {
		return nil
	}
	result := make([]int16, len(vals))
	copy(result, vals)
	return result
}

// RemoveAll removes all values for the given key and returns them.
func (m *CharInt16SetMultimap) RemoveAll(key uint16) []int16 {
	vals, ok := m.data[key]
	if !ok {
		return nil
	}
	delete(m.data, key)
	m.size -= len(vals)
	return vals
}

// ContainsKey returns true if the multimap contains the given key.
func (m *CharInt16SetMultimap) ContainsKey(key uint16) bool {
	_, ok := m.data[key]
	return ok
}

// ContainsKeyValue returns true if the multimap contains the given key-value pair.
func (m *CharInt16SetMultimap) ContainsKeyValue(key uint16, value int16) bool {
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
func (m *CharInt16SetMultimap) KeysCount() int {
	return len(m.data)
}

// Size returns the total number of values across all keys.
func (m *CharInt16SetMultimap) Size() int {
	return m.size
}

// IsEmpty returns true if the multimap contains no values.
func (m *CharInt16SetMultimap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the multimap.
func (m *CharInt16SetMultimap) Clear() {
	m.data = make(map[uint16][]int16)
	m.size = 0
}

// ForEach calls the given function for each key-value pair.
func (m *CharInt16SetMultimap) ForEach(f func(uint16, int16)) {
	for key, vals := range m.data {
		for _, val := range vals {
			f(key, val)
		}
	}
}

// ForEachKeyValues calls the given function for each key with a copy of its values.
func (m *CharInt16SetMultimap) ForEachKeyValues(f func(uint16, []int16)) {
	for key, vals := range m.data {
		copied := make([]int16, len(vals))
		copy(copied, vals)
		f(key, copied)
	}
}

// Keys returns a slice of all distinct keys.
func (m *CharInt16SetMultimap) Keys() []uint16 {
	result := make([]uint16, 0, len(m.data))
	for key := range m.data {
		result = append(result, key)
	}
	return result
}

// Values returns a slice of all values across all keys.
func (m *CharInt16SetMultimap) Values() []int16 {
	result := make([]int16, 0, m.size)
	for _, vals := range m.data {
		result = append(result, vals...)
	}
	return result
}

// Select returns a new multimap containing only key-value pairs that satisfy the predicate.
func (m *CharInt16SetMultimap) Select(predicate func(uint16, int16) bool) *CharInt16SetMultimap {
	result := NewCharInt16SetMultimap()
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
func (m *CharInt16SetMultimap) Reject(predicate func(uint16, int16) bool) *CharInt16SetMultimap {
	result := NewCharInt16SetMultimap()
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
func (m *CharInt16SetMultimap) String() string {
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
func (m *CharInt16SetMultimap) Equals(other *CharInt16SetMultimap) bool {
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
func (m *CharInt16SetMultimap) KeysToSlice() []uint16 {
	return m.Keys()
}

// ValuesToSlice returns all values as a slice.
func (m *CharInt16SetMultimap) ValuesToSlice() []int16 {
	return m.Values()
}

// WithKeyValue adds a key-value pair and returns the multimap (fluent API).
func (m *CharInt16SetMultimap) WithKeyValue(key uint16, value int16) *CharInt16SetMultimap {
	m.Put(key, value)
	return m
}

// WithoutKey removes all values for the key and returns the multimap (fluent API).
func (m *CharInt16SetMultimap) WithoutKey(key uint16) *CharInt16SetMultimap {
	m.RemoveAll(key)
	return m
}
