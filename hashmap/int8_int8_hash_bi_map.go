
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

// Int8Int8HashBiMap is a bidirectional map with int8 keys and int8 values.
// Both key-to-value and value-to-key lookups are O(1).
type Int8Int8HashBiMap struct {
	forward *Int8Int8HashMap
	reverse *Int8Int8HashMap
}

// NewInt8Int8HashBiMap creates a new empty Int8Int8HashBiMap with default capacity.
func NewInt8Int8HashBiMap() *Int8Int8HashBiMap {
	return &Int8Int8HashBiMap{
		forward: NewInt8Int8HashMap(),
		reverse: NewInt8Int8HashMap(),
	}
}

// NewInt8Int8HashBiMapWithCapacity creates a new empty Int8Int8HashBiMap with the given initial capacity.
func NewInt8Int8HashBiMapWithCapacity(capacity int) *Int8Int8HashBiMap {
	return &Int8Int8HashBiMap{
		forward: NewInt8Int8HashMapWithCapacity(capacity),
		reverse: NewInt8Int8HashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *Int8Int8HashBiMap) Put(key int8, value int8) (int8, bool) {
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
func (m *Int8Int8HashBiMap) Get(key int8) (int8, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *Int8Int8HashBiMap) GetKey(value int8) (int8, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *Int8Int8HashBiMap) Remove(key int8) (int8, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *Int8Int8HashBiMap) RemoveValue(value int8) (int8, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *Int8Int8HashBiMap) ContainsKey(key int8) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Int8Int8HashBiMap) ContainsValue(value int8) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *Int8Int8HashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Int8Int8HashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *Int8Int8HashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *Int8Int8HashBiMap) ForEach(f func(int8, int8)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *Int8Int8HashBiMap) Keys() iter.Seq[int8] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Int8Int8HashBiMap) Values() iter.Seq[int8] {
	return m.forward.Values()
}

// Inverse returns a new Int8Int8HashBiMap with keys and values swapped.
func (m *Int8Int8HashBiMap) Inverse() *Int8Int8HashBiMap {
	result := NewInt8Int8HashBiMap()
	m.forward.ForEach(func(k int8, v int8) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *Int8Int8HashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k int8, v int8) {
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
func (m *Int8Int8HashBiMap) Equals(other *Int8Int8HashBiMap) bool {
	return m.forward.Equals(other.forward)
}
