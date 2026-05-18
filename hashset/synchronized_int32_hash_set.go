
package hashset

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedInt32HashSet is a thread-safe wrapper around Int32HashSet.
//
// Read methods hold an RLock; writes hold a Lock. Methods that take a
// caller-supplied function (Select, ForEach, AnySatisfy, …) snapshot
// the backing set under RLock and release it before invoking the
// callback, so the callback is free to call back into the wrapper
// without deadlocking.
//
// Methods that return a new set (Select, Reject, Union, Intersect,
// Difference, SymmetricDifference) return an unwrapped *Int32HashSet;
// the caller owns it.
type SynchronizedInt32HashSet struct {
	delegate *Int32HashSet
	mu       sync.RWMutex
}

// NewSynchronizedInt32HashSet creates a new thread-safe empty set.
func NewSynchronizedInt32HashSet() *SynchronizedInt32HashSet {
	return &SynchronizedInt32HashSet{delegate: NewInt32HashSet()}
}

// NewSynchronizedInt32HashSetFrom wraps an existing set. The
// wrapper takes ownership — callers must not mutate the delegate
// directly without locking.
func NewSynchronizedInt32HashSetFrom(s *Int32HashSet) *SynchronizedInt32HashSet {
	return &SynchronizedInt32HashSet{delegate: s}
}

// SynchronizedInt32HashSetOf constructs a synchronized set from values.
func SynchronizedInt32HashSetOf(values ...int32) *SynchronizedInt32HashSet {
	s := NewInt32HashSet()
	for _, v := range values {
		s.Add(v)
	}
	return &SynchronizedInt32HashSet{delegate: s}
}

// snapshot returns a defensive copy of the set's elements under RLock.
func (s *SynchronizedInt32HashSet) snapshot() []int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

// ── writes ────────────────────────────────────────────────────────────

func (s *SynchronizedInt32HashSet) Add(value int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Add(value)
}

func (s *SynchronizedInt32HashSet) AddAll(values ...int32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.AddAll(values...)
}

func (s *SynchronizedInt32HashSet) Remove(value int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.delegate.Remove(value)
}

func (s *SynchronizedInt32HashSet) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (s *SynchronizedInt32HashSet) Contains(value int32) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Contains(value)
}

func (s *SynchronizedInt32HashSet) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.Size()
}

func (s *SynchronizedInt32HashSet) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.IsEmpty()
}

func (s *SynchronizedInt32HashSet) ToSlice() []int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToSlice()
}

func (s *SynchronizedInt32HashSet) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.String()
}

// ── iteration ────────────────────────────────────────────────────────

// All returns an iter.Seq over a snapshot. Iteration is lock-free.
func (s *SynchronizedInt32HashSet) All() iter.Seq[int32] {
	snapshot := s.snapshot()
	return func(yield func(int32) bool) {
		for _, v := range snapshot {
			if !yield(v) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (s *SynchronizedInt32HashSet) ForEach(f func(int32)) {
	for _, v := range s.snapshot() {
		f(v)
	}
}

func (s *SynchronizedInt32HashSet) AnySatisfy(predicate func(int32) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (s *SynchronizedInt32HashSet) AllSatisfy(predicate func(int32) bool) bool {
	for _, v := range s.snapshot() {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (s *SynchronizedInt32HashSet) NoneSatisfy(predicate func(int32) bool) bool {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (s *SynchronizedInt32HashSet) Detect(predicate func(int32) bool) (int32, bool) {
	for _, v := range s.snapshot() {
		if predicate(v) {
			return v, true
		}
	}
	var zero int32
	return zero, false
}

// ── functional that return a new set ─────────────────────────────────

func (s *SynchronizedInt32HashSet) Select(predicate func(int32) bool) *Int32HashSet {
	snapshot := s.snapshot()
	result := NewInt32HashSet()
	for _, v := range snapshot {
		if predicate(v) {
			result.Add(v)
		}
	}
	return result
}

func (s *SynchronizedInt32HashSet) Reject(predicate func(int32) bool) *Int32HashSet {
	snapshot := s.snapshot()
	result := NewInt32HashSet()
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
func (s *SynchronizedInt32HashSet) lockPair(other *SynchronizedInt32HashSet) func() {
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

func (s *SynchronizedInt32HashSet) Union(other *SynchronizedInt32HashSet) *Int32HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Union(other.delegate)
}

func (s *SynchronizedInt32HashSet) Intersect(other *SynchronizedInt32HashSet) *Int32HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Intersect(other.delegate)
}

func (s *SynchronizedInt32HashSet) Difference(other *SynchronizedInt32HashSet) *Int32HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Difference(other.delegate)
}

func (s *SynchronizedInt32HashSet) SymmetricDifference(other *SynchronizedInt32HashSet) *Int32HashSet {
	release := s.lockPair(other)
	defer release()
	return s.delegate.SymmetricDifference(other.delegate)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (s *SynchronizedInt32HashSet) With(value int32) *SynchronizedInt32HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.With(value)
	return s
}

func (s *SynchronizedInt32HashSet) WithAll(values ...int32) *SynchronizedInt32HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithAll(values...)
	return s
}

func (s *SynchronizedInt32HashSet) Without(value int32) *SynchronizedInt32HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.Without(value)
	return s
}

func (s *SynchronizedInt32HashSet) WithoutAll(values ...int32) *SynchronizedInt32HashSet {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delegate.WithoutAll(values...)
	return s
}

// ── conversions ───────────────────────────────────────────────────────

func (s *SynchronizedInt32HashSet) ToImmutable() *ImmutableInt32HashSet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delegate.ToImmutable()
}

// Equals compares by contents. Locks are acquired in pointer-address
// order to prevent deadlocks under concurrent A.Equals(B) / B.Equals(A).
func (s *SynchronizedInt32HashSet) Equals(other *SynchronizedInt32HashSet) bool {
	release := s.lockPair(other)
	defer release()
	return s.delegate.Equals(other.delegate)
}
