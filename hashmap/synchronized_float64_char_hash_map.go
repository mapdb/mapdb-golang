
package hashmap

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedFloat64CharHashMap is a thread-safe wrapper around Float64CharHashMap.
//
// Read methods hold RLock; writes hold Lock. Functional methods
// (ForEach, Select, AnySatisfy, …) snapshot (keys, values) under
// RLock and run the callback unlocked, so callbacks may freely call
// back into this wrapper.
//
// Methods whose signature takes a callback AND mutates (UpdateValue)
// hold the write lock while invoking the callback — the callback
// must not re-enter the wrapper in that case. This matches the
// Java EC synchronized-collection convention.
type SynchronizedFloat64CharHashMap struct {
	delegate *Float64CharHashMap
	mu       sync.RWMutex
}

// NewSynchronizedFloat64CharHashMap wraps a mutable map with synchronization.
func NewSynchronizedFloat64CharHashMap() *SynchronizedFloat64CharHashMap {
	return &SynchronizedFloat64CharHashMap{delegate: NewFloat64CharHashMap()}
}

// NewSynchronizedFloat64CharHashMapWithCapacity wraps a new map with the given initial capacity.
func NewSynchronizedFloat64CharHashMapWithCapacity(capacity int) *SynchronizedFloat64CharHashMap {
	return &SynchronizedFloat64CharHashMap{delegate: NewFloat64CharHashMapWithCapacity(capacity)}
}

// NewSynchronizedFloat64CharHashMapFrom wraps an existing map with synchronization.
// The wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedFloat64CharHashMapFrom(m *Float64CharHashMap) *SynchronizedFloat64CharHashMap {
	return &SynchronizedFloat64CharHashMap{delegate: m}
}

// snapshot returns (keys, values) slices in matching order, taken under RLock.
func (m *SynchronizedFloat64CharHashMap) snapshot() (keys []float64, values []uint16) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice(), m.delegate.ValuesToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *SynchronizedFloat64CharHashMap) Put(key float64, value uint16) (uint16, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Put(key, value)
}

// Remove deletes the entry for the given key. Returns the previous value and true if found.
func (m *SynchronizedFloat64CharHashMap) Remove(key float64) (uint16, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Remove(key)
}

// Clear removes all entries.
func (m *SynchronizedFloat64CharHashMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.Clear()
}

// AddToValue increments the value for the given key by `amount`,
// inserting it if absent. Holds the write lock; returns the new value.
func (m *SynchronizedFloat64CharHashMap) AddToValue(key float64, amount uint16) uint16 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.AddToValue(key, amount)
}

// UpdateValue applies f to the current (or initial) value under the
// write lock. The callback must not re-enter this wrapper — it will
// deadlock. Prefer Get + Put on caller side if re-entry is needed.
func (m *SynchronizedFloat64CharHashMap) UpdateValue(key float64, initial uint16, f func(uint16) uint16) uint16 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.UpdateValue(key, initial, f)
}

// ── simple reads ──────────────────────────────────────────────────────

// Get returns the value for the given key and true if found.
func (m *SynchronizedFloat64CharHashMap) Get(key float64) (uint16, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *SynchronizedFloat64CharHashMap) GetOrDefault(key float64, defaultValue uint16) uint16 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *SynchronizedFloat64CharHashMap) ContainsKey(key float64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if any entry's value matches.
func (m *SynchronizedFloat64CharHashMap) ContainsValue(value uint16) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *SynchronizedFloat64CharHashMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *SynchronizedFloat64CharHashMap) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.IsEmpty()
}

// SumOfValues returns the sum of all values, under RLock.
func (m *SynchronizedFloat64CharHashMap) SumOfValues() uint16 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.SumOfValues()
}

// KeysToSlice returns a copy of all keys.
func (m *SynchronizedFloat64CharHashMap) KeysToSlice() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns a copy of all values.
func (m *SynchronizedFloat64CharHashMap) ValuesToSlice() []uint16 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *SynchronizedFloat64CharHashMap) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq2 over a snapshot of all key-value pairs.
// Iteration is lock-free.
func (m *SynchronizedFloat64CharHashMap) All() iter.Seq2[float64, uint16] {
	keys, values := m.snapshot()
	return func(yield func(float64, uint16) bool) {
		for i := range keys {
			if !yield(keys[i], values[i]) {
				return
			}
		}
	}
}

