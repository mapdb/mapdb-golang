
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

// CharInt32HashBiMap is a bidirectional map with uint16 keys and int32 values.
// Both key-to-value and value-to-key lookups are O(1).
type CharInt32HashBiMap struct {
	forward *CharInt32HashMap
	reverse *Int32CharHashMap
}

// NewCharInt32HashBiMap creates a new empty CharInt32HashBiMap with default capacity.
func NewCharInt32HashBiMap() *CharInt32HashBiMap {
	return &CharInt32HashBiMap{
		forward: NewCharInt32HashMap(),
		reverse: NewInt32CharHashMap(),
	}
}

// NewCharInt32HashBiMapWithCapacity creates a new empty CharInt32HashBiMap with the given initial capacity.
func NewCharInt32HashBiMapWithCapacity(capacity int) *CharInt32HashBiMap {
	return &CharInt32HashBiMap{
		forward: NewCharInt32HashMapWithCapacity(capacity),
		reverse: NewInt32CharHashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *CharInt32HashBiMap) Put(key uint16, value int32) (int32, bool) {
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
func (m *CharInt32HashBiMap) Get(key uint16) (int32, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *CharInt32HashBiMap) GetKey(value int32) (uint16, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *CharInt32HashBiMap) Remove(key uint16) (int32, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *CharInt32HashBiMap) RemoveValue(value int32) (uint16, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *CharInt32HashBiMap) ContainsKey(key uint16) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *CharInt32HashBiMap) ContainsValue(value int32) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *CharInt32HashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *CharInt32HashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *CharInt32HashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *CharInt32HashBiMap) ForEach(f func(uint16, int32)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *CharInt32HashBiMap) Keys() iter.Seq[uint16] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *CharInt32HashBiMap) Values() iter.Seq[int32] {
	return m.forward.Values()
}

// Inverse returns a new Int32CharHashBiMap with keys and values swapped.
func (m *CharInt32HashBiMap) Inverse() *Int32CharHashBiMap {
	result := NewInt32CharHashBiMap()
	m.forward.ForEach(func(k uint16, v int32) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *CharInt32HashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k uint16, v int32) {
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
func (m *CharInt32HashBiMap) Equals(other *CharInt32HashBiMap) bool {
	return m.forward.Equals(other.forward)
}
