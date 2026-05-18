
package bag

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedInt32HashBag is a thread-safe wrapper around Int32HashBag.
//
// Read methods hold an RLock; writes hold a Lock. Functional methods
// (ForEach/Select/Reject/AnySatisfy/…) snapshot (value, count) pairs
// under RLock, release, and run the callback against the snapshot so
// the callback may safely re-enter the wrapper.
type SynchronizedInt32HashBag struct {
	delegate *Int32HashBag
	mu       sync.RWMutex
}

// NewSynchronizedInt32HashBag creates a new thread-safe empty bag.
func NewSynchronizedInt32HashBag() *SynchronizedInt32HashBag {
	return &SynchronizedInt32HashBag{delegate: NewInt32HashBag()}
}

// NewSynchronizedInt32HashBagFrom wraps an existing bag. The
// wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedInt32HashBagFrom(b *Int32HashBag) *SynchronizedInt32HashBag {
	return &SynchronizedInt32HashBag{delegate: b}
}

// snapshotDistinct returns (values, counts) for every distinct element,
// held only briefly under RLock.
func (b *SynchronizedInt32HashBag) snapshotDistinct() (values []int32, counts []int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for v, c := range b.delegate.AllWithOccurrences() {
		values = append(values, v)
		counts = append(counts, c)
	}
	return
}

// ── writes ────────────────────────────────────────────────────────────

func (b *SynchronizedInt32HashBag) Add(value int32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Add(value)
}

func (b *SynchronizedInt32HashBag) AddOccurrences(value int32, occurrences int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.AddOccurrences(value, occurrences)
}

func (b *SynchronizedInt32HashBag) Remove(value int32) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.Remove(value)
}

func (b *SynchronizedInt32HashBag) RemoveOccurrences(value int32, occurrences int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveOccurrences(value, occurrences)
}

func (b *SynchronizedInt32HashBag) RemoveAll(value int32) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveAll(value)
}

func (b *SynchronizedInt32HashBag) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (b *SynchronizedInt32HashBag) OccurrencesOf(value int32) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.OccurrencesOf(value)
}

func (b *SynchronizedInt32HashBag) Contains(value int32) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Contains(value)
}

func (b *SynchronizedInt32HashBag) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Size()
}

func (b *SynchronizedInt32HashBag) SizeDistinct() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.SizeDistinct()
}

func (b *SynchronizedInt32HashBag) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.IsEmpty()
}

func (b *SynchronizedInt32HashBag) ToSlice() []int32 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToSlice()
}

func (b *SynchronizedInt32HashBag) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All yields every occurrence (multiplicity preserved).
func (b *SynchronizedInt32HashBag) All() iter.Seq[int32] {
	values, counts := b.snapshotDistinct()
	return func(yield func(int32) bool) {
		for i, v := range values {
			for j := 0; j < counts[i]; j++ {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// AllDistinct yields each distinct value exactly once.
func (b *SynchronizedInt32HashBag) AllDistinct() iter.Seq[int32] {
	values, _ := b.snapshotDistinct()
	return func(yield func(int32) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// AllWithOccurrences yields (value, count) pairs for each distinct value.
func (b *SynchronizedInt32HashBag) AllWithOccurrences() iter.Seq2[int32, int] {
	values, counts := b.snapshotDistinct()
	return func(yield func(int32, int) bool) {
		for i, v := range values {
			if !yield(v, counts[i]) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (b *SynchronizedInt32HashBag) ForEach(f func(int32)) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		for j := 0; j < counts[i]; j++ {
			f(v)
		}
	}
}

func (b *SynchronizedInt32HashBag) ForEachWithOccurrences(f func(int32, int)) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		f(v, counts[i])
	}
}

func (b *SynchronizedInt32HashBag) AnySatisfy(predicate func(int32) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (b *SynchronizedInt32HashBag) AllSatisfy(predicate func(int32) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (b *SynchronizedInt32HashBag) NoneSatisfy(predicate func(int32) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (b *SynchronizedInt32HashBag) Detect(predicate func(int32) bool) (int32, bool) {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return v, true
		}
	}
	var zero int32
	return zero, false
}

// ── functional that return new bags ──────────────────────────────────

func (b *SynchronizedInt32HashBag) Select(predicate func(int32) bool) *Int32HashBag {
	values, counts := b.snapshotDistinct()
	result := NewInt32HashBag()
	for i, v := range values {
		if predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

func (b *SynchronizedInt32HashBag) Reject(predicate func(int32) bool) *Int32HashBag {
	values, counts := b.snapshotDistinct()
	result := NewInt32HashBag()
	for i, v := range values {
		if !predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

// TopOccurrences returns the n most frequent elements. Returns the
// exact same shape as the underlying bag.
func (b *SynchronizedInt32HashBag) TopOccurrences(n int) []struct {
	Value int32
	Count int
} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.TopOccurrences(n)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (b *SynchronizedInt32HashBag) With(value int32) *SynchronizedInt32HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.With(value)
	return b
}

func (b *SynchronizedInt32HashBag) WithAll(values ...int32) *SynchronizedInt32HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.WithAll(values...)
	return b
}

func (b *SynchronizedInt32HashBag) Without(value int32) *SynchronizedInt32HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Without(value)
	return b
}

func (b *SynchronizedInt32HashBag) WithoutAll(values ...int32) *SynchronizedInt32HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.WithoutAll(values...)
	return b
}

// ── conversions & equals ──────────────────────────────────────────────

func (b *SynchronizedInt32HashBag) ToImmutable() *ImmutableInt32HashBag {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (b *SynchronizedInt32HashBag) Equals(other *SynchronizedInt32HashBag) bool {
	if b == other {
		b.mu.RLock()
		defer b.mu.RUnlock()
		return b.delegate.Equals(other.delegate)
	}
	first, second := b, other
	if uintptr(unsafe.Pointer(b)) > uintptr(unsafe.Pointer(other)) {
		first, second = other, b
	}
	first.mu.RLock()
	defer first.mu.RUnlock()
	second.mu.RLock()
	defer second.mu.RUnlock()
	return b.delegate.Equals(other.delegate)
}
