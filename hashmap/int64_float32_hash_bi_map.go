
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

// Int64Float32HashBiMap is a bidirectional map with int64 keys and float32 values.
// Both key-to-value and value-to-key lookups are O(1).
type Int64Float32HashBiMap struct {
	forward *Int64Float32HashMap
	reverse *Float32Int64HashMap
}

// NewInt64Float32HashBiMap creates a new empty Int64Float32HashBiMap with default capacity.
func NewInt64Float32HashBiMap() *Int64Float32HashBiMap {
	return &Int64Float32HashBiMap{
		forward: NewInt64Float32HashMap(),
		reverse: NewFloat32Int64HashMap(),
	}
}

// NewInt64Float32HashBiMapWithCapacity creates a new empty Int64Float32HashBiMap with the given initial capacity.
func NewInt64Float32HashBiMapWithCapacity(capacity int) *Int64Float32HashBiMap {
	return &Int64Float32HashBiMap{
		forward: NewInt64Float32HashMapWithCapacity(capacity),
		reverse: NewFloat32Int64HashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *Int64Float32HashBiMap) Put(key int64, value float32) (float32, bool) {
	// If this value is already mapped to a different key, remove that old key->value pair
	if oldKey, ok := m.reverse.Get(value); ok {
		if !(oldKey == key) {
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
func (m *Int64Float32HashBiMap) Get(key int64) (float32, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *Int64Float32HashBiMap) GetKey(value float32) (int64, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *Int64Float32HashBiMap) Remove(key int64) (float32, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *Int64Float32HashBiMap) RemoveValue(value float32) (int64, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *Int64Float32HashBiMap) ContainsKey(key int64) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Int64Float32HashBiMap) ContainsValue(value float32) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *Int64Float32HashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Int64Float32HashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *Int64Float32HashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *Int64Float32HashBiMap) ForEach(f func(int64, float32)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *Int64Float32HashBiMap) Keys() iter.Seq[int64] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Int64Float32HashBiMap) Values() iter.Seq[float32] {
	return m.forward.Values()
}

// Inverse returns a new Float32Int64HashBiMap with keys and values swapped.
func (m *Int64Float32HashBiMap) Inverse() *Float32Int64HashBiMap {
	result := NewFloat32Int64HashBiMap()
	m.forward.ForEach(func(k int64, v float32) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *Int64Float32HashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k int64, v float32) {
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
func (m *Int64Float32HashBiMap) Equals(other *Int64Float32HashBiMap) bool {
	return m.forward.Equals(other.forward)
}
