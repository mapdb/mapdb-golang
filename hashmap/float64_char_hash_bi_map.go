
package hashmap

import (
	"fmt"
	"iter"
	"math"
	"strings"
)

// Float64CharHashBiMap is a bidirectional map with float64 keys and uint16 values.
// Both key-to-value and value-to-key lookups are O(1).
type Float64CharHashBiMap struct {
	forward *Float64CharHashMap
	reverse *CharFloat64HashMap
}

// NewFloat64CharHashBiMap creates a new empty Float64CharHashBiMap with default capacity.
func NewFloat64CharHashBiMap() *Float64CharHashBiMap {
	return &Float64CharHashBiMap{
		forward: NewFloat64CharHashMap(),
		reverse: NewCharFloat64HashMap(),
	}
}

// NewFloat64CharHashBiMapWithCapacity creates a new empty Float64CharHashBiMap with the given initial capacity.
func NewFloat64CharHashBiMapWithCapacity(capacity int) *Float64CharHashBiMap {
	return &Float64CharHashBiMap{
		forward: NewFloat64CharHashMapWithCapacity(capacity),
		reverse: NewCharFloat64HashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *Float64CharHashBiMap) Put(key float64, value uint16) (uint16, bool) {
	// If this value is already mapped to a different key, remove that old key->value pair
	if oldKey, ok := m.reverse.Get(value); ok {
		if !(math.Float64bits(oldKey) == math.Float64bits(key)) {
			m.forward.Remove(oldKey)
		}
	}

	// If this key already has a value, remove the old value->key reverse mapping
	oldVal, existed := m.forward.Get(key)
	if existed {
		m.reverse.Remove(oldVal)
	}

	m.forward.Put(key, value)
	m.reverse.Put(value, key)
	return oldVal, existed
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *Float64CharHashBiMap) Get(key float64) (uint16, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *Float64CharHashBiMap) GetKey(value uint16) (float64, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *Float64CharHashBiMap) Remove(key float64) (uint16, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *Float64CharHashBiMap) RemoveValue(value uint16) (float64, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *Float64CharHashBiMap) ContainsKey(key float64) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Float64CharHashBiMap) ContainsValue(value uint16) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *Float64CharHashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Float64CharHashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *Float64CharHashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *Float64CharHashBiMap) ForEach(f func(float64, uint16)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *Float64CharHashBiMap) Keys() iter.Seq[float64] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Float64CharHashBiMap) Values() iter.Seq[uint16] {
	return m.forward.Values()
}

// Inverse returns a new CharFloat64HashBiMap with keys and values swapped.
func (m *Float64CharHashBiMap) Inverse() *CharFloat64HashBiMap {
	result := NewCharFloat64HashBiMap()
	m.forward.ForEach(func(k float64, v uint16) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *Float64CharHashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k float64, v uint16) {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, v)
		first = false
	})
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other bi-map has the same key-value pairs.
func (m *Float64CharHashBiMap) Equals(other *Float64CharHashBiMap) bool {
	return m.forward.Equals(other.forward)
}
