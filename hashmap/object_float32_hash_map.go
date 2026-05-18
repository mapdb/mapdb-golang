
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	objectFloat32HashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// ObjectFloat32HashMap is an open-addressing hash map with generic comparable keys and float32 values.
// The value type is specialized to avoid boxing overhead.
type ObjectFloat32HashMap[K comparable] struct {
	keys     []K
	values   []float32
	occupied []bool
	size     int
}

// NewObjectFloat32HashMap creates a new empty ObjectFloat32HashMap with default capacity.
func NewObjectFloat32HashMap[K comparable]() *ObjectFloat32HashMap[K] {
	return NewObjectFloat32HashMapWithCapacity[K](objectFloat32HashMapDefaultCapacity)
}

// NewObjectFloat32HashMapWithCapacity creates a new empty ObjectFloat32HashMap with the given initial capacity.
func NewObjectFloat32HashMapWithCapacity[K comparable](capacity int) *ObjectFloat32HashMap[K] {
	cap := nextPowerOfTwoObjectFloat32HashMap(capacity)
	return &ObjectFloat32HashMap[K]{
		keys:     make([]K, cap),
		values:   make([]float32, cap),
		occupied: make([]bool, cap),
		size:     0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *ObjectFloat32HashMap[K]) Put(key K, value float32) (float32, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			m.keys[idx] = key
			m.values[idx] = value
			m.occupied[idx] = true
			m.size++
			return 0.0, false
		}
		if m.keys[idx] == key {
			old := m.values[idx]
			m.values[idx] = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *ObjectFloat32HashMap[K]) Get(key K) (float32, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0.0, false
	}
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			return 0.0, false
		}
		if m.keys[idx] == key {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *ObjectFloat32HashMap[K]) GetOrDefault(key K, defaultValue float32) float32 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *ObjectFloat32HashMap[K]) Remove(key K) (float32, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0.0, false
	}
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			return 0.0, false
		}
		if m.keys[idx] == key {
			old := m.values[idx]
			m.occupied[idx] = false
			var zeroK K
			m.keys[idx] = zeroK
			m.values[idx] = 0.0
			m.size--
			m.rehashFromObjectFloat32HashMap(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *ObjectFloat32HashMap[K]) ContainsKey(key K) bool {
	_, ok := m.Get(key)
	return ok
}

// Size returns the number of key-value pairs in the map.
func (m *ObjectFloat32HashMap[K]) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *ObjectFloat32HashMap[K]) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *ObjectFloat32HashMap[K]) Clear() {
	var zeroK K
	for i := range m.occupied {
		m.occupied[i] = false
		m.keys[i] = zeroK
		m.values[i] = 0.0
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ObjectFloat32HashMap[K]) All() iter.Seq2[K, float32] {
	return func(yield func(K, float32) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.keys[i], m.values[i]) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *ObjectFloat32HashMap[K]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.keys[i]) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *ObjectFloat32HashMap[K]) Values() iter.Seq[float32] {
	return func(yield func(float32) bool) {
		for i := range m.occupied {
			if m.occupied[i] {
				if !yield(m.values[i]) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *ObjectFloat32HashMap[K]) ForEach(f func(K, float32)) {
	for i := range m.occupied {
		if m.occupied[i] {
			f(m.keys[i], m.values[i])
		}
	}
}

// Select returns a new map containing only entries that satisfy the predicate.
func (m *ObjectFloat32HashMap[K]) Select(predicate func(K, float32) bool) *ObjectFloat32HashMap[K] {
	result := NewObjectFloat32HashMap[K]()
	for i := range m.occupied {
		if m.occupied[i] && predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// Reject returns a new map containing only entries that do not satisfy the predicate.
func (m *ObjectFloat32HashMap[K]) Reject(predicate func(K, float32) bool) *ObjectFloat32HashMap[K] {
	result := NewObjectFloat32HashMap[K]()
	for i := range m.occupied {
		if m.occupied[i] && !predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// String returns a string representation of the map.
func (m *ObjectFloat32HashMap[K]) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range m.occupied {
		if m.occupied[i] {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v: %v", m.keys[i], m.values[i])
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *ObjectFloat32HashMap[K]) needsResize() bool {
	return (m.size+1)*4 > len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *ObjectFloat32HashMap[K]) resize() {
	oldKeys := m.keys
	oldValues := m.values
	oldOccupied := m.occupied
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = objectFloat32HashMapDefaultCapacity
	}
	m.keys = make([]K, newCap)
	m.values = make([]float32, newCap)
	m.occupied = make([]bool, newCap)
	m.size = 0

	for i := range oldOccupied {
		if oldOccupied[i] {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func (m *ObjectFloat32HashMap[K]) rehashFromObjectFloat32HashMap(deleted int, mask int) {
	idx := (deleted + 1) & mask
	for m.occupied[idx] {
		ideal := int(hashComparable(m.keys[idx])) & mask
		if (idx-ideal+len(m.keys))&mask > (idx-deleted+len(m.keys))&mask {
		} else {
			m.keys[deleted] = m.keys[idx]
			m.values[deleted] = m.values[idx]
			m.occupied[deleted] = true
			m.occupied[idx] = false
			var zeroK K
			m.keys[idx] = zeroK
			m.values[idx] = 0.0
			deleted = idx
		}
		idx = (idx + 1) & mask
	}
}

func nextPowerOfTwoObjectFloat32HashMap(n int) int {
	if n <= 0 {
		return 16
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}
