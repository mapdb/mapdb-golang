
package bitset

import (
	"sync"
)

// SynchronizedBitSet is a thread-safe wrapper around BitSet.
type SynchronizedBitSet struct {
	delegate *BitSet
	mu       sync.RWMutex
}

// NewSynchronizedBitSet creates a new thread-safe empty BitSet.
func NewSynchronizedBitSet() *SynchronizedBitSet {
	return &SynchronizedBitSet{delegate: NewBitSet()}
}

func (b *SynchronizedBitSet) Set(bit int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Set(bit)
}

func (b *SynchronizedBitSet) Clear(bit int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Clear(bit)
}

func (b *SynchronizedBitSet) Flip(bit int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.Flip(bit)
}

func (b *SynchronizedBitSet) Get(bit int) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Get(bit)
}

func (b *SynchronizedBitSet) Cardinality() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Cardinality()
}

func (b *SynchronizedBitSet) Length() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.Length()
}

func (b *SynchronizedBitSet) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.IsEmpty()
}

func (b *SynchronizedBitSet) ClearAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delegate.ClearAll()
}

func (b *SynchronizedBitSet) Intersects(other *SynchronizedBitSet) bool {
	b.mu.RLock()
	thisBits := b.delegate.ToSlice()
	b.mu.RUnlock()
	other.mu.RLock()
	otherBits := other.delegate.ToSlice()
	other.mu.RUnlock()
	// Build a quick lookup for other bits
	otherSet := make(map[int]bool)
	for _, bit := range otherBits {
		otherSet[bit] = true
	}
	for _, bit := range thisBits {
		if otherSet[bit] {
			return true
		}
	}
	return false
}

func (b *SynchronizedBitSet) AndInPlace(other *SynchronizedBitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	other.mu.RLock()
	otherCopy := other.delegate
	other.mu.RUnlock()
	b.delegate.AndInPlace(otherCopy)
}

func (b *SynchronizedBitSet) OrInPlace(other *SynchronizedBitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	other.mu.RLock()
	otherCopy := other.delegate
	other.mu.RUnlock()
	b.delegate.OrInPlace(otherCopy)
}

func (b *SynchronizedBitSet) XorInPlace(other *SynchronizedBitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	other.mu.RLock()
	otherCopy := other.delegate
	other.mu.RUnlock()
	b.delegate.XorInPlace(otherCopy)
}

func (b *SynchronizedBitSet) AndNotInPlace(other *SynchronizedBitSet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	other.mu.RLock()
	otherCopy := other.delegate
	other.mu.RUnlock()
	b.delegate.AndNotInPlace(otherCopy)
}

func (b *SynchronizedBitSet) NextSetBit(from int) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.NextSetBit(from)
}

func (b *SynchronizedBitSet) ToSlice() []int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.ToSlice()
}

func (b *SynchronizedBitSet) Equals(other *SynchronizedBitSet) bool {
	b.mu.RLock()
	thisBits := b.delegate.ToSlice()
	b.mu.RUnlock()
	other.mu.RLock()
	otherBits := other.delegate.ToSlice()
	other.mu.RUnlock()
	if len(thisBits) != len(otherBits) {
		return false
	}
	for i, bit := range thisBits {
		if otherBits[i] != bit {
			return false
		}
	}
	return true
}

func (b *SynchronizedBitSet) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.delegate.String()
}
