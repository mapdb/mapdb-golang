
package hashmap

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedFloat32Int8HashMap is a thread-safe wrapper around Float32Int8HashMap.
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
type SynchronizedFloat32Int8HashMap struct {
	delegate *Float32Int8HashMap
	mu       sync.RWMutex
}

// NewSynchronizedFloat32Int8HashMap wraps a mutable map with synchronization.
func NewSynchronizedFloat32Int8HashMap() *SynchronizedFloat32Int8HashMap {
	return &SynchronizedFloat32Int8HashMap{delegate: NewFloat32Int8HashMap()}
}

// NewSynchronizedFloat32Int8HashMapWithCapacity wraps a new map with the given initial capacity.
func NewSynchronizedFloat32Int8HashMapWithCapacity(capacity int) *SynchronizedFloat32Int8HashMap {
	return &SynchronizedFloat32Int8HashMap{delegate: NewFloat32Int8HashMapWithCapacity(capacity)}
}

// NewSynchronizedFloat32Int8HashMapFrom wraps an existing map with synchronization.
// The wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedFloat32Int8HashMapFrom(m *Float32Int8HashMap) *SynchronizedFloat32Int8HashMap {
	return &SynchronizedFloat32Int8HashMap{delegate: m}
}

// snapshot returns (keys, values) slices in matching order, taken under RLock.
func (m *SynchronizedFloat32Int8HashMap) snapshot() (keys []float32, values []int8) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice(), m.delegate.ValuesToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *SynchronizedFloat32Int8HashMap) Put(key float32, value int8) (int8, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Put(key, value)
}

// Remove deletes the entry for the given key. Returns the previous value and true if found.
func (m *SynchronizedFloat32Int8HashMap) Remove(key float32) (int8, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Remove(key)
}

// Clear removes all entries.
func (m *SynchronizedFloat32Int8HashMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.Clear()
}

// AddToValue increments the value for the given key by `amount`,
// inserting it if absent. Holds the write lock; returns the new value.
func (m *SynchronizedFloat32Int8HashMap) AddToValue(key float32, amount int8) int8 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.AddToValue(key, amount)
}

// UpdateValue applies f to the current (or initial) value under the
// write lock. The callback must not re-enter this wrapper — it will
// deadlock. Prefer Get + Put on caller side if re-entry is needed.
func (m *SynchronizedFloat32Int8HashMap) UpdateValue(key float32, initial int8, f func(int8) int8) int8 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.UpdateValue(key, initial, f)
}

// ── simple reads ──────────────────────────────────────────────────────

// Get returns the value for the given key and true if found.
func (m *SynchronizedFloat32Int8HashMap) Get(key float32) (int8, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *SynchronizedFloat32Int8HashMap) GetOrDefault(key float32, defaultValue int8) int8 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *SynchronizedFloat32Int8HashMap) ContainsKey(key float32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if any entry's value matches.
func (m *SynchronizedFloat32Int8HashMap) ContainsValue(value int8) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *SynchronizedFloat32Int8HashMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *SynchronizedFloat32Int8HashMap) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.IsEmpty()
}

// SumOfValues returns the sum of all values, under RLock.
func (m *SynchronizedFloat32Int8HashMap) SumOfValues() int8 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.SumOfValues()
}

// KeysToSlice returns a copy of all keys.
func (m *SynchronizedFloat32Int8HashMap) KeysToSlice() []float32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns a copy of all values.
func (m *SynchronizedFloat32Int8HashMap) ValuesToSlice() []int8 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *SynchronizedFloat32Int8HashMap) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq2 over a snapshot of all key-value pairs.
// Iteration is lock-free.
func (m *SynchronizedFloat32Int8HashMap) All() iter.Seq2[float32, int8] {
	keys, values := m.snapshot()
	return func(yield func(float32, int8) bool) {
		for i := range keys {
			if !yield(keys[i], values[i]) {
				return
			}
		}
	}
}

