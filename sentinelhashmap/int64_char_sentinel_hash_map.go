
package sentinelhashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int64CharSentinelHashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
	int64CharSentinelHashMapEmptyKey   = int64(0)
	int64CharSentinelHashMapRemovedKey = int64(1)
)

// Int64CharSentinelHashMap is a sentinel-based open-addressing hash map with int64 keys and uint16 values.
// It uses sentinel values (0=empty, 1=removed) to track slot state.
// Keys 0 and 1 are stored separately in dedicated fields.
type Int64CharSentinelHashMap struct {
	keys   []int64
	values []uint16
	size   int

	// Sentinel key storage — keys 0 and 1 are valid user keys but also
	// serve as empty/removed markers in the table, so we store them separately.
	zeroKeyPresent bool
	zeroKeyValue   uint16
	oneKeyPresent  bool
	oneKeyValue    uint16
}

// NewInt64CharSentinelHashMap creates a new empty Int64CharSentinelHashMap with default capacity.
func NewInt64CharSentinelHashMap() *Int64CharSentinelHashMap {
	return NewInt64CharSentinelHashMapWithCapacity(int64CharSentinelHashMapDefaultCapacity)
}

// NewInt64CharSentinelHashMapWithCapacity creates a new empty Int64CharSentinelHashMap with the given initial capacity.
func NewInt64CharSentinelHashMapWithCapacity(capacity int) *Int64CharSentinelHashMap {
	cap := nextPowerOfTwoInt64CharSentinelHashMap(capacity)
	return &Int64CharSentinelHashMap{
		keys:   make([]int64, cap),
		values: make([]uint16, cap),
		size:   0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Int64CharSentinelHashMap) Put(key int64, value uint16) (uint16, bool) {
	if key == int64CharSentinelHashMapEmptyKey {
		old := m.zeroKeyValue
		existed := m.zeroKeyPresent
		m.zeroKeyValue = value
		if !m.zeroKeyPresent {
			m.zeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
	if key == int64CharSentinelHashMapRemovedKey {
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

func (m *Int64CharSentinelHashMap) putRegular(key int64, value uint16) (uint16, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := int64CharSentinelHashMapEmptyKey
	removed := int64CharSentinelHashMapRemovedKey
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
func (m *Int64CharSentinelHashMap) Get(key int64) (uint16, bool) {
	if key == int64CharSentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			return m.zeroKeyValue, true
		}
		return 0, false
	}
	if key == int64CharSentinelHashMapRemovedKey {
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
	empty := int64CharSentinelHashMapEmptyKey

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
func (m *Int64CharSentinelHashMap) GetOrDefault(key int64, defaultValue uint16) uint16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *Int64CharSentinelHashMap) Remove(key int64) (uint16, bool) {
	if key == int64CharSentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			old := m.zeroKeyValue
			m.zeroKeyPresent = false
			m.zeroKeyValue = 0
			m.size--
			return old, true
		}
		return 0, false
	}
	if key == int64CharSentinelHashMapRemovedKey {
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

func (m *Int64CharSentinelHashMap) removeRegular(key int64) (uint16, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := int64CharSentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0, false
		}
		if k == key {
			old := m.values[idx]
			m.keys[idx] = int64CharSentinelHashMapRemovedKey
			m.values[idx] = 0
			m.size--
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *Int64CharSentinelHashMap) ContainsKey(key int64) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *Int64CharSentinelHashMap) ContainsValue(value uint16) bool {
	empty := int64CharSentinelHashMapEmptyKey
	removed := int64CharSentinelHashMapRemovedKey
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
func (m *Int64CharSentinelHashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *Int64CharSentinelHashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *Int64CharSentinelHashMap) Clear() {
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
func (m *Int64CharSentinelHashMap) All() iter.Seq2[int64, uint16] {
	return func(yield func(int64, uint16) bool) {
		if m.zeroKeyPresent {
			if !yield(0, m.zeroKeyValue) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(int64CharSentinelHashMapRemovedKey, m.oneKeyValue) {
				return
			}
		}
		empty := int64CharSentinelHashMapEmptyKey
		removed := int64CharSentinelHashMapRemovedKey
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
func (m *Int64CharSentinelHashMap) Keys() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		if m.zeroKeyPresent {
			if !yield(0) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(int64CharSentinelHashMapRemovedKey) {
				return
			}
		}
		empty := int64CharSentinelHashMapEmptyKey
		removed := int64CharSentinelHashMapRemovedKey
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
func (m *Int64CharSentinelHashMap) Values() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
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
		empty := int64CharSentinelHashMapEmptyKey
		removed := int64CharSentinelHashMapRemovedKey
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
func (m *Int64CharSentinelHashMap) ForEach(f func(int64, uint16)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *Int64CharSentinelHashMap) Select(predicate func(int64, uint16) bool) *Int64CharSentinelHashMap {
	result := NewInt64CharSentinelHashMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *Int64CharSentinelHashMap) Reject(predicate func(int64, uint16) bool) *Int64CharSentinelHashMap {
	result := NewInt64CharSentinelHashMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *Int64CharSentinelHashMap) AnySatisfy(predicate func(int64, uint16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *Int64CharSentinelHashMap) AllSatisfy(predicate func(int64, uint16) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *Int64CharSentinelHashMap) NoneSatisfy(predicate func(int64, uint16) bool) bool {
	return !m.AnySatisfy(predicate)
}

// String returns a string representation of the map.
func (m *Int64CharSentinelHashMap) String() string {
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

func (m *Int64CharSentinelHashMap) hashKey(key int64) uint64 {
	return func() uint64 { h := uint64(key) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *Int64CharSentinelHashMap) needsResize() bool {
	// Count only regular entries (not sentinel entries) for load factor
	regularEntries := m.size
	if m.zeroKeyPresent {
		regularEntries--
	}
	if m.oneKeyPresent {
		regularEntries--
	}
	return (regularEntries+1)*4 >= len(m.keys)*3 // 0.75 load factor, integer math
}

func (m *Int64CharSentinelHashMap) resize() {
	oldKeys := m.keys
	oldValues := m.values
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = int64CharSentinelHashMapDefaultCapacity
	}

	// Save sentinel state
	savedSize := m.size
	savedZeroPresent := m.zeroKeyPresent
	savedZeroValue := m.zeroKeyValue
	savedOnePresent := m.oneKeyPresent
	savedOneValue := m.oneKeyValue

	m.keys = make([]int64, newCap)
	m.values = make([]uint16, newCap)
	m.size = 0
	m.zeroKeyPresent = false
	m.oneKeyPresent = false

	// Re-insert sentinel entries
	if savedZeroPresent {
		m.Put(0, savedZeroValue)
	}
	if savedOnePresent {
		m.Put(int64CharSentinelHashMapRemovedKey, savedOneValue)
	}

	// Re-insert regular entries
	empty := int64CharSentinelHashMapEmptyKey
	removed := int64CharSentinelHashMapRemovedKey
	for i := range oldKeys {
		if oldKeys[i] != empty && oldKeys[i] != removed {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
	_ = savedSize
}

func nextPowerOfTwoInt64CharSentinelHashMap(n int) int {
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
