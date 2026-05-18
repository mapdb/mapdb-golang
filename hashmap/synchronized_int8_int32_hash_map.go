
package hashmap

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedInt8Int32HashMap is a thread-safe wrapper around Int8Int32HashMap.
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
type SynchronizedInt8Int32HashMap struct {
	delegate *Int8Int32HashMap
	mu       sync.RWMutex
}

// NewSynchronizedInt8Int32HashMap wraps a mutable map with synchronization.
func NewSynchronizedInt8Int32HashMap() *SynchronizedInt8Int32HashMap {
	return &SynchronizedInt8Int32HashMap{delegate: NewInt8Int32HashMap()}
}

// NewSynchronizedInt8Int32HashMapWithCapacity wraps a new map with the given initial capacity.
func NewSynchronizedInt8Int32HashMapWithCapacity(capacity int) *SynchronizedInt8Int32HashMap {
	return &SynchronizedInt8Int32HashMap{delegate: NewInt8Int32HashMapWithCapacity(capacity)}
}

// NewSynchronizedInt8Int32HashMapFrom wraps an existing map with synchronization.
// The wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedInt8Int32HashMapFrom(m *Int8Int32HashMap) *SynchronizedInt8Int32HashMap {
	return &SynchronizedInt8Int32HashMap{delegate: m}
}

// snapshot returns (keys, values) slices in matching order, taken under RLock.
func (m *SynchronizedInt8Int32HashMap) snapshot() (keys []int8, values []int32) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice(), m.delegate.ValuesToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

// Put inserts or updates a key-value pair. Returns the previous value and true if the key existed.
func (m *SynchronizedInt8Int32HashMap) Put(key int8, value int32) (int32, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Put(key, value)
}

// Remove deletes the entry for the given key. Returns the previous value and true if found.
func (m *SynchronizedInt8Int32HashMap) Remove(key int8) (int32, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.Remove(key)
}

// Clear removes all entries.
func (m *SynchronizedInt8Int32HashMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.Clear()
}

// AddToValue increments the value for the given key by `amount`,
// inserting it if absent. Holds the write lock; returns the new value.
func (m *SynchronizedInt8Int32HashMap) AddToValue(key int8, amount int32) int32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.AddToValue(key, amount)
}

// UpdateValue applies f to the current (or initial) value under the
// write lock. The callback must not re-enter this wrapper — it will
// deadlock. Prefer Get + Put on caller side if re-entry is needed.
func (m *SynchronizedInt8Int32HashMap) UpdateValue(key int8, initial int32, f func(int32) int32) int32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.delegate.UpdateValue(key, initial, f)
}

// ── simple reads ──────────────────────────────────────────────────────

// Get returns the value for the given key and true if found.
func (m *SynchronizedInt8Int32HashMap) Get(key int8) (int32, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Get(key)
}

// GetOrDefault returns the value for the given key if present, or the default value.
func (m *SynchronizedInt8Int32HashMap) GetOrDefault(key int8, defaultValue int32) int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.GetOrDefault(key, defaultValue)
}

// ContainsKey returns true if the map contains the given key.
func (m *SynchronizedInt8Int32HashMap) ContainsKey(key int8) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsKey(key)
}

// ContainsValue returns true if any entry's value matches.
func (m *SynchronizedInt8Int32HashMap) ContainsValue(value int32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ContainsValue(value)
}

// Size returns the number of key-value pairs.
func (m *SynchronizedInt8Int32HashMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.Size()
}

// IsEmpty returns true if the map contains no entries.
func (m *SynchronizedInt8Int32HashMap) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.IsEmpty()
}

// SumOfValues returns the sum of all values, under RLock.
func (m *SynchronizedInt8Int32HashMap) SumOfValues() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.SumOfValues()
}

// KeysToSlice returns a copy of all keys.
func (m *SynchronizedInt8Int32HashMap) KeysToSlice() []int8 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.KeysToSlice()
}

// ValuesToSlice returns a copy of all values.
func (m *SynchronizedInt8Int32HashMap) ValuesToSlice() []int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ValuesToSlice()
}

// String returns a string representation.
func (m *SynchronizedInt8Int32HashMap) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All returns an iter.Seq2 over a snapshot of all key-value pairs.
// Iteration is lock-free.
func (m *SynchronizedInt8Int32HashMap) All() iter.Seq2[int8, int32] {
	keys, values := m.snapshot()
	return func(yield func(int8, int32) bool) {
		for i := range keys {
			if !yield(keys[i], values[i]) {
				return
			}
		}
	}
}