// Keys returns an iter.Seq over a snapshot of keys.
func (m *SynchronizedFloat64CharHashMap) Keys() iter.Seq[float64] {
	keys, _ := m.snapshot()
	return func(yield func(float64) bool) {
		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq over a snapshot of values.
func (m *SynchronizedFloat64CharHashMap) Values() iter.Seq[uint16] {
	_, values := m.snapshot()
	return func(yield func(uint16) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional (callback) methods over snapshot ──────────────────────

// ForEach iterates entries over a snapshot. Callback runs unlocked.
func (m *SynchronizedFloat64CharHashMap) ForEach(f func(float64, uint16)) {
	keys, values := m.snapshot()
	for i := range keys {
		f(keys[i], values[i])
	}
}

// ForEachKey iterates keys over a snapshot. Callback runs unlocked.
func (m *SynchronizedFloat64CharHashMap) ForEachKey(f func(float64)) {
	keys, _ := m.snapshot()
	for _, k := range keys {
		f(k)
	}
}

// ForEachValue iterates values over a snapshot. Callback runs unlocked.
func (m *SynchronizedFloat64CharHashMap) ForEachValue(f func(uint16)) {
	_, values := m.snapshot()
	for _, v := range values {
		f(v)
	}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *SynchronizedFloat64CharHashMap) AnySatisfy(predicate func(float64, uint16) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every entry satisfies the predicate.
func (m *SynchronizedFloat64CharHashMap) AllSatisfy(predicate func(float64, uint16) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *SynchronizedFloat64CharHashMap) NoneSatisfy(predicate func(float64, uint16) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *SynchronizedFloat64CharHashMap) Count(predicate func(float64, uint16) bool) int {
	keys, values := m.snapshot()
	n := 0
	for i := range keys {
		if predicate(keys[i], values[i]) {
			n++
		}
	}
	return n
}

// Detect returns any entry satisfying the predicate, or zero values and false.
func (m *SynchronizedFloat64CharHashMap) Detect(predicate func(float64, uint16) bool) (float64, uint16, bool) {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return keys[i], values[i], true
		}
	}
	var zeroK float64
	var zeroV uint16
	return zeroK, zeroV, false
}

// InjectInto folds entries into an accumulator, callback unlocked.
func (m *SynchronizedFloat64CharHashMap) InjectInto(initial uint16, f func(uint16, float64, uint16) uint16) uint16 {
	keys, values := m.snapshot()
	acc := initial
	for i := range keys {
		acc = f(acc, keys[i], values[i])
	}
	return acc
}

// ── functional that return a new map ─────────────────────────────────

// Select returns a new (unsynchronized) map with entries satisfying predicate.
func (m *SynchronizedFloat64CharHashMap) Select(predicate func(float64, uint16) bool) *Float64CharHashMap {
	keys, values := m.snapshot()
	result := NewFloat64CharHashMap()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// Reject returns a new (unsynchronized) map with entries NOT satisfying predicate.
func (m *SynchronizedFloat64CharHashMap) Reject(predicate func(float64, uint16) bool) *Float64CharHashMap {
	keys, values := m.snapshot()
	result := NewFloat64CharHashMap()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// ── fluent mutators ───────────────────────────────────────────────────

func (m *SynchronizedFloat64CharHashMap) WithKeyValue(key float64, value uint16) *SynchronizedFloat64CharHashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithKeyValue(key, value)
	return m
}

func (m *SynchronizedFloat64CharHashMap) WithoutKey(key float64) *SynchronizedFloat64CharHashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutKey(key)
	return m
}

// WithoutAllKeys is variadic for caller convenience; internally the
// slice is passed straight through since the underlying method already
// accepts a slice.
func (m *SynchronizedFloat64CharHashMap) WithoutAllKeys(keys ...float64) *SynchronizedFloat64CharHashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutAllKeys(keys)
	return m
}

// ── conversions & equals ──────────────────────────────────────────────

func (m *SynchronizedFloat64CharHashMap) ToImmutable() *ImmutableFloat64CharHashMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (m *SynchronizedFloat64CharHashMap) Equals(other *SynchronizedFloat64CharHashMap) bool {
	if m == other {
		m.mu.RLock()
		defer m.mu.RUnlock()
		return m.delegate.Equals(other.delegate)
	}
	first, second := m, other
	if uintptr(unsafe.Pointer(m)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, m
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	return m.delegate.Equals(other.delegate)
}

// Entry is NOT wrapped — the Entry API on the mutable map is designed
// for lock-free fast paths and would require returning a
// delegate-bound handle, which would race with other callers through
// the wrapper. If you need atomic check-and-modify under the synchronized
// wrapper, use UpdateValue or take the wrapper's lock externally and
// call into the delegate directly via a helper.
