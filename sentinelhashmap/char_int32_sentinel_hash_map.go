
package sentinelhashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	charInt32SentinelHashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
	charInt32SentinelHashMapEmptyKey   = uint16(0)
	charInt32SentinelHashMapRemovedKey = uint16(1)
)

// CharInt32SentinelHashMap is a sentinel-based open-addressing hash map with uint16 keys and int32 values.
// It uses sentinel values (0=empty, 1=removed) to track slot state.
// Keys 0 and 1 are stored separately in dedicated fields.
type CharInt32SentinelHashMap struct {
	keys   []uint16
	values []int32
	size   int

	// Sentinel key storage — keys 0 and 1 are valid user keys but also
	// serve as empty/removed markers in the table, so we store them separately.
	zeroKeyPresent bool
	zeroKeyValue   int32
	oneKeyPresent  bool
	oneKeyValue    int32
}

// NewCharInt32SentinelHashMap creates a new empty CharInt32SentinelHashMap with default capacity.
func NewCharInt32SentinelHashMap() *CharInt32SentinelHashMap {
	return NewCharInt32SentinelHashMapWithCapacity(charInt32SentinelHashMapDefaultCapacity)
}

// NewCharInt32SentinelHashMapWithCapacity creates a new empty CharInt32SentinelHashMap with the given initial capacity.
func NewCharInt32SentinelHashMapWithCapacity(capacity int) *CharInt32SentinelHashMap {
	cap := nextPowerOfTwoCharInt32SentinelHashMap(capacity)
	return &CharInt32SentinelHashMap{
		keys:   make([]uint16, cap),
		values: make([]int32, cap),
		size:   0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *CharInt32SentinelHashMap) Put(key uint16, value int32) (int32, bool) {
	if key == charInt32SentinelHashMapEmptyKey {
		old := m.zeroKeyValue
		existed := m.zeroKeyPresent
		m.zeroKeyValue = value
		if !m.zeroKeyPresent {
			m.zeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
	if key == charInt32SentinelHashMapRemovedKey {
		old := m.oneKeyValue
		existed := m.oneKeyPresent
		m.oneKeyValue = value
		if !m.oneKeyPresent {
			m.oneKeyPresent = true
			m.size++
		}
		return old, existed
	}
	return m.putRegular(key, value)
}

func (m *CharInt32SentinelHashMap) putRegular(key uint16, value int32) (int32, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := charInt32SentinelHashMapEmptyKey
	removed := charInt32SentinelHashMapRemovedKey
	firstRemoved := -1

	for {
		k := m.keys[idx]
		if k == empty {
			if firstRemoved >= 0 {
				idx = firstRemoved
			}
			m.keys[idx] = key
			m.values[idx] = value
			m.size++
			return 0, false
		}
		if k == removed {
			if firstRemoved < 0 {
				firstRemoved = idx
			}
			idx = (idx + 1) & mask
			continue
		}
		if k == key {
			old := m.values[idx]
			m.values[idx] = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *CharInt32SentinelHashMap) Get(key uint16) (int32, bool) {
	if key == charInt32SentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			return m.zeroKeyValue, true
		}
		return 0, false
	}
	if key == charInt32SentinelHashMapRemovedKey {
		if m.oneKeyPresent {
			return m.oneKeyValue, true
		}
		return 0, false
	}
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := charInt32SentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0, false
		}
		if k == key {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *CharInt32SentinelHashMap) GetOrDefault(key uint16, defaultValue int32) int32 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *CharInt32SentinelHashMap) Remove(key uint16) (int32, bool) {
	if key == charInt32SentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			old := m.zeroKeyValue
			m.zeroKeyPresent = false
			m.zeroKeyValue = 0
			m.size--
			return old, true
		}
		return 0, false
	}
	if key == charInt32SentinelHashMapRemovedKey {
		if m.oneKeyPresent {
			old := m.oneKeyValue
			m.oneKeyPresent = false
			m.oneKeyValue = 0
			m.size--
			return old, true
		}
		return 0, false
	}
	return m.removeRegular(key)
}

func (m *CharInt32SentinelHashMap) removeRegular(key uint16) (int32, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := charInt32SentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0, false
		}
		if k == key {
			old := m.values[idx]
			m.keys[idx] = charInt32SentinelHashMapRemovedKey
			m.values[idx] = 0
			m.size--
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *CharInt32SentinelHashMap) ContainsKey(key uint16) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *CharInt32SentinelHashMap) ContainsValue(value int32) bool {
	empty := charInt32SentinelHashMapEmptyKey
	removed := charInt32SentinelHashMapRemovedKey
	if m.zeroKeyPresent && m.zeroKeyValue == value {
		return true
	}
	if m.oneKeyPresent && m.oneKeyValue == value {
		return true
	}
	for i := range m.keys {
		if m.keys[i] != empty && m.keys[i] != removed && m.values[i] == value {
			return true
		}
	}
	return false
}