// Keys returns an iter.Seq over a snapshot of keys.
func (m *SynchronizedInt8Int32HashMap) Keys() iter.Seq[int8] {
	keys, _ := m.snapshot()
	return func(yield func(int8) bool) {
		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iter.Seq over a snapshot of values.
func (m *SynchronizedInt8Int32HashMap) Values() iter.Seq[int32] {
	_, values := m.snapshot()
	return func(yield func(int32) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional (callback) methods over snapshot ──────────────────────

// ForEach iterates entries over a snapshot. Callback runs unlocked.
func (m *SynchronizedInt8Int32HashMap) ForEach(f func(int8, int32)) {
	keys, values := m.snapshot()
	for i := range keys {
		f(keys[i], values[i])
	}
}

// ForEachKey iterates keys over a snapshot. Callback runs unlocked.
func (m *SynchronizedInt8Int32HashMap) ForEachKey(f func(int8)) {
	keys, _ := m.snapshot()
	for _, k := range keys {
		f(k)
	}
}

// ForEachValue iterates values over a snapshot. Callback runs unlocked.
func (m *SynchronizedInt8Int32HashMap) ForEachValue(f func(int32)) {
	_, values := m.snapshot()
	for _, v := range values {
		f(v)
	}
}

// AnySatisfy returns true if any entry satisfies the predicate.
func (m *SynchronizedInt8Int32HashMap) AnySatisfy(predicate func(int8, int32) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every entry satisfies the predicate.
func (m *SynchronizedInt8Int32HashMap) AllSatisfy(predicate func(int8, int32) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// NoneSatisfy returns true if no entry satisfies the predicate.
func (m *SynchronizedInt8Int32HashMap) NoneSatisfy(predicate func(int8, int32) bool) bool {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return false
		}
	}
	return true
}

// Count returns the number of entries satisfying the predicate.
func (m *SynchronizedInt8Int32HashMap) Count(predicate func(int8, int32) bool) int {
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
func (m *SynchronizedInt8Int32HashMap) Detect(predicate func(int8, int32) bool) (int8, int32, bool) {
	keys, values := m.snapshot()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			return keys[i], values[i], true
		}
	}
	var zeroK int8
	var zeroV int32
	return zeroK, zeroV, false
}

// InjectInto folds entries into an accumulator, callback unlocked.
func (m *SynchronizedInt8Int32HashMap) InjectInto(initial int32, f func(int32, int8, int32) int32) int32 {
	keys, values := m.snapshot()
	acc := initial
	for i := range keys {
		acc = f(acc, keys[i], values[i])
	}
	return acc
}

// ── functional that return a new map ─────────────────────────────────

// Select returns a new (unsynchronized) map with entries satisfying predicate.
func (m *SynchronizedInt8Int32HashMap) Select(predicate func(int8, int32) bool) *Int8Int32HashMap {
	keys, values := m.snapshot()
	result := NewInt8Int32HashMap()
	for i := range keys {
		if predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// Reject returns a new (unsynchronized) map with entries NOT satisfying predicate.
func (m *SynchronizedInt8Int32HashMap) Reject(predicate func(int8, int32) bool) *Int8Int32HashMap {
	keys, values := m.snapshot()
	result := NewInt8Int32HashMap()
	for i := range keys {
		if !predicate(keys[i], values[i]) {
			result.Put(keys[i], values[i])
		}
	}
	return result
}

// ── fluent mutators ───────────────────────────────────────────────────

func (m *SynchronizedInt8Int32HashMap) WithKeyValue(key int8, value int32) *SynchronizedInt8Int32HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithKeyValue(key, value)
	return m
}

func (m *SynchronizedInt8Int32HashMap) WithoutKey(key int8) *SynchronizedInt8Int32HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutKey(key)
	return m
}

// WithoutAllKeys is variadic for caller convenience; internally the
// slice is passed straight through since the underlying method already
// accepts a slice.
func (m *SynchronizedInt8Int32HashMap) WithoutAllKeys(keys ...int8) *SynchronizedInt8Int32HashMap {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delegate.WithoutAllKeys(keys)
	return m
}

// ── conversions & equals ──────────────────────────────────────────────

func (m *SynchronizedInt8Int32HashMap) ToImmutable() *ImmutableInt8Int32HashMap {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (m *SynchronizedInt8Int32HashMap) Equals(other *SynchronizedInt8Int32HashMap) bool {
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
