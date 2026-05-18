
package deque

import (
	"fmt"
	"strings"
)

// Int64ArrayDeque is a double-ended queue of int64 values, backed by a
// power-of-two ring buffer. AddFirst, AddLast, RemoveFirst, RemoveLast,
// PeekFirst and PeekLast are all O(1) amortised.
//
// The public API is intentionally identical to the slice-backed deque
// that preceded it: callers that iterate via ToSlice or ForEach see
// elements in logical front-to-back order regardless of where head
// happens to sit in the underlying buffer.
type Int64ArrayDeque struct {
	items []int64 // len == capacity, always a power of two; indexed modulo cap
	head  int     // index of the front element (0 when empty)
	size  int     // number of logical elements
}

// initialDequeCap is the smallest power-of-two capacity allocated for a
// freshly constructed deque. It matches the previous slice-backed
// implementation's starting capacity so behaviour around the first few
// reallocations stays comparable.
const initialInt64DequeCap = 16

// ceilPow2 rounds n up to the next power of two, with a floor of
// initialInt64DequeCap. Used when sizing the buffer to fit a
// caller-supplied slice in Int64ArrayDequeOf.
func ceilPow2Int64Deque(n int) int {
	cap := initialInt64DequeCap
	for cap < n {
		cap <<= 1
	}
	return cap
}

// NewInt64ArrayDeque creates a new empty Int64ArrayDeque.
func NewInt64ArrayDeque() *Int64ArrayDeque {
	return &Int64ArrayDeque{items: make([]int64, initialInt64DequeCap)}
}

// Int64ArrayDequeOf creates a new Int64ArrayDeque from the given values in
// front-to-back order.
func Int64ArrayDequeOf(values ...int64) *Int64ArrayDeque {
	d := &Int64ArrayDeque{
		items: make([]int64, ceilPow2Int64Deque(len(values))),
		size:  len(values),
	}
	copy(d.items, values)
	return d
}

// grow doubles the backing buffer and repacks elements so that head is at 0.
// Called lazily when size would exceed capacity.
func (d *Int64ArrayDeque) grow() {
	newCap := len(d.items) * 2
	if newCap == 0 {
		newCap = initialInt64DequeCap
	}
	next := make([]int64, newCap)
	// Copy tail segment (head..end), then wrap segment (0..head) so that
	// logical order is preserved and head resets to 0.
	n := copy(next, d.items[d.head:])
	copy(next[n:], d.items[:d.head])
	d.items = next
	d.head = 0
}

// AddFirst prepends a value to the front of the deque. O(1) amortised.
func (d *Int64ArrayDeque) AddFirst(value int64) {
	if d.size == len(d.items) {
		d.grow()
	}
	mask := len(d.items) - 1
	d.head = (d.head - 1) & mask
	d.items[d.head] = value
	d.size++
}

// AddLast appends a value to the back of the deque. O(1) amortised.
func (d *Int64ArrayDeque) AddLast(value int64) {
	if d.size == len(d.items) {
		d.grow()
	}
	mask := len(d.items) - 1
	d.items[(d.head+d.size)&mask] = value
	d.size++
}

// RemoveFirst removes and returns the front element, or an error if empty. O(1).
func (d *Int64ArrayDeque) RemoveFirst() (int64, error) {
	if d.size == 0 {
		return 0, fmt.Errorf("Int64ArrayDeque: RemoveFirst on empty deque")
	}
	mask := len(d.items) - 1
	v := d.items[d.head]
	d.items[d.head] = 0 // let GC reclaim references if int64 ever carries them
	d.head = (d.head + 1) & mask
	d.size--
	return v, nil
}

// RemoveLast removes and returns the back element, or an error if empty. O(1).
func (d *Int64ArrayDeque) RemoveLast() (int64, error) {
	if d.size == 0 {
		return 0, fmt.Errorf("Int64ArrayDeque: RemoveLast on empty deque")
	}
	mask := len(d.items) - 1
	d.size--
	idx := (d.head + d.size) & mask
	v := d.items[idx]
	d.items[idx] = 0
	return v, nil
}

// PeekFirst returns the front element without removing it, or an error if empty.
func (d *Int64ArrayDeque) PeekFirst() (int64, error) {
	if d.size == 0 {
		return 0, fmt.Errorf("Int64ArrayDeque: PeekFirst on empty deque")
	}
	return d.items[d.head], nil
}

// PeekLast returns the back element without removing it, or an error if empty.
func (d *Int64ArrayDeque) PeekLast() (int64, error) {
	if d.size == 0 {
		return 0, fmt.Errorf("Int64ArrayDeque: PeekLast on empty deque")
	}
	mask := len(d.items) - 1
	return d.items[(d.head+d.size-1)&mask], nil
}

// Size returns the number of elements in the deque.
func (d *Int64ArrayDeque) Size() int { return d.size }

// IsEmpty returns true if the deque contains no elements.
func (d *Int64ArrayDeque) IsEmpty() bool { return d.size == 0 }

// Clear removes all elements. The backing buffer is retained.
func (d *Int64ArrayDeque) Clear() {
	// Wipe slots so retained references are released. Cheap for value types.
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		d.items[(d.head+i)&mask] = 0
	}
	d.head = 0
	d.size = 0
}

// Contains returns true if the deque contains the given value.
func (d *Int64ArrayDeque) Contains(value int64) bool {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		v := d.items[(d.head+i)&mask]
		if v == value {
			return true
		}
	}
	return false
}

// ForEach applies the function to each element from front to back.
func (d *Int64ArrayDeque) ForEach(f func(int64)) {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		f(d.items[(d.head+i)&mask])
	}
}

// AnySatisfy returns true if any element satisfies the predicate.
func (d *Int64ArrayDeque) AnySatisfy(predicate func(int64) bool) bool {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		if predicate(d.items[(d.head+i)&mask]) {
			return true
		}
	}
	return false
}

// AllSatisfy returns true if every element satisfies the predicate.
func (d *Int64ArrayDeque) AllSatisfy(predicate func(int64) bool) bool {
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		if !predicate(d.items[(d.head+i)&mask]) {
			return false
		}
	}
	return true
}

// ToSlice returns a copy of the elements in front-to-back order.
func (d *Int64ArrayDeque) ToSlice() []int64 {
	out := make([]int64, d.size)
	if d.size == 0 {
		return out
	}
	cap := len(d.items)
	tail := cap - d.head
	if d.size <= tail {
		copy(out, d.items[d.head:d.head+d.size])
	} else {
		n := copy(out, d.items[d.head:])
		copy(out[n:], d.items[:d.size-n])
	}
	return out
}

// Equals returns true if the other deque has the same elements in the same order.
func (d *Int64ArrayDeque) Equals(other *Int64ArrayDeque) bool {
	if d.size != other.size {
		return false
	}
	dMask := len(d.items) - 1
	oMask := len(other.items) - 1
	for i := 0; i < d.size; i++ {
		a := d.items[(d.head+i)&dMask]
		b := other.items[(other.head+i)&oMask]
		if !(a == b) {
			return false
		}
	}
	return true
}

// String returns a string representation in front-to-back order.
func (d *Int64ArrayDeque) String() string {
	if d.size == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	mask := len(d.items) - 1
	for i := 0; i < d.size; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "%v", d.items[(d.head+i)&mask])
	}
	sb.WriteString("]")
	return sb.String()
}
