
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

// Int64Int64HashBiMap is a bidirectional map with int64 keys and int64 values.
// Both key-to-value and value-to-key lookups are O(1).
type Int64Int64HashBiMap struct {
	forward *Int64Int64HashMap
	reverse *Int64Int64HashMap
}

// NewInt64Int64HashBiMap creates a new empty Int64Int64HashBiMap with default capacity.
func NewInt64Int64HashBiMap() *Int64Int64HashBiMap {
	return &Int64Int64HashBiMap{
		forward: NewInt64Int64HashMap(),
		reverse: NewInt64Int64HashMap(),
	}
}

// NewInt64Int64HashBiMapWithCapacity creates a new empty Int64Int64HashBiMap with the given initial capacity.
func NewInt64Int64HashBiMapWithCapacity(capacity int) *Int64Int64HashBiMap {
	return &Int64Int64HashBiMap{
		forward: NewInt64Int64HashMapWithCapacity(capacity),
		reverse: NewInt64Int64HashMapWithCapacity(capacity),
	}
}

// Put inserts or updates a key-value pair in both directions.
// If the key already existed, the old value mapping is removed from the reverse map.
// If the value already existed as a value for a different key, that old key mapping is removed.
// Returns the previous value and true if the key existed.
func (m *Int64Int64HashBiMap) Put(key int64, value int64) (int64, bool) {
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
func (m *Int64Int64HashBiMap) Get(key int64) (int64, bool) {
	return m.forward.Get(key)
}

// GetKey returns the key for the given value and true if found, or the zero value and false if not.
func (m *Int64Int64HashBiMap) GetKey(value int64) (int64, bool) {
	return m.reverse.Get(value)
}

// Remove deletes the entry for the given key from both directions.
// Returns the previous value and true if the key existed.
func (m *Int64Int64HashBiMap) Remove(key int64) (int64, bool) {
	oldVal, existed := m.forward.Remove(key)
	if existed {
		m.reverse.Remove(oldVal)
	}
	return oldVal, existed
}

// RemoveValue deletes the entry for the given value from both directions.
// Returns the previous key and true if the value existed.
func (m *Int64Int64HashBiMap) RemoveValue(value int64) (int64, bool) {
	oldKey, existed := m.reverse.Remove(value)
	if existed {
		m.forward.Remove(oldKey)
	}
	return oldKey, existed
}

// ContainsKey returns true if the map contains the given key.
func (m *Int64Int64HashBiMap) ContainsKey(key int64) bool {
	return m.forward.ContainsKey(key)
}

// ContainsValue returns true if the map contains the given value.
func (m *Int64Int64HashBiMap) ContainsValue(value int64) bool {
	return m.reverse.ContainsKey(value)
}

// Size returns the number of key-value pairs in the map.
func (m *Int64Int64HashBiMap) Size() int {
	return m.forward.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *Int64Int64HashBiMap) IsEmpty() bool {
	return m.forward.IsEmpty()
}

// Clear removes all entries from both directions.
func (m *Int64Int64HashBiMap) Clear() {
	m.forward.Clear()
	m.reverse.Clear()
}

// ForEach calls the given function for each key-value pair.
func (m *Int64Int64HashBiMap) ForEach(f func(int64, int64)) {
	m.forward.ForEach(f)
}

// Keys returns an iter.Seq that yields all keys.
func (m *Int64Int64HashBiMap) Keys() iter.Seq[int64] {
	return m.forward.Keys()
}

// Values returns an iter.Seq that yields all values.
func (m *Int64Int64HashBiMap) Values() iter.Seq[int64] {
	return m.forward.Values()
}

// Inverse returns a new Int64Int64HashBiMap with keys and values swapped.
func (m *Int64Int64HashBiMap) Inverse() *Int64Int64HashBiMap {
	result := NewInt64Int64HashBiMap()
	m.forward.ForEach(func(k int64, v int64) {
		result.Put(v, k)
	})
	return result
}

// String returns a string representation of the bi-map.
func (m *Int64Int64HashBiMap) String() string {
	if m.forward.Size() == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	m.forward.ForEach(func(k int64, v int64) {
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
func (m *Int64Int64HashBiMap) Equals(other *Int64Int64HashBiMap) bool {
	return m.forward.Equals(other.forward)
}
