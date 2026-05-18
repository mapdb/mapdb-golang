
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int32ObjectHashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// Int32ObjectHashMap is an open-addressing hash map with int32 keys and generic values.
// The key type is specialized to avoid boxing overhead.
type Int32ObjectHashMap[V any] struct {
	keys     []int32
	values   []V
	occupied []bool
	size     int
}

// NewInt32ObjectHashMap creates a new empty Int32ObjectHashMap with default capacity.
func NewInt32ObjectHashMap[V any]() *Int32ObjectHashMap[V] {
	return NewInt32ObjectHashMapWithCapacity[V](int32ObjectHashMapDefaultCapacity)
}

// NewInt32ObjectHashMapWithCapacity creates a new empty Int32ObjectHashMap with the given initial capacity.
func NewInt32ObjectHashMapWithCapacity[V any](capacity int) *Int32ObjectHashMap[V] {
	cap := nextPowerOfTwoInt32ObjectHashMap(capacity)
	return &Int32ObjectHashMap[V]{
		keys:     make([]int32, cap),
		values:   make([]V, cap),
		occupied: make([]bool, cap),
		size:     0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Int32ObjectHashMap[V]) Put(key int32, value V) (V, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.occupied[idx] {
			m.keys[idx] = key
			m.values[idx] = value
			m.occupied[idx] = true
			m.size++
			var zero V
			return zero, false
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
func (m *Int32ObjectHashMap[V]) Get(key int32) (V, bool) {
	cap := len(m.keys)
	if cap == 0 {
		var zero V
		return zero, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.occupied[idx] {
			var zero V
			return zero, false
		}
		if m.keys[idx] == key {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *Int32ObjectHashMap[V]) GetOrDefault(key int32, defaultValue V) V {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *Int32ObjectHashMap[V]) Remove(key int32) (V, bool) {
	cap := len(m.keys)
	if cap == 0 {
		var zero V
		return zero, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.occupied[idx] {
			var zero V
			return zero, false
		}
		if m.keys[idx] == key {
			old := m.values[idx]
			m.occupied[idx] = false
			m.keys[idx] = 0
			var zeroV V
			m.values[idx] = zeroV
			m.size--
			m.rehashFrom(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *Int32ObjectHashMap[V]) ContainsKey(key int32) bool {
	_, ok := m.Get(key)
	return ok
}

// Size returns the number of key-value pairs in the map.
func (m *Int32ObjectHashMap[V]) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *Int32ObjectHashMap[V]) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *Int32ObjectHashMap[V]) Clear() {
	var zeroV V
	for i := range m.occupied {
		m.occupied[i] = false
		m.keys[i] = 0
		m.values[i] = zeroV
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Int32ObjectHashMap[V]) All() iter.Seq2[int32, V] {
	return func(yield func(int32, V) bool) {
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
func (m *Int32ObjectHashMap[V]) Keys() iter.Seq[int32] {
	return func(yield func(int32) bool) {
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
func (m *Int32ObjectHashMap[V]) Values() iter.Seq[V] {
	return func(yield func(V) bool) {
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
func (m *Int32ObjectHashMap[V]) ForEach(f func(int32, V)) {
	for i := range m.occupied {
		if m.occupied[i] {
			f(m.keys[i], m.values[i])
		}
	}
}

// Select returns a new map containing only entries that satisfy the predicate.
func (m *Int32ObjectHashMap[V]) Select(predicate func(int32, V) bool) *Int32ObjectHashMap[V] {
	result := NewInt32ObjectHashMap[V]()
	for i := range m.occupied {
		if m.occupied[i] && predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// Reject returns a new map containing only entries that do not satisfy the predicate.
func (m *Int32ObjectHashMap[V]) Reject(predicate func(int32, V) bool) *Int32ObjectHashMap[V] {
	result := NewInt32ObjectHashMap[V]()
	for i := range m.occupied {
		if m.occupied[i] && !predicate(m.keys[i], m.values[i]) {
			result.Put(m.keys[i], m.values[i])
		}
	}
	return result
}

// String returns a string representation of the map.
func (m *Int32ObjectHashMap[V]) String() string {
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

func (m *Int32ObjectHashMap[V]) hashKey(key int32) uint64 {
	return func() uint64 { h := uint64(uint32(key)) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *Int32ObjectHashMap[V]) needsResize() bool {
	return (m.size+1)*4 > len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *Int32ObjectHashMap[V]) resize() {
	oldKeys := m.keys
	oldValues := m.values
	oldOccupied := m.occupied
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = int32ObjectHashMapDefaultCapacity
	}
	m.keys = make([]int32, newCap)
	m.values = make([]V, newCap)
	m.occupied = make([]bool, newCap)
	m.size = 0

	for i := range oldOccupied {
		if oldOccupied[i] {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
}

func (m *Int32ObjectHashMap[V]) rehashFrom(deleted int, mask int) {
	idx := (deleted + 1) & mask
	for m.occupied[idx] {
		ideal := int(m.hashKey(m.keys[idx])) & mask
		if (idx-ideal+len(m.keys))&mask > (idx-deleted+len(m.keys))&mask {
		} else {
			m.keys[deleted] = m.keys[idx]
			m.values[deleted] = m.values[idx]
			m.occupied[deleted] = true
			m.occupied[idx] = false
			m.keys[idx] = 0
			var zeroV V
			m.values[idx] = zeroV
			deleted = idx
		}
		idx = (idx + 1) & mask
	}
}

func nextPowerOfTwoInt32ObjectHashMap(n int) int {
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
