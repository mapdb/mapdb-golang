
package deque

import (
	"sync"
)

// SynchronizedFloat64ArrayDeque is a thread-safe wrapper around Float64ArrayDeque.
type SynchronizedFloat64ArrayDeque struct {
	delegate *Float64ArrayDeque
	mu       sync.RWMutex
}

// NewSynchronizedFloat64ArrayDeque creates a new thread-safe empty deque.
func NewSynchronizedFloat64ArrayDeque() *SynchronizedFloat64ArrayDeque {
	return &SynchronizedFloat64ArrayDeque{delegate: NewFloat64ArrayDeque()}
}

func (d *SynchronizedFloat64ArrayDeque) AddFirst(value float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delegate.AddFirst(value)
}

func (d *SynchronizedFloat64ArrayDeque) AddLast(value float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delegate.AddLast(value)
}

func (d *SynchronizedFloat64ArrayDeque) RemoveFirst() (float64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delegate.RemoveFirst()
}

func (d *SynchronizedFloat64ArrayDeque) RemoveLast() (float64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.delegate.RemoveLast()
}

func (d *SynchronizedFloat64ArrayDeque) PeekFirst() (float64, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.PeekFirst()
}

func (d *SynchronizedFloat64ArrayDeque) PeekLast() (float64, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.PeekLast()
}

func (d *SynchronizedFloat64ArrayDeque) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.Size()
}

func (d *SynchronizedFloat64ArrayDeque) IsEmpty() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.IsEmpty()
}

func (d *SynchronizedFloat64ArrayDeque) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.delegate.Clear()
}

func (d *SynchronizedFloat64ArrayDeque) Contains(value float64) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.Contains(value)
}

func (d *SynchronizedFloat64ArrayDeque) ForEach(f func(float64)) {
	d.mu.RLock()
	snapshot := d.delegate.ToSlice()
	d.mu.RUnlock()
	for _, v := range snapshot {
		f(v)
	}
}

func (d *SynchronizedFloat64ArrayDeque) AnySatisfy(predicate func(float64) bool) bool {
	d.mu.RLock()
	snapshot := d.delegate.ToSlice()
	d.mu.RUnlock()
	for _, v := range snapshot {
		if predicate(v) {
			return true
		}
	}
	return false
}

func (d *SynchronizedFloat64ArrayDeque) AllSatisfy(predicate func(float64) bool) bool {
	d.mu.RLock()
	snapshot := d.delegate.ToSlice()
	d.mu.RUnlock()
	for _, v := range snapshot {
		if !predicate(v) {
			return false
		}
	}
	return true
}

func (d *SynchronizedFloat64ArrayDeque) ToSlice() []float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.ToSlice()
}

func (d *SynchronizedFloat64ArrayDeque) Equals(other *SynchronizedFloat64ArrayDeque) bool {
	d.mu.RLock()
	thisSlice := d.delegate.ToSlice()
	d.mu.RUnlock()
	other.mu.RLock()
	otherSlice := other.delegate.ToSlice()
	other.mu.RUnlock()
	if len(thisSlice) != len(otherSlice) {
		return false
	}
	for i, v := range thisSlice {
		if otherSlice[i] != v {
			return false
		}
	}
	return true
}

func (d *SynchronizedFloat64ArrayDeque) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate.String()
}
