
package sentinelhashmap

import (
	"fmt"
	"iter"
	"math"
	"strings"
)

const (
	charFloat32SentinelHashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
	charFloat32SentinelHashMapEmptyKey   = uint16(0)
	charFloat32SentinelHashMapRemovedKey = uint16(1)
)

// CharFloat32SentinelHashMap is a sentinel-based open-addressing hash map with uint16 keys and float32 values.
// It uses sentinel values (0=empty, 1=removed) to track slot state.
// Keys 0 and 1 are stored separately in dedicated fields.
type CharFloat32SentinelHashMap struct {
	keys   []uint16
	values []float32
	size   int

	// Sentinel key storage — keys 0 and 1 are valid user keys but also
	// serve as empty/removed markers in the table, so we store them separately.
	zeroKeyPresent bool
	zeroKeyValue   float32
	oneKeyPresent  bool
	oneKeyValue    float32
}

// NewCharFloat32SentinelHashMap creates a new empty CharFloat32SentinelHashMap with default capacity.
func NewCharFloat32SentinelHashMap() *CharFloat32SentinelHashMap {
	return NewCharFloat32SentinelHashMapWithCapacity(charFloat32SentinelHashMapDefaultCapacity)
}

// NewCharFloat32SentinelHashMapWithCapacity creates a new empty CharFloat32SentinelHashMap with the given initial capacity.
func NewCharFloat32SentinelHashMapWithCapacity(capacity int) *CharFloat32SentinelHashMap {
	cap := nextPowerOfTwoCharFloat32SentinelHashMap(capacity)
	return &CharFloat32SentinelHashMap{
		keys:   make([]uint16, cap),
		values: make([]float32, cap),
		size:   0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *CharFloat32SentinelHashMap) Put(key uint16, value float32) (float32, bool) {
	if key == charFloat32SentinelHashMapEmptyKey {
		old := m.zeroKeyValue
		existed := m.zeroKeyPresent
		m.zeroKeyValue = value
		if !m.zeroKeyPresent {
			m.zeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
	if key == charFloat32SentinelHashMapRemovedKey {
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

func (m *CharFloat32SentinelHashMap) putRegular(key uint16, value float32) (float32, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := charFloat32SentinelHashMapEmptyKey
	removed := charFloat32SentinelHashMapRemovedKey
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
			return 0.0, false
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
func (m *CharFloat32SentinelHashMap) Get(key uint16) (float32, bool) {
	if key == charFloat32SentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			return m.zeroKeyValue, true
		}
		return 0.0, false
	}
	if key == charFloat32SentinelHashMapRemovedKey {
		if m.oneKeyPresent {
			return m.oneKeyValue, true
		}
		return 0.0, false
	}
	cap := len(m.keys)
	if cap == 0 {
		return 0.0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := charFloat32SentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0.0, false
		}
		if k == key {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *CharFloat32SentinelHashMap) GetOrDefault(key uint16, defaultValue float32) float32 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *CharFloat32SentinelHashMap) Remove(key uint16) (float32, bool) {
	if key == charFloat32SentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			old := m.zeroKeyValue
			m.zeroKeyPresent = false
			m.zeroKeyValue = 0.0
			m.size--
			return old, true
		}
		return 0.0, false
	}
	if key == charFloat32SentinelHashMapRemovedKey {
		if m.oneKeyPresent {
			old := m.oneKeyValue
			m.oneKeyPresent = false
			m.oneKeyValue = 0.0
			m.size--
			return old, true
		}
		return 0.0, false
	}
	return m.removeRegular(key)
}

func (m *CharFloat32SentinelHashMap) removeRegular(key uint16) (float32, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0.0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := charFloat32SentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0.0, false
		}
		if k == key {
			old := m.values[idx]
			m.keys[idx] = charFloat32SentinelHashMapRemovedKey
			m.values[idx] = 0.0
			m.size--
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *CharFloat32SentinelHashMap) ContainsKey(key uint16) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *CharFloat32SentinelHashMap) ContainsValue(value float32) bool {
	empty := charFloat32SentinelHashMapEmptyKey
	removed := charFloat32SentinelHashMapRemovedKey
	if m.zeroKeyPresent && math.Float32bits(m.zeroKeyValue) == math.Float32bits(value) {
		return true
	}
	if m.oneKeyPresent && math.Float32bits(m.oneKeyValue) == math.Float32bits(value) {
		return true
	}
	for i := range m.keys {
		if m.keys[i] != empty && m.keys[i] != removed && math.Float32bits(m.values[i]) == math.Float32bits(value) {
			return true
		}
	}
	return false
}

// Size returns the number of key-value pairs in the map.
func (m *CharFloat32SentinelHashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *CharFloat32SentinelHashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *CharFloat32SentinelHashMap) Clear() {
	for i := range m.keys {
		m.keys[i] = 0
		m.values[i] = 0.0
	}
	m.zeroKeyPresent = false
	m.zeroKeyValue = 0.0
	m.oneKeyPresent = false
	m.oneKeyValue = 0.0
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *CharFloat32SentinelHashMap) All() iter.Seq2[uint16, float32] {
	return func(yield func(uint16, float32) bool) {
		if m.zeroKeyPresent {
			if !yield(0, m.zeroKeyValue) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(charFloat32SentinelHashMapRemovedKey, m.oneKeyValue) {
				return
			}
		}
		empty := charFloat32SentinelHashMapEmptyKey
		removed := charFloat32SentinelHashMapRemovedKey
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
func (m *CharFloat32SentinelHashMap) Keys() iter.Seq[uint16] {
	return func(yield func(uint16) bool) {
		if m.zeroKeyPresent {
			if !yield(0) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(charFloat32SentinelHashMapRemovedKey) {
				return
			}
		}
		empty := charFloat32SentinelHashMapEmptyKey
		removed := charFloat32SentinelHashMapRemovedKey
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
func (m *CharFloat32SentinelHashMap) Values() iter.Seq[float32] {
	return func(yield func(float32) bool) {
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
		empty := charFloat32SentinelHashMapEmptyKey
		removed := charFloat32SentinelHashMapRemovedKey
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
func (m *CharFloat32SentinelHashMap) ForEach(f func(uint16, float32)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *CharFloat32SentinelHashMap) Select(predicate func(uint16, float32) bool) *CharFloat32SentinelHashMap {
	result := NewCharFloat32SentinelHashMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *CharFloat32SentinelHashMap) Reject(predicate func(uint16, float32) bool) *CharFloat32SentinelHashMap {
	result := NewCharFloat32SentinelHashMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *CharFloat32SentinelHashMap) AnySatisfy(predicate func(uint16, float32) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *CharFloat32SentinelHashMap) AllSatisfy(predicate func(uint16, float32) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *CharFloat32SentinelHashMap) NoneSatisfy(predicate func(uint16, float32) bool) bool {
	return !m.AnySatisfy(predicate)
}

// String returns a string representation of the map.
func (m *CharFloat32SentinelHashMap) String() string {
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

func (m *CharFloat32SentinelHashMap) hashKey(key uint16) uint64 {
	return func() uint64 { h := uint64(key) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *CharFloat32SentinelHashMap) needsResize() bool {
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

func (m *CharFloat32SentinelHashMap) resize() {
	oldKeys := m.keys
	oldValues := m.values
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = charFloat32SentinelHashMapDefaultCapacity
	}

	// Save sentinel state
	savedSize := m.size
	savedZeroPresent := m.zeroKeyPresent
	savedZeroValue := m.zeroKeyValue
	savedOnePresent := m.oneKeyPresent
	savedOneValue := m.oneKeyValue

	m.keys = make([]uint16, newCap)
	m.values = make([]float32, newCap)
	m.size = 0
	m.zeroKeyPresent = false
	m.oneKeyPresent = false

	// Re-insert sentinel entries
	if savedZeroPresent {
		m.Put(0, savedZeroValue)
	}
	if savedOnePresent {
		m.Put(charFloat32SentinelHashMapRemovedKey, savedOneValue)
	}

	// Re-insert regular entries
	empty := charFloat32SentinelHashMapEmptyKey
	removed := charFloat32SentinelHashMapRemovedKey
	for i := range oldKeys {
		if oldKeys[i] != empty && oldKeys[i] != removed {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
	_ = savedSize
}

func nextPowerOfTwoCharFloat32SentinelHashMap(n int) int {
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