// Size returns the number of key-value pairs in the map.
func (m *CharInt32SentinelHashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *CharInt32SentinelHashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *CharInt32SentinelHashMap) Clear() {
	for i := range m.keys {
		m.keys[i] = 0
		m.values[i] = 0
	}
	m.zeroKeyPresent = false
	m.zeroKeyValue = 0
	m.oneKeyPresent = false
	m.oneKeyValue = 0
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *CharInt32SentinelHashMap) All() iter.Seq2[uint16, int32] {
	return func(yield func(uint16, int32) bool) {
		if m.zeroKeyPresent {
			if !yield(0, m.zeroKeyValue) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(charInt32SentinelHashMapRemovedKey, m.oneKeyValue) {
				return
			}
		}
		empty := charInt32SentinelHashMapEmptyKey
		removed := charInt32SentinelHashMapRemovedKey
		for i := range m.keys {
			if m.keys[i] != empty && m.keys[i] != removed {
				if !yield(m.keys[i], m.values[i]) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *CharInt32SentinelHashMap) Keys() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		if m.zeroKeyPresent {
			if !yield(0) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(charInt32SentinelHashMapRemovedKey) {
				return
			}
		}
		empty := charInt32SentinelHashMapEmptyKey
		removed := charInt32SentinelHashMapRemovedKey
		for i := range m.keys {
			if m.keys[i] != empty && m.keys[i] != removed {
				if !yield(m.keys[i]) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *CharInt32SentinelHashMap) Values() iter.Seq[int32] {
	return func(yield func(int32) bool) {
		if m.zeroKeyPresent {
			if !yield(m.zeroKeyValue) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(m.oneKeyValue) {
				return
			}
		}
		empty := charInt32SentinelHashMapEmptyKey
		removed := charInt32SentinelHashMapRemovedKey
		for i := range m.keys {
			if m.keys[i] != empty && m.keys[i] != removed {
				if !yield(m.values[i]) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *CharInt32SentinelHashMap) ForEach(f func(uint16, int32)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *CharInt32SentinelHashMap) Select(predicate func(uint16, int32) bool) *CharInt32SentinelHashMap {
	result := NewCharInt32SentinelHashMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *CharInt32SentinelHashMap) Reject(predicate func(uint16, int32) bool) *CharInt32SentinelHashMap {
	result := NewCharInt32SentinelHashMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *CharInt32SentinelHashMap) AnySatisfy(predicate func(uint16, int32) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *CharInt32SentinelHashMap) AllSatisfy(predicate func(uint16, int32) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *CharInt32SentinelHashMap) NoneSatisfy(predicate func(uint16, int32) bool) bool {
	return !m.AnySatisfy(predicate)
}

// String returns a string representation of the map.
func (m *CharInt32SentinelHashMap) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for k, v := range m.All() {
		if !first {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v: %v", k, v)
		first = false
	}
	sb.WriteString("}")
	return sb.String()
}

func (m *CharInt32SentinelHashMap) hashKey(key uint16) uint64 {
	return func() uint64 { h := uint64(key) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *CharInt32SentinelHashMap) needsResize() bool {
	// Count only regular entries (not sentinel entries) for load factor
	regularEntries := m.size
	if m.zeroKeyPresent {
		regularEntries--
	}
	if m.oneKeyPresent {
		regularEntries--
	}
	return (regularEntries+1)*4 > len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *CharInt32SentinelHashMap) resize() {
	oldKeys := m.keys
	oldValues := m.values
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = charInt32SentinelHashMapDefaultCapacity
	}

	// Save sentinel state
	savedSize := m.size
	savedZeroPresent := m.zeroKeyPresent
	savedZeroValue := m.zeroKeyValue
	savedOnePresent := m.oneKeyPresent
	savedOneValue := m.oneKeyValue

	m.keys = make([]uint16, newCap)
	m.values = make([]int32, newCap)
	m.size = 0
	m.zeroKeyPresent = false
	m.oneKeyPresent = false

	// Re-insert sentinel entries
	if savedZeroPresent {
		m.Put(0, savedZeroValue)
	}
	if savedOnePresent {
		m.Put(charInt32SentinelHashMapRemovedKey, savedOneValue)
	}

	// Re-insert regular entries
	empty := charInt32SentinelHashMapEmptyKey
	removed := charInt32SentinelHashMapRemovedKey
	for i := range oldKeys {
		if oldKeys[i] != empty && oldKeys[i] != removed {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
	_ = savedSize
}

func nextPowerOfTwoCharInt32SentinelHashMap(n int) int {
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
