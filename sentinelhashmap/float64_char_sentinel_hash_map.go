
package sentinelhashmap

import (
	"fmt"
	"iter"
	"math"
	"strings"
	"unsafe"
)

const (
	float64CharSentinelHashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
	float64CharSentinelHashMapEmptyKey   = float64(0)
	float64CharSentinelHashMapRemovedKey = float64(1)
)

// Float64CharSentinelHashMap is a sentinel-based open-addressing hash map with float64 keys and uint16 values.
// It uses sentinel values (0=empty, 1=removed) to track slot state.
// Keys 0 and 1 are stored separately in dedicated fields.
type Float64CharSentinelHashMap struct {
	keys   []float64
	values []uint16
	size   int

	// Sentinel key storage — keys 0 and 1 are valid user keys but also
	// serve as empty/removed markers in the table, so we store them separately.
	zeroKeyPresent bool
	zeroKeyValue   uint16
	oneKeyPresent  bool
	oneKeyValue    uint16
}

// NewFloat64CharSentinelHashMap creates a new empty Float64CharSentinelHashMap with default capacity.
func NewFloat64CharSentinelHashMap() *Float64CharSentinelHashMap {
	return NewFloat64CharSentinelHashMapWithCapacity(float64CharSentinelHashMapDefaultCapacity)
}

