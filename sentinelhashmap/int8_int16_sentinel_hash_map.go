
package sentinelhashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int8Int16SentinelHashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
	int8Int16SentinelHashMapEmptyKey   = int8(0)
	int8Int16SentinelHashMapRemovedKey = int8(1)
)

// Int8Int16SentinelHashMap is a sentinel-based open-addressing hash map with int8 keys and int16 values.
// It uses sentinel values (0=empty, 1=removed) to track slot state.
// Keys 0 and 1 are stored separately in dedicated fields.
type Int8Int16SentinelHashMap struct {
	keys   []int8
	values []int16
	size   int

	// Sentinel key storage — keys 0 and 1 are valid user keys but also
	// serve as empty/removed markers in the table, so we store them separately.
	zeroKeyPresent bool
	zeroKeyValue   int16
	oneKeyPresent  bool
	oneKeyValue    int16
}

// NewInt8Int16SentinelHashMap creates a new empty Int8Int16SentinelHashMap with default capacity.
func NewInt8Int16SentinelHashMap() *Int8Int16SentinelHashMap {
	return NewInt8Int16SentinelHashMapWithCapacity(int8Int16SentinelHashMapDefaultCapacity)
}

// NewInt8Int16SentinelHashMapWithCapacity creates a new empty Int8Int16SentinelHashMap with the given initial capacity.
func NewInt8Int16SentinelHashMapWithCapacity(capacity int) *Int8Int16SentinelHashMap {
	cap := nextPowerOfTwoInt8Int16SentinelHashMap(capacity)
	return &Int8Int16SentinelHashMap{
		keys:   make([]int8, cap),
		values: make([]int16, cap),
		size:   0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Int8Int16SentinelHashMap) Put(key int8, value int16) (int16, bool) {
	if key == int8Int16SentinelHashMapEmptyKey {
		old := m.zeroKeyValue
		existed := m.zeroKeyPresent
		m.zeroKeyValue = value
		if !m.zeroKeyPresent {
			m.zeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
	if key == int8Int16SentinelHashMapRemovedKey {
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

func (m *Int8Int16SentinelHashMap) putRegular(key int8, value int16) (int16, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := int8Int16SentinelHashMapEmptyKey
	removed := int8Int16SentinelHashMapRemovedKey
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
func (m *Int8Int16SentinelHashMap) Get(key int8) (int16, bool) {
	if key == int8Int16SentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			return m.zeroKeyValue, true
		}
		return 0, false
	}
	if key == int8Int16SentinelHashMapRemovedKey {
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
	empty := int8Int16SentinelHashMapEmptyKey

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
func (m *Int8Int16SentinelHashMap) GetOrDefault(key int8, defaultValue int16) int16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *Int8Int16SentinelHashMap) Remove(key int8) (int16, bool) {
	if key == int8Int16SentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			old := m.zeroKeyValue
			m.zeroKeyPresent = false
			m.zeroKeyValue = 0
			m.size--
			return old, true
		}
		return 0, false
	}
	if key == int8Int16SentinelHashMapRemovedKey {
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

func (m *Int8Int16SentinelHashMap) removeRegular(key int8) (int16, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := int8Int16SentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0, false
		}
		if k == key {
			old := m.values[idx]
			m.keys[idx] = int8Int16SentinelHashMapRemovedKey
			m.values[idx] = 0
			m.size--
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *Int8Int16SentinelHashMap) ContainsKey(key int8) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *Int8Int16SentinelHashMap) ContainsValue(value int16) bool {
	empty := int8Int16SentinelHashMapEmptyKey
	removed := int8Int16SentinelHashMapRemovedKey
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
func (m *Int8Int16SentinelHashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *Int8Int16SentinelHashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *Int8Int16SentinelHashMap) Clear() {
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
func (m *Int8Int16SentinelHashMap) All() iter.Seq2[int8, int16] {
	return func(yield func(int8, int16) bool) {
		if m.zeroKeyPresent {
			if !yield(0, m.zeroKeyValue) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(int8Int16SentinelHashMapRemovedKey, m.oneKeyValue) {
				return
			}
		}
		empty := int8Int16SentinelHashMapEmptyKey
		removed := int8Int16SentinelHashMapRemovedKey
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
func (m *Int8Int16SentinelHashMap) Keys() iter.Seq[int8] {
	return func(yield func(int8) bool) {
		if m.zeroKeyPresent {
			if !yield(0) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(int8Int16SentinelHashMapRemovedKey) {
				return
			}
		}
		empty := int8Int16SentinelHashMapEmptyKey
		removed := int8Int16SentinelHashMapRemovedKey
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
func (m *Int8Int16SentinelHashMap) Values() iter.Seq[int16] {
	return func(yield func(int16) bool) {
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
		empty := int8Int16SentinelHashMapEmptyKey
		removed := int8Int16SentinelHashMapRemovedKey
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
func (m *Int8Int16SentinelHashMap) ForEach(f func(int8, int16)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *Int8Int16SentinelHashMap) Select(predicate func(int8, int16) bool) *Int8Int16SentinelHashMap {
	result := NewInt8Int16SentinelHashMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *Int8Int16SentinelHashMap) Reject(predicate func(int8, int16) bool) *Int8Int16SentinelHashMap {
	result := NewInt8Int16SentinelHashMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *Int8Int16SentinelHashMap) AnySatisfy(predicate func(int8, int16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *Int8Int16SentinelHashMap) AllSatisfy(predicate func(int8, int16) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *Int8Int16SentinelHashMap) NoneSatisfy(predicate func(int8, int16) bool) bool {
	return !m.AnySatisfy(predicate)
}

// String returns a string representation of the map.
func (m *Int8Int16SentinelHashMap) String() string {
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

func (m *Int8Int16SentinelHashMap) hashKey(key int8) uint64 {
	return func() uint64 { h := uint64(key) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *Int8Int16SentinelHashMap) needsResize() bool {
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

func (m *Int8Int16SentinelHashMap) resize() {
	oldKeys := m.keys
	oldValues := m.values
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = int8Int16SentinelHashMapDefaultCapacity
	}

	// Save sentinel state
	savedSize := m.size
	savedZeroPresent := m.zeroKeyPresent
	savedZeroValue := m.zeroKeyValue
	savedOnePresent := m.oneKeyPresent
	savedOneValue := m.oneKeyValue

	m.keys = make([]int8, newCap)
	m.values = make([]int16, newCap)
	m.size = 0
	m.zeroKeyPresent = false
	m.oneKeyPresent = false

	// Re-insert sentinel entries
	if savedZeroPresent {
		m.Put(0, savedZeroValue)
	}
	if savedOnePresent {
		m.Put(int8Int16SentinelHashMapRemovedKey, savedOneValue)
	}

	// Re-insert regular entries
	empty := int8Int16SentinelHashMapEmptyKey
	removed := int8Int16SentinelHashMapRemovedKey
	for i := range oldKeys {
		if oldKeys[i] != empty && oldKeys[i] != removed {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
	_ = savedSize
}

func nextPowerOfTwoInt8Int16SentinelHashMap(n int) int {
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
