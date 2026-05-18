
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	objectInt16HashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// ObjectInt16HashMap is an open-addressing hash map with generic comparable keys and int16 values.
// The value type is specialized to avoid boxing overhead.
type ObjectInt16HashMap[K comparable] struct {
	keys     []K
	values   []int16
	occupied []bool
	size     int
}

// NewObjectInt16HashMap creates a new empty ObjectInt16HashMap with default capacity.
func NewObjectInt16HashMap[K comparable]() *ObjectInt16HashMap[K] {
	return NewObjectInt16HashMapWithCapacity[K](objectInt16HashMapDefaultCapacity)
}

// NewObjectInt16HashMapWithCapacity creates a new empty ObjectInt16HashMap with the given initial capacity.
func NewObjectInt16HashMapWithCapacity[K comparable](capacity int) *ObjectInt16HashMap[K] {
	cap := nextPowerOfTwoObjectInt16HashMap(capacity)
	return &ObjectInt16HashMap[K]{
		keys:     make([]K, cap),
		values:   make([]int16, cap),
		occupied: make([]bool, cap),
		size:     0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *ObjectInt16HashMap[K]) Put(key K, value int16) (int16, bool) {
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
			return 0, false
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
func (m *ObjectInt16HashMap[K]) Get(key K) (int16, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			return 0, false
		}
		if m.keys[idx] == key {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *ObjectInt16HashMap[K]) GetOrDefault(key K, defaultValue int16) int16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *ObjectInt16HashMap[K]) Remove(key K) (int16, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(hashComparable(key)) & mask

	for {
		if !m.occupied[idx] {
			return 0, false
		}
		if m.keys[idx] == key {
			old := m.values[idx]
			m.occupied[idx] = false
			var zeroK K
			m.keys[idx] = zeroK
			m.values[idx] = 0
			m.size--
			m.rehashFromObjectInt16HashMap(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *ObjectInt16HashMap[K]) ContainsKey(key K) bool {
	_, ok := m.Get(key)
	return ok
}

// Size returns the number of key-value pairs in the map.
func (m *ObjectInt16HashMap[K]) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *ObjectInt16HashMap[K]) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *ObjectInt16HashMap[K]) Clear() {
	var zeroK K
	for i := range m.occupied {
		m.occupied[i] = false
		m.keys[i] = zeroK
		m.values[i] = 0
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ObjectInt16HashMap[K]) All() iter.Seq2[K, int16] {
	return func(yield func(K, int16) bool) {
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
func (m *ObjectInt16HashMap[K]) Keys() iter.Seq[K] {
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
func (m *ObjectInt16HashMap[K]) Values() iter.Seq[int16] {
	return func(yield func(int16) bool) {
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
func (m *ObjectInt16HashMap[K]) ForEach(f func(K, int16)) {
	for i := range m.occupied {
		if m.occupied[i] {
			f(m.keys[i], m.values[i])
		}
	}
}

// Select returns a new map containing only entries that satisfy the predicate.
func (m *ObjectInt16HashMap[K]) Select(predicate func(K, int16) bool) *ObjectInt16HashMap[K] {
	result := NewObjectInt16HashMap[K]()
	for i := range m.occupied {
		if m.occupied[i] && predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// Reject returns a new map containing only entries that do not satisfy the predicate.
func (m *ObjectInt16HashMap[K]) Reject(predicate func(K, int16) bool) *ObjectInt16HashMap[K] {
	result := NewObjectInt16HashMap[K]()
	for i := range m.occupied {
		if m.occupied[i] && !predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// String returns a string representation of the map.
func (m *ObjectInt16HashMap[K]) String() string {
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

func (m *ObjectInt16HashMap[K]) needsResize() bool {
	return (m.size+1)*4 > len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *ObjectInt16HashMap[K]) resize() {
	oldKeys := m.keys
	oldValues := m.values
	oldOccupied := m.occupied
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = objectInt16HashMapDefaultCapacity
	}
	m.keys = make([]K, newCap)
	m.values = make([]int16, newCap)
	m.occupied = make([]bool, newCap)
	m.size = 0

	for i := range oldOccupied {
		if oldOccupied[i] {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func (m *ObjectInt16HashMap[K]) rehashFromObjectInt16HashMap(deleted int, mask int) {
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
			m.values[idx] = 0
			deleted = idx
		}
		idx = (idx + 1) & mask
	}
}

func nextPowerOfTwoObjectInt16HashMap(n int) int {
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
