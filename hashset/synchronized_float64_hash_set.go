
package hashset

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedFloat64HashSet is a thread-safe wrapper around Float64HashSet.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take a
// caller-supplied function (Select, ForEach, AnySatisfy, …) snapshot
// the backing set under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a new set (Select, Reject, Union, Intersect,
// Difference, SymmetricDifference) return an unwrapped *Float64HashSet;
// the caller owns it.
type SynchronizedFloat64HashSet struct {
	delegate *Float64HashSet
	mu       sync.RWMutex
}

// NewSynchronizedFloat64HashSet creates a new thread-safe empty set.
func NewSynchronizedFloat64HashSet() *SynchronizedFloat64HashSet {
	return &SynchronizedFloat64HashSet{delegate: NewFloat64HashSet()}
}

// NewSynchronizedFloat64HashSetFrom wraps an existing set. The
// wrapper takes ownership — callers must not mutate the delegate
// directly without locking.
func NewSynchronizedFloat64HashSetFrom(s *Float64HashSet) *SynchronizedFloat64HashSet {
	return &SynchronizedFloat64HashSet{delegate: s}
}

// SynchronizedFloat64HashSetOf constructs a synchronized set from values.
func SynchronizedFloat64HashSetOf(values ...float64) *SynchronizedFloat64HashSet {
	s := NewFloat64HashSet()
	for _, v := range values {
		s.Add(v)
	}
	return &SynchronizedFloat64HashSet{delegate: s}
}

// snapshot returns a defensive copy of the set's elements under RLock.
func (s *SynchronizedFloat64HashSet) snapshot() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *SynchronizedFloat64HashSet) Add(value float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Add(value)
}

func (s *SynchronizedFloat64HashSet) AddAll(values ...float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAll(values...)
}

func (s *SynchronizedFloat64HashSet) Remove(value float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Remove(value)
}

func (s *SynchronizedFloat64HashSet) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *SynchronizedFloat64HashSet) Contains(value float64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
}

func (s *SynchronizedFloat64HashSet) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Size()
}

func (s *SynchronizedFloat64HashSet) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.IsEmpty()
}

func (s *SynchronizedFloat64HashSet) ToSlice() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

func (s *SynchronizedFloat64HashSet) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.String()
}

// ── iteration ────────────────────────────────────────────────────────

// All returns an iter.Seq over a snapshot. Iteration is lock-free.
func (s *SynchronizedFloat64HashSet) All() iter.Seq[float64] {
	snapshot := s.snapshot()
	return func(yield func(float64) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (s *SynchronizedFloat64HashSet) ForEach(f func(float64)) {
	for _, v := range s.snapshot() {
		f(v)
	}
}

func (s *SynchronizedFloat64HashSet) AnySatisfy(predicate func(float64) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *SynchronizedFloat64HashSet) AllSatisfy(predicate func(float64) bool) bool {
	for _, v := range s.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *SynchronizedFloat64HashSet) NoneSatisfy(predicate func(float64) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *SynchronizedFloat64HashSet) Detect(predicate func(float64) bool) (float64, bool) {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero float64
	return zero, false
}

// ── functional that return a new set ─────────────────────────────────

func (s *SynchronizedFloat64HashSet) Select(predicate func(float64) bool) *Float64HashSet {
	snapshot := s.snapshot()
	result := NewFloat64HashSet()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

func (s *SynchronizedFloat64HashSet) Reject(predicate func(float64) bool) *Float64HashSet {
	snapshot := s.snapshot()
	result := NewFloat64HashSet()
	for _, v := range snapshot {
		if !predicate(v) {
			result.Add(v)
		}
	}
	return result
}

// ── set operations (two-lock, deadlock-safe) ──────────────────────────

// lockPair acquires two RLocks in pointer-address order and returns
// a release function. Guarantees no A.op(B) ⟷ B.op(A) deadlock.
func (s *SynchronizedFloat64HashSet) lockPair(other *SynchronizedFloat64HashSet) func() {
	if s == other {
		s.mu.RLock()
		return func() { s.mu.RUnlock() }
	}
	first, second := s, other
	if uintptr(unsafe.Pointer(s)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, s
	}
	first.mu.RLock()
	second.mu.RLock()
	return func() { second.mu.RUnlock(); first.mu.RUnlock() }
}

func (s *SynchronizedFloat64HashSet) Union(other *SynchronizedFloat64HashSet) *Float64HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Union(other.delegate)
}

func (s *SynchronizedFloat64HashSet) Intersect(other *SynchronizedFloat64HashSet) *Float64HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Intersect(other.delegate)
}

func (s *SynchronizedFloat64HashSet) Difference(other *SynchronizedFloat64HashSet) *Float64HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Difference(other.delegate)
}

func (s *SynchronizedFloat64HashSet) SymmetricDifference(other *SynchronizedFloat64HashSet) *Float64HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.SymmetricDifference(other.delegate)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (s *SynchronizedFloat64HashSet) With(value float64) *SynchronizedFloat64HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.With(value)
	return s
}

func (s *SynchronizedFloat64HashSet) WithAll(values ...float64) *SynchronizedFloat64HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithAll(values...)
	return s
}

func (s *SynchronizedFloat64HashSet) Without(value float64) *SynchronizedFloat64HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Without(value)
	return s
}

func (s *SynchronizedFloat64HashSet) WithoutAll(values ...float64) *SynchronizedFloat64HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithoutAll(values...)
	return s
}

// ── conversions ───────────────────────────────────────────────────────

func (s *SynchronizedFloat64HashSet) ToImmutable() *ImmutableFloat64HashSet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent deadlocks under concurrent A.Equals(B) / B.Equals(A).
func (s *SynchronizedFloat64HashSet) Equals(other *SynchronizedFloat64HashSet) bool {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Equals(other.delegate)
}