// Keys returns an iter.Seq over a snapshot of keys.
func (m *SynchronizedFloat32Int8HashMap) Keys() iter.Seq[float32] {
	keys, _ := m.snapshot()
	return func(yield func(float32) bool) {
		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq over a snapshot of values.
func (m *SynchronizedFloat32Int8HashMap) Values() iter.Seq[int8] {
	_, values := m.snapshot()
	return func(yield func(int8) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional (callback) methods over snapshot ──────────────────────

// ForEach iterates entries over a snapshot. Callback runs unlocked.
func (m *SynchronizedFloat32Int8HashMap) ForEach(f func(float32, int8)) {
	keys, values := m.snapshot()
	for i := range keys {
		f(keys[i], values[i])
	}
}

// ForEachKey iterates keys over a snapshot. Callback runs unlocked.
func (m *SynchronizedFloat32Int8HashMap) ForEachKey(f func(float32)) {
	keys, _ := m.snapshot()
	for _, k := range keys {
		f(k)
	}
}

// ForEachValue iterates values over a snapshot. Callback runs unlocked.
func (m *SynchronizedFloat32Int8HashMap) ForEachValue(f func(int8)) {
	_, values := m.snapshot()
	for _, v := range values {
		f(v)
	}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *SynchronizedFloat32Int8HashMap) AnySatisfy(predicate func(float32, int8) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every entry satisfies the predicate.
func (m *SynchronizedFloat32Int8HashMap) AllSatisfy(predicate func(float32, int8) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *SynchronizedFloat32Int8HashMap) NoneSatisfy(predicate func(float32, int8) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *SynchronizedFloat32Int8HashMap) Count(predicate func(float32, int8) bool) int {
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
func (m *SynchronizedFloat32Int8HashMap) Detect(predicate func(float32, int8) bool) (float32, int8, bool) {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return keys[i], values[i], true
		}
	}
	var zeroK float32
	var zeroV int8
	return zeroK, zeroV, false
}

// InjectInto folds entries into an accumulator, callback unlocked.
func (m *SynchronizedFloat32Int8HashMap) InjectInto(initial int8, f func(int8, float32, int8) int8) int8 {
	keys, values := m.snapshot()
	acc := initial
	for i := range keys {
		acc = f(acc, keys[i], values[i])
	}
	return acc
}

// ── functional that return a new map ─────────────────────────────────

// Select returns a new (unsynchronized) map with entries satisfying predicate.
func (m *SynchronizedFloat32Int8HashMap) Select(predicate func(float32, int8) bool) *Float32Int8HashMap {
	keys, values := m.snapshot()
	result := NewFloat32Int8HashMap()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// Reject returns a new (unsynchronized) map with entries NOT satisfying predicate.
func (m *SynchronizedFloat32Int8HashMap) Reject(predicate func(float32, int8) bool) *Float32Int8HashMap {
	keys, values := m.snapshot()
	result := NewFloat32Int8HashMap()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// ── fluent mutators ───────────────────────────────────────────────────

func (m *SynchronizedFloat32Int8HashMap) WithKeyValue(key float32, value int8) *SynchronizedFloat32Int8HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithKeyValue(key, value)
	return m
}

func (m *SynchronizedFloat32Int8HashMap) WithoutKey(key float32) *SynchronizedFloat32Int8HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutKey(key)
	return m
}

// WithoutAllKeys is variadic for caller convenience; internally the
// slice is passed straight through since the underlying method already
// accepts a slice.
func (m *SynchronizedFloat32Int8HashMap) WithoutAllKeys(keys ...float32) *SynchronizedFloat32Int8HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutAllKeys(keys)
	return m
}

// ── conversions & equals ──────────────────────────────────────────────

func (m *SynchronizedFloat32Int8HashMap) ToImmutable() *ImmutableFloat32Int8HashMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (m *SynchronizedFloat32Int8HashMap) Equals(other *SynchronizedFloat32Int8HashMap) bool {
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