// NewFloat64CharSentinelHashMapWithCapacity creates a new empty Float64CharSentinelHashMap with the given initial capacity.
func NewFloat64CharSentinelHashMapWithCapacity(capacity int) *Float64CharSentinelHashMap {
	cap := nextPowerOfTwoFloat64CharSentinelHashMap(capacity)
	return &Float64CharSentinelHashMap{
		keys:   make([]float64, cap),
		values: make([]uint16, cap),
		size:   0,
	}
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Float64CharSentinelHashMap) Put(key float64, value uint16) (uint16, bool) {
	if key == float64CharSentinelHashMapEmptyKey {
		old := m.zeroKeyValue
		existed := m.zeroKeyPresent
		m.zeroKeyValue = value
		if !m.zeroKeyPresent {
			m.zeroKeyPresent = true
			m.size++
		}
		return old, existed
	}
	if key == float64CharSentinelHashMapRemovedKey {
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

func (m *Float64CharSentinelHashMap) putRegular(key float64, value uint16) (uint16, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.keys)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := float64CharSentinelHashMapEmptyKey
	removed := float64CharSentinelHashMapRemovedKey
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
		if math.Float64bits(k) == math.Float64bits(key) {
			old := m.values[idx]
			m.values[idx] = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *Float64CharSentinelHashMap) Get(key float64) (uint16, bool) {
	if key == float64CharSentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			return m.zeroKeyValue, true
		}
		return 0, false
	}
	if key == float64CharSentinelHashMapRemovedKey {
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
	empty := float64CharSentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0, false
		}
		if math.Float64bits(k) == math.Float64bits(key) {
			return m.values[idx], true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *Float64CharSentinelHashMap) GetOrDefault(key float64, defaultValue uint16) uint16 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *Float64CharSentinelHashMap) Remove(key float64) (uint16, bool) {
	if key == float64CharSentinelHashMapEmptyKey {
		if m.zeroKeyPresent {
			old := m.zeroKeyValue
			m.zeroKeyPresent = false
			m.zeroKeyValue = 0
			m.size--
			return old, true
		}
		return 0, false
	}
	if key == float64CharSentinelHashMapRemovedKey {
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

func (m *Float64CharSentinelHashMap) removeRegular(key float64) (uint16, bool) {
	cap := len(m.keys)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask
	empty := float64CharSentinelHashMapEmptyKey

	for {
		k := m.keys[idx]
		if k == empty {
			return 0, false
		}
		if math.Float64bits(k) == math.Float64bits(key) {
			old := m.values[idx]
			m.keys[idx] = float64CharSentinelHashMapRemovedKey
			m.values[idx] = 0
			m.size--
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *Float64CharSentinelHashMap) ContainsKey(key float64) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *Float64CharSentinelHashMap) ContainsValue(value uint16) bool {
	empty := float64CharSentinelHashMapEmptyKey
	removed := float64CharSentinelHashMapRemovedKey
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
func (m *Float64CharSentinelHashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *Float64CharSentinelHashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *Float64CharSentinelHashMap) Clear() {
	for i := range m.keys {
		m.keys[i] = 0.0
		m.values[i] = 0
	}
	m.zeroKeyPresent = false
	m.zeroKeyValue = 0
	m.oneKeyPresent = false
	m.oneKeyValue = 0
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Float64CharSentinelHashMap) All() iter.Seq2[float64, uint16] {
	return func(yield func(float64, uint16) bool) {
		if m.zeroKeyPresent {
			if !yield(0.0, m.zeroKeyValue) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(float64CharSentinelHashMapRemovedKey, m.oneKeyValue) {
				return
			}
		}
		empty := float64CharSentinelHashMapEmptyKey
		removed := float64CharSentinelHashMapRemovedKey
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
func (m *Float64CharSentinelHashMap) Keys() iter.Seq[float64] {
	return func(yield func(float64) bool) {
		if m.zeroKeyPresent {
			if !yield(0.0) {
				return
			}
		}
		if m.oneKeyPresent {
			if !yield(float64CharSentinelHashMapRemovedKey) {
				return
			}
		}
		empty := float64CharSentinelHashMapEmptyKey
		removed := float64CharSentinelHashMapRemovedKey
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
func (m *Float64CharSentinelHashMap) Values() iter.Seq[uint16] {
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
		empty := float64CharSentinelHashMapEmptyKey
		removed := float64CharSentinelHashMapRemovedKey
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
func (m *Float64CharSentinelHashMap) ForEach(f func(float64, uint16)) {
	for k, v := range m.All() {
		f(k, v)
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *Float64CharSentinelHashMap) Select(predicate func(float64, uint16) bool) *Float64CharSentinelHashMap {
	result := NewFloat64CharSentinelHashMap()
	for k, v := range m.All() {
		if predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *Float64CharSentinelHashMap) Reject(predicate func(float64, uint16) bool) *Float64CharSentinelHashMap {
	result := NewFloat64CharSentinelHashMap()
	for k, v := range m.All() {
		if !predicate(k, v) {
			result.Put(k, v)
		}
	}
	return result
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *Float64CharSentinelHashMap) AnySatisfy(predicate func(float64, uint16) bool) bool {
	for k, v := range m.All() {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *Float64CharSentinelHashMap) AllSatisfy(predicate func(float64, uint16) bool) bool {
	for k, v := range m.All() {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *Float64CharSentinelHashMap) NoneSatisfy(predicate func(float64, uint16) bool) bool {
	return !m.AnySatisfy(predicate)
}

// String returns a string representation of the map.
func (m *Float64CharSentinelHashMap) String() string {
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

func (m *Float64CharSentinelHashMap) hashKey(key float64) uint64 {
	return func() uint64 { h := *(*uint64)(unsafe.Pointer(&key)) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *Float64CharSentinelHashMap) needsResize() bool {
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

func (m *Float64CharSentinelHashMap) resize() {
	oldKeys := m.keys
	oldValues := m.values
	newCap := len(oldKeys) * 2
	if newCap == 0 {
		newCap = float64CharSentinelHashMapDefaultCapacity
	}

	// Save sentinel state
	savedSize := m.size
	savedZeroPresent := m.zeroKeyPresent
	savedZeroValue := m.zeroKeyValue
	savedOnePresent := m.oneKeyPresent
	savedOneValue := m.oneKeyValue

	m.keys = make([]float64, newCap)
	m.values = make([]uint16, newCap)
	m.size = 0
	m.zeroKeyPresent = false
	m.oneKeyPresent = false

	// Re-insert sentinel entries
	if savedZeroPresent {
		m.Put(0.0, savedZeroValue)
	}
	if savedOnePresent {
		m.Put(float64CharSentinelHashMapRemovedKey, savedOneValue)
	}

	// Re-insert regular entries
	empty := float64CharSentinelHashMapEmptyKey
	removed := float64CharSentinelHashMapRemovedKey
	for i := range oldKeys {
		if oldKeys[i] != empty && oldKeys[i] != removed {
			m.Put(oldKeys[i], oldValues[i])
		}
	}
	_ = savedSize
}

func nextPowerOfTwoFloat64CharSentinelHashMap(n int) int {
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
