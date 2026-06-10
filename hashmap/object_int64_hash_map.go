
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	objectInt64HashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// ObjectInt64HashMap is an open-addressing hash map with generic comparable keys and int64 values.
// The value type is specialized to avoid boxing overhead.
type ObjectInt64HashMap[K comparable] struct {
	keys     []K
	values   []int64
	occupied []bool
	size     int
}

// NewObjectInt64HashMap creates a new empty ObjectInt64HashMap with default capacity.
func NewObjectInt64HashMap[K comparable]() *ObjectInt64HashMap[K] {
	return NewObjectInt64HashMapWithCapacity[K](objectInt64HashMapDefaultCapacity)
}

// NewObjectInt64HashMapWithCapacity creates a new empty ObjectInt64HashMap with the given initial capacity.
func NewObjectInt64HashMapWithCapacity[K comparable](capacity int) *ObjectInt64HashMap[K] {
	cap := nextPowerOfTwoObjectInt64HashMap(capacity)
	return &ObjectInt64HashMap[K]{
		keys:     make([]K, cap),
		values:   make([]int64, cap),
		occupied: make([]bool, cap),
		size:     0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *ObjectInt64HashMap[K]) Put(key K, value int64) (int64, bool) {
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
func (m *ObjectInt64HashMap[K]) Get(key K) (int64, bool) {
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
func (m *ObjectInt64HashMap[K]) GetOrDefault(key K, defaultValue int64) int64 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *ObjectInt64HashMap[K]) Remove(key K) (int64, bool) {
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
			m.rehashFromObjectInt64HashMap(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *ObjectInt64HashMap[K]) ContainsKey(key K) bool {
	_, ok := m.Get(key)
	return ok
}

// Size returns the number of key-value pairs in the map.
func (m *ObjectInt64HashMap[K]) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *ObjectInt64HashMap[K]) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *ObjectInt64HashMap[K]) Clear() {
	var zeroK K
	for i := range m.occupied {
		m.occupied[i] = false
		m.keys[i] = zeroK
		m.values[i] = 0
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *ObjectInt64HashMap[K]) All() iter.Seq2[K, int64] {
	return func(yield func(K, int64) bool) {
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
func (m *ObjectInt64HashMap[K]) Keys() iter.Seq[K] {
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
func (m *ObjectInt64HashMap[K]) Values() iter.Seq[int64] {
	return func(yield func(int64) bool) {
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
func (m *ObjectInt64HashMap[K]) ForEach(f func(K, int64)) {
	for i := range m.occupied {
		if m.occupied[i] {
			f(m.keys[i], m.values[i])
		}
	}
}

// Select returns a new map containing only entries that satisfy the predicate.
func (m *ObjectInt64HashMap[K]) Select(predicate func(K, int64) bool) *ObjectInt64HashMap[K] {
	result := NewObjectInt64HashMap[K]()
	for i := range m.occupied {
		if m.occupied[i] && predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// Reject returns a new map containing only entries that do not satisfy the predicate.
func (m *ObjectInt64HashMap[K]) Reject(predicate func(K, int64) bool) *ObjectInt64HashMap[K] {
	result := NewObjectInt64HashMap[K]()
	for i := range m.occupied {
		if m.occupied[i] && !predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// String returns a string representation of the map.
func (m *ObjectInt64HashMap[K]) String() string {
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

func (m *ObjectInt64HashMap[K]) needsResize() bool {
	return (m.size+1)*4 >= len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *ObjectInt64HashMap[K]) resize() {
	oldKeys := m.keys
	oldValues := m.values
	oldOccupied := m.occupied
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = objectInt64HashMapDefaultCapacity
	}
	m.keys = make([]K, newCap)
	m.values = make([]int64, newCap)
	m.occupied = make([]bool, newCap)
	m.size = 0

	for i := range oldOccupied {
		if oldOccupied[i] {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func (m *ObjectInt64HashMap[K]) rehashFromObjectInt64HashMap(deleted int, mask int) {
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

func nextPowerOfTwoObjectInt64HashMap(n int) int {
	if n <= 0 {
		return 16
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32 // no-op on 32-bit platforms (Go shifts are width-defined), required on 64-bit
	n++
	return n
}
