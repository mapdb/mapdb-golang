
package bag

import (
	"iter"
	"sync"
	"unsafe"
)

// SynchronizedFloat64HashBag is a thread-safe wrapper around Float64HashBag.
//
// Read methods hold an RLock; writes hold a Lock. Functional methods
// (ForEach/Select/Reject/AnySatisfy/…) snapshot (value, count) pairs
// under RLock, release, and run the callback against the snapshot so
// the callback may safely re-enter the wrapper.
type SynchronizedFloat64HashBag struct {
	delegate *Float64HashBag
	mu       sync.RWMutex
}

// NewSynchronizedFloat64HashBag creates a new thread-safe empty bag.
func NewSynchronizedFloat64HashBag() *SynchronizedFloat64HashBag {
	return &SynchronizedFloat64HashBag{delegate: NewFloat64HashBag()}
}

// NewSynchronizedFloat64HashBagFrom wraps an existing bag. The
// wrapper takes ownership — do not mutate the delegate directly.
func NewSynchronizedFloat64HashBagFrom(b *Float64HashBag) *SynchronizedFloat64HashBag {
	return &SynchronizedFloat64HashBag{delegate: b}
}

// snapshotDistinct returns (values, counts) for every distinct element,
// held only briefly under RLock.
func (b *SynchronizedFloat64HashBag) snapshotDistinct() (values []float64, counts []int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for v, c := range b.delegate.AllWithOccurrences() {
		values = append(values, v)
		counts = append(counts, c)
	}
	return
}

// ── writes ────────────────────────────────────────────────────────────

func (b *SynchronizedFloat64HashBag) Add(value float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Add(value)
}

func (b *SynchronizedFloat64HashBag) AddOccurrences(value float64, occurrences int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.AddOccurrences(value, occurrences)
}

func (b *SynchronizedFloat64HashBag) Remove(value float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.Remove(value)
}

func (b *SynchronizedFloat64HashBag) RemoveOccurrences(value float64, occurrences int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveOccurrences(value, occurrences)
}

func (b *SynchronizedFloat64HashBag) RemoveAll(value float64) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.delegate.RemoveAll(value)
}

func (b *SynchronizedFloat64HashBag) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Clear()
}

// ── simple reads ──────────────────────────────────────────────────────

func (b *SynchronizedFloat64HashBag) OccurrencesOf(value float64) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.OccurrencesOf(value)
}

func (b *SynchronizedFloat64HashBag) Contains(value float64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Contains(value)
}

func (b *SynchronizedFloat64HashBag) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Size()
}

func (b *SynchronizedFloat64HashBag) SizeDistinct() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.SizeDistinct()
}

func (b *SynchronizedFloat64HashBag) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.IsEmpty()
}

func (b *SynchronizedFloat64HashBag) ToSlice() []float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToSlice()
}

func (b *SynchronizedFloat64HashBag) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.String()
}

// ── iteration (snapshot-based) ────────────────────────────────────────

// All yields every occurrence (multiplicity preserved).
func (b *SynchronizedFloat64HashBag) All() iter.Seq[float64] {
	values, counts := b.snapshotDistinct()
	return func(yield func(float64) bool) {
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
func (b *SynchronizedFloat64HashBag) AllDistinct() iter.Seq[float64] {
	values, _ := b.snapshotDistinct()
	return func(yield func(float64) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// AllWithOccurrences yields (value, count) pairs for each distinct value.
func (b *SynchronizedFloat64HashBag) AllWithOccurrences() iter.Seq2[float64, int] {
	values, counts := b.snapshotDistinct()
	return func(yield func(float64, int) bool) {
		for i, v := range values {
			if !yield(v, counts[i]) {
				return
			}
		}
	}
}

// ── functional over snapshot ──────────────────────────────────────────

func (b *SynchronizedFloat64HashBag) ForEach(f func(float64)) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		for j := 0; j < counts[i]; j++ {
			f(v)
		}
	}
}

func (b *SynchronizedFloat64HashBag) ForEachWithOccurrences(f func(float64, int)) {
	values, counts := b.snapshotDistinct()
	for i, v := range values {
		f(v, counts[i])
	}
}

func (b *SynchronizedFloat64HashBag) AnySatisfy(predicate func(float64) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (b *SynchronizedFloat64HashBag) AllSatisfy(predicate func(float64) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (b *SynchronizedFloat64HashBag) NoneSatisfy(predicate func(float64) bool) bool {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return false
		}
	}
	return true
}

func (b *SynchronizedFloat64HashBag) Detect(predicate func(float64) bool) (float64, bool) {
	values, _ := b.snapshotDistinct()
	for _, v := range values {
		if predicate(v) {
			return v, true
		}
	}
	var zero float64
	return zero, false
}

// ── functional that return new bags ──────────────────────────────────

func (b *SynchronizedFloat64HashBag) Select(predicate func(float64) bool) *Float64HashBag {
	values, counts := b.snapshotDistinct()
	result := NewFloat64HashBag()
	for i, v := range values {
		if predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

func (b *SynchronizedFloat64HashBag) Reject(predicate func(float64) bool) *Float64HashBag {
	values, counts := b.snapshotDistinct()
	result := NewFloat64HashBag()
	for i, v := range values {
		if !predicate(v) {
			result.AddOccurrences(v, counts[i])
		}
	}
	return result
}

// TopOccurrences returns the n most frequent elements. Returns the
// exact same shape as the underlying bag.
func (b *SynchronizedFloat64HashBag) TopOccurrences(n int) []struct {
	Value float64
	Count int
} {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.TopOccurrences(n)
}

// ── fluent mutators ───────────────────────────────────────────────────

func (b *SynchronizedFloat64HashBag) With(value float64) *SynchronizedFloat64HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.With(value)
	return b
}

func (b *SynchronizedFloat64HashBag) WithAll(values ...float64) *SynchronizedFloat64HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.WithAll(values...)
	return b
}

func (b *SynchronizedFloat64HashBag) Without(value float64) *SynchronizedFloat64HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Without(value)
	return b
}

func (b *SynchronizedFloat64HashBag) WithoutAll(values ...float64) *SynchronizedFloat64HashBag {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.WithoutAll(values...)
	return b
}

// ── conversions & equals ──────────────────────────────────────────────

func (b *SynchronizedFloat64HashBag) ToImmutable() *ImmutableFloat64HashBag {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToImmutable()
}

// Equals compares by contents. Locks acquired in pointer-address
// order to prevent A.Equals(B) / B.Equals(A) deadlocks.
func (b *SynchronizedFloat64HashBag) Equals(other *SynchronizedFloat64HashBag) bool {
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
