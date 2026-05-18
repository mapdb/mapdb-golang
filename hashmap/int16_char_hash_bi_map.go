
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

// Int16CharHashBiMap is a bidirectional map with int16 keys and uint16 values.
// Both key-to-value and value-to-key lookups are O(1).
type Int16CharHashBiMap struct {
	forward *Int16CharHashMap
	reverse *CharInt16HashMap
}

// NewInt16CharHashBiMap creates a new empty Int16CharHashBiMap with default capacity.
func NewInt16CharHashBiMap() *Int16CharHashBiMap {
	return &Int16CharHashBiMap{
		forward: NewInt16CharHashMap(),
		reverse: NewCharInt16HashMap(),
	}
}

// NewInt16CharHashBiMapWithCapacity creates a new empty Int16CharHashBiMap with the given initial capacity.
func NewInt16CharHashBiMapWithCapacity(capacity int) *Int16CharHashBiMap {
	return &Int16CharHashBiMap{
		forward: NewInt16CharHashMapWithCapacity(capacity),
		reverse: NewCharInt16HashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *Int16CharHashBiMap) Put(key int16, value uint16) (uint16, bool) {
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
func (m *Int16CharHashBiMap) Get(key int16) (uint16, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *Int16CharHashBiMap) GetKey(value uint16) (int16, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *Int16CharHashBiMap) Remove(key int16) (uint16, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *Int16CharHashBiMap) RemoveValue(value uint16) (int16, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *Int16CharHashBiMap) ContainsKey(key int16) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Int16CharHashBiMap) ContainsValue(value uint16) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *Int16CharHashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Int16CharHashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *Int16CharHashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *Int16CharHashBiMap) ForEach(f func(int16, uint16)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *Int16CharHashBiMap) Keys() iter.Seq[int16] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Int16CharHashBiMap) Values() iter.Seq[uint16] {
	return m.forward.Values()
}

// Inverse returns a new CharInt16HashBiMap with keys and values swapped.
func (m *Int16CharHashBiMap) Inverse() *CharInt16HashBiMap {
	result := NewCharInt16HashBiMap()
	m.forward.ForEach(func(k int16, v uint16) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *Int16CharHashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k int16, v uint16) {
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
func (m *Int16CharHashBiMap) Equals(other *Int16CharHashBiMap) bool {
	return m.forward.Equals(other.forward)
}
