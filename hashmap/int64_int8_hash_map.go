
package hashmap

import (
	"fmt"
	"iter"
	"strings"
)

const (
	int64Int8HashMapDefaultCapacity = 16
	// Load factor 3/4 = 0.75, using integer math to avoid float conversion per insert.
)

// int64Int8HashMapEntry holds a single slot in the hash map for cache locality.
type int64Int8HashMapEntry struct {
	key      int64
	value    int8
	occupied bool
}

// Int64Int8HashMap is an open-addressing hash map with int64 keys and int8 values.
type Int64Int8HashMap struct {
	entries []int64Int8HashMapEntry
	size    int
}

// NewInt64Int8HashMap creates a new empty Int64Int8HashMap with default capacity.
func NewInt64Int8HashMap() *Int64Int8HashMap {
	return NewInt64Int8HashMapWithCapacity(int64Int8HashMapDefaultCapacity)
}

// NewInt64Int8HashMapWithCapacity creates a new empty Int64Int8HashMap with the given initial capacity.
func NewInt64Int8HashMapWithCapacity(capacity int) *Int64Int8HashMap {
	cap := nextPowerOfTwoInt64Int8HashMap(capacity)
	return &Int64Int8HashMap{
		entries: make([]int64Int8HashMapEntry, cap),
		size:    0,
	}
}

// Int64Int8HashMapOf creates a new Int64Int8HashMap from key-value pairs.
func Int64Int8HashMapOf(pairs ...struct {
	Key   int64
	Value int8
}) *Int64Int8HashMap {
	m := NewInt64Int8HashMapWithCapacity(len(pairs) * 2)
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}
	return m
}

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *Int64Int8HashMap) Put(key int64, value int8) (int8, bool) {
	if m.needsResize() {
		m.resize()
	}
	cap := len(m.entries)
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.entries[idx].occupied {
			m.entries[idx].key = key
			m.entries[idx].value = value
			m.entries[idx].occupied = true
			m.size++
			return 0, false
		}
		if m.entries[idx].key == key {
			old := m.entries[idx].value
			m.entries[idx].value = value
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// Get returns the value for the given key and true if found, or the zero value and false if not.
func (m *Int64Int8HashMap) Get(key int64) (int8, bool) {
	cap := len(m.entries)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.entries[idx].occupied {
			return 0, false
		}
		if m.entries[idx].key == key {
			return m.entries[idx].value, true
		}
		idx = (idx + 1) & mask
	}
}

// GetOrDefault returns the value for the given key if present, or the default value otherwise.
func (m *Int64Int8HashMap) GetOrDefault(key int64, defaultValue int8) int8 {
	if v, ok := m.Get(key); ok {
		return v
	}
	return defaultValue
}

// Remove deletes the entry for the given key. Returns the previous value and true if the key existed.
func (m *Int64Int8HashMap) Remove(key int64) (int8, bool) {
	cap := len(m.entries)
	if cap == 0 {
		return 0, false
	}
	mask := cap - 1
	idx := int(m.hashKey(key)) & mask

	for {
		if !m.entries[idx].occupied {
			return 0, false
		}
		if m.entries[idx].key == key {
			old := m.entries[idx].value
			m.entries[idx].occupied = false
			m.entries[idx].key = 0
			m.entries[idx].value = 0
			m.size--
			// Backward-shift deletion: the sibling of linear probing that
			// closes the hole by pulling each subsequent probed-past entry
			// one slot back until we reach an empty slot or an entry whose
			// preferred index equals its current index. This is distinct
			// from Robin Hood hashing (which is an insertion strategy).
			m.rehashFrom(idx, mask)
			return old, true
		}
		idx = (idx + 1) & mask
	}
}

// ContainsKey returns true if the map contains the given key.
func (m *Int64Int8HashMap) ContainsKey(key int64) bool {
	_, ok := m.Get(key)
	return ok
}

// ContainsValue returns true if the map contains the given value.
func (m *Int64Int8HashMap) ContainsValue(value int8) bool {
	for i := range m.entries {
		if m.entries[i].occupied && m.entries[i].value == value {
			return true
		}
	}
	return false
}

// Size returns the number of key-value pairs in the map.
func (m *Int64Int8HashMap) Size() int {
	return m.size
}

// IsEmpty returns true if the map contains no entries.
func (m *Int64Int8HashMap) IsEmpty() bool {
	return m.size == 0
}

// Clear removes all entries from the map.
func (m *Int64Int8HashMap) Clear() {
	for i := range m.entries {
		m.entries[i] = int64Int8HashMapEntry{}
	}
	m.size = 0
}

// All returns an iter.Seq2 that yields all key-value pairs.
func (m *Int64Int8HashMap) All() iter.Seq2[int64, int8] {
	return func(yield func(int64, int8) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].key, m.entries[i].value) {
					return
				}
			}
		}
	}
}

