
package hashmap

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedCharInt64HashMap is a thread-safe wrapper around CharInt64HashMap.
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
type SynchronizedCharInt64HashMap struct {
	delegate *CharInt64HashMap
	mu       sync.RWMutex
}

// NewSynchronizedCharInt64HashMap wraps a mutable map with synchronization.
func NewSynchronizedCharInt64HashMap() *SynchronizedCharInt64HashMap {
	return &SynchronizedCharInt64HashMap{delegate: NewCharInt64HashMap()}
}

// NewSynchronizedCharInt64HashMapWithCapacity wraps a new map with the given initial capacity.
func NewSynchronizedCharInt64HashMapWithCapacity(capacity int) *SynchronizedCharInt64HashMap {
	return &SynchronizedCharInt64HashMap{delegate: NewCharInt64HashMapWithCapacity(capacity)}
}

// NewSynchronizedCharInt64HashMapFrom wraps an existing map with synchronization.
// The wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedCharInt64HashMapFrom(m *CharInt64HashMap) *SynchronizedCharInt64HashMap {
	return &SynchronizedCharInt64HashMap{delegate: m}
}

// snapshot returns (keys, values) slices in matching order, taken under RLock.
func (m *SynchronizedCharInt64HashMap) snapshot() (keys []uint16, values []int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice(), m.delegate.ValuesToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *SynchronizedCharInt64HashMap) Put(key uint16, value int64) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Put(key, value)
}

// Remove deletes the entry for the given key. Returns the previous value and true if found.
func (m *SynchronizedCharInt64HashMap) Remove(key uint16) (int64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Remove(key)
}

// Clear removes all entries.
func (m *SynchronizedCharInt64HashMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.Clear()
}

// AddToValue increments the value for the given key by `amount`,
// inserting it if absent. Holds the write lock; returns the new value.
func (m *SynchronizedCharInt64HashMap) AddToValue(key uint16, amount int64) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.AddToValue(key, amount)
}

// UpdateValue applies f to the current (or initial) value under the
// write lock. The callback must not re-enter this wrapper — it will
// deadlock. Prefer Get + Put on caller side if re-entry is needed.
func (m *SynchronizedCharInt64HashMap) UpdateValue(key uint16, initial int64, f func(int64) int64) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.UpdateValue(key, initial, f)
}

// ── simple reads ──────────────────────────────────────────────────────

// Get returns the value for the given key and true if found.
func (m *SynchronizedCharInt64HashMap) Get(key uint16) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *SynchronizedCharInt64HashMap) GetOrDefault(key uint16, defaultValue int64) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *SynchronizedCharInt64HashMap) ContainsKey(key uint16) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if any entry's value matches.
func (m *SynchronizedCharInt64HashMap) ContainsValue(value int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *SynchronizedCharInt64HashMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *SynchronizedCharInt64HashMap) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.IsEmpty()
}

// SumOfValues returns the sum of all values, under RLock.
func (m *SynchronizedCharInt64HashMap) SumOfValues() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.SumOfValues()
}

// KeysToSlice returns a copy of all keys.
func (m *SynchronizedCharInt64HashMap) KeysToSlice() []uint16 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns a copy of all values.
func (m *SynchronizedCharInt64HashMap) ValuesToSlice() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *SynchronizedCharInt64HashMap) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq2 over a snapshot of all key-value pairs.
// Iteration is lock-free.
func (m *SynchronizedCharInt64HashMap) All() iter.Seq2[uint16, int64] {
	keys, values := m.snapshot()
	return func(yield func(uint16, int64) bool) {
		for i := range keys {
			if !yield(keys[i], values[i]) {
				return
			}
		}
	}
}

// Keys returns an iter.Seq over a snapshot of keys.
func (m *SynchronizedCharInt64HashMap) Keys() iter.Seq[uint16] {
	keys, _ := m.snapshot()
	return func(yield func(uint16) bool) {
		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq over a snapshot of values.
func (m *SynchronizedCharInt64HashMap) Values() iter.Seq[int64] {
	_, values := m.snapshot()
	return func(yield func(int64) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional (callback) methods over snapshot ──────────────────────

// ForEach iterates entries over a snapshot. Callback runs unlocked.
func (m *SynchronizedCharInt64HashMap) ForEach(f func(uint16, int64)) {
	keys, values := m.snapshot()
	for i := range keys {
		f(keys[i], values[i])
	}
}

// ForEachKey iterates keys over a snapshot. Callback runs unlocked.
func (m *SynchronizedCharInt64HashMap) ForEachKey(f func(uint16)) {
	keys, _ := m.snapshot()
	for _, k := range keys {
		f(k)
	}
}

// ForEachValue iterates values over a snapshot. Callback runs unlocked.
func (m *SynchronizedCharInt64HashMap) ForEachValue(f func(int64)) {
	_, values := m.snapshot()
	for _, v := range values {
		f(v)
	}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *SynchronizedCharInt64HashMap) AnySatisfy(predicate func(uint16, int64) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every entry satisfies the predicate.
func (m *SynchronizedCharInt64HashMap) AllSatisfy(predicate func(uint16, int64) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *SynchronizedCharInt64HashMap) NoneSatisfy(predicate func(uint16, int64) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *SynchronizedCharInt64HashMap) Count(predicate func(uint16, int64) bool) int {
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
func (m *SynchronizedCharInt64HashMap) Detect(predicate func(uint16, int64) bool) (uint16, int64, bool) {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return keys[i], values[i], true
		}
	}
	var zeroK uint16
	var zeroV int64
	return zeroK, zeroV, false
}

// InjectInto folds entries into an accumulator, callback unlocked.
func (m *SynchronizedCharInt64HashMap) InjectInto(initial int64, f func(int64, uint16, int64) int64) int64 {
	keys, values := m.snapshot()
	acc := initial
	for i := range keys {
		acc = f(acc, keys[i], values[i])
	}
	return acc
}

// ── functional that return a new map ─────────────────────────────────

// Select returns a new (unsynchronized) map with entries satisfying predicate.
func (m *SynchronizedCharInt64HashMap) Select(predicate func(uint16, int64) bool) *CharInt64HashMap {
	keys, values := m.snapshot()
	result := NewCharInt64HashMap()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// Reject returns a new (unsynchronized) map with entries NOT satisfying predicate.
func (m *SynchronizedCharInt64HashMap) Reject(predicate func(uint16, int64) bool) *CharInt64HashMap {
	keys, values := m.snapshot()
	result := NewCharInt64HashMap()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// ── fluent mutators ───────────────────────────────────────────────────

func (m *SynchronizedCharInt64HashMap) WithKeyValue(key uint16, value int64) *SynchronizedCharInt64HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithKeyValue(key, value)
	return m
}

func (m *SynchronizedCharInt64HashMap) WithoutKey(key uint16) *SynchronizedCharInt64HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutKey(key)
	return m
}

// WithoutAllKeys is variadic for caller convenience; internally the
// slice is passed straight through since the underlying method already
// accepts a slice.
func (m *SynchronizedCharInt64HashMap) WithoutAllKeys(keys ...uint16) *SynchronizedCharInt64HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutAllKeys(keys)
	return m
}

// ── conversions & equals ──────────────────────────────────────────────

func (m *SynchronizedCharInt64HashMap) ToImmutable() *ImmutableCharInt64HashMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (m *SynchronizedCharInt64HashMap) Equals(other *SynchronizedCharInt64HashMap) bool {
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
