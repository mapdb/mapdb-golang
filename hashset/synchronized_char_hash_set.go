
package hashset

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedCharHashSet is a thread-safe wrapper around CharHashSet.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take a
// caller-supplied function (Select, ForEach, AnySatisfy, …) snapshot
// the backing set under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a new set (Select, Reject, Union, Intersect,
// Difference, SymmetricDifference) return an unwrapped *CharHashSet;
// the caller owns it.
type SynchronizedCharHashSet struct {
	delegate *CharHashSet
	mu       sync.RWMutex
}

// NewSynchronizedCharHashSet creates a new thread-safe empty set.
func NewSynchronizedCharHashSet() *SynchronizedCharHashSet {
	return &SynchronizedCharHashSet{delegate: NewCharHashSet()}
}

// NewSynchronizedCharHashSetFrom wraps an existing set. The
// wrapper takes ownership — callers must not mutate the delegate
// directly without locking.
func NewSynchronizedCharHashSetFrom(s *CharHashSet) *SynchronizedCharHashSet {
	return &SynchronizedCharHashSet{delegate: s}
}

// SynchronizedCharHashSetOf constructs a synchronized set from values.
func SynchronizedCharHashSetOf(values ...uint16) *SynchronizedCharHashSet {
	s := NewCharHashSet()
	for _, v := range values {
		s.Add(v)
	}
	return &SynchronizedCharHashSet{delegate: s}
}

// snapshot returns a defensive copy of the set's elements under RLock.
func (s *SynchronizedCharHashSet) snapshot() []uint16 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *SynchronizedCharHashSet) Add(value uint16) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Add(value)
}

func (s *SynchronizedCharHashSet) AddAll(values ...uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAll(values...)
}

func (s *SynchronizedCharHashSet) Remove(value uint16) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Remove(value)
}

func (s *SynchronizedCharHashSet) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *SynchronizedCharHashSet) Contains(value uint16) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
}

func (s *SynchronizedCharHashSet) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Size()
}

func (s *SynchronizedCharHashSet) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.IsEmpty()
}

func (s *SynchronizedCharHashSet) ToSlice() []uint16 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

func (s *SynchronizedCharHashSet) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.String()
}

// ── iteration ────────────────────────────────────────────────────────

// All returns an iter.Seq over a snapshot. Iteration is lock-free.
func (s *SynchronizedCharHashSet) All() iter.Seq[uint16] {
	snapshot := s.snapshot()
	return func(yield func(uint16) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (s *SynchronizedCharHashSet) ForEach(f func(uint16)) {
	for _, v := range s.snapshot() {
		f(v)
	}
}

func (s *SynchronizedCharHashSet) AnySatisfy(predicate func(uint16) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *SynchronizedCharHashSet) AllSatisfy(predicate func(uint16) bool) bool {
	for _, v := range s.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *SynchronizedCharHashSet) NoneSatisfy(predicate func(uint16) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *SynchronizedCharHashSet) Detect(predicate func(uint16) bool) (uint16, bool) {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero uint16
	return zero, false
}

// ── functional that return a new set ─────────────────────────────────

func (s *SynchronizedCharHashSet) Select(predicate func(uint16) bool) *CharHashSet {
	snapshot := s.snapshot()
	result := NewCharHashSet()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

func (s *SynchronizedCharHashSet) Reject(predicate func(uint16) bool) *CharHashSet {
	snapshot := s.snapshot()
	result := NewCharHashSet()
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
func (s *SynchronizedCharHashSet) lockPair(other *SynchronizedCharHashSet) func() {
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

func (s *SynchronizedCharHashSet) Union(other *SynchronizedCharHashSet) *CharHashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Union(other.delegate)
}

func (s *SynchronizedCharHashSet) Intersect(other *SynchronizedCharHashSet) *CharHashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Intersect(other.delegate)
}

func (s *SynchronizedCharHashSet) Difference(other *SynchronizedCharHashSet) *CharHashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Difference(other.delegate)
}

func (s *SynchronizedCharHashSet) SymmetricDifference(other *SynchronizedCharHashSet) *CharHashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.SymmetricDifference(other.delegate)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (s *SynchronizedCharHashSet) With(value uint16) *SynchronizedCharHashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.With(value)
	return s
}

func (s *SynchronizedCharHashSet) WithAll(values ...uint16) *SynchronizedCharHashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithAll(values...)
	return s
}

func (s *SynchronizedCharHashSet) Without(value uint16) *SynchronizedCharHashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Without(value)
	return s
}

func (s *SynchronizedCharHashSet) WithoutAll(values ...uint16) *SynchronizedCharHashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithoutAll(values...)
	return s
}

// ── conversions ───────────────────────────────────────────────────────

func (s *SynchronizedCharHashSet) ToImmutable() *ImmutableCharHashSet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent deadlocks under concurrent A.Equals(B) / B.Equals(A).
func (s *SynchronizedCharHashSet) Equals(other *SynchronizedCharHashSet) bool {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Equals(other.delegate)
}
