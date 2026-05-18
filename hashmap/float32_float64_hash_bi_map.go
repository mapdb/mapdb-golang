
package hashmap

import (
	"fmt"
	"iter"
	"math"
	"strings"
)

// Float32Float64HashBiMap is a bidirectional map with float32 keys and float64 values.
// Both key-to-value and value-to-key lookups are O(1).
type Float32Float64HashBiMap struct {
	forward *Float32Float64HashMap
	reverse *Float64Float32HashMap
}

// NewFloat32Float64HashBiMap creates a new empty Float32Float64HashBiMap with default capacity.
func NewFloat32Float64HashBiMap() *Float32Float64HashBiMap {
	return &Float32Float64HashBiMap{
		forward: NewFloat32Float64HashMap(),
		reverse: NewFloat64Float32HashMap(),
	}
}

// NewFloat32Float64HashBiMapWithCapacity creates a new empty Float32Float64HashBiMap with the given initial capacity.
func NewFloat32Float64HashBiMapWithCapacity(capacity int) *Float32Float64HashBiMap {
	return &Float32Float64HashBiMap{
		forward: NewFloat32Float64HashMapWithCapacity(capacity),
		reverse: NewFloat64Float32HashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *Float32Float64HashBiMap) Put(key float32, value float64) (float64, bool) {
	// If this value is already mapped to a different key, remove that old key->value pair
	if oldKey, ok := m.reverse.Get(value); ok {
		if !(math.Float32bits(oldKey) == math.Float32bits(key)) {
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
func (m *Float32Float64HashBiMap) Get(key float32) (float64, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *Float32Float64HashBiMap) GetKey(value float64) (float32, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *Float32Float64HashBiMap) Remove(key float32) (float64, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *Float32Float64HashBiMap) RemoveValue(value float64) (float32, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *Float32Float64HashBiMap) ContainsKey(key float32) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Float32Float64HashBiMap) ContainsValue(value float64) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *Float32Float64HashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Float32Float64HashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *Float32Float64HashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *Float32Float64HashBiMap) ForEach(f func(float32, float64)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *Float32Float64HashBiMap) Keys() iter.Seq[float32] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Float32Float64HashBiMap) Values() iter.Seq[float64] {
	return m.forward.Values()
}

// Inverse returns a new Float64Float32HashBiMap with keys and values swapped.
func (m *Float32Float64HashBiMap) Inverse() *Float64Float32HashBiMap {
	result := NewFloat64Float32HashBiMap()
	m.forward.ForEach(func(k float32, v float64) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *Float32Float64HashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k float32, v float64) {
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
func (m *Float32Float64HashBiMap) Equals(other *Float32Float64HashBiMap) bool {
	return m.forward.Equals(other.forward)
}