// Keys returns an iter.Seq that yields all keys.
func (m *Int64Int8HashMap) Keys() iter.Seq[int64] {
	return func(yield func(int64) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].key) {
					return
				}
			}
		}
	}
}

// Values returns an iter.Seq that yields all values.
func (m *Int64Int8HashMap) Values() iter.Seq[int8] {
	return func(yield func(int8) bool) {
		for i := range m.entries {
			if m.entries[i].occupied {
				if !yield(m.entries[i].value) {
					return
				}
			}
		}
	}
}

// ForEach calls the given function for each key-value pair.
func (m *Int64Int8HashMap) ForEach(f func(int64, int8)) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key, m.entries[i].value)
		}
	}
}

// ForEachKey calls the given function for each key.
func (m *Int64Int8HashMap) ForEachKey(f func(int64)) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].key)
		}
	}
}

// ForEachValue calls the given function for each value.
func (m *Int64Int8HashMap) ForEachValue(f func(int8)) {
	for i := range m.entries {
		if m.entries[i].occupied {
			f(m.entries[i].value)
		}
	}
}

// Select returns a new map containing only the key-value pairs that satisfy the predicate.
func (m *Int64Int8HashMap) Select(predicate func(int64, int8) bool) *Int64Int8HashMap {
	result := NewInt64Int8HashMap()
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			result.Put(m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// Reject returns a new map containing only the key-value pairs that do not satisfy the predicate.
func (m *Int64Int8HashMap) Reject(predicate func(int64, int8) bool) *Int64Int8HashMap {
	result := NewInt64Int8HashMap()
	for i := range m.entries {
		if m.entries[i].occupied && !predicate(m.entries[i].key, m.entries[i].value) {
			result.Put(m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// Detect returns the first key-value pair that satisfies the predicate, or zero values and false.
func (m *Int64Int8HashMap) Detect(predicate func(int64, int8) bool) (int64, int8, bool) {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return m.entries[i].key, m.entries[i].value, true
		}
	}
	return 0, 0, false
}

// AnySatisfy returns true if any key-value pair satisfies the predicate.
func (m *Int64Int8HashMap) AnySatisfy(predicate func(int64, int8) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if all key-value pairs satisfy the predicate.
func (m *Int64Int8HashMap) AllSatisfy(predicate func(int64, int8) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && !predicate(m.entries[i].key, m.entries[i].value) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no key-value pair satisfies the predicate.
func (m *Int64Int8HashMap) NoneSatisfy(predicate func(int64, int8) bool) bool {
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			return false
		}
	}
	return true
}

// Count returns the number of key-value pairs that satisfy the predicate.
func (m *Int64Int8HashMap) Count(predicate func(int64, int8) bool) int {
	count := 0
	for i := range m.entries {
		if m.entries[i].occupied && predicate(m.entries[i].key, m.entries[i].value) {
			count++
		}
	}
	return count
}

// String returns a string representation of the map.
func (m *Int64Int8HashMap) String() string {
	if m.size == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	first := true
	for i := range m.entries {
		if m.entries[i].occupied {
			if !first {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%v: %v", m.entries[i].key, m.entries[i].value)
			first = false
		}
	}
	sb.WriteString("}")
	return sb.String()
}

// Equals returns true if the other map has the same key-value pairs.
func (m *Int64Int8HashMap) Equals(other *Int64Int8HashMap) bool {
	if m.size != other.size {
		return false
	}
	for i := range m.entries {
		if m.entries[i].occupied {
			v, ok := other.Get(m.entries[i].key)
			if !ok || !(v == m.entries[i].value) {
				return false
			}
		}
	}
	return true
}

// KeysToSlice returns all keys as a slice.
func (m *Int64Int8HashMap) KeysToSlice() []int64 {
	result := make([]int64, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].key)
		}
	}
	return result
}

// ValuesToSlice returns all values as a slice.
func (m *Int64Int8HashMap) ValuesToSlice() []int8 {
	result := make([]int8, 0, m.size)
	for i := range m.entries {
		if m.entries[i].occupied {
			result = append(result, m.entries[i].value)
		}
	}
	return result
}

// ToImmutable returns an immutable copy of this map.
func (m *Int64Int8HashMap) ToImmutable() *ImmutableInt64Int8HashMap {
	return ImmutableInt64Int8HashMapFrom(m)
}

// InjectInto performs a left fold over all key-value pairs.
func (m *Int64Int8HashMap) InjectInto(initial int8, f func(int8, int64, int8) int8) int8 {
	result := initial
	for i := range m.entries {
		if m.entries[i].occupied {
			result = f(result, m.entries[i].key, m.entries[i].value)
		}
	}
	return result
}

// AddToValue adds the given amount to the value for the key.
// If the key is not present, inserts it with the given amount as value.
// Returns the new value.
func (m *Int64Int8HashMap) AddToValue(key int64, amount int8) int8 {
	if v, ok := m.Get(key); ok {
		newVal := v + amount
		m.Put(key, newVal)
		return newVal
	}
	m.Put(key, amount)
	return amount
}

// UpdateValue updates the value for the key using the function.
// If key is absent, inserts initialValue first then applies the function.
// Returns the new value.
func (m *Int64Int8HashMap) UpdateValue(key int64, initialValue int8, f func(int8) int8) int8 {
	if v, ok := m.Get(key); ok {
		newVal := f(v)
		m.Put(key, newVal)
		return newVal
	}
	newVal := f(initialValue)
	m.Put(key, newVal)
	return newVal
}

// WithKeyValue returns the map after putting the key-value pair (fluent API).
func (m *Int64Int8HashMap) WithKeyValue(key int64, value int8) *Int64Int8HashMap {
	m.Put(key, value)
	return m
}

// WithoutKey returns the map after removing the key (fluent API).
func (m *Int64Int8HashMap) WithoutKey(key int64) *Int64Int8HashMap {
	m.Remove(key)
	return m
}

// WithoutAllKeys removes all given keys (fluent API).
func (m *Int64Int8HashMap) WithoutAllKeys(keys []int64) *Int64Int8HashMap {
	for _, k := range keys {
		m.Remove(k)
	}
	return m
}

// SumOfValues returns the sum of all values.
func (m *Int64Int8HashMap) SumOfValues() int8 {
	var sum int8
	for i := range m.entries {
		if m.entries[i].occupied {
			sum += m.entries[i].value
		}
	}
	return sum
}

// Entry returns a handle for in-place check-and-modify operations on the
// given key. The handle is not thread-safe: external synchronisation (the
// SynchronizedInt64Int8HashMap wrapper's Lock / RLock, or your own mutex) is required
// when multiple goroutines share the same underlying map. The name is
// modelled on Rust's std::collections::hash_map::Entry, not on Java's
// ConcurrentMap.compute; there is no internal locking, no CAS, and no
// atomicity guarantee across callback invocation.
func (m *Int64Int8HashMap) Entry(key int64) Int64Int8Entry {
	return Int64Int8Entry{m: m, key: key}
}

// Int64Int8Entry provides in-place check-and-modify operations for a single
// key. Not thread-safe — see Int64Int8HashMap.Entry.
type Int64Int8Entry struct {
	m   *Int64Int8HashMap
	key int64
}

// OrInsert inserts the default value if the key is absent, and returns the current value.
func (e Int64Int8Entry) OrInsert(defaultValue int8) int8 {
	if v, ok := e.m.Get(e.key); ok {
		return v
	}
	e.m.Put(e.key, defaultValue)
	return defaultValue
}

// OrInsertWith inserts the value from the function if the key is absent,
// and returns the current value.
func (e Int64Int8Entry) OrInsertWith(f func() int8) int8 {
	if v, ok := e.m.Get(e.key); ok {
		return v
	}
	val := f()
	e.m.Put(e.key, val)
	return val
}

// AndModify calls f with a pointer to the value if the key is present,
// and returns the entry for fluent chaining. If the key is absent, f is
// not called and the entry is returned unchanged.
//
// CAUTION: f must not call Put / OrInsert / OrInsertWith on the same map.
// Those calls may trigger a resize that reallocates the underlying
// entries slice, leaving f's pointer dangling into the old slice. To
// guard against silent data loss this path panics if it detects a
// resize happened during f — see the post-call check below.
func (e Int64Int8Entry) AndModify(f func(*int8)) Int64Int8Entry {
	cap := len(e.m.entries)
	if cap == 0 {
		return e
	}
	mask := cap - 1
	idx := int(e.m.hashKey(e.key)) & mask
	for {
		if !e.m.entries[idx].occupied {
			return e
		}
		if e.m.entries[idx].key == e.key {
			// Detect backing-slice identity before and after the callback.
			// If the slice header changed (resize) or length changed (rehash),
			// the pointer we passed to f aliased the pre-resize storage and
			// the mutation is lost. Panic rather than silently dropping data.
			prevPtr := &e.m.entries[0]
			prevLen := len(e.m.entries)
			f(&e.m.entries[idx].value)
			if prevLen != len(e.m.entries) || prevPtr != &e.m.entries[0] {
				panic("Int64Int8Entry.AndModify: map was resized during callback — do not mutate the map from within AndModify")
			}
			return e
		}
		idx = (idx + 1) & mask
	}
}

func (m *Int64Int8HashMap) hashKey(key int64) uint64 {
	return func() uint64 { h := uint64(key) * 0x9E3779B97F4A7C15; return h ^ (h >> 32) }()
}

func (m *Int64Int8HashMap) needsResize() bool {
	return (m.size+1)*4 >= len(m.entries)*3 // 0.75 load factor, integer math
}

func (m *Int64Int8HashMap) resize() {
	oldEntries := m.entries
	newCap := len(oldEntries) * 2
	if newCap == 0 {
		newCap = int64Int8HashMapDefaultCapacity
	}
	m.entries = make([]int64Int8HashMapEntry, newCap)
	m.size = 0

	for i := range oldEntries {
		if oldEntries[i].occupied {
			m.Put(oldEntries[i].key, oldEntries[i].value)
		}
	}
}

// rehashFrom fixes the invariant after a deletion using backward-shift.
func (m *Int64Int8HashMap) rehashFrom(deleted int, mask int) {
	c := len(m.entries)
	idx := (deleted + 1) & mask
	for m.entries[idx].occupied {
		ideal := int(m.hashKey(m.entries[idx].key)) & mask
		distCurrent := (idx - ideal + c) & mask
		distGap := (deleted - ideal + c) & mask
		if distCurrent > distGap {
			m.entries[deleted] = m.entries[idx]
			m.entries[idx] = int64Int8HashMapEntry{}
			deleted = idx
		}
		idx = (idx + 1) & mask
		if idx == deleted {
			break
		}
	}
}

func nextPowerOfTwoInt64Int8HashMap(n int) int {
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
